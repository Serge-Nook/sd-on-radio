const fs = require("node:fs");
const path = require("node:path");

const DEFAULT_SETTINGS = Object.freeze({
  stations: [],
  last_station: "",
  volume: 0.8,
});

function isValidHttpUrl(value) {
  if (typeof value !== "string") {
    return false;
  }

  try {
    const url = new URL(value);
    return (url.protocol === "http:" || url.protocol === "https:") && Boolean(url.hostname);
  } catch {
    return false;
  }
}

function normalizeStation(station) {
  if (!station || typeof station !== "object") {
    return null;
  }

  const name = typeof station.name === "string" ? station.name.trim() : "";
  const url = typeof station.url === "string" ? station.url.trim() : "";

  if (!name || name.length > 80 || url.length > 2048 || !isValidHttpUrl(url)) {
    return null;
  }

  return { name, url };
}

function normalizeSettings(value) {
  const source = value && typeof value === "object" ? value : {};
  const stations = Array.isArray(source.stations)
    ? source.stations.map(normalizeStation).filter(Boolean)
    : [];
  const volume = Number.isFinite(source.volume)
    ? Math.min(1, Math.max(0, source.volume))
    : DEFAULT_SETTINGS.volume;
  const lastStation = typeof source.last_station === "string"
    ? source.last_station.slice(0, 80)
    : DEFAULT_SETTINGS.last_station;

  return {
    stations,
    last_station: lastStation,
    volume,
  };
}

function writeSettings(settingsPath, value) {
  const settings = normalizeSettings(value);
  const settingsDirectory = path.dirname(settingsPath);
  const temporaryPath = `${settingsPath}.${process.pid}.${Date.now()}.tmp`;

  fs.mkdirSync(settingsDirectory, { recursive: true, mode: 0o700 });
  fs.writeFileSync(temporaryPath, `${JSON.stringify(settings, null, 2)}\n`, {
    encoding: "utf8",
    mode: 0o600,
  });
  fs.renameSync(temporaryPath, settingsPath);

  return settings;
}

function readSettings(settingsPath) {
  try {
    const value = JSON.parse(fs.readFileSync(settingsPath, "utf8"));
    return normalizeSettings(value);
  } catch (error) {
    if (error.code !== "ENOENT" && !(error instanceof SyntaxError)) {
      throw error;
    }

    return writeSettings(settingsPath, DEFAULT_SETTINGS);
  }
}

module.exports = {
  DEFAULT_SETTINGS,
  isValidHttpUrl,
  normalizeSettings,
  readSettings,
  writeSettings,
};
