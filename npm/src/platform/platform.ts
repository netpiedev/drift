export type SupportedPlatform =
  | "linux-amd64"
  | "linux-arm64"
  | "darwin-amd64"
  | "darwin-arm64"
  | "windows-amd64";

export function detectPlatform(): SupportedPlatform {
  const os = process.platform;
  const arch = process.arch;

  if (os === "linux" && arch === "x64") return "linux-amd64";
  if (os === "linux" && arch === "arm64") return "linux-arm64";
  if (os === "darwin" && arch === "x64") return "darwin-amd64";
  if (os === "darwin" && arch === "arm64") return "darwin-arm64";
  if (os === "win32" && arch === "x64") return "windows-amd64";

  throw new Error(`Unsupported platform: ${os}-${arch}`);
}
