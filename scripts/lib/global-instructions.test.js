"use strict";

const assert = require("node:assert/strict");
const test = require("node:test");
const {
  buildAll,
  checkGenerated,
  generatedNotice,
  parseOverlay,
  render
} = require("./global-instructions");

function renderWith(template, overlay) {
  const parsed = parseOverlay(overlay, "overlay.md");
  assert.deepEqual(parsed.errors, []);
  return render(template, parsed.slots, "overlay.md");
}

test("committed global instructions match their sources", () => {
  assert.deepEqual(checkGenerated(), []);
});

test("every harness renders without errors", () => {
  for (const result of buildAll()) {
    assert.deepEqual(result.errors, [], result.harness);
    assert.ok(result.output.startsWith(`${generatedNotice}\n`), result.harness);
    assert.doesNotMatch(result.output, /\{\{[a-z0-9_]+\}\}/, result.harness);
  }
});

test("inline and block slots render, and empty block slots vanish", () => {
  const { output, errors } = renderWith(
    "# {{title}}\n\n{{extra}}\n\n## Rules\n\n- run {{invoke}}review\n",
    "<!-- slot: title -->\nFILE.md\n<!-- slot: extra -->\n\n<!-- slot: invoke -->\n$\n"
  );
  assert.deepEqual(errors, []);
  assert.equal(output, `${generatedNotice}\n\n# FILE.md\n\n## Rules\n\n- run $review\n`);
});

test("blank-line runs outside removed slots are preserved", () => {
  const { output, errors } = renderWith(
    "a\n\n{{gone}}\n\nb\n\n\nc\n",
    "<!-- slot: gone -->\n"
  );
  assert.deepEqual(errors, []);
  assert.equal(output, `${generatedNotice}\n\na\n\nb\n\n\nc\n`);
});

test("missing, unused, and multi-line inline slots are errors", () => {
  const missing = renderWith("{{a}} {{b}}\n", "<!-- slot: a -->\nx\n");
  assert.ok(missing.errors.some((error) => error.includes("missing slot b")));

  const unused = renderWith("{{a}}\n", "<!-- slot: a -->\nx\n<!-- slot: b -->\ny\n");
  assert.ok(unused.errors.some((error) => error.includes("slot b is not used")));

  const inline = renderWith("see {{a}}\n", "<!-- slot: a -->\nx\ny\n");
  assert.ok(inline.errors.some((error) => error.includes("multi-line slot a used inline")));
});

test("overlay content outside a slot and duplicate slots are errors", () => {
  const stray = parseOverlay("stray\n<!-- slot: a -->\nx\n", "overlay.md");
  assert.ok(stray.errors.some((error) => error.includes("content before the first slot")));

  const duplicate = parseOverlay("<!-- slot: a -->\nx\n<!-- slot: a -->\ny\n", "overlay.md");
  assert.ok(duplicate.errors.some((error) => error.includes("duplicate slot a")));
});
