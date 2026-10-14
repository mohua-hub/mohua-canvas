const { app, BrowserWindow, dialog, ipcMain, protocol, safeStorage, shell } = require("electron");
const path = require("node:path");
const { fileURLToPath } = require("node:url");
const { operations, importSchema, id } = require("./operations.cjs");
const { startService } = require("./service.cjs");

protocol.registerSchemesAsPrivileged([{scheme:"mohua-media",privileges:{standard:true,secure:true,supportFetchAPI:true,stream:true}}]);
const repository="https://github.com/mohua-hub/mohua-canvas";
const failure=(code,message)=>({ok:false,error:{code,message}});
const invalid=()=>failure("INVALID_INPUT","输入内容无效");
let window,service,quitting=false;
const dev=!app.isPackaged && process.env.MOHUA_DEV==="1";
const rendererPath=path.resolve(__dirname,"../../dist/renderer/index.html");
function trusted(event) {
  if (!window || event.sender!==window.webContents || event.senderFrame!==event.sender.mainFrame) return false;
  try {
    const current=new URL(event.senderFrame.url);
    if (dev) return current.origin==="http://127.0.0.1:5173";
    return current.protocol==="file:" && fileURLToPath(current)===rendererPath;
  } catch { return false; }
}
async function guarded(event,fn) {
  if (!trusted(event)) return failure("FORBIDDEN","此来源无法调用桌面能力");
  if (!service?.available()) return failure("SERVICE_UNAVAILABLE","本机服务不可用，请重新启动应用");
  try { return await fn(); } catch { return failure("SERVICE_UNAVAILABLE","本机服务未能完成请求"); }
}
function registerIPC() {
  ipcMain.handle("mohua:invoke",(event,operation,input)=>guarded(event,async()=>{
    if (typeof operation!=="string" || !Object.hasOwn(operations,operation)) return invalid();
    const op=operations[operation], parsed=op.schema.safeParse(input);
    if (!parsed.success) return invalid();
    const json=op.method==="POST" ? JSON.stringify(op.body(parsed.data)) : undefined;
    if (json && Buffer.byteLength(json)>8*1024*1024) return invalid();
    const res=await service.request(op.method,op.path(parsed.data),json,json ? {"Content-Type":"application/json"} : {});
    const result=await res.json();
    if (typeof result.ok!=="boolean") return failure("SERVICE_UNAVAILABLE","本机服务响应无效");
    return result;
  }));
  ipcMain.handle("mohua:import-asset",(event,input)=>guarded(event,async()=>{
    const parsed=importSchema.safeParse(input);if(!parsed.success)return invalid();
    const res=await service.request("POST","/assets/import",Buffer.from(parsed.data.bytes),{"Content-Type":"application/octet-stream","X-Asset-Name":encodeURIComponent(parsed.data.name)});
    return res.json();
  }));
  ipcMain.handle("mohua:open-repository",(event)=>guarded(event,async()=>{await shell.openExternal(repository);return {ok:true,data:null}}));
}
function createWindow() {
  window=new BrowserWindow({
    width:1440,height:940,minWidth:1000,minHeight:700,backgroundColor:"#f4f2ed",title:"墨华画布",
    icon:path.join(__dirname,"../resources/icon.ico"),
    webPreferences:{preload:path.join(__dirname,"../preload/index.cjs"),sandbox:true,contextIsolation:true,nodeIntegration:false,webSecurity:true},
  });
  window.webContents.setWindowOpenHandler(()=>({action:"deny"}));
  window.webContents.on("will-navigate",event=>event.preventDefault());
  window.webContents.on("will-attach-webview",event=>event.preventDefault());
  window.webContents.on("will-prevent-unload",event=>{
    const choice=dialog.showMessageBoxSync(window,{
      type:"question",buttons:["返回保存","放弃修改并关闭"],defaultId:0,cancelId:0,
      title:"保留这次创作？",message:"当前画布有未保存的修改。",detail:"返回工作台保存，或放弃本次修改并关闭窗口。",
    });
    if(choice===0)event.preventDefault();
  });
  window.webContents.session.setPermissionRequestHandler((_webContents,_permission,callback)=>callback(false));
  window.webContents.session.setPermissionCheckHandler(()=>false);
  window.webContents.session.webRequest.onHeadersReceived((details,callback)=>{
    const csp=dev
      ? "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' mohua-media: data:; media-src mohua-media:; connect-src 'self' ws://127.0.0.1:5173; object-src 'none'; base-uri 'none'; frame-src 'none'"
      : "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' mohua-media: data:; media-src mohua-media:; connect-src 'none'; object-src 'none'; base-uri 'none'; frame-src 'none'";
    callback({responseHeaders:{...details.responseHeaders,"Content-Security-Policy":[csp]}});
  });
  window.on("closed",()=>{window=null});
  if (dev) window.loadURL("http://127.0.0.1:5173"); else window.loadFile(rendererPath);
}
async function setup() {
  service=await startService({app,safeStorage,onUnexpectedExit:()=>{
    if (!quitting) { dialog.showErrorBox("本机服务已停止","请重新启动墨华画布。未保存的编辑请先保留。"); }
  }});
  protocol.handle("mohua-media",async request=>{
    try {
      const target=new URL(request.url),assetID=target.pathname.slice(1);
      if (target.hostname!=="asset" || target.search || target.hash || !id.safeParse(assetID).success || !["GET","HEAD"].includes(request.method) || !service.available()) return new Response(null,{status:404});
      const headers={};const range=request.headers.get("Range");if(range)headers.Range=range;
      const res=await service.request(request.method,"/assets/"+assetID+"/content",undefined,headers);
      const output=new Headers();
      for (const h of ["Content-Type","Content-Length","Content-Range","Accept-Ranges","Last-Modified"]) { const value=res.headers.get(h);if(value)output.set(h,value) }
      output.set("X-Content-Type-Options","nosniff");
      return new Response(res.body,{status:res.status,headers:output});
    } catch { return new Response(null,{status:503}) }
  });
  registerIPC();createWindow();
}
if (!app.requestSingleInstanceLock()) app.quit();
else {
  app.on("second-instance",()=>{if(window){if(window.isMinimized())window.restore();window.focus()}});
  app.whenReady().then(setup).catch(async()=>{
    dialog.showErrorBox("无法启动墨华画布","请确认 Go 已安装（开发模式），安全凭据存储可用，且数据目录可写。");
    await service?.stop();app.quit();
  });
  app.on("window-all-closed",()=>app.quit());
  app.on("before-quit",event=>{
    if(quitting || !service)return;
    event.preventDefault();
    if(window){window.close();return}
    quitting=true;service.stop().finally(()=>app.quit());
  });
}
