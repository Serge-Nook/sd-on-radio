import crypto from "node:crypto";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import electronExecutable from "electron";

const APPIMAGETOOL_URL = "https://github.com/AppImage/appimagetool/releases/download/1.9.1/appimagetool-x86_64.AppImage";
const APPIMAGETOOL_SHA256 = "ed4ce84f0d9caff66f50bcca6ff6f35aae54ce8135408b3fa33abfc3cb384eb0";
const APPIMAGE_RUNTIME_URL = "https://github.com/AppImage/type2-runtime/releases/download/20251108/runtime-x86_64";
const APPIMAGE_RUNTIME_SHA256 = "2fca8b443c92510f1483a883f60061ad09b46b978b2631c807cd873a47ec260d";

const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const packageJson = JSON.parse(fs.readFileSync(path.join(repositoryRoot, "package.json"), "utf8"));
const outputDirectory = path.join(repositoryRoot, "dist");
const outputPath = path.join(
  outputDirectory,
  `SD-ON-RADIO-${packageJson.version}-x86_64.AppImage`,
);
const cacheDirectory = path.join(repositoryRoot, ".cache", "appimage");
const appImageToolPath = path.join(cacheDirectory, "appimagetool-x86_64.AppImage");
const appImageRuntimePath = path.join(cacheDirectory, "runtime-x86_64");
const temporaryDirectory = fs.mkdtempSync(path.join(os.tmpdir(), "sd-on-radio-appimage-"));
const appDirectory = path.join(temporaryDirectory, "SD-ON-RADIO.AppDir");
const runtimeDirectory = path.join(appDirectory, "usr", "lib", "sd-on-radio");
const packagedSourceDirectory = path.join(runtimeDirectory, "resources", "app");

function fileSha256(filePath) {
  return crypto.createHash("sha256").update(fs.readFileSync(filePath)).digest("hex");
}

async function downloadVerified(url, destination, expectedSha256) {
  if (fs.existsSync(destination) && fileSha256(destination) === expectedSha256) {
    return;
  }

  fs.mkdirSync(path.dirname(destination), { recursive: true });
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Не удалось скачать ${url}: HTTP ${response.status}`);
  }

  const temporaryPath = `${destination}.${process.pid}.download`;
  fs.writeFileSync(temporaryPath, Buffer.from(await response.arrayBuffer()));
  const actualSha256 = fileSha256(temporaryPath);
  if (actualSha256 !== expectedSha256) {
    fs.rmSync(temporaryPath, { force: true });
    throw new Error(`Контрольная сумма ${path.basename(destination)} не совпадает`);
  }

  fs.renameSync(temporaryPath, destination);
  fs.chmodSync(destination, 0o755);
}

function writePackagedApplication() {
  fs.cpSync(path.dirname(electronExecutable), runtimeDirectory, { recursive: true });
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

  const appRun = `#!/bin/sh
set -eu
APP_DIR="\${APPDIR:-$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)}"
SANDBOX_ARGUMENT="--disable-setuid-sandbox"
if [ -r /proc/sys/kernel/unprivileged_userns_clone ] && [ "$(cat /proc/sys/kernel/unprivileged_userns_clone)" != "1" ]; then
  SANDBOX_ARGUMENT="--no-sandbox"
fi
exec "$APP_DIR/usr/lib/sd-on-radio/electron" "$SANDBOX_ARGUMENT" "$APP_DIR/usr/lib/sd-on-radio/resources/app" "$@"
`;
  fs.writeFileSync(path.join(appDirectory, "AppRun"), appRun, { mode: 0o755 });

  const desktopEntry = `[Desktop Entry]
Type=Application
Name=SD-ON RADIO
Comment=Интернет-радио для Steam Deck
Exec=sd-on-radio
Icon=sd-on-radio
Terminal=false
Categories=Audio;AudioVideo;Player;
StartupNotify=true
StartupWMClass=SD-ON RADIO
X-AppImage-Version=${packageJson.version}
`;
  fs.writeFileSync(path.join(appDirectory, "sd-on-radio.desktop"), desktopEntry);
  fs.mkdirSync(path.join(appDirectory, "usr", "share", "applications"), { recursive: true });
  fs.writeFileSync(
    path.join(appDirectory, "usr", "share", "applications", "sd-on-radio.desktop"),
    desktopEntry,
  );

  const iconSource = path.join(repositoryRoot, "src", "renderer", "assets", "icon.svg");
  fs.copyFileSync(iconSource, path.join(appDirectory, "sd-on-radio.svg"));
  const iconDirectory = path.join(
    appDirectory,
    "usr",
    "share",
    "icons",
    "hicolor",
    "scalable",
    "apps",
  );
  fs.mkdirSync(iconDirectory, { recursive: true });
  fs.copyFileSync(iconSource, path.join(iconDirectory, "sd-on-radio.svg"));
  fs.symlinkSync("sd-on-radio.svg", path.join(appDirectory, ".DirIcon"));
}

try {
  await Promise.all([
    downloadVerified(APPIMAGETOOL_URL, appImageToolPath, APPIMAGETOOL_SHA256),
    downloadVerified(APPIMAGE_RUNTIME_URL, appImageRuntimePath, APPIMAGE_RUNTIME_SHA256),
  ]);
  writePackagedApplication();
  fs.mkdirSync(outputDirectory, { recursive: true });

  const result = spawnSync(appImageToolPath, [
    "--runtime-file",
    appImageRuntimePath,
    appDirectory,
    outputPath,
  ], {
    env: {
      ...process.env,
      APPIMAGE_EXTRACT_AND_RUN: "1",
      ARCH: "x86_64",
      SOURCE_DATE_EPOCH: "0",
    },
    stdio: "inherit",
  });
  if (result.status !== 0) {
    throw new Error("Не удалось создать AppImage");
  }

  fs.chmodSync(outputPath, 0o755);
  const sizeMiB = (fs.statSync(outputPath).size / 1024 / 1024).toFixed(1);
  console.log(`Создан ${outputPath} (${sizeMiB} MiB)`);
} finally {
  fs.rmSync(temporaryDirectory, { recursive: true, force: true });
}
