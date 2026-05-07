import os from "node:os";
import path from "node:path";

export const PACKAGE_NAME = "@netpiedev/drift";
export const BINARY_NAME = process.platform === "win32" ? "drift.exe" : "drift";
export const CACHE_ROOT = process.env.DRIFT_CACHE_DIR || path.join(os.homedir(), ".cache", "netpiedev-drift");
export const BINARY_BASE_URL = process.env.DRIFT_BINARY_BASE_URL || "https://github.com/netpiedev/drift/releases/download";
