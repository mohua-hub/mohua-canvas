import { useCallback, useEffect, useRef, useState, useSyncExternalStore } from "react";
import { AudioLines, BookOpen, Check, ChevronRight, Film, FolderOpen, Image, Layers3, LoaderCircle, Plus, Redo2, Save, Settings2, Sparkles, Trash2, Undo2, X } from "lucide-react";
import type { Asset, Channel, Job, Kind, LibraryEntry, Plugin, Project, ProjectSummary } from "../../contracts/desktop";
import { APIClient, APIError, errorMessage } from "../services/api";
import { CanvasController } from "../features/canvas/controller";
import { Canvas } from "../features/canvas/Canvas";
import { Inspector } from "../features/canvas/Inspector";
import type { loadNodePlugins } from "../features/canvas/node-plugins";
import { Settings } from "../features/settings/Settings";
import { Assets } from "../features/assets/Assets";
import { Library } from "../features/library/Library";
import { Agent } from "../features/agent/Agent";

type Panel="node"|"assets"|"library"|"agent";
const icons={text:Layers3,image:Image,video:Film,audio:AudioLines};
const labels:Record<Kind,string>={text:"文本",image:"图片",video:"视频",audio:"音频"};
const pending=(j:Job)=>j.status==="queued"||j.status==="running";
export function App({api,controller,nodePlugins}:{api:APIClient;controller:CanvasController;nodePlugins:ReturnType<typeof loadNodePlugins>}){
  const state=useSyncExternalStore(controller.subscribe,controller.getSnapshot),projectID=state.project?.id;
  const [projects,setProjects]=useState<ProjectSummary[]>([]),[channels,setChannels]=useState<Channel[]>([]);
  const [assets,setAssets]=useState<Asset[]>([]),[library,setLibrary]=useState<LibraryEntry[]>([]),[plugins,setPlugins]=useState<Plugin[]>([]);
  const [jobs,setJobs]=useState<Job[]>([]),[panel,setPanel]=useState<Panel>("node"),[settings,setSettings]=useState(false);
  const [notice,setNotice]=useState(""),[saving,setSaving]=useState(false),[working,setWorking]=useState(false),[submitting,setSubmitting]=useState(0);
  const [creating,setCreating]=useState(false),[newName,setNewName]=useState("未命名画布"),[renaming,setRenaming]=useState(false),[name,setName]=useState("");
  const [deleting,setDeleting]=useState<ProjectSummary|null>(null),[deleteTransition,setDeleteTransition]=useState<ProjectSummary|null>(null);
  const [transition,setTransition]=useState<null|(()=>Promise<void>)>(null);
  const saveLock=useRef<Promise<Project>|null>(null),actionLock=useRef(false);
  const submittingNodes=useRef(new Set<string>());
  const receipts=useRef(new Map<string,{signature:string;id:string}>());
  const onError=useCallback((error:unknown)=>setNotice(error instanceof APIError&&error.code==="CONFLICT"?"保存或引用发生冲突，本地编辑已保留。可新建副本保存，或重新打开项目放弃本地修改。":errorMessage(error)),[]);
  const refreshProjects=useCallback(async()=>setProjects(await api.call("listProjects",{})),[api]);
  const refreshAssets=useCallback(async()=>setAssets(await api.call("listAssets",{})),[api]);
  const refreshLibrary=useCallback(async()=>setLibrary(await api.call("listLibrary",{})),[api]);
  const refreshChannels=useCallback(async()=>setChannels(await api.call("listChannels",{})),[api]);
  const refresh=useCallback(async()=>{
    const results=await Promise.allSettled([refreshProjects(),refreshAssets(),refreshLibrary(),refreshChannels(),api.call("listPlugins",{}).then(setPlugins)]);
    for(const result of results)if(result.status==="rejected")onError(result.reason);
  },[api,refreshProjects,refreshAssets,refreshLibrary,refreshChannels,onError]);
  useEffect(()=>{void refresh()},[refresh]);
  useEffect(()=>{
    const handler=(event:BeforeUnloadEvent)=>{if(controller.getSnapshot().dirty||saveLock.current||submittingNodes.current.size){event.preventDefault();event.returnValue=""}};
    window.addEventListener("beforeunload",handler);return()=>window.removeEventListener("beforeunload",handler);
  },[controller]);
  useEffect(()=>{
    receipts.current.clear();setJobs([]);
    if(!projectID)return;
    let disposed=false,timer:ReturnType<typeof setTimeout>;const loadedAssets=new Set<string>();
    const poll=async()=>{
      try{
        const next=await api.call("listJobs",{projectId:projectID});
        if(disposed)return;
        setJobs(next);
        for(const job of next)if(job.status==="succeeded")controller.applyJob({nodeId:job.nodeId,job});
        const newAssets=next.filter(j=>j.status==="succeeded"&&j.assetId&&!loadedAssets.has(j.assetId));
        if(newAssets.length){await refreshAssets();newAssets.forEach(j=>loadedAssets.add(j.assetId))}
      }catch(e){if(!disposed)onError(e)}
      if(!disposed)timer=setTimeout(()=>void poll(),2000);
    };
    void poll();return()=>{disposed=true;clearTimeout(timer)};
  },[api,controller,projectID,refreshAssets,onError]);
  const save=useCallback(async():Promise<Project>=>{
    if(saveLock.current)return saveLock.current;
    const snapshot=controller.getSnapshot();
    if(!snapshot.project)throw new Error("请先创建项目");
    if(!snapshot.dirty)return snapshot.project;
    setSaving(true);
    const request=api.call("saveProject",{id:snapshot.project.id,name:snapshot.name,expectedRevision:snapshot.project.revision,document:snapshot.document}).then(project=>{
      controller.acknowledge(project,snapshot.sequence);setProjects(ps=>ps.map(p=>p.id===project.id?project:p));return project;
    });
    saveLock.current=request;
    try{return await request}finally{saveLock.current=null;setSaving(false)}
  },[api,controller]);
  const perform=useCallback(async(action:()=>Promise<unknown>)=>{
    if(actionLock.current)return;
    actionLock.current=true;setWorking(true);
    try{await action()}catch(e){onError(e)}finally{actionLock.current=false;setWorking(false)}
  },[onError]);
  const requestTransition=(action:()=>Promise<void>)=>{
    if(submittingNodes.current.size){onError(new Error("正在提交生成任务，请稍后切换项目"));return}
    if(controller.getSnapshot().dirty||saveLock.current){setTransition(()=>action);return}
    void perform(action);
  };
  const open=async(id:string)=>{const project=await api.call("openProject",{id});controller.load(project);setPanel("node")};
  const create=async()=>{
    const project=await api.call("createProject",{name:newName});controller.load(project);
    setCreating(false);setPanel("node");await refreshProjects();
  };
  const copy=async()=>{
    const snapshot=controller.getSnapshot();if(!snapshot.project)return;
    const project=await api.call("createProject",{name:(snapshot.name.slice(0,28)||"画布")+" · 副本"});
    const document={...snapshot.document,nodes:snapshot.document.nodes.map(n=>({...n,jobId:""}))};
    const stored=await api.call("saveProject",{id:project.id,name:project.name,expectedRevision:project.revision,document});
    controller.load(stored);await refreshProjects();
  };
  const submit=async(nodeId:string,override?:string):Promise<Job>=>{
    if(actionLock.current||submittingNodes.current.has(nodeId))throw new Error("请等待当前操作完成");
    submittingNodes.current.add(nodeId);setSubmitting(submittingNodes.current.size);
    try{
    const before=controller.getSnapshot(),id=before.project?.id;
    const node=before.document.nodes.find(n=>n.id===nodeId);
    if(!node||!id)throw new Error("节点不存在");
    const prompt=override??[controller.upstreamText(nodeId),node.prompt||node.content].filter(Boolean).join("\n\n");
    if(!node.channelId||!node.model||!prompt.trim())throw new Error("请选择渠道、模型并填写提示词");
    await save();
    const current=controller.getSnapshot();
    if(current.project?.id!==id||current.dirty)throw new Error("画布在保存期间发生了修改，请再次生成");
    const signature=JSON.stringify({nodeId,channelId:node.channelId,model:node.model,kind:node.kind,prompt});
    let receipt=receipts.current.get(nodeId);
    if(!receipt||receipt.signature!==signature){receipt={signature,id:crypto.randomUUID()};receipts.current.set(nodeId,receipt)}
    const job=await api.call("submitGeneration",{requestId:receipt.id,projectId:id,nodeId,channelId:node.channelId,model:node.model,kind:node.kind,prompt});
    if(controller.getSnapshot().project?.id===id){controller.trackJob(nodeId,job);setJobs(js=>[job,...js.filter(j=>j.id!==job.id)]);receipts.current.delete(nodeId)}
    return job;
    }finally{submittingNodes.current.delete(nodeId);setSubmitting(submittingNodes.current.size)}
  };
  const cancel=async(id:string)=>{const job=await api.call("cancelJob",{id});setJobs(js=>js.map(j=>j.id===id?job:j))};
  const add=(kind:Kind)=>{
    try{const v=state.document.viewport;controller.createNode({kind,title:labels[kind]+"节点",x:(300-v.x)/v.zoom,y:(150-v.y)/v.zoom});setPanel("node")}catch(e){onError(e)}
  };
  useEffect(()=>{
    const handler=(event:KeyboardEvent)=>{
      if(actionLock.current||transition||settings||creating||deleting)return;
      const editing=(event.target as HTMLElement)?.closest("input,textarea,select,[contenteditable]");
      if((event.ctrlKey||event.metaKey)&&event.key.toLowerCase()==="s"){event.preventDefault();void save().catch(onError);return}
      if(editing)return;
      try{
        if((event.ctrlKey||event.metaKey)&&event.key.toLowerCase()==="z"){event.preventDefault();event.shiftKey?controller.redo():controller.undo()}
        if(event.key==="Delete"&&controller.getSnapshot().selected.length)controller.removeNodes({ids:controller.getSnapshot().selected});
      }catch(e){onError(e)}
    };
    window.addEventListener("keydown",handler);return()=>window.removeEventListener("keydown",handler);
  },[controller,save,onError,transition,settings,creating,deleting]);
  const selected=state.document.nodes.find(n=>n.id===state.selected[0]);
  const tabs:[Panel,string,typeof Layers3][]=[["node","节点",Layers3],["assets","素材",Image],["library","创作库",BookOpen],["agent","助手",Sparkles]];
  return <main className="workspace">
    <aside className="project-rail"><button className="brand" onClick={()=>void api.openRepository().catch(onError)} title="打开 GitHub 仓库"><span className="brand-symbol">墨</span><div><strong>墨华画布</strong><small>MOHUA CANVAS</small></div></button>
      <div className="rail-heading"><span className="eyebrow">YOUR WORKSPACE</span><button className="icon-button" aria-label="新建项目" onClick={()=>setCreating(true)}><Plus size={17}/></button></div>
      <nav className="project-list" aria-label="项目">{projects.map(p=><div key={p.id} className={"project-item "+(p.id===projectID?"active":"")}><button disabled={working} onClick={()=>requestTransition(()=>open(p.id))}><FolderOpen size={15}/><span>{p.name}</span></button><button className="project-delete icon-button" aria-label={"删除项目 "+p.name} onClick={()=>{if(p.id===projectID&&controller.getSnapshot().dirty){setDeleteTransition(p);return}setDeleting(p)}}><Trash2 size={13}/></button></div>)}
      {projects.length===0?<p className="muted small">新建画布，开始你的第一份创作。</p>:null}</nav>
      <div className="rail-foot"><div className="local-status"><span className="status-dot"/>本地工作空间</div><button className="button secondary full" onClick={()=>setSettings(true)}><Settings2 size={15}/>渠道与插件</button></div>
    </aside>
    <section className="studio">
      <header className="studio-header"><div className="breadcrumb"><span>工作空间</span><ChevronRight size={13}/>
        {renaming?<form onSubmit={e=>{e.preventDefault();try{controller.rename(name);setRenaming(false)}catch(e){onError(e)}}}><input autoFocus value={name} onChange={e=>setName(e.target.value)} aria-label="项目名称"/><button className="icon-button" aria-label="确认名称"><Check size={15}/></button><button type="button" className="icon-button" aria-label="取消重命名" onClick={()=>setRenaming(false)}><X size={15}/></button></form>:
          <button className="project-title" disabled={!projectID} title="重命名项目" onClick={()=>{setName(state.name);setRenaming(true)}}>{state.name||"选择你的画布"}</button>}
        {projectID?<span className={"save-status "+(state.dirty?"unsaved":"")}>{saving?"保存中":state.dirty?"未保存":"已保存"}</span>:null}</div>
        <div className="header-actions"><button className="text-button" disabled={!projectID||working||saving||submitting>0} onClick={()=>void perform(copy)}>另存副本</button>
          <button className="button primary" disabled={!projectID||saving||working} onClick={()=>void save().catch(onError)}>{saving?<LoaderCircle className="spin" size={15}/>:<Save size={15}/>}保存<span className="shortcut">Ctrl S</span></button></div>
      </header>
      {notice?<div className="notice" role="alert"><span>{notice}</span><button className="text-button" onClick={()=>void refresh()}>刷新</button><button className="icon-button" aria-label="关闭提示" onClick={()=>setNotice("")}><X size={15}/></button></div>:null}
      <div className="studio-body" inert={working}><section className="canvas-workspace">
        {projectID?<><div className="canvas-toolbar"><span className="toolbar-label">添加</span>{nodePlugins.plugins.map(p=>{const Icon=icons[p.kind];return <button key={p.id} onClick={()=>add(p.kind)}><Icon size={16}/>{p.label}</button>})}<span className="toolbar-separator"/><button aria-label="撤销" disabled={!state.undoCount} onClick={()=>controller.undo()}><Undo2 size={16}/></button><button aria-label="重做" disabled={!state.redoCount} onClick={()=>controller.redo()}><Redo2 size={16}/></button></div>
        <Canvas controller={controller} jobs={jobs} nodeTypes={nodePlugins.nodeTypes} plugins={nodePlugins.plugins} onError={onError}/></>:
        <div className="workspace-welcome"><span className="welcome-stamp">墨</span><span className="eyebrow">THINK FREELY. CREATE LOCALLY.</span><h1>让想法，有处落笔。</h1><p>把文字、图像、声音与故事，<br/>连接在一张自由生长的画布上。</p><button className="button primary" onClick={()=>setCreating(true)}><Plus size={16}/>创建第一张画布</button></div>}
        <footer className="canvas-footer"><span>{state.document.nodes.length} 个节点 · {state.document.edges.length} 条连接</span><span>{jobs.filter(pending).length?jobs.filter(pending).length+" 个生成任务进行中":"灵感，从这里延展"}</span></footer>
      </section><aside className="details-rail"><nav className="panel-tabs" aria-label="工具面板">{tabs.map(([id,label,Icon])=><button className={panel===id?"active":""} key={id} onClick={()=>setPanel(id)}><Icon size={16}/><span>{label}</span></button>)}</nav>
        <div className="panel-content">
          {panel==="node"?(selected?<Inspector key={selected.id} node={selected} controller={controller} channels={channels} assets={assets} library={library} job={jobs.find(j=>j.id===selected.jobId)} onGenerate={submit} onCancel={cancel} onError={onError}/>:<div className="panel-empty"><Layers3 size={30} strokeWidth={1.1}/><h3>每个节点，一片灵感。</h3><p>选择画布节点，编辑内容或生成作品。</p></div>):null}
          {panel==="assets"?<Assets api={api} assets={assets} selected={selected} controller={controller} onRefresh={refreshAssets} onError={onError}/>:null}
          {panel==="library"?<Library api={api} entries={library} onRefresh={refreshLibrary} onError={onError}/>:null}
          {projectID?<div hidden={panel!=="agent"}><Agent key={projectID} controller={controller} channels={channels} library={library} jobs={jobs} onSubmit={submit} onError={onError}/></div>:panel==="agent"?<div className="panel-empty"><Sparkles size={28}/><p>打开画布后，与助手一起创作。</p></div>:null}
          {jobs.some(pending)?<div className="task-list"><span className="eyebrow">GENERATING</span>{jobs.filter(pending).map(j=><div key={j.id}><LoaderCircle size={13} className="spin"/><span>{state.document.nodes.find(n=>n.id===j.nodeId)?.title??"生成任务"}</span><button className="text-button" onClick={()=>void cancel(j.id).catch(onError)}>取消</button></div>)}</div>:null}
        </div>
      </aside></div>
    </section>
    {settings?<Settings api={api} channels={channels} plugins={plugins} onRefresh={refreshChannels} onClose={()=>setSettings(false)} onError={onError}/>:null}
    {creating?<div className="modal-backdrop"><form className="small-modal" role="dialog" aria-modal="true" aria-labelledby="create-title" onSubmit={e=>{e.preventDefault();requestTransition(create)}}><span className="eyebrow">A NEW BEGINNING</span><h2 id="create-title">新建画布</h2><label>项目名称<input autoFocus required value={newName} onChange={e=>setNewName(e.target.value)}/></label><div className="form-actions"><button className="button primary" disabled={working}>创建画布</button><button className="text-button" type="button" onClick={()=>setCreating(false)}>取消</button></div></form></div>:null}
    {transition?<div className="modal-backdrop"><section className="small-modal" role="dialog" aria-modal="true" aria-labelledby="save-title"><h2 id="save-title">保留这次创作？</h2><p>当前画布有未保存的修改。</p><div className="form-actions">
      <button className="button primary" disabled={working||saving} onClick={()=>void perform(async()=>{await save();if(controller.getSnapshot().dirty)throw new Error("保存期间产生了新编辑，请再次保存");await transition();setTransition(null)})}>保存并继续</button>
      <button className="text-button danger" disabled={working||saving} onClick={()=>void perform(async()=>{await transition();setTransition(null)})}>放弃修改</button>
      <button className="text-button" disabled={working} onClick={()=>setTransition(null)}>取消</button></div></section></div>:null}
    {deleteTransition?<div className="modal-backdrop"><section className="small-modal" role="dialog" aria-modal="true" aria-labelledby="delete-save-title"><h2 id="delete-save-title">删除前保留修改？</h2><p>当前画布有未保存的修改。删除会同时放弃这些修改。</p><div className="form-actions"><button className="button danger" disabled={working||saving||submitting>0} onClick={()=>{setDeleting(deleteTransition);setDeleteTransition(null)}}>继续删除</button><button className="text-button" onClick={()=>{setDeleteTransition(null);void save().catch(onError)}}>保存并取消删除</button><button className="text-button" onClick={()=>setDeleteTransition(null)}>取消</button></div></section></div>:null}
    {deleting?<div className="modal-backdrop"><section className="small-modal" role="dialog" aria-modal="true" aria-labelledby="delete-title"><h2 id="delete-title">删除这张画布？</h2><p>「{deleting.name}」和它的任务历史将被删除。素材库会保留。</p><div className="form-actions"><button className="button danger" disabled={working||saving||submitting>0} onClick={()=>void perform(async()=>{await api.call("deleteProject",{id:deleting.id});if(projectID===deleting.id)controller.clear();setDeleting(null);await refreshProjects()})}>确认删除</button><button className="text-button" disabled={working} onClick={()=>setDeleting(null)}>取消</button></div></section></div>:null}
  </main>;
}
