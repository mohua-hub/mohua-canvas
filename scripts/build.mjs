import { spawn } from "node:child_process";
import { mkdir } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import path from "node:path";
const root=path.resolve(path.dirname(fileURLToPath(import.meta.url)),"..");
async function run(command,args,env=process.env){
  await new Promise((resolve,reject)=>{
    const child=spawn(command,args,{cwd:root,env,stdio:"inherit",windowsHide:true});
    child.on("error",reject);child.on("exit",code=>code===0?resolve():reject(new Error("BUILD_FAILED")));
  });
}
await mkdir(path.join(root,"dist/bin"),{recursive:true});
await run(process.execPath,[path.join(root,"node_modules/vite/bin/vite.js"),"build"]);
const binary=process.platform==="win32"?"mohua-service.exe":"mohua-service";
await run("go",["build","-trimpath","-ldflags=-s -w","-o",path.join("dist/bin",binary),"./cmd/mohua"],{...process.env,CGO_ENABLED:"0"});
