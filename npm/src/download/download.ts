import fs from "node:fs/promises";
import { createWriteStream } from "node:fs";
import path from "node:path";
import { pipeline } from "node:stream/promises";

import { BINARY_BASE_URL, BINARY_NAME, CACHE_ROOT } from "../core/constants.js";
import { detectPlatform } from "../platform/platform.js";

function packageVersion(): string {
  return process.env.npm_package_version || "0.1.0";
}

export function cachedBinaryPath(): string {
  const platform = detectPlatform();
  return path.join(CACHE_ROOT, `v${packageVersion()}`, platform, BINARY_NAME);
}

export async function ensureBinary(): Promise<string> {
  const target = cachedBinaryPath();
  try {
    await fs.access(target);
    return target;
  } catch {
    // continue
  }

  await fs.mkdir(path.dirname(target), { recursive: true });
  const platform = detectPlatform();
  const archiveName = `${platform}-${BINARY_NAME}`;
  const url = `${BINARY_BASE_URL}/v${packageVersion()}/${archiveName}`;

  const response = await fetch(url);
  if (!response.ok || !response.body) {
    throw new Error(`Failed to download Drift binary from ${url}: HTTP ${response.status}`);
  }

  await pipeline(response.body as unknown as NodeJS.ReadableStream, createWriteStream(target));
  if (process.platform !== "win32") {
    await fs.chmod(target, 0o755);
  }

  return target;
}
