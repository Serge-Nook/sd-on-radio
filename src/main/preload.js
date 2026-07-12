const { contextBridge, ipcRenderer } = require("electron");

contextBridge.exposeInMainWorld("radioAPI", {
  loadSettings: () => ipcRenderer.invoke("settings:load"),
  saveSettings: (settings) => ipcRenderer.invoke("settings:save", settings),
  getAppInfo: () => ipcRenderer.invoke("app:info"),
  openWebsite: () => ipcRenderer.invoke("app:open-website"),
});
