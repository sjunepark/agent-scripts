"use strict";

const crypto = require("node:crypto");
const fs = require("node:fs");
const path = require("node:path");

const TREE_HASH_ALGORITHM = "tree-sha256-v2";
const RECONCILER_STATE_RELATIVE_PATH = ".agents/.global-skill-state.json";

const DEFAULT_ROOT_POLICY = Object.freeze([
  Object.freeze({
    id: "shared",
    relativePath: ".agents/skills",
    kind: "managed",
    lockRelativePath: ".agents/.skill-lock.json"
  }),
  Object.freeze({
    id: "claude",
    relativePath: ".claude/skills",
    kind: "managed",
    lockRelativePath: ".claude/.skill-lock.json"
  }),
  Object.freeze({
    id: "pi",
    relativePath: ".pi/agent/skills",
    kind: "legacy",
    canonicalRootId: "shared"
  }),
  Object.freeze({
    id: "codex-system",
    relativePath: ".codex/skills/.system",
    kind: "protected"
  }),
  Object.freeze({
    id: "codex-plugin-cache",
    relativePath: ".codex/plugins/cache",
    kind: "protected"
  })
]);

const allowedRootKinds = new Set(["legacy", "managed", "protected"]);

function compareText(left, right) {
  if (left === right) return 0;
  return left < right ? -1 : 1;
}

function isObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function canonicalSourceIdentity(source) {
  if (typeof source !== "string" || source.length === 0) return null;
  const shorthand = source.match(/^([A-Za-z0-9_.-]+)\/([A-Za-z0-9_.-]+?)(?:\.git)?$/);
  if (shorthand) return `github:${shorthand[1].toLowerCase()}/${shorthand[2].toLowerCase()}`;
  let parsed;
  try {
    parsed = new URL(source);
  } catch {
    return source.replace(/\.git$/, "");
  }
  if (parsed.hostname.toLowerCase() === "github.com") {
    const parts = parsed.pathname.replace(/^\/+|\/+$/g, "").split("/");
    if (parts.length >= 2) {
      return `github:${parts[0].toLowerCase()}/${parts[1].replace(/\.git$/, "").toLowerCase()}`;
    }
  }
  return `${parsed.protocol}//${parsed.host}${parsed.pathname.replace(/\.git$/, "").replace(/\/$/, "")}`;
}

function isPathWithin(candidate, parent) {
  const relative = path.relative(parent, candidate);
  return relative === "" || (!relative.startsWith("..") && !path.isAbsolute(relative));
}

function pathEntryExists(candidate) {
  try {
    fs.lstatSync(candidate);
    return true;
  } catch (error) {
    if (error.code === "ENOENT") return false;
    throw error;
  }
}

function resolveHomePath(homeDir, relativePath, label) {
  if (typeof relativePath !== "string" || relativePath.length === 0 || path.isAbsolute(relativePath)) {
    throw new Error(`${label} must be a non-empty path relative to homeDir`);
  }
  const resolved = path.resolve(homeDir, relativePath);
  if (!isPathWithin(resolved, homeDir)) {
    throw new Error(`${label} must stay inside homeDir`);
  }
  return resolved;
}

function assertMutationPathWithin(candidate, homeDir, label = "mutation path") {
  if (!path.isAbsolute(candidate) || !path.isAbsolute(homeDir)) {
    throw new Error(`${label} and homeDir must be absolute paths`);
  }
  const resolvedCandidate = path.resolve(candidate);
  const resolvedHome = path.resolve(homeDir);
  if (!isPathWithin(resolvedCandidate, resolvedHome)) {
    throw new Error(`${label} must stay inside homeDir: ${resolvedCandidate}`);
  }

  let canonicalHome;
  try {
    canonicalHome = fs.realpathSync(resolvedHome);
  } catch (error) {
    throw new Error(`homeDir must exist and be readable: ${error.message}`);
  }

  let existingAncestor = resolvedCandidate;
  while (true) {
    try {
      fs.lstatSync(existingAncestor);
      break;
    } catch (error) {
      if (error.code !== "ENOENT") {
        throw new Error(`${label} ancestor is unreadable: ${existingAncestor}: ${error.message}`);
      }
      const parent = path.dirname(existingAncestor);
      if (parent === existingAncestor) {
        throw new Error(`${label} has no existing ancestor: ${resolvedCandidate}`);
      }
      existingAncestor = parent;
    }
  }

  let canonicalAncestor;
  try {
    canonicalAncestor = fs.realpathSync(existingAncestor);
  } catch (error) {
    throw new Error(`${label} ancestor cannot be resolved: ${existingAncestor}: ${error.message}`);
  }
  if (!isPathWithin(canonicalAncestor, canonicalHome)) {
    throw new Error(`${label} resolves outside homeDir: ${resolvedCandidate}`);
  }
  return { canonicalHome, existingAncestor, canonicalAncestor, resolvedCandidate };
}

function assertModeledDirectory({
  directory,
  homeDir,
  label = "modeled directory",
  allowMissing = false,
  expectedCanonicalPath = null
}) {
  assertMutationPathWithin(directory, homeDir, label);
  const resolvedHome = path.resolve(homeDir);
  const resolvedDirectory = path.resolve(directory);
  const relative = path.relative(resolvedHome, resolvedDirectory);
  const components = relative === "" ? [] : relative.split(path.sep);
  let current = resolvedHome;
  for (const component of components) {
    current = path.join(current, component);
    let stat;
    try {
      stat = fs.lstatSync(current);
    } catch (error) {
      if (error.code === "ENOENT" && allowMissing) {
        return { exists: false, canonicalPath: null };
      }
      if (error.code === "ENOENT") throw new Error(`${label} does not exist: ${directory}`);
      throw new Error(`${label} is unreadable: ${current}: ${error.message}`);
    }
    if (stat.isSymbolicLink()) {
      throw new Error(`${label} must not traverse a symlink: ${current}`);
    }
    if (!stat.isDirectory()) {
      throw new Error(`${label} must be a directory: ${current}`);
    }
  }
  const canonicalPath = fs.realpathSync(resolvedDirectory);
  if (expectedCanonicalPath && canonicalPath !== expectedCanonicalPath) {
    throw new Error(`${label} canonical identity changed: ${directory}`);
  }
  return { exists: true, canonicalPath };
}

