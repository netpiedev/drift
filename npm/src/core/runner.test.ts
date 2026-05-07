import { describe, expect, test } from "bun:test";

import { detectPlatform } from "../platform/platform.js";

describe("platform detection", () => {
  test("returns a supported platform string", () => {
    const value = detectPlatform();
    expect(value.length).toBeGreaterThan(5);
  });
});
