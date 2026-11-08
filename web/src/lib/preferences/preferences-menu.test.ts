import { shouldKeepPreferencesOpen } from "./preferences-menu.ts";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

function node(slot: string | null, parent?: ReturnType<typeof node>) {
  return {
    slot,
    closest(selector: string) {
      let wants = "";
      if (selector.includes("select-content")) wants = "select-content";
      else if (selector.includes("select-item")) wants = "select-item";
      if (slot && wants === slot) return this;
      return parent?.closest(selector) ?? null;
    },
  };
}

test("dropdown click inside the preferences panel is not an outside click", () => {
  const menu = node("select-content");
  const item = node("select-item", menu);
  assert.equal(shouldKeepPreferencesOpen({ target: item }), true);
  assert.equal(
    shouldKeepPreferencesOpen({
      target: null,
      detail: { originalEvent: { target: item } },
    }),
    true,
  );
});

test("a click on the preferences panel itself still counts as outside", () => {
  assert.equal(shouldKeepPreferencesOpen({ target: node(null) }), false);
  assert.equal(shouldKeepPreferencesOpen({ target: null }), false);
});

test("page font follows the selected preference instead of a hardcoded face", () => {
  const css = readFileSync(new URL("../../app/globals.css", import.meta.url), "utf8");
  const bodyRule = css.match(/body\s*\{[^}]*\}/)?.[0] ?? "";
  assert.match(bodyRule, /font-family:\s*var\(--font-active\)/);
  assert.doesNotMatch(bodyRule, /JetBrains Mono/);
});

test("centered layout is narrower than the window on a normal laptop", () => {
  const layout = readFileSync(new URL("../../app/(main)/layout.tsx", import.meta.url), "utf8");
  assert.match(layout, /max-w-\[calc\(100%-8rem\)\]/);
  assert.doesNotMatch(layout, /max-w-screen-2xl/);
});

test("next font loader can read every localFont call", () => {
  const registry = readFileSync(new URL("../fonts/registry.ts", import.meta.url), "utf8");
  assert.doesNotMatch(registry, /\.\.\.fontOpts/);
  assert.match(registry, /display: "swap"/);
});
