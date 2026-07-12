const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const test = require("node:test");
const {
  isValidHttpUrl,
  normalizeSettings,
  readSettings,
  writeSettings,
} = require("../src/main/settings");

test("принимает только HTTP(S)-потоки", () => {
  assert.equal(isValidHttpUrl("https://example.com/radio.mp3"), true);
  assert.equal(isValidHttpUrl("http://127.0.0.1:8000/stream"), true);
  assert.equal(isValidHttpUrl("file:///tmp/audio.mp3"), false);
  assert.equal(isValidHttpUrl("javascript:alert(1)"), false);
  assert.equal(isValidHttpUrl("example.com/stream"), false);
});

test("нормализует пользовательские настройки", () => {
  assert.deepEqual(normalizeSettings({
    stations: [
      { name: "  Моё радио  ", url: " https://example.com/live " },
      { name: "", url: "https://example.com/invalid" },
      { name: "FTP", url: "ftp://example.com/live" },
    ],
    last_station: "Моё радио",
    volume: 4,
  }), {
    stations: [{ name: "Моё радио", url: "https://example.com/live" }],
    last_station: "Моё радио",
    volume: 1,
  });
});

test("создаёт и атомарно обновляет settings.json", (context) => {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), "sd-on-radio-test-"));
  const settingsPath = path.join(directory, "nested", "settings.json");
  context.after(() => fs.rmSync(directory, { recursive: true, force: true }));

  assert.deepEqual(readSettings(settingsPath), {
    stations: [],
    last_station: "",
    volume: 0.8,
  });

  writeSettings(settingsPath, {
    stations: [{ name: "Test", url: "https://example.com/radio" }],
    last_station: "Test",
    volume: 0.45,
  });

  assert.deepEqual(readSettings(settingsPath), {
    stations: [{ name: "Test", url: "https://example.com/radio" }],
    last_station: "Test",
    volume: 0.45,
  });
});
