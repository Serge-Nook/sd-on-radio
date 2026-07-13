const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const test = require("node:test");
const vm = require("node:vm");

test("renderer-скрипты не создают конфликтующих глобальных объявлений", () => {
  const context = vm.createContext({ window: {} });
  const rendererDirectory = path.join(__dirname, "..", "src", "renderer");

  for (const fileName of ["stations.js", "player.js"]) {
    const source = fs.readFileSync(path.join(rendererDirectory, fileName), "utf8");
    vm.runInContext(source, context, { filename: fileName });
  }

  vm.runInContext(
    "const { RadioPlayer } = window.SdOnRadio; globalThis.RadioPlayerType = typeof RadioPlayer;",
    context,
  );

  assert.equal(context.RadioPlayerType, "function");
});