function normalizedRootPolicy(homeDir, rootPolicy) {
  if (!path.isAbsolute(homeDir)) throw new Error("homeDir must be an absolute path");
  if (!Array.isArray(rootPolicy) || rootPolicy.length === 0) {
    throw new Error("rootPolicy must be a non-empty array");
  }

  const ids = new Set();
  return rootPolicy.map((root, index) => {
    const label = `rootPolicy[${index}]`;
    if (!isObject(root)) throw new Error(`${label} must be an object`);
    if (typeof root.id !== "string" || !/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(root.id)) {
      throw new Error(`${label}.id must be a portable lowercase name`);
    }
    if (ids.has(root.id)) throw new Error(`duplicate root id: ${root.id}`);
    ids.add(root.id);
    if (!allowedRootKinds.has(root.kind)) {
      throw new Error(`${label}.kind must be managed, legacy, or protected`);
    }

    return {
      id: root.id,
      kind: root.kind,
      relativePath: root.relativePath,
      path: resolveHomePath(homeDir, root.relativePath, `${label}.relativePath`),
      lockPath: root.lockRelativePath
        ? resolveHomePath(homeDir, root.lockRelativePath, `${label}.lockRelativePath`)
        : null,
      canonicalRootId: root.canonicalRootId ?? null
    };
  });
}

function updateHashField(hash, value) {
  const buffer = Buffer.isBuffer(value) ? value : Buffer.from(String(value));
  hash.update(String(buffer.length));
  hash.update(":");
  hash.update(buffer);
}

function hashPathInto(hash, target, relativePath) {
  const stat = fs.lstatSync(target);
  const portablePath = relativePath.split(path.sep).join("/");

  if (stat.isSymbolicLink()) {
    updateHashField(hash, "symlink");
    updateHashField(hash, portablePath);
    updateHashField(hash, fs.readlinkSync(target));
    return;
  }
  if (stat.isFile()) {
    updateHashField(hash, "file");
    updateHashField(hash, portablePath);
    updateHashField(hash, (stat.mode & 0o111) === 0 ? "non-executable" : "executable");
    updateHashField(hash, fs.readFileSync(target));
    return;
  }
  if (!stat.isDirectory()) {
    updateHashField(hash, "other");
    updateHashField(hash, portablePath);
    return;
  }

  updateHashField(hash, "directory");
  updateHashField(hash, portablePath);
  for (const entry of fs.readdirSync(target, { withFileTypes: true }).sort((a, b) =>
    compareText(a.name, b.name)
  )) {
    hashPathInto(hash, path.join(target, entry.name), path.join(relativePath, entry.name));
  }
}

function hashSkillDirectory(skillDir) {
  const stat = fs.lstatSync(skillDir);
  if (!stat.isDirectory()) throw new Error(`skill path is not a directory: ${skillDir}`);
  const hash = crypto.createHash("sha256");
  hashPathInto(hash, skillDir, ".");
  return { hashAlgorithm: TREE_HASH_ALGORITHM, hash: hash.digest("hex") };
}

function hashObservedPath(target) {
  const stat = fs.lstatSync(target);
  if (stat.isDirectory()) return hashSkillDirectory(target);
  const hash = crypto.createHash("sha256");
  hashPathInto(hash, target, ".");
  return { hashAlgorithm: TREE_HASH_ALGORITHM, hash: hash.digest("hex") };
}

function inspectEntry(entryPath, root, scannableRootPaths) {
  const stat = fs.lstatSync(entryPath);
  const base = {
    name: path.basename(entryPath),
    root: root.id,
    path: entryPath,
    kind: "other",
    hashAlgorithm: null,
    hash: null,
    hashStatus: "unsupported-entry-type",
    linkTarget: null,
    resolvedTarget: null
  };

  if (stat.isSymbolicLink()) {
    base.kind = "symlink";
    base.linkTarget = fs.readlinkSync(entryPath);
    try {
      base.resolvedTarget = fs.realpathSync(entryPath);
    } catch {
      base.hashStatus = "broken-symlink";
      return base;
    }

    if (!scannableRootPaths.some((rootPath) => isPathWithin(base.resolvedTarget, rootPath))) {
      base.hashStatus = "target-outside-explicit-roots";
      return base;
    }

    try {
      Object.assign(base, hashObservedPath(base.resolvedTarget));
      base.hashStatus = "verified";
    } catch {
      base.hashStatus = "target-unreadable";
    }
    return base;
  }

  if (stat.isDirectory()) base.kind = "directory";
  else if (stat.isFile()) base.kind = "file";

  try {
    Object.assign(base, hashObservedPath(entryPath));
    base.hashStatus = "verified";
  } catch {
    base.hashStatus = "unreadable";
  }
  return base;
}

function parseLockFile(lockPath, rootId) {
  let parsed;
  try {
    parsed = JSON.parse(fs.readFileSync(lockPath, "utf8"));
  } catch (error) {
    return {
      locks: [],
      errors: [{ root: rootId, path: lockPath, reason: "lock-file-invalid", detail: error.message }]
    };
  }

  if (!isObject(parsed) || !isObject(parsed.skills)) {
    return {
      locks: [],
      errors: [{ root: rootId, path: lockPath, reason: "lock-file-invalid" }]
    };
  }

  const locks = [];
  const errors = [];
  for (const name of Object.keys(parsed.skills).sort()) {
    const record = parsed.skills[name];
    if (!isObject(record)) {
      errors.push({ root: rootId, path: lockPath, reason: "lock-record-invalid", skill: name });
      continue;
    }
    locks.push({
      name,
      root: rootId,
      path: lockPath,
      source: record.sourceUrl ?? record.source ?? null,
      ref: record.ref ?? null,
      skillPath: record.skillPath ?? record.path ?? null,
      hash: record.contentHash ?? record.computedHash ?? null,
      hashAlgorithm: record.hashAlgorithm ?? null
    });
  }
  return { locks, errors };
}

