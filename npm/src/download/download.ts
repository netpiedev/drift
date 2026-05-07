import fs from "node:fs/promises";
import { createWriteStream } from "node:fs";
import path from "node:path";
import { pipeline } from "node:stream/promises";

import { BINARY_BASE_URL, BINARY_NAME, CACHE_ROOT } from "../core/constants.js";
import { detectPlatform, type SupportedPlatform } from "../platform/platform.js";

function packageVersion(): string {
  return process.env.npm_package_version || "0.1.0";
}

export function cachedBinaryPath(): string {
  const platform = detectPlatform();
  return path.join(CACHE_ROOT, `v${packageVersion()}`, platform, BINARY_NAME);
}

type ReleaseAsset = {
  name: string;
  browser_download_url: string;
};

function candidateAssetNames(platform: SupportedPlatform): string[] {
  const candidates: Record<SupportedPlatform, string[]> = {
    "linux-amd64": [
      "linux-amd64-drift",
      "drift_linux_amd64",
      "drift-Linux-x86_64",
      "drift-linux-amd64"
    ],
    "linux-arm64": [
      "linux-arm64-drift",
      "drift_linux_arm64",
      "drift-Linux-arm64",
      "drift-linux-arm64",
      "drift-Linux-aarch64"
    ],
    "darwin-amd64": [
      "darwin-amd64-drift",
      "drift_darwin_amd64",
      "drift-Darwin-x86_64",
      "drift-darwin-amd64"
    ],
    "darwin-arm64": [
      "darwin-arm64-drift",
      "drift_darwin_arm64",
      "drift-Darwin-arm64",
      "drift-darwin-arm64",
      "drift-Darwin-aarch64"
    ],
    "windows-amd64": [
      "windows-amd64-drift.exe",
      "windows-amd64-drift",
      "drift_windows_amd64.exe",
      "drift-Windows-x86_64.exe",
      "drift-windows-amd64.exe"
    ]
  };
  return candidates[platform];
}

async function downloadTo(target: string, url: string): Promise<boolean> {
  const response = await fetch(url);
  if (!response.ok || !response.body) {
    return false;
  }
  await pipeline(response.body as unknown as NodeJS.ReadableStream, createWriteStream(target));
  if (process.platform !== "win32") {
    await fs.chmod(target, 0o755);
  }
  return true;
}

async function fetchReleaseAssets(version: string): Promise<ReleaseAsset[]> {
  const apiURL = `https://api.github.com/repos/netpiedev/drift/releases/tags/v${version}`;
  const response = await fetch(apiURL, {
    headers: {
      Accept: "application/vnd.github+json"
    }
  });
  if (!response.ok) {
    return [];
  }
  const json = (await response.json()) as { assets?: ReleaseAsset[] };
  return Array.isArray(json.assets) ? json.assets : [];
}

function selectAssetFromRelease(assets: ReleaseAsset[], platform: SupportedPlatform): ReleaseAsset | null {
  if (assets.length === 0) {
    return null;
  }

  const desired = new Set(candidateAssetNames(platform).map((n) => n.toLowerCase()));
  for (const asset of assets) {
    if (desired.has(asset.name.toLowerCase())) {
      return asset;
    }
  }

  const [os, arch] = platform.split("-");
  for (const asset of assets) {
    const name = asset.name.toLowerCase();
    if (!name.includes("drift")) {
      continue;
    }
    if (name.includes(os) && name.includes(arch)) {
      return asset;
    }
  }

  return null;
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

  const version = packageVersion();
  const platform = detectPlatform();
  const names = candidateAssetNames(platform);

  for (const name of names) {
    const url = `${BINARY_BASE_URL}/v${version}/${name}`;
    if (await downloadTo(target, url)) {
      return target;
    }
  }

  const assets = await fetchReleaseAssets(version);
  const matched = selectAssetFromRelease(assets, platform);
  if (matched) {
    const lower = matched.name.toLowerCase();
    if (lower.endsWith(".tar.gz") || lower.endsWith(".zip") || lower.endsWith(".tgz")) {
      throw new Error(
        `Found release asset '${matched.name}', but compressed archives are not supported by the wrapper downloader yet. ` +
          `Expected raw binary artifact for ${platform}.`
      );
    }
    if (await downloadTo(target, matched.browser_download_url)) {
      return target;
    }
  }

  const known = assets.length > 0 ? assets.map((a) => a.name).join(", ") : "<none>";
  throw new Error(
    `Failed to download Drift binary for ${platform} version v${version}. ` +
      `Tried names: ${names.join(", ")}. Release assets: ${known}`
  );
}
