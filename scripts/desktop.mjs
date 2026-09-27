import { spawnSync } from "node:child_process";
import { createRequire } from "node:module";
import { cpSync, copyFileSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join, basename } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const web = join(root, "web");
const tauri = join(root, "src-tauri");
const require = createRequire(join(web, "package.json"));
const version = readFileSync(join(root, "VERSION"), "utf8").trim().replace(/^v/, "");
if (!/^\d+\.\d+\.\d+$/.test(version)) throw new Error("VERSION 必须是 vX.Y.Z 或 X.Y.Z");

function run(command, args, options = {}) {
    const result = spawnSync(command, args, { cwd: root, stdio: "inherit", ...options });
    if (result.error) throw result.error;
    if (result.status !== 0) process.exit(result.status ?? 1);
}

function syncVersion() {
    for (const [file, spaces] of [[join(web, "package.json"), 4], [join(tauri, "tauri.conf.json"), 2]]) {
        const config = JSON.parse(readFileSync(file, "utf8"));
        config.version = version;
        writeFileSync(file, JSON.stringify(config, null, spaces) + "\n");
    }
    const cargo = join(tauri, "Cargo.toml");
    writeFileSync(cargo, readFileSync(cargo, "utf8").replace(/^version = ".*"$/m, `version = "${version}"`));
}

function prepareRuntime() {
    const targets = {
        "win32-x64": "x86_64-pc-windows-msvc",
        "darwin-x64": "x86_64-apple-darwin",
        "darwin-arm64": "aarch64-apple-darwin",
        "linux-x64": "x86_64-unknown-linux-gnu",
    };
    const target = targets[`${process.platform}-${process.arch}`];
    if (!target || (process.env.TAURI_ENV_TARGET_TRIPLE && process.env.TAURI_ENV_TARGET_TRIPLE !== target)) {
        throw new Error("请在目标平台和架构上使用原生 Node 构建桌面端");
    }
    mkdirSync(join(tauri, "binaries"), { recursive: true });
    mkdirSync(join(tauri, "resources"), { recursive: true });
    copyFileSync(process.execPath, join(tauri, "binaries", `node-${target}${process.platform === "win32" ? ".exe" : ""}`));
    writeFileSync(join(tauri, "resources", "desktop-build.json"), JSON.stringify({ version, buildId: "development" }));
}

const [command = "dev", ...args] = process.argv.slice(2);
syncVersion();

if (command === "prepare") {
    prepareRuntime();
    run(process.execPath, [require.resolve("next/dist/bin/next"), "build"], {
        cwd: web,
        env: { ...process.env, NEXT_TELEMETRY_DISABLED: "1" },
    });
    const standalone = join(web, ".next", "standalone");
    cpSync(join(web, "public"), join(standalone, "public"), { recursive: true });
    cpSync(join(web, ".next", "static"), join(standalone, ".next", "static"), { recursive: true });
    copyFileSync(join(root, "desktop", "server.cjs"), join(standalone, "desktop-server.cjs"));
    copyFileSync(join(root, "LICENSE"), join(standalone, "LICENSE"));
    const license = await fetch(`https://raw.githubusercontent.com/nodejs/node/${process.version}/LICENSE`);
    if (!license.ok) throw new Error(`无法获取 Node ${process.version} 的许可证`);
    writeFileSync(join(standalone, "NODE-LICENSE.txt"), await license.text());
    const { create } = require("tar");
    await create({
        cwd: standalone,
        file: join(tauri, "resources", "server.tar.gz"),
        gzip: true,
        portable: true,
        follow: true,
        filter: (path) => !basename(path).startsWith(".env") && !path.replaceAll("\\", "/").includes(".next/cache"),
    }, ["."]);
    const buildId = readFileSync(join(web, ".next", "BUILD_ID"), "utf8").trim();
    writeFileSync(join(tauri, "resources", "desktop-build.json"), JSON.stringify({ version, buildId }));
    console.log(`已打包 Next standalone、静态资源与 Node ${process.version}`);
} else if (command !== "sync") {
    if (command === "dev" || command === "build") prepareRuntime();
    run(process.execPath, [join(web, "node_modules", "@tauri-apps", "cli", "tauri.js"), command, ...args]);
}
