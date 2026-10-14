import { useEffect, useState } from "react";
import { Play, Square, Trash2 } from "lucide-react";
import type { Asset, CanvasNode, Channel, Job, LibraryEntry } from "../../../contracts/desktop";
import type { CanvasController } from "./controller";

export function Inspector({node,controller,channels,assets,library,job,onGenerate,onCancel,onError}:{
  node:CanvasNode;controller:CanvasController;channels:Channel[];assets:Asset[];library:LibraryEntry[];job?:Job;
  onGenerate:(id:string)=>Promise<unknown>;onCancel:(id:string)=>Promise<unknown>;onError:(error:unknown)=>void;
}){
  const [draft,setDraft]=useState(node),[busy,setBusy]=useState(false);
  useEffect(()=>setDraft(node),[node]);
  const patch=(value:Partial<CanvasNode>)=>setDraft(d=>({...d,...value}));
  const channel=channels.find(c=>c.id===draft.channelId);
  const models=channel?.models.filter(m=>m.kind===node.kind)??[];
  const commit=()=>controller.editNode({id:node.id,title:draft.title,content:draft.content,prompt:draft.prompt,channelId:draft.channelId,model:draft.model});
  const act=async(fn:()=>unknown)=>{setBusy(true);try{await fn()}catch(e){onError(e)}finally{setBusy(false)}};
  const pending=job?.status==="queued"||job?.status==="running";
  return <div className="inspector">
    <div className="panel-heading"><span className="eyebrow">NODE DETAILS</span><span className="tag">{node.kind}</span></div>
    <label>节点名称<input value={draft.title} onChange={e=>patch({title:e.target.value})}/></label>
    {node.kind==="text"?<label>文本内容<textarea rows={8} value={draft.content} onChange={e=>patch({content:e.target.value})} placeholder="在这里整理想法…"/></label>:null}
    <label>生成提示词<textarea rows={5} value={draft.prompt} onChange={e=>patch({prompt:e.target.value})} placeholder="描述你想要生成的内容…"/></label>
    <label>从创作库填入<select defaultValue="" onChange={e=>{const entry=library.find(p=>p.id===e.target.value);if(entry)patch({prompt:entry.content});e.target.value=""}}>
      <option value="">选择提示词</option>{library.filter(e=>e.kind==="prompt").map(e=><option key={e.id} value={e.id}>{e.name}</option>)}
    </select></label>
    <label>模型渠道<select value={draft.channelId} onChange={e=>patch({channelId:e.target.value,model:""})}>
      <option value="">选择渠道</option>{channels.filter(c=>c.models.some(m=>m.kind===node.kind)).map(c=><option key={c.id} value={c.id}>{c.name}</option>)}
    </select></label>
    <label>模型<select value={draft.model} onChange={e=>patch({model:e.target.value})}>
      <option value="">选择模型</option>{models.map(m=><option key={m.name} value={m.name}>{m.name}</option>)}
    </select></label>
    <button className="button secondary full" disabled={busy} onClick={()=>void act(commit)}>应用节点设置</button>
    <button className="button primary full" disabled={busy||pending||!draft.channelId||!draft.model} onClick={()=>void act(async()=>{commit();await onGenerate(node.id)})}><Play size={14}/>生成{node.kind==="text"?"文本":node.kind==="image"?"图片":node.kind==="video"?"视频":"音频"}</button>
    <p className="muted small">生成会先保存画布。连入此节点的上游文本会加入提示词。</p>
    {pending?<button className="button secondary full" onClick={()=>void act(()=>onCancel(job!.id))}><Square size={13}/>取消生成</button>:null}
    {job?.errorCode?<p className="inline-error">{job.errorCode} · 请检查渠道配置后重新生成</p>:null}
    {node.kind!=="text"?<label>附加已有素材<select value={node.assetId} onChange={e=>void act(()=>{const asset=assets.find(a=>a.id===e.target.value);if(asset)controller.attachAsset({nodeId:node.id,asset})})}>
      <option value="">选择素材</option>{assets.filter(a=>a.mimeType.startsWith(node.kind+"/")).map(a=><option key={a.id} value={a.id}>{a.name}</option>)}
    </select></label>:null}
    <div className="panel-divider"/>
    <button className="text-button danger" onClick={()=>void act(()=>controller.removeNodes({ids:[node.id]}))}><Trash2 size={14}/>删除节点</button>
  </div>;
}
