import type { Asset, CanvasDocument, CanvasNode, Job, Kind, Project } from "../../../contracts/desktop";

const bytes = (s: string) => new TextEncoder().encode(s).length;
const idPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;
export const kinds: readonly Kind[] = ["text", "image", "video", "audio"];
export type CreateNode = { kind: Kind; title: string; content?: string; prompt?: string; x: number; y: number };
export interface CanvasCapabilityPort {
  getSummary(): { projectId: string; name: string; nodes: CanvasNode[]; edges: CanvasDocument["edges"] };
  createNode(input: CreateNode): string;
  connectNodes(input: { source: string; target: string }): string;
  removeNodes(input: { ids: string[] }): void;
}
export type CanvasState = {
  project: Project | null; name: string; document: CanvasDocument; selected: string[];
  dirty: boolean; sequence: number; undoCount: number; redoCount: number;
};
const empty = (): CanvasDocument => ({ version:1, nodes:[], edges:[], viewport:{x:0,y:0,zoom:1} });
export class CanvasController implements CanvasCapabilityPort {
  private state: CanvasState = {project:null,name:"",document:empty(),selected:[],dirty:false,sequence:0,undoCount:0,redoCount:0};
  private listeners = new Set<() => void>();
  private past: CanvasDocument[] = [];
  private future: CanvasDocument[] = [];
  private gesture: CanvasDocument | null = null;
  private consumedJobs=new Set<string>();
  readonly subscribe = (fn: () => void) => { this.listeners.add(fn); return () => { this.listeners.delete(fn); }; };
  readonly getSnapshot = () => this.state;
  private emit(patch: Partial<CanvasState>) {
    this.state = {...this.state,...patch,undoCount:this.past.length,redoCount:this.future.length};
    this.listeners.forEach(fn => fn());
  }
  private commit(document: CanvasDocument, record = true) {
    if (record) { this.past = [...this.past.slice(-49), this.state.document]; this.future = []; }
    this.emit({document,dirty:true,sequence:this.state.sequence+1});
  }
  load(project: Project) {
    this.past=[];this.future=[];this.gesture=null;this.consumedJobs.clear();
    this.emit({project,name:project.name,document:project.document,selected:[],dirty:false,sequence:0});
  }
  clear() { this.past=[];this.future=[];this.gesture=null;this.consumedJobs.clear();this.emit({project:null,name:"",document:empty(),selected:[],dirty:false,sequence:0}); }
  acknowledge(project: Project, sequence: number) {
    if (project.id!==this.state.project?.id) return;
    this.emit({project,dirty:this.state.sequence!==sequence});
  }
  rename(name: string) {
    if (!name.trim() || bytes(name)>120) throw new Error("项目名称需为 1–120 字节");
    this.emit({name,dirty:true,sequence:this.state.sequence+1});
  }
  getSummary() {
    return {projectId:this.state.project?.id??"",name:this.state.name,nodes:structuredClone(this.state.document.nodes),edges:structuredClone(this.state.document.edges)};
  }
  createNode(input: CreateNode) {
    if (!this.state.project) throw new Error("请先创建或打开项目");
    if (!kinds.includes(input.kind) || !input.title.trim() || bytes(input.title)>120 ||
      !Number.isFinite(input.x) || !Number.isFinite(input.y) || Math.abs(input.x)>1e6 || Math.abs(input.y)>1e6 ||
      bytes(input.content??"")>32000 || bytes(input.prompt??"")>32000 || this.state.document.nodes.length>=2000) throw new Error("节点内容或位置无效");
    const id=crypto.randomUUID();
    const node: CanvasNode={id,kind:input.kind,x:input.x,y:input.y,title:input.title,content:input.content??"",prompt:input.prompt??"",channelId:"",model:"",assetId:"",jobId:""};
    this.commit({...this.state.document,nodes:[...this.state.document.nodes,node]});
    this.emit({selected:[id]});return id;
  }
  editNode(input: {id:string;title:string;content:string;prompt:string;channelId:string;model:string}) {
    const node=this.state.document.nodes.find(n=>n.id===input.id);
    if (!node || !input.title.trim() || bytes(input.title)>120 || bytes(input.content)>32000 || bytes(input.prompt)>32000 ||
      bytes(input.model)>200 || (input.channelId!=="" && !idPattern.test(input.channelId))) throw new Error("节点内容无效");
    this.commit({...this.state.document,nodes:this.state.document.nodes.map(n=>n.id===input.id?{...n,...input}:n)});
  }
  connectNodes({source,target}: {source:string;target:string}) {
    const {nodes,edges}=this.state.document;
    if (!nodes.some(n=>n.id===source) || !nodes.some(n=>n.id===target) || source===target ||
      edges.length>=4000 || edges.some(e=>e.source===source&&e.target===target)) throw new Error("无法连接这些节点");
    const visited=new Set<string>();
    const stack=[target];
    while(stack.length) { const id=stack.pop()!;if(id===source)throw new Error("连线不能形成循环");if(visited.has(id))continue;visited.add(id);edges.forEach(e=>{if(e.source===id)stack.push(e.target)}); }
    const id=crypto.randomUUID();this.commit({...this.state.document,edges:[...edges,{id,source,target}]});return id;
  }
  removeNodes({ids}: {ids:string[]}) {
    if (ids.length===0 || ids.some(id=>!idPattern.test(id) || !this.state.document.nodes.some(n=>n.id===id))) throw new Error("节点不存在");
    const removed=new Set(ids);
    this.commit({...this.state.document,nodes:this.state.document.nodes.filter(n=>!removed.has(n.id)),edges:this.state.document.edges.filter(e=>!removed.has(e.source)&&!removed.has(e.target))});
    this.emit({selected:this.state.selected.filter(id=>!removed.has(id))});
  }
  removeEdges(ids: string[]) { const removed=new Set(ids);this.commit({...this.state.document,edges:this.state.document.edges.filter(e=>!removed.has(e.id))}); }
  select(ids: string[]) {
    if(ids.length===this.state.selected.length && ids.every(id=>this.state.selected.includes(id)))return;
    this.emit({selected:ids});
  }
  beginGesture() { this.gesture=this.state.document; }
  moveNode(id:string,x:number,y:number) {
    if(!Number.isFinite(x)||!Number.isFinite(y)||Math.abs(x)>1e6||Math.abs(y)>1e6)return;
    this.commit({...this.state.document,nodes:this.state.document.nodes.map(n=>n.id===id?{...n,x,y}:n)},false);
  }
  endGesture() {
    if(this.gesture && this.gesture!==this.state.document){this.past=[...this.past.slice(-49),this.gesture];this.future=[];this.emit({})}this.gesture=null;
  }
  viewport(viewport:CanvasDocument["viewport"]) {
    const previous=this.state.document.viewport;
    if(previous.x===viewport.x&&previous.y===viewport.y&&previous.zoom===viewport.zoom)return;
    this.commit({...this.state.document,viewport},false);
  }
  attachAsset({nodeId,asset}: {nodeId:string;asset:Asset}) {
    const node=this.state.document.nodes.find(n=>n.id===nodeId);
    if(!node || !asset.mimeType.startsWith(node.kind+"/"))throw new Error("素材类型与节点不匹配");
    this.commit({...this.state.document,nodes:this.state.document.nodes.map(n=>n.id===nodeId?{...n,assetId:asset.id}:n)});
  }
  trackJob(nodeId:string,job:Job) {
    if(job.projectId!==this.state.project?.id || job.nodeId!==nodeId)return;
    this.commit({...this.state.document,nodes:this.state.document.nodes.map(n=>n.id===nodeId?{...n,jobId:job.id}:n)});
  }
  applyJob({nodeId,job}: {nodeId:string;job:Job}) {
    const node=this.state.document.nodes.find(n=>n.id===nodeId);
    if(!node || this.consumedJobs.has(job.id) || node.jobId!==job.id || node.kind!==job.kind || job.projectId!==this.state.project?.id || job.status!=="succeeded")return;
    this.consumedJobs.add(job.id);
    const content=job.kind==="text"?job.text:node.content;
    this.commit({...this.state.document,nodes:this.state.document.nodes.map(n=>n.id===nodeId?{...n,content,assetId:job.assetId,jobId:""}:n)});
  }
  upstreamText(nodeId:string) {
    const {nodes,edges}=this.state.document;const seen=new Set<string>();const queue=[nodeId];const text:string[]=[];
    while(queue.length){const id=queue.shift()!;for(const e of edges){if(e.target!==id||seen.has(e.source))continue;seen.add(e.source);queue.push(e.source);const n=nodes.find(n=>n.id===e.source);if(n?.kind==="text"&&n.content)text.push(n.content)}}
    const result=text.join("\n\n");if(bytes(result)>32000)throw new Error("上游文本过长，请精简后生成");return result;
  }
  private restored(document:CanvasDocument):CanvasDocument {
    return {...document,viewport:this.state.document.viewport,nodes:document.nodes.map(n=>this.consumedJobs.has(n.jobId)?{...n,jobId:""}:n)};
  }
  undo() { const previous=this.past.pop();if(!previous)return;this.future.push(this.state.document);this.commit(this.restored(previous),false); }
  redo() { const next=this.future.pop();if(!next)return;this.past.push(this.state.document);this.commit(this.restored(next),false); }
}
