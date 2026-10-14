const { spawn, execFile } = require("node:child_process");
const { randomBytes } = require("node:crypto");
const { mkdir, readFile, writeFile } = require("node:fs/promises");
const path = require("node:path");

async function masterKey(dataDir, safeStorage) {
  if (!safeStorage.isEncryptionAvailable()) throw new Error("CREDENTIAL_STORAGE_UNAVAILABLE");
  await mkdir(dataDir, { recursive:true });
  const location = path.join(dataDir, "vault-key.bin");
  let sealed;
  try { sealed = await readFile(location); } catch (error) {
    if (error.code !== "ENOENT") throw new Error("CREDENTIAL_STORAGE_FAILED");
    const key = randomBytes(32).toString("base64");
    await writeFile(location, safeStorage.encryptString(key), { flag:"wx", mode:0o600 });
    return key;
  }
  const key = safeStorage.decryptString(sealed);
  if (Buffer.from(key, "base64").length !== 32) throw new Error("CREDENTIAL_STORAGE_FAILED");
  return key;
}
async function startService({app,safeStorage,onUnexpectedExit}) {
  const root = path.resolve(__dirname, "../..");
  const dataDir = path.join(app.getPath("userData"),"workspace-v1");
  const key = await masterKey(dataDir,safeStorage);
  const token = randomBytes(32).toString("hex");
  const executable = app.isPackaged ? path.join(process.resourcesPath,"bin",process.platform === "win32" ? "mohua-service.exe" : "mohua-service") : "go";
  const child = spawn(executable,app.isPackaged ? [] : ["run","./cmd/mohua"],{
    cwd:app.isPackaged ? process.resourcesPath : root, windowsHide:true, stdio:["pipe","pipe","pipe"],
    env:{...process.env,MOHUA_DATA_DIR:dataDir,MOHUA_SERVICE_TOKEN:token,MOHUA_VAULT_KEY:key,
      MOHUA_PLUGIN_CONFIG:app.isPackaged ? path.join(process.resourcesPath,"config/plugins.json") : path.join(root,"config/plugins.json")},
  });
  let closing = false, ready = false, port, alive = true;
  const available = () => alive && ready && !closing;
  child.stderr.on("data", () => {}); // Do not forward subprocess output or credentials.
  child.stdin.on("error", () => {});
  child.on("error", () => { alive=false; });
  child.on("exit", () => { alive=false; if (ready && !closing) onUnexpectedExit(); });
  async function stop() {
    if (closing) return;
    closing=true; child.stdin.end();
    if (!alive) return;
    await new Promise(resolve => {
      const timer=setTimeout(() => {
        if (process.platform==="win32" && child.pid) execFile("taskkill",["/PID",String(child.pid),"/T","/F"],{windowsHide:true},()=>{});
        else child.kill("SIGKILL");
        resolve();
      },10000);
      child.once("exit",()=>{clearTimeout(timer);resolve()});
    });
  }
  try {
    port = await new Promise((resolve,reject) => {
      let buffer="";
      const timer=setTimeout(()=>reject(new Error("SERVICE_START_TIMEOUT")),60000);
      const finish = (error,value) => { clearTimeout(timer); child.stdout.removeListener("data",read); error ? reject(error) : resolve(value) };
      const read = chunk => {
        buffer+=chunk.toString("utf8");
        if (buffer.length>4096) { finish(new Error("SERVICE_START_FAILED")); return; }
        const match=buffer.match(/(?:^|\n)MOHUA_READY=(\d+)\r?\n/);
        if (match) { const parsed=Number(match[1]); if (parsed>0 && parsed<=65535) finish(null,parsed); else finish(new Error("SERVICE_START_FAILED")) }
      };
      child.stdout.on("data",read);
      child.once("error",()=>finish(new Error("SERVICE_EXECUTABLE_UNAVAILABLE")));
      child.once("exit",()=>finish(new Error("SERVICE_START_FAILED")));
    });
    const health=await fetch("http://127.0.0.1:"+port+"/v1/health",{headers:{Authorization:"Bearer "+token},signal:AbortSignal.timeout(10000)});
    const result=await health.json();
    if (!health.ok || !result.ok || result.data.version!==1 || !alive) throw new Error("SERVICE_HEALTH_FAILED");
    ready=true;
  } catch(error) { await stop(); throw error; }
  return {
    stop, available,
    async request(method, suffix, body, extraHeaders = {}) {
      if (!available()) throw new Error("SERVICE_UNAVAILABLE");
      return fetch("http://127.0.0.1:"+port+"/v1"+suffix,{
        method,body,headers:{Authorization:"Bearer "+token,...extraHeaders},
        signal:AbortSignal.timeout(120000),
      });
    },
  };
}
module.exports={startService};