function parseReconcilerStateFromValue(parsed, statePath) {
  if (!isObject(parsed) || parsed.version !== 1 || !Array.isArray(parsed.records)) {
    return {
      records: [],
      errors: [{ root: "shared", path: statePath, reason: "reconciler-state-invalid" }]
    };
  }

  const records = [];
  const errors = [];
  const keys = new Set();
  for (const record of parsed.records) {
    const valid =
      isObject(record) &&
      new Set(["shared", "claude"]).has(record.root) &&
      typeof record.skill === "string" &&
      /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(record.skill) &&
      typeof record.source === "string" &&
      record.source.length > 0 &&
      record.hashAlgorithm === TREE_HASH_ALGORITHM &&
      typeof record.hash === "string" &&
      /^[a-f0-9]{64}$/.test(record.hash) &&
      typeof record.recordedAt === "string";
    const key = valid ? `${record.root}\0${record.skill}` : null;
    if (!valid || keys.has(key)) {
      errors.push({
        root: valid ? record.root : "shared",
        skill: valid ? record.skill : null,
        path: statePath,
        reason: "reconciler-state-record-invalid"
      });
      continue;
    }
    keys.add(key);
    records.push({
      root: record.root,
      skill: record.skill,
      source: record.source,
      hashAlgorithm: record.hashAlgorithm,
      hash: record.hash,
      recordedAt: record.recordedAt
    });
  }
  return { records: records.sort(compareStateRecords), errors };
}

function parseReconcilerState(statePath) {
  let parsed;
  try {
    parsed = JSON.parse(fs.readFileSync(statePath, "utf8"));
  } catch (error) {
    return {
      records: [],
      errors: [{ root: "shared", path: statePath, reason: "reconciler-state-invalid", detail: error.message }]
    };
  }
  return parseReconcilerStateFromValue(parsed, statePath);
}

function inspectSkillRoots({ homeDir, rootPolicy = DEFAULT_ROOT_POLICY }) {
  if (typeof homeDir !== "string" || homeDir.length === 0 || !path.isAbsolute(homeDir)) {
    throw new Error("homeDir must be a non-empty absolute path");
  }
  const resolvedHome = path.resolve(homeDir);
  let canonicalHome;
  try {
    canonicalHome = fs.realpathSync(resolvedHome);
  } catch (error) {
    throw new Error(`homeDir must exist and be readable: ${error.message}`);
  }
  const roots = normalizedRootPolicy(resolvedHome, rootPolicy).sort((a, b) =>
    compareText(a.id, b.id)
  );
  const protectedPaths = roots.filter(({ kind }) => kind === "protected").map(({ path }) => path);
  const entries = [];
  const protectedRoots = [];
  const errors = [];
  const locks = [];
  const reconcilerState = [];
  const readLockPaths = new Set();

  for (const root of roots) {
    let stat;
    let lookupFailed = false;
    try {
      stat = fs.lstatSync(root.path);
    } catch (error) {
      if (error.code !== "ENOENT") {
        errors.push({ root: root.id, path: root.path, reason: "root-unreadable", detail: error.message });
        lookupFailed = true;
      }
      stat = null;
    }
    root.exists = stat !== null;
    root.canonicalPath = null;
    root.safeToRead = !root.exists && !lookupFailed;
    if (root.exists) {
      try {
        root.canonicalPath = fs.realpathSync(root.path);
        root.safeToRead = isPathWithin(root.canonicalPath, canonicalHome);
        if (!root.safeToRead) {
          errors.push({
            root: root.id,
            path: root.path,
            reason: "root-outside-home",
            detail: root.canonicalPath
          });
        }
      } catch (error) {
        errors.push({
          root: root.id,
          path: root.path,
          reason: "root-unreadable",
          detail: error.message
        });
        root.safeToRead = false;
      }
    } else if (!lookupFailed && root.kind !== "protected") {
      try {
        assertModeledDirectory({
          directory: root.path,
          homeDir: resolvedHome,
          label: `${root.id} root`,
          allowMissing: true
        });
      } catch (error) {
        errors.push({
          root: root.id,
          path: root.path,
          reason: "root-unsafe-ancestor",
          detail: error.message
        });
        root.safeToRead = false;
      }
    }
  }

  const scannableRootPaths = roots
    .filter(({ kind, safeToRead, canonicalPath }) =>
      kind !== "protected" && safeToRead && canonicalPath
    )
    .map(({ canonicalPath }) => canonicalPath);

  for (const root of roots) {
    let stat = null;
    if (root.exists && root.safeToRead) {
      try {
        stat = fs.lstatSync(root.path);
      } catch (error) {
        errors.push({
          root: root.id,
          path: root.path,
          reason: "root-unreadable",
          detail: error.message
        });
      }
    }

    if (root.kind === "protected") {
      if (root.exists && root.safeToRead) protectedRoots.push({ root: root.id, path: root.path });
      continue;
    }
    if (!root.safeToRead || (root.exists && !stat)) continue;
    if (root.exists && (!stat.isDirectory() || stat.isSymbolicLink())) {
      errors.push({ root: root.id, path: root.path, reason: "root-not-directory" });
      continue;
    }
    if (root.exists) {
      let children;
      try {
        children = fs.readdirSync(root.path, { withFileTypes: true }).sort((a, b) =>
          compareText(a.name, b.name)
        );
      } catch (error) {
        errors.push({
          root: root.id,
          path: root.path,
          reason: "root-unreadable",
          detail: error.message
        });
        children = [];
      }
      for (const child of children) {
        const childPath = path.join(root.path, child.name);
        if (protectedPaths.some((protectedPath) => isPathWithin(protectedPath, childPath))) {
          continue;
        }
        try {
          entries.push(inspectEntry(childPath, root, scannableRootPaths));
        } catch (error) {
          errors.push({
            root: root.id,
            path: childPath,
            reason: "entry-unreadable",
            detail: error.message
          });
        }
      }
    }

    if (root.lockPath && !readLockPaths.has(root.lockPath) && fs.existsSync(root.lockPath)) {
      readLockPaths.add(root.lockPath);
      let canonicalLockPath;
      try {
        canonicalLockPath = fs.realpathSync(root.lockPath);
      } catch (error) {
        errors.push({
          root: root.id,
          path: root.lockPath,
          reason: "lock-file-invalid",
          detail: error.message
        });
      }
      if (canonicalLockPath && !isPathWithin(canonicalLockPath, canonicalHome)) {
        errors.push({
          root: root.id,
          path: root.lockPath,
          reason: "lock-file-outside-home",
          detail: canonicalLockPath
        });
      } else if (canonicalLockPath) {
        const parsed = parseLockFile(root.lockPath, root.id);
        locks.push(...parsed.locks);
        errors.push(...parsed.errors);
      }
    }
  }

  const reconcilerStatePath = resolveHomePath(
    resolvedHome,
    RECONCILER_STATE_RELATIVE_PATH,
    "reconciler state path"
  );
  if (fs.existsSync(reconcilerStatePath)) {
    let canonicalStatePath;
    try {
      canonicalStatePath = fs.realpathSync(reconcilerStatePath);
    } catch (error) {
      errors.push({
        root: "shared",
        path: reconcilerStatePath,
        reason: "reconciler-state-invalid",
        detail: error.message
      });
    }
    if (canonicalStatePath && !isPathWithin(canonicalStatePath, canonicalHome)) {
      errors.push({
        root: "shared",
        path: reconcilerStatePath,
        reason: "reconciler-state-outside-home",
        detail: canonicalStatePath
      });
    } else if (canonicalStatePath) {
      const parsed = parseReconcilerState(reconcilerStatePath);
      if (parsed.errors.length === 0) reconcilerState.push(...parsed.records);
      errors.push(...parsed.errors);
    }
  }

  return {
    homeDir: resolvedHome,
    roots,
    entries: entries.sort(compareStateRecords),
    locks: locks.sort(compareStateRecords),
    reconcilerState: reconcilerState.sort(compareStateRecords),
    reconcilerStatePath,
    protectedRoots: protectedRoots.sort(compareStateRecords),
    errors: errors.sort(compareStateRecords)
  };
}

