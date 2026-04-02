const { app, BrowserWindow, ipcMain } = require('electron');
const path = require('path');
const fs = require('fs');
const { spawn } = require('child_process');

const isDev = process.env.NODE_ENV === 'development';
const appDir = path.join(__dirname, 'app');
const resourcesDir = path.join(__dirname, 'resources');

function createWindow() {
  const win = new BrowserWindow({
    width: 1000,
    height: 800,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false
    }
  });

  win.loadFile(path.join(appDir, 'index.html'));
  if (isDev) win.webContents.openDevTools();
}

app.whenReady().then(() => {
  createWindow();
  app.on('activate', function () {
    if (BrowserWindow.getAllWindows().length === 0) createWindow();
  });
});

app.on('window-all-closed', function () {
  if (process.platform !== 'darwin') app.quit();
});

// IPC: spawn backend processes on demand
ipcMain.handle('spawn-backend', async (event, args) => {
  // args: { goBinary: 'worker'|'worker.exe', pythonExe: 'app'|'app.exe' }
  const goPath = path.join(resourcesDir, args.goBinary);
  const pyPath = path.join(resourcesDir, 'python', args.pythonExe);
  // ensure executables exist
  if (!fs.existsSync(goPath)) throw new Error('Go binary missing: ' + goPath);
  if (!fs.existsSync(pyPath)) throw new Error('Python exe missing: ' + pyPath);

  // spawn Go server (if needed) or use it to produce batches; here we spawn the Go binary as a child process
  const goProc = spawn(goPath, ['-dir', path.join(appDir, 'uploads'), '-batch', '4'], { stdio: ['ignore','pipe','pipe'] });

  // spawn python generator and pipe stdin/stdout
  const pyProc = spawn(pyPath, [], { stdio: ['pipe','pipe','pipe'] });

  // return PIDs so renderer can track
  return { goPid: goProc.pid, pyPid: pyProc.pid };
});
