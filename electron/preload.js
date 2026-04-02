const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('desktopAPI', {
  spawnBackend: (opts) => ipcRenderer.invoke('spawn-backend', opts)
});