function expectedGlobalPlacements(desiredEntries) {
  const placements = [];
  for (const entry of desiredEntries) {
    const agents = new Set(entry.agents || []);
    if (agents.has("claude-code")) {
      placements.push({
        skill: entry.name,
        root: "claude",
        targetAgent: "claude-code",
        manager: entry.manager,
        source: entry.source,
        sourceId: entry.sourceId,
        fullDepth: entry.fullDepth === true
      });
    }
    if (agents.has("codex") || agents.has("pi")) {
      placements.push({
        skill: entry.name,
        root: "shared",
        targetAgent: "codex",
        manager: entry.manager,
        source: entry.source,
        sourceId: entry.sourceId,
        fullDepth: entry.fullDepth === true
      });
    }
  }
  return placements.sort(compareStateRecords);
}

function compareStateRecords(left, right) {
  const leftKey = [left.skill ?? left.name ?? "", left.root ?? "", left.type ?? "", left.path ?? ""];
  const rightKey = [right.skill ?? right.name ?? "", right.root ?? "", right.type ?? "", right.path ?? ""];
  return compareText(leftKey.join("\0"), rightKey.join("\0"));
}

function issueFromPlacement(type, placement, root, reason, extra = {}) {
  return {
    type,
    skill: placement.skill,
    root: placement.root,
    path: root.targetPath ?? root.path,
    modeledRootPath: root.path,
    expectedRootCanonicalPath: root.canonicalPath,
    reason,
    manager: placement.manager,
    source: placement.source,
    sourceId: placement.sourceId,
    targetAgent: placement.targetAgent,
    fullDepth: placement.fullDepth,
    ...extra
  };
}

function lockForSkill(locks, skill, rootId) {
  return locks.find((lock) => lock.name === skill && lock.root === rootId) ?? null;
}

function reconcilerRecordForSkill(records, skill, rootId) {
  return records.find((record) => record.skill === skill && record.root === rootId) ?? null;
}

function classifyExpectedEntry({ placement, root, current, expected, lock, reconcilerRecord }) {
  if (!current) return issueFromPlacement("missing", placement, root, "expected-entry-absent");
  if (current.kind === "symlink") {
    return issueFromPlacement(
      "misplaced",
      placement,
      root,
      "expected-copy-is-symlink",
      { resolvedTarget: current.resolvedTarget }
    );
  }
  if (current.kind !== "directory") {
    return issueFromPlacement("modified", placement, root, "expected-skill-is-not-directory");
  }
  if (!expected?.hash || expected.hashAlgorithm !== TREE_HASH_ALGORITHM) {
    if (placement.manager !== "skills-cli") {
      return issueFromPlacement(
        "protected",
        placement,
        root,
        "externally-managed-content-not-verified"
      );
    }
    return issueFromPlacement(
      "unclassified",
      placement,
      root,
      "expected-hash-unverifiable"
    );
  }
  if (current.hashStatus !== "verified" || current.hashAlgorithm !== TREE_HASH_ALGORITHM) {
    return issueFromPlacement("unclassified", placement, root, "current-hash-unverifiable");
  }
  if (current.hash === expected.hash) return null;
  if (reconcilerRecord) {
    if (reconcilerRecord.source !== placement.source) {
      return issueFromPlacement(
        "unclassified",
        placement,
        root,
        "reconciler-state-source-mismatch"
      );
    }
    if (
      reconcilerRecord.hashAlgorithm === TREE_HASH_ALGORITHM &&
      current.hash === reconcilerRecord.hash
    ) {
      return issueFromPlacement(
        "outdated",
        placement,
        root,
        "installed-content-is-older-than-expected",
        {
          currentKind: current.kind,
          currentHashAlgorithm: current.hashAlgorithm,
          currentHash: current.hash
        }
      );
    }
    return issueFromPlacement(
      "modified",
      placement,
      root,
      "installed-content-differs-from-reconciler-state"
    );
  }
  if (
    lock &&
    lock.source !== null &&
    canonicalSourceIdentity(lock.source) !== canonicalSourceIdentity(placement.source)
  ) {
    return issueFromPlacement("unclassified", placement, root, "source-provenance-mismatch");
  }
  if (!lock?.hash || lock.hashAlgorithm !== TREE_HASH_ALGORITHM) {
    return issueFromPlacement(
      "unclassified",
      placement,
      root,
      "lock-hash-unverifiable",
      {
        currentKind: current.kind,
        currentHashAlgorithm: current.hashAlgorithm,
        currentHash: current.hash
      }
    );
  }
  if (current.hash === lock.hash) {
    return issueFromPlacement(
      "outdated",
      placement,
      root,
      "installed-content-is-older-than-expected",
      {
        currentKind: current.kind,
        currentHashAlgorithm: current.hashAlgorithm,
        currentHash: current.hash
      }
    );
  }
  return issueFromPlacement("modified", placement, root, "installed-content-differs-from-lock");
}

