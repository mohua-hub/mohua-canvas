import type { Asset, DesktopAPI, Operations, Result } from "../../contracts/desktop";

export class APIError extends Error {
  constructor(readonly code: string, message: string) { super(message); }
}
function unwrap<T>(result: Result<T>): T {
  if (!result.ok) throw new APIError(result.error.code, result.error.message);
  return result.data;
}
export class APIClient {
  constructor(private readonly desktop: DesktopAPI) {
    if (!desktop || desktop.version !== 1) throw new Error("请在 Electron 桌面应用中打开墨华画布");
  }
  async call<K extends keyof Operations>(operation: K, input: Operations[K]["input"]): Promise<Operations[K]["output"]> {
    return unwrap(await this.desktop.invoke(operation, input));
  }
  async importAsset(file: File): Promise<Asset> {
    return unwrap(await this.desktop.importAsset({ name: file.name, bytes: await file.arrayBuffer() }));
  }
  async openRepository() { return unwrap(await this.desktop.openRepository()); }
}
export function assetURL(id: string) { return "mohua-media://asset/" + id; }
export function errorMessage(error: unknown) { return error instanceof Error ? error.message : "操作未能完成"; }
