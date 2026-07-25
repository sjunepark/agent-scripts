import assert from "node:assert/strict";
import test from "node:test";

import { renderReport } from "../src/cli.js";

test("routes CSV reports through the formatter", () => {
  const report = renderReport([{ name: "alpha", total: 2 }], "csv");

  assert.equal(report, "name,total\nalpha,2");
});
