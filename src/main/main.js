const path = require("node:path");
const os = require("node:os");
const { app, BrowserWindow, ipcMain, shell } = require("electron");
const { readSettings, writeSettings } = require("./settings");

const APP_DIRECTORY = process.env.SD_ON_RADIO_DATA_DIR
  || path.join(os.homedir(), ".local", "share", "sd-on-radio");
const SETTINGS_PATH = path.join(APP_DIRECTORY, "settings.json");
const WEBSITE_URL = "https://sd-on.ru/";

let mainWindow;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 800,
    minWidth: 900,
    minHeight: 600,
    backgroundColor: "#090b10",
    autoHideMenuBar: true,
    show: false,
    title: "SD-ON RADIO",
    icon: path.join(__dirname, "..", "renderer", "assets", "icon.svg"),
    webPreferences: {
      preload: path.join(__dirname, "preload.js"),
      contextIsolation: true,
      nodeIntegration: false,
      sandbox: true,
    },
  });

  mainWindow.loadFile(path.join(__dirname, "..", "renderer", "index.html"));
  mainWindow.once("ready-to-show", () => mainWindow.show());
  mainWindow.webContents.setWindowOpenHandler(() => ({ action: "deny" }));
  mainWindow.webContents.on("will-navigate", (event) => event.preventDefault());
  mainWindow.on("closed", () => {
    mainWindow = undefined;
  });
}

app.setName("SD-ON RADIO");

app.whenReady().then(() => {
  ipcMain.handle("settings:load", () => readSettings(SETTINGS_PATH));
  ipcMain.handle("settings:save", (_event, value) => writeSettings(SETTINGS_PATH, value));
  ipcMain.handle("app:info", () => ({
    name: "SD-ON RADIO",
    version: app.getVersion(),
  }));
  ipcMain.handle("app:open-website", () => shell.openExternal(WEBSITE_URL));

  createWindow();

  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    }
  });
});

app.on("window-all-closed", () => app.quit());
