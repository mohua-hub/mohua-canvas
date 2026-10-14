import { spawn } from "node:child_process";
import { createRequire } from "node:module";
import path from "node:path";
import { fileURLToPath } from "node:url";
const require=createRequire(import.meta.url);
const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),"..");
const vite=spawn(process.execPath,[path.join(root,"node_modules/vite/bin/vite.js")],{cwd:root,stdio:"inherit",windowsHide:true});
let desktop,closing=false;
function stop(code=0){if(closing)return;closing=true;desktop?.kill();vite.kill();process.exitCode=code;}
vite.on("error",()=>stop(1));vite.on("exit",code=>stop(code??1));
process.on("SIGINT",()=>stop());process.on("SIGTERM",()=>stop());
const deadline=Date.now()+30000;
while(!closing){
  try { const res=await fetch("http://127.0.0.1:5173",{signal:AbortSignal.timeout(1000)});if(res.ok)break; } catch {}
  if(Date.now()>deadline){console.error("VITE_START_TIMEOUT");stop(1);break;}
  await new Promise(resolve=>setTimeout(resolve,250));
}
if(!closing){
  desktop=spawn(require("electron"),[root],{cwd:root,stdio:"inherit",windowsHide:true,env:{...process.env,MOHUA_DEV:"1"}});
  desktop.on("error",()=>stop(1));desktop.on("exit",code=>stop(code??0));
}
