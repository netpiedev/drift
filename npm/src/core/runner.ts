import fs from "node:fs/promises";
import path from "node:path";
import { execa } from "execa";

import { BINARY_NAME } from "./constants.js";
import { ensureBinary } from "../download/download.js";

async function localBinaryCandidates(): Promise<string[]> {
  const cwd = process.cwd();
  const candidates = [
    process.env.DRIFT_BINARY_PATH || "",
    path.join(cwd, "core", "bin", BINARY_NAME),
    path.join(cwd, "bin", BINARY_NAME),
    path.join(cwd, BINARY_NAME)
  ].filter(Boolean);
  return candidates;
}

async function exists(filePath: string): Promise<boolean> {
  try {
    await fs.access(filePath);
    return true;
  } catch {
    return false;
  }
}

export async function resolveBinaryPath(): Promise<string> {
  for (const candidate of await localBinaryCandidates()) {
    if (await exists(candidate)) {
      return candidate;
    }
  }
  return ensureBinary();
}

export async function runDrift(args: string[]): Promise<number> {
  const binary = await resolveBinaryPath();

  const subprocess = execa(binary, args, {
    stdio: "inherit",
    preferLocal: false
  });

  try {
    await subprocess;
    return 0;
  } catch (error: any) {
    return typeof error?.exitCode === "number" ? error.exitCode : 1;
  }
}
