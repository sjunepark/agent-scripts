"use strict";

const crypto = require("node:crypto");
const fs = require("node:fs");
const path = require("node:path");

const TREE_HASH_ALGORITHM = "tree-sha256-v1";

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

function isPathWithin(candidate, parent) {
  const relative = path.relative(parent, candidate);
  return relative === "" || (!relative.startsWith("..") && !path.isAbsolute(relative));
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

  return {
    homeDir: resolvedHome,
    roots,
    entries: entries.sort(compareStateRecords),
    locks: locks.sort(compareStateRecords),
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
    path: root.path,
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

function classifyExpectedEntry({ placement, root, current, expected, lock }) {
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
  if (lock && lock.source !== null && lock.source !== placement.source) {
    return issueFromPlacement("unclassified", placement, root, "source-provenance-mismatch");
  }
  if (current.hash === expected.hash) return null;
  if (!lock?.hash || lock.hashAlgorithm !== TREE_HASH_ALGORITHM) {
    return issueFromPlacement("unclassified", placement, root, "lock-hash-unverifiable");
  }
  if (current.hash === lock.hash) {
    return issueFromPlacement("outdated", placement, root, "installed-content-is-older-than-expected");
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
    const classification = classifyExpectedEntry({
      placement,
      root: { ...root, path: path.join(root.path, placement.skill) },
      current,
      expected,
      lock
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
      issues.push({
        type: "verified-legacy-duplicate",
        skill: entry.name,
        root: entry.root,
        path: entry.path,
        reason: equivalentSymlink ? "symlink-targets-canonical-entry" : "content-matches-canonical-entry",
        manager: desired.manager,
        canonicalPath: canonical.path
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
  for (const root of report.inventory.roots) {
    if (isPathWithin(quarantineRoot, root.path) || isPathWithin(root.path, quarantineRoot)) {
      throw new Error(`quarantineRoot must not overlap skill root ${root.path}`);
    }
  }

  const apply = [];
  const prune = [];
  const restore = [];
  const manual = [];
  const blocked = [];
  const operationKeys = new Set();

  for (const issue of report.issues) {
    if (issue.type === "missing" || issue.type === "outdated") {
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
            fullDepth: issue.fullDepth
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
        mustNotExist: true
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
    restore: restore.sort(compareStateRecords),
    manual: manual.sort(compareStateRecords),
    blocked: blocked.sort(compareStateRecords)
  };
}

module.exports = {
  DEFAULT_ROOT_POLICY,
  TREE_HASH_ALGORITHM,
  expectedGlobalPlacements,
  hashSkillDirectory,
  inspectSkillRoots,
  planGlobalSkillOperations,
  reconcileGlobalSkillState
};
