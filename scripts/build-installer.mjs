import crypto from "node:crypto";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import electronExecutable from "electron";

const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const packageJson = JSON.parse(fs.readFileSync(path.join(repositoryRoot, "package.json"), "utf8"));
const outputDirectory = path.join(repositoryRoot, "dist");
const outputPath = path.join(outputDirectory, "sd-on-radio-installer.sh");
const temporaryDirectory = fs.mkdtempSync(path.join(os.tmpdir(), "sd-on-radio-build-"));
const appDirectory = path.join(temporaryDirectory, "app");
const runtimeDirectory = path.dirname(electronExecutable);
const packagedSourceDirectory = path.join(appDirectory, "resources", "app");

try {
  fs.cpSync(runtimeDirectory, appDirectory, { recursive: true });
  fs.mkdirSync(packagedSourceDirectory, { recursive: true });
  fs.cpSync(path.join(repositoryRoot, "src"), path.join(packagedSourceDirectory, "src"), {
    recursive: true,
  });
  fs.writeFileSync(
    path.join(packagedSourceDirectory, "package.json"),
    `${JSON.stringify({
      name: packageJson.name,
      version: packageJson.version,
      description: packageJson.description,
      main: packageJson.main,
      author: packageJson.author,
      license: packageJson.license,
    }, null, 2)}\n`,
  );

  const launcher = `#!/bin/sh
set -eu
APP_DIR="$(dirname "$(readlink -f "$0")")"
SANDBOX_ARGUMENT="--disable-setuid-sandbox"
if [ -r /proc/sys/kernel/unprivileged_userns_clone ] && [ "$(cat /proc/sys/kernel/unprivileged_userns_clone)" != "1" ]; then
  SANDBOX_ARGUMENT="--no-sandbox"
fi
exec "$APP_DIR/electron" "$SANDBOX_ARGUMENT" "$APP_DIR/resources/app" "$@"
`;
  fs.writeFileSync(path.join(appDirectory, "sd-on-radio"), launcher, { mode: 0o755 });

  const payloadPath = path.join(temporaryDirectory, "payload.tar.gz");
  const tarResult = spawnSync("tar", [
    "--sort=name",
    "--mtime=@0",
    "--owner=0",
    "--group=0",
    "--numeric-owner",
    "-czf",
    payloadPath,
    "-C",
    temporaryDirectory,
    "app",
  ], { stdio: "inherit" });
  if (tarResult.status !== 0) {
    throw new Error("Не удалось создать payload.tar.gz");
  }

  const payload = fs.readFileSync(payloadPath);
  const payloadSha256 = crypto.createHash("sha256").update(payload).digest("hex");
  const header = fs.readFileSync(
    path.join(repositoryRoot, "scripts", "installer-header.sh"),
    "utf8",
  )
    .replaceAll("__APP_VERSION__", packageJson.version)
    .replace("__PAYLOAD_SHA256__", payloadSha256);

  fs.mkdirSync(outputDirectory, { recursive: true });
  fs.writeFileSync(outputPath, header);
  fs.appendFileSync(outputPath, payload);
  fs.chmodSync(outputPath, 0o755);

  const sizeMiB = (fs.statSync(outputPath).size / 1024 / 1024).toFixed(1);
  console.log(`Создан ${outputPath} (${sizeMiB} MiB)`);
} finally {
  fs.rmSync(temporaryDirectory, { recursive: true, force: true });
}
