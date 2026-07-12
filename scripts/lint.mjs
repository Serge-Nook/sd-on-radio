import fs from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const repositoryRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const roots = ["src", "scripts", "test"];
const files = [];

function collect(directory) {
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    const entryPath = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      collect(entryPath);
    } else if (entry.name.endsWith(".js") || entry.name.endsWith(".mjs")) {
      files.push(entryPath);
    }
  }
}

for (const root of roots) {
  collect(path.join(repositoryRoot, root));
}

let failed = false;
for (const file of files) {
  const result = spawnSync(process.execPath, ["--check", file], { stdio: "inherit" });
  if (result.status !== 0) {
    failed = true;
  }
}

if (failed) {
  process.exitCode = 1;
} else {
  console.log(`Проверено JavaScript-файлов: ${files.length}`);
}