function reconcileGlobalSkillState({
  desiredEntries,
  knownSkillNames = [],
  expectedContent = {},
  inventory
}) {
  const placements = expectedGlobalPlacements(desiredEntries);
  const rootById = new Map(inventory.roots.map((root) => [root.id, root]));
  const entryByKey = new Map(inventory.entries.map((entry) => [`${entry.root}\0${entry.name}`, entry]));
  const desiredByName = new Map(desiredEntries.map((entry) => [entry.name, entry]));
  const knownNames = new Set([...knownSkillNames, ...desiredByName.keys()]);
  const placementByKey = new Map(placements.map((placement) => [
    `${placement.root}\0${placement.skill}`,
    placement
  ]));
  const cleanExpectedKeys = new Set();
  const issues = [];

  for (const error of inventory.errors) {
    issues.push({
      type: "unclassified",
      skill: error.skill ?? null,
      root: error.root,
      path: error.path,
      reason: error.reason,
      detail: error.detail ?? null
    });
  }
  for (const protectedRoot of inventory.protectedRoots) {
    issues.push({
      type: "protected",
      skill: null,
      root: protectedRoot.root,
      path: protectedRoot.path,
      reason: "runtime-owned-root"
    });
  }

  for (const placement of placements) {
    const root = rootById.get(placement.root);
    if (!root) throw new Error(`root policy does not define expected root ${placement.root}`);
    const key = `${placement.root}\0${placement.skill}`;
    const current = entryByKey.get(key);
    const expected = expectedContent[placement.skill];
    const lock = lockForSkill(inventory.locks, placement.skill, placement.root);
    const reconcilerRecord = reconcilerRecordForSkill(
      inventory.reconcilerState ?? [],
      placement.skill,
      placement.root
    );
    const classification = classifyExpectedEntry({
      placement,
      root: { ...root, targetPath: path.join(root.path, placement.skill) },
      current,
      expected,
      lock,
      reconcilerRecord
    });
    if (classification) issues.push(classification);
    else cleanExpectedKeys.add(key);
  }

  for (const entry of inventory.entries) {
    const key = `${entry.root}\0${entry.name}`;
    if (placementByKey.has(key)) continue;
    const root = rootById.get(entry.root);
    const desired = desiredByName.get(entry.name);

    if (root.kind === "managed") {
      if (desired) {
        issues.push({
          type: "misplaced",
          skill: entry.name,
          root: entry.root,
          path: entry.path,
          reason: "skill-is-in-unexpected-managed-root",
          manager: desired.manager
        });
      } else if (knownNames.has(entry.name)) {
        issues.push({
          type: "unexpected-managed",
          skill: entry.name,
          root: entry.root,
          path: entry.path,
          reason: "known-skill-is-not-selected"
        });
      } else {
        issues.push({
          type: "unclassified",
          skill: entry.name,
          root: entry.root,
          path: entry.path,
          reason: "unknown-managed-entry"
        });
      }
      continue;
    }

    if (root.kind !== "legacy") continue;
    if (!desired) {
      issues.push({
        type: knownNames.has(entry.name) ? "misplaced" : "unclassified",
        skill: entry.name,
        root: entry.root,
        path: entry.path,
        reason: knownNames.has(entry.name) ? "known-skill-is-not-selected" : "unknown-legacy-entry"
      });
      continue;
    }

    const canonicalRootId = root.canonicalRootId;
    const canonicalKey = `${canonicalRootId}\0${entry.name}`;
    const canonical = entryByKey.get(canonicalKey);
    let canonicalResolvedPath = canonical?.path ?? null;
    try {
      if (canonicalResolvedPath) canonicalResolvedPath = fs.realpathSync(canonicalResolvedPath);
    } catch {
      canonicalResolvedPath = null;
    }
    const equivalentSymlink =
      entry.kind === "symlink" && canonicalResolvedPath && entry.resolvedTarget === canonicalResolvedPath;
    const equivalentHash =
      entry.hashStatus === "verified" &&
      canonical?.hashStatus === "verified" &&
      entry.hash === canonical.hash;
    if (cleanExpectedKeys.has(canonicalKey) && (equivalentSymlink || equivalentHash)) {
      const canonicalRoot = rootById.get(canonicalRootId);
      issues.push({
        type: "verified-legacy-duplicate",
        skill: entry.name,
        root: entry.root,
        path: entry.path,
        reason: equivalentSymlink ? "symlink-targets-canonical-entry" : "content-matches-canonical-entry",
        manager: desired.manager,
        canonicalPath: canonical.path,
        modeledRootPath: root.path,
        expectedRootCanonicalPath: root.canonicalPath,
        canonicalRootPath: canonicalRoot.path,
        expectedCanonicalRootCanonicalPath: canonicalRoot.canonicalPath,
        entryKind: entry.kind,
        hashAlgorithm: entry.hashAlgorithm,
        hash: entry.hash,
        resolvedTarget: entry.resolvedTarget,
        canonicalHashAlgorithm: canonical.hashAlgorithm,
        canonicalHash: canonical.hash
      });
    } else {
      issues.push({
        type: "misplaced",
        skill: entry.name,
        root: entry.root,
        path: entry.path,
        reason: "legacy-entry-is-not-a-verified-duplicate",
        manager: desired.manager
      });
    }
  }

  return {
    inventory,
    placements,
    issues: issues.sort(compareStateRecords)
  };
}

