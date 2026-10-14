export type Kind = "text" | "image" | "video" | "audio";
export type CanvasNode = {
  id: string; kind: Kind; x: number; y: number; title: string;
  content: string; prompt: string; channelId: string; model: string; assetId: string; jobId: string;
};
export type CanvasEdge = { id: string; source: string; target: string };
export type CanvasDocument = { version: 1; nodes: CanvasNode[]; edges: CanvasEdge[]; viewport: { x: number; y: number; zoom: number } };
export type ProjectSummary = { id: string; name: string; revision: number; createdAt: string; updatedAt: string };
export type Project = ProjectSummary & { document: CanvasDocument };
export type Model = { name: string; kind: Kind };
export type Channel = { id: string; name: string; pluginId: string; baseUrl: string; models: Model[]; hasCredential: boolean };
export type ConfigureChannel = { id?: string; name: string; pluginId: string; baseUrl: string; models: Model[]; apiKey?: string };
export type Asset = { id: string; name: string; mimeType: string; bytes: number; createdAt: string };
export type Job = {
  id: string; projectId: string; nodeId: string; kind: Kind;
  status: "queued" | "running" | "succeeded" | "failed" | "cancelled";
  text: string; assetId: string; errorCode: string; createdAt: string; updatedAt: string;
};
export type LibraryEntry = { id: string; kind: "prompt" | "skill"; name: string; content: string; createdAt: string; updatedAt: string };
export type Plugin = { id: string; type: "storage" | "media" | "provider"; version: string; apiVersion: string; capabilities: Kind[]; status: string };
export type Result<T> = { ok: true; data: T } | { ok: false; error: { code: string; message: string } };
type IDInput = { id: string };
export type Operations = {
  listProjects: { input: Record<string, never>; output: ProjectSummary[] };
  createProject: { input: { name: string }; output: Project };
  openProject: { input: IDInput; output: Project };
  saveProject: { input: { id: string; name: string; expectedRevision: number; document: CanvasDocument }; output: Project };
  deleteProject: { input: IDInput; output: null };
  listChannels: { input: Record<string, never>; output: Channel[] };
  configureChannel: { input: ConfigureChannel; output: Channel };
  deleteChannel: { input: IDInput; output: null };
  listAssets: { input: Record<string, never>; output: Asset[] };
  deleteAsset: { input: IDInput; output: null };
  listPlugins: { input: Record<string, never>; output: Plugin[] };
  submitGeneration: { input: { requestId: string; projectId: string; nodeId: string; channelId: string; model: string; kind: Kind; prompt: string }; output: Job };
  listJobs: { input: { projectId: string }; output: Job[] };
  getJob: { input: IDInput; output: Job };
  cancelJob: { input: IDInput; output: Job };
  listLibrary: { input: Record<string, never>; output: LibraryEntry[] };
  saveLibraryEntry: { input: { id?: string; kind: "prompt" | "skill"; name: string; content: string }; output: LibraryEntry };
  deleteLibraryEntry: { input: IDInput; output: null };
};
export interface DesktopAPI {
  readonly version: 1;
  invoke<K extends keyof Operations>(operation: K, input: Operations[K]["input"]): Promise<Result<Operations[K]["output"]>>;
  importAsset(input: { name: string; bytes: ArrayBuffer }): Promise<Result<Asset>>;
  openRepository(): Promise<Result<null>>;
}
declare global { interface Window { mohua: DesktopAPI } }
