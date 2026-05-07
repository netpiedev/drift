#!/usr/bin/env node

import { ensureBinary } from "../download/download.js";

async function main(): Promise<void> {
  if (process.env.DRIFT_SKIP_POSTINSTALL === "1") {
    return;
  }
  try {
    await ensureBinary();
  } catch {
    // Best-effort postinstall. Runtime resolution retries.
  }
}

await main();