function planGlobalSkillOperations({ report, quarantineRoot }) {
  if (!path.isAbsolute(quarantineRoot)) throw new Error("quarantineRoot must be an absolute path");
  if (
    quarantineRoot === report.inventory.homeDir ||
    !isPathWithin(quarantineRoot, report.inventory.homeDir)
  ) {
    throw new Error("quarantineRoot must be a child of the inspected homeDir");
  }
  const quarantineBase = path.join(report.inventory.homeDir, ".skill-quarantine");
  for (const root of report.inventory.roots) {
    if (isPathWithin(quarantineRoot, root.path) || isPathWithin(root.path, quarantineRoot)) {
      throw new Error(`quarantineRoot must not overlap skill root ${root.path}`);
    }
  }
  if (quarantineRoot === quarantineBase || !isPathWithin(quarantineRoot, quarantineBase)) {
    throw new Error("quarantineRoot must be a run directory under homeDir/.skill-quarantine");
  }

  const apply = [];
  const prune = [];
  const replace = [];
  const restore = [];
  const manual = [];
  const blocked = [];
  const operationKeys = new Set();
  const unsafeRootReasons = new Set([
    "root-not-directory",
    "root-outside-home",
    "root-unreadable",
    "root-unsafe-ancestor"
  ]);
  const unsafeRootIds = new Set(
    report.inventory.errors
      .filter(({ reason }) => unsafeRootReasons.has(reason))
      .map(({ root }) => root)
  );

  for (const issue of report.issues) {
    if (
      issue.type === "unclassified" &&
      issue.reason === "lock-hash-unverifiable" &&
      issue.manager === "skills-cli" &&
      issue.currentKind === "directory" &&
      issue.currentHashAlgorithm === TREE_HASH_ALGORITHM &&
      issue.currentHash &&
      !unsafeRootIds.has(issue.root)
    ) {
      replace.push({
        type: "quarantine-unverified",
        skill: issue.skill,
        root: issue.root,
        from: issue.path,
        to: path.join(quarantineRoot, issue.root, issue.skill),
        mustNotExist: true,
        expectedKind: issue.currentKind,
        expectedHashAlgorithm: issue.currentHashAlgorithm,
        expectedHash: issue.currentHash,
        expectedResolvedTarget: null,
        modeledRootPath: issue.modeledRootPath,
        expectedRootCanonicalPath: issue.expectedRootCanonicalPath,
        source: issue.source,
        sourceId: issue.sourceId,
        targetAgent: issue.targetAgent,
        fullDepth: issue.fullDepth
      });
    }
    if (issue.type === "missing" || issue.type === "outdated") {
      if (unsafeRootIds.has(issue.root)) {
        blocked.push({
          type: "blocked",
          skill: issue.skill,
          root: issue.root,
          path: issue.path,
          issue: issue.type,
          reason: "modeled-root-is-unsafe"
        });
        continue;
      }
      if (issue.manager === "skills-cli") {
        const key = `${issue.skill}\0${issue.root}\0${issue.type}`;
        if (!operationKeys.has(key)) {
          operationKeys.add(key);
          apply.push({
            type: issue.type === "missing" ? "install" : "update",
            skill: issue.skill,
            source: issue.source,
            sourceId: issue.sourceId,
            root: issue.root,
            targetAgent: issue.targetAgent,
            fullDepth: issue.fullDepth,
            targetPath: issue.path,
            modeledRootPath: issue.modeledRootPath,
            expectedRootCanonicalPath: issue.expectedRootCanonicalPath,
            expectedCurrentKind: issue.currentKind ?? null,
            expectedCurrentHashAlgorithm: issue.currentHashAlgorithm ?? null,
            expectedCurrentHash: issue.currentHash ?? null
          });
        }
      } else {
        manual.push({
          type: "manual-attention",
          skill: issue.skill,
          root: issue.root,
          manager: issue.manager,
          issue: issue.type,
          reason: "installation-manager-boundary"
        });
      }
      continue;
    }

    if (issue.manager && issue.manager !== "skills-cli") {
      manual.push({
        type: "manual-attention",
        skill: issue.skill,
        root: issue.root,
        manager: issue.manager,
        issue: issue.type,
        reason: "installation-manager-boundary"
      });
      continue;
    }

    if (issue.type === "verified-legacy-duplicate") {
      const destination = path.join(quarantineRoot, issue.root, issue.skill);
      prune.push({
        type: "quarantine",
        skill: issue.skill,
        root: issue.root,
        from: issue.path,
        to: destination,
        mustNotExist: true,
        expectedKind: issue.entryKind,
        expectedHashAlgorithm: issue.hashAlgorithm,
        expectedHash: issue.hash,
        expectedResolvedTarget: issue.resolvedTarget,
        canonicalPath: issue.canonicalPath,
        modeledRootPath: issue.modeledRootPath,
        expectedRootCanonicalPath: issue.expectedRootCanonicalPath,
        canonicalRootPath: issue.canonicalRootPath,
        expectedCanonicalRootCanonicalPath: issue.expectedCanonicalRootCanonicalPath,
        expectedCanonicalHashAlgorithm: issue.canonicalHashAlgorithm,
        expectedCanonicalHash: issue.canonicalHash
      });
      restore.push({
        type: "restore",
        skill: issue.skill,
        from: destination,
        to: issue.path,
        mustNotExist: true
      });
      continue;
    }

    if (issue.type !== "protected") {
      blocked.push({
        type: "blocked",
        skill: issue.skill,
        root: issue.root,
        path: issue.path,
        issue: issue.type,
        reason: issue.reason
      });
    }
  }

  return {
    apply: apply.sort(compareStateRecords),
    prune: prune.sort(compareStateRecords),
    replace: replace.sort(compareStateRecords),
    restore: restore.sort(compareStateRecords),
    manual: manual.sort(compareStateRecords),
    blocked: blocked.sort(compareStateRecords)
  };
}

function assertQuarantineOperationSafe(operation, homeDir, quarantineRoot) {
  if (!isPathWithin(operation.from, homeDir)) {
    throw new Error(`quarantine source must stay inside homeDir: ${operation.from}`);
  }
  if (!isPathWithin(operation.to, quarantineRoot)) {
    throw new Error(`quarantine destination must stay inside quarantineRoot: ${operation.to}`);
  }
  if (path.dirname(operation.from) !== operation.modeledRootPath) {
    throw new Error(`quarantine source left its modeled root: ${operation.from}`);
  }
  assertModeledDirectory({
    directory: operation.modeledRootPath,
    homeDir,
    label: "quarantine source root",
    expectedCanonicalPath: operation.expectedRootCanonicalPath
  });
  assertMutationPathWithin(operation.from, homeDir, "quarantine source");
  assertMutationPathWithin(operation.to, homeDir, "quarantine destination");
  if (!fs.existsSync(operation.from)) {
    throw new Error(`quarantine source no longer exists: ${operation.from}`);
  }
  if (fs.existsSync(operation.to)) {
    throw new Error(`quarantine destination already exists: ${operation.to}`);
  }

  const stat = fs.lstatSync(operation.from);
  const currentKind = stat.isSymbolicLink()
    ? "symlink"
    : stat.isDirectory()
      ? "directory"
      : stat.isFile()
        ? "file"
        : "other";
  if (currentKind !== operation.expectedKind) {
    throw new Error(`quarantine source type changed: ${operation.from}`);
  }
  if (operation.expectedResolvedTarget !== null && operation.expectedResolvedTarget !== undefined) {
    let resolvedTarget;
    try {
      resolvedTarget = fs.realpathSync(operation.from);
    } catch {
      resolvedTarget = null;
    }
    if (resolvedTarget !== operation.expectedResolvedTarget) {
      throw new Error(`quarantine symlink target changed: ${operation.from}`);
    }
  }
  if (operation.expectedHash) {
    const current = stat.isSymbolicLink()
      ? hashObservedPath(fs.realpathSync(operation.from))
      : hashObservedPath(operation.from);
    if (
      current.hashAlgorithm !== operation.expectedHashAlgorithm ||
      current.hash !== operation.expectedHash
    ) {
      throw new Error(`quarantine source content changed: ${operation.from}`);
    }
  }
  if (operation.type === "quarantine") {
    if (!operation.canonicalPath || !fs.existsSync(operation.canonicalPath)) {
      throw new Error(`canonical entry no longer exists: ${operation.canonicalPath}`);
    }
    if (path.dirname(operation.canonicalPath) !== operation.canonicalRootPath) {
      throw new Error(`canonical entry left its modeled root: ${operation.canonicalPath}`);
    }
    assertModeledDirectory({
      directory: operation.canonicalRootPath,
      homeDir,
      label: "canonical managed root",
      expectedCanonicalPath: operation.expectedCanonicalRootCanonicalPath
    });
    assertMutationPathWithin(operation.canonicalPath, homeDir, "canonical entry");
    const canonical = hashObservedPath(operation.canonicalPath);
    if (
      canonical.hashAlgorithm !== operation.expectedCanonicalHashAlgorithm ||
      canonical.hash !== operation.expectedCanonicalHash
    ) {
      throw new Error(`canonical entry content changed: ${operation.canonicalPath}`);
    }
  }
}

