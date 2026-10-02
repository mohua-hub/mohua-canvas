import { readdir, readFile, writeFile } from "node:fs/promises";
import { basename, join } from "node:path";

const [tag, artifactsDir, releaseFile] = process.argv.slice(2);
if (!tag || !artifactsDir || !releaseFile) throw new Error("Usage: node scripts/create-updater-manifest.mjs <tag> <artifacts-dir> <release-json>");
const releaseMetadata = JSON.parse(await readFile(releaseFile, "utf8"));
if (releaseMetadata.draft) throw new Error("Publish the GitHub Release before generating updater URLs");
const releaseAssets = releaseMetadata.assets;

async function listFiles(directory) {
    const entries = await readdir(directory, { withFileTypes: true });
    const nested = await Promise.all(entries.map((entry) => {
        const path = join(directory, entry.name);
        return entry.isDirectory() ? listFiles(path) : [path];
    }));
    return nested.flat();
}

const files = await listFiles(artifactsDir);
function findAsset(description, matches) {
    const found = files.filter((path) => matches(basename(path)));
    if (found.length !== 1) throw new Error(`Expected one ${description} updater asset, found ${found.length}`);
    return found[0];
}

async function platform(id, description, matches) {
    const packagePath = findAsset(description, matches);
    const signature = (await readFile(`${packagePath}.sig`, "utf8")).trim();
    if (!signature) throw new Error(`${description} updater signature is empty`);
    const asset = releaseAssets.find((item) => item.label === basename(packagePath));
    if (!asset?.browser_download_url) throw new Error(`${description} published asset is missing`);
    return [id, { signature, url: asset.browser_download_url }];
}

const windows = await platform("windows-x86_64", "Windows NSIS", (name) => /-setup\.exe$/i.test(name));

const version = tag.replace(/^v/, "");
const changelog = await readFile(new URL("../CHANGELOG.md", import.meta.url), "utf8");
const release = changelog.split(/^## /m).find((section) => section.split(/\r?\n/, 1)[0].split(" ")[0].replace(/^v/, "") === version);
if (!release) throw new Error(`CHANGELOG.md 缺少 ${tag}`);
const notes = release.split(/\r?\n/).slice(1).filter((line) => line.startsWith("+ ")).map((line) => line.slice(2)).join("\n");
const manifest = {
    version,
    notes,
    pub_date: new Date().toISOString(),
    platforms: Object.fromEntries([windows]),
};

await writeFile(join(artifactsDir, "latest.json"), JSON.stringify(manifest, null, 2) + "\n");
