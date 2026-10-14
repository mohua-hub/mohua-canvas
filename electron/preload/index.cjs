const { contextBridge, ipcRenderer } = require("electron");
contextBridge.exposeInMainWorld("mohua",Object.freeze({
  version:1,
  invoke:(operation,input)=>ipcRenderer.invoke("mohua:invoke",operation,input),
  importAsset:input=>ipcRenderer.invoke("mohua:import-asset",input),
  openRepository:()=>ipcRenderer.invoke("mohua:open-repository"),
}));