function writeJsonAtomic(file, value, homeDir = null) {
  if (homeDir) assertMutationPathWithin(file, homeDir, "state file");
  const temporary = `${file}.tmp-${process.pid}-${crypto.randomBytes(6).toString("hex")}`;
  if (homeDir) assertMutationPathWithin(temporary, homeDir, "temporary state file");
  fs.writeFileSync(temporary, `${JSON.stringify(value, null, 2)}\n`, { flag: "wx" });
  if (homeDir) {
    assertMutationPathWithin(file, homeDir, "state file");
    assertMutationPathWithin(temporary, homeDir, "temporary state file");
  }
  fs.renameSync(temporary, file);
}

function writeReconcilerState({ homeDir, records }) {
  if (!Array.isArray(records)) throw new Error("reconciler state records must be an array");
  const statePath = resolveHomePath(
    path.resolve(homeDir),
    RECONCILER_STATE_RELATIVE_PATH,
    "reconciler state path"
  );
  const parsed = parseReconcilerStateFromValue({ version: 1, records }, statePath);
  if (parsed.errors.length > 0) throw new Error("reconciler state records are invalid");
  assertModeledDirectory({
    directory: path.dirname(statePath),
    homeDir,
    label: "reconciler state parent",
    allowMissing: true
  });
  fs.mkdirSync(path.dirname(statePath), { recursive: true });
  assertModeledDirectory({
    directory: path.dirname(statePath),
    homeDir,
    label: "reconciler state parent"
  });
  writeJsonAtomic(statePath, { version: 1, records: parsed.records }, homeDir);
  return statePath;
}

function quarantinePlannedEntries({ report, operations, quarantineRoot, createdAt }) {
  if (!Array.isArray(operations)) throw new Error("operations must be an array");
  if (operations.length === 0) return null;
  if (!path.isAbsolute(quarantineRoot)) throw new Error("quarantineRoot must be an absolute path");
  const homeDir = report.inventory.homeDir;
  if (!isPathWithin(quarantineRoot, homeDir) || quarantineRoot === homeDir) {
    throw new Error("quarantineRoot must be a child of the inspected homeDir");
  }
  const quarantineBase = path.join(homeDir, ".skill-quarantine");
  if (quarantineRoot === quarantineBase || !isPathWithin(quarantineRoot, quarantineBase)) {
    throw new Error("quarantineRoot must be a run directory under homeDir/.skill-quarantine");
  }
  for (const root of report.inventory.roots) {
    if (isPathWithin(quarantineRoot, root.path) || isPathWithin(root.path, quarantineRoot)) {
      throw new Error(`quarantineRoot must not overlap skill root ${root.path}`);
    }
  }
  if (fs.existsSync(quarantineRoot)) {
    throw new Error(`quarantineRoot already exists: ${quarantineRoot}`);
  }
  assertModeledDirectory({
    directory: quarantineRoot,
    homeDir,
    label: "quarantine root",
    allowMissing: true
  });

  for (const operation of operations) {
    if (!new Set(["quarantine", "quarantine-unverified", "quarantine-update"]).has(operation.type)) {
      throw new Error(`unsupported quarantine operation: ${operation.type}`);
    }
    assertQuarantineOperationSafe(operation, homeDir, quarantineRoot);
  }

  fs.mkdirSync(quarantineRoot, { recursive: true });
  assertModeledDirectory({ directory: quarantineRoot, homeDir, label: "quarantine root" });
  const manifestPath = path.join(quarantineRoot, "manifest.json");
  const manifest = {
    version: 2,
    createdAt,
    homeDir,
    quarantineRoot,
    quarantineCanonicalPath: fs.realpathSync(quarantineRoot),
    roots: Object.fromEntries(
      report.inventory.roots
        .filter(({ id }) => operations.some(({ root }) => root === id))
        .map(({ id, path: rootPath, canonicalPath }) => [
          id,
          { path: rootPath, canonicalPath }
        ])
    ),
    entries: operations.map((operation) => ({
      skill: operation.skill,
      root: operation.root,
      reason: operation.type,
      originalPath: operation.from,
      quarantinedPath: operation.to,
      status: "pending"
    }))
  };
  writeJsonAtomic(manifestPath, manifest, homeDir);

  for (const [index, operation] of operations.entries()) {
    assertQuarantineOperationSafe(operation, homeDir, quarantineRoot);
    assertModeledDirectory({
      directory: path.dirname(operation.to),
      homeDir,
      label: "quarantine destination parent",
      allowMissing: true
    });
    fs.mkdirSync(path.dirname(operation.to), { recursive: true });
    assertModeledDirectory({
      directory: path.dirname(operation.to),
      homeDir,
      label: "quarantine destination parent"
    });
    assertMutationPathWithin(operation.from, homeDir, "quarantine source");
    assertMutationPathWithin(operation.to, homeDir, "quarantine destination");
    fs.renameSync(operation.from, operation.to);
    manifest.entries[index].status = "quarantined";
    writeJsonAtomic(manifestPath, manifest, homeDir);
  }

  return { manifestPath, manifest };
}

const quarantineVerifiedDuplicates = quarantinePlannedEntries;

