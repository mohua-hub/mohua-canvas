import { useState } from "react";
import { Plus, Trash2 } from "lucide-react";
import type { LibraryEntry } from "../../../contracts/desktop";
import type { APIClient } from "../../services/api";

export function Library({api,entries,onRefresh,onError}:{
  api:APIClient;entries:LibraryEntry[];onRefresh:()=>Promise<void>;onError:(e:unknown)=>void;
}){
  const [selected,setSelected]=useState<LibraryEntry|null>(null),[name,setName]=useState(""),[content,setContent]=useState("");
  const [kind,setKind]=useState<"prompt"|"skill">("prompt"),[busy,setBusy]=useState(false),[deleting,setDeleting]=useState(false);
  const select=(e:LibraryEntry|null)=>{setSelected(e);setName(e?.name??"");setContent(e?.content??"");setKind(e?.kind??"prompt");setDeleting(false)};
  const save=async()=>{setBusy(true);try{const e=await api.call("saveLibraryEntry",{...(selected?{id:selected.id}:{}),kind,name,content});await onRefresh();select(e)}catch(e){onError(e)}finally{setBusy(false)}};
  const remove=async()=>{if(!selected)return;setBusy(true);try{await api.call("deleteLibraryEntry",{id:selected.id});await onRefresh();select(null)}catch(e){onError(e)}finally{setBusy(false)}};
  return <div className="library-panel">
    <div className="panel-heading"><h3>创作库</h3><button className="icon-button" aria-label="新建创作条目" disabled={busy} onClick={()=>select(null)}><Plus size={17}/></button></div>
    <div className="library-list">{entries.map(e=><button className={"library-item "+(selected?.id===e.id?"active":"")} key={e.id} disabled={busy} onClick={()=>select(e)}><span className="tag">{e.kind==="prompt"?"提示词":"Skill"}</span><strong>{e.name}</strong></button>)}
    {entries.length===0?<p className="muted small">收藏常用提示词和创作方法。</p>:null}</div>
    <form onSubmit={e=>{e.preventDefault();void save()}}>
      <label>类型<select value={kind} onChange={e=>setKind(e.target.value as "prompt"|"skill")}><option value="prompt">提示词</option><option value="skill">Agent Skill</option></select></label>
      <label>名称<input required value={name} onChange={e=>setName(e.target.value)} placeholder="给创作方法起个名字"/></label>
      <label>内容<textarea required rows={10} value={content} onChange={e=>setContent(e.target.value)} placeholder="写下可以复用的指令…"/></label>
      <p className="muted small">提示词用于节点生成；Skill 作为 Agent 的补充指令。</p>
      <button className="button primary full" disabled={busy}>{busy?"保存中…":"保存到创作库"}</button>
    </form>
    {selected?<button className="text-button danger" disabled={busy} onClick={()=>setDeleting(true)}><Trash2 size={14}/>删除条目</button>:null}
    {deleting?<div className="confirm-inline"><p>确认删除「{selected?.name}」？</p><button className="button danger" disabled={busy} onClick={()=>void remove()}>确认删除</button><button className="text-button" onClick={()=>setDeleting(false)}>取消</button></div>:null}
  </div>;
}