function restoreQuarantine({ manifestPath, homeDir }) {
  if (!path.isAbsolute(homeDir)) throw new Error("homeDir must be an absolute path");
  if (!path.isAbsolute(manifestPath)) throw new Error("manifestPath must be an absolute path");
  const resolvedHome = path.resolve(homeDir);
  const resolvedManifest = path.resolve(manifestPath);
  if (!isPathWithin(resolvedManifest, resolvedHome)) {
    throw new Error("restore manifest must stay inside homeDir");
  }
  assertModeledDirectory({
    directory: path.dirname(resolvedManifest),
    homeDir: resolvedHome,
    label: "restore manifest parent"
  });

  let manifest;
  try {
    manifest = JSON.parse(fs.readFileSync(resolvedManifest, "utf8"));
  } catch (error) {
    throw new Error(`could not read restore manifest: ${error.message}`);
  }
  if (
    !isObject(manifest) ||
    manifest.version !== 2 ||
    manifest.homeDir !== resolvedHome ||
    !path.isAbsolute(manifest.quarantineRoot) ||
    typeof manifest.quarantineCanonicalPath !== "string" ||
    manifest.quarantineRoot === path.join(resolvedHome, ".skill-quarantine") ||
    !isPathWithin(manifest.quarantineRoot, path.join(resolvedHome, ".skill-quarantine")) ||
    !isPathWithin(resolvedManifest, manifest.quarantineRoot) ||
    !isObject(manifest.roots) ||
    !Array.isArray(manifest.entries)
  ) {
    throw new Error("restore manifest is invalid or belongs to a different homeDir");
  }
  assertModeledDirectory({
    directory: manifest.quarantineRoot,
    homeDir: resolvedHome,
    label: "restore quarantine root",
    expectedCanonicalPath: manifest.quarantineCanonicalPath
  });

  const restored = [];
  for (const entry of manifest.entries) {
    if (!isObject(entry)) throw new Error("restore manifest contains an invalid entry");
    if (entry.status === "restored" || entry.status === "not-moved") continue;
    if (!new Set(["pending", "quarantined"]).has(entry.status)) {
      throw new Error(`restore entry has unsupported status for ${entry.skill ?? "unknown"}`);
    }
    const originalRootRecord = manifest.roots[entry.root];
    const originalRoot = originalRootRecord?.path;
    if (
      typeof entry.skill !== "string" ||
      !/^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(entry.skill) ||
      !isObject(originalRootRecord) ||
      typeof originalRoot !== "string" ||
      typeof originalRootRecord.canonicalPath !== "string" ||
      !path.isAbsolute(originalRoot) ||
      !isPathWithin(originalRoot, resolvedHome) ||
      !path.isAbsolute(entry.originalPath) ||
      path.dirname(entry.originalPath) !== originalRoot ||
      path.basename(entry.originalPath) !== entry.skill ||
      !path.isAbsolute(entry.quarantinedPath) ||
      entry.quarantinedPath !== path.join(manifest.quarantineRoot, entry.root, entry.skill)
    ) {
      throw new Error(`restore entry escapes its allowed roots: ${entry.skill ?? "unknown"}`);
    }
    assertModeledDirectory({
      directory: originalRoot,
      homeDir: resolvedHome,
      label: "restore destination root",
      allowMissing: true,
      expectedCanonicalPath: originalRootRecord.canonicalPath
    });
    assertModeledDirectory({
      directory: path.dirname(entry.quarantinedPath),
      homeDir: resolvedHome,
      label: "quarantined entry parent"
    });
    assertMutationPathWithin(entry.originalPath, resolvedHome, "restore destination");
    assertMutationPathWithin(entry.quarantinedPath, resolvedHome, "quarantined entry");
    if (entry.status === "pending" || entry.status === "quarantined") {
      const originalExists = pathEntryExists(entry.originalPath);
      const quarantinedExists = pathEntryExists(entry.quarantinedPath);
      if (originalExists && !quarantinedExists) {
        entry.status = entry.status === "pending" ? "not-moved" : "restored";
        writeJsonAtomic(resolvedManifest, manifest, resolvedHome);
        continue;
      }
      if (!originalExists && quarantinedExists) {
        if (entry.status === "pending") {
          entry.status = "quarantined";
          writeJsonAtomic(resolvedManifest, manifest, resolvedHome);
        }
      } else if (entry.status === "quarantined" && originalExists && quarantinedExists) {
        throw new Error(`restore destination already exists: ${entry.originalPath}`);
      } else if (entry.status === "quarantined" && !originalExists && !quarantinedExists) {
        throw new Error(`quarantined entry is missing: ${entry.quarantinedPath}`);
      } else {
        throw new Error(`restore entry has ambiguous filesystem state: ${entry.skill}`);
      }
    }
    if (!pathEntryExists(entry.quarantinedPath)) {
      throw new Error(`quarantined entry is missing: ${entry.quarantinedPath}`);
    }
    if (pathEntryExists(entry.originalPath)) {
      throw new Error(`restore destination already exists: ${entry.originalPath}`);
    }
    assertModeledDirectory({
      directory: originalRoot,
      homeDir: resolvedHome,
      label: "restore destination root",
      allowMissing: true,
      expectedCanonicalPath: originalRootRecord.canonicalPath
    });
    fs.mkdirSync(path.dirname(entry.originalPath), { recursive: true });
    assertModeledDirectory({
      directory: originalRoot,
      homeDir: resolvedHome,
      label: "restore destination root",
      expectedCanonicalPath: originalRootRecord.canonicalPath
    });
    assertMutationPathWithin(entry.originalPath, resolvedHome, "restore destination");
    assertMutationPathWithin(entry.quarantinedPath, resolvedHome, "quarantined entry");
    fs.renameSync(entry.quarantinedPath, entry.originalPath);
    entry.status = "restored";
    writeJsonAtomic(resolvedManifest, manifest, resolvedHome);
    restored.push(entry.originalPath);
  }
  return { manifestPath: resolvedManifest, restored };
}

module.exports = {
  DEFAULT_ROOT_POLICY,
  RECONCILER_STATE_RELATIVE_PATH,
  TREE_HASH_ALGORITHM,
  assertModeledDirectory,
  assertMutationPathWithin,
  expectedGlobalPlacements,
  hashSkillDirectory,
  inspectSkillRoots,
  pathEntryExists,
  planGlobalSkillOperations,
  quarantinePlannedEntries,
  quarantineVerifiedDuplicates,
  restoreQuarantine,
  reconcileGlobalSkillState,
  writeReconcilerState
};
