import { useRef, useState } from "react";
import { Upload, FileImage, Trash2, Link2 } from "lucide-react";
import type { Asset, CanvasNode } from "../../../contracts/desktop";
import { assetURL, type APIClient } from "../../services/api";
import type { CanvasController } from "../canvas/controller";

export function Assets({api,assets,selected,controller,onRefresh,onError}:{
  api:APIClient;assets:Asset[];selected?:CanvasNode;controller:CanvasController;onRefresh:()=>Promise<void>;onError:(e:unknown)=>void;
}){
  const input=useRef<HTMLInputElement>(null),[busy,setBusy]=useState(false),[deleting,setDeleting]=useState<Asset|null>(null);
  const importFile=async(file:File)=>{setBusy(true);try{await api.importAsset(file);await onRefresh()}catch(e){onError(e)}finally{setBusy(false)}};
  const remove=async()=>{if(!deleting)return;setBusy(true);try{await api.call("deleteAsset",{id:deleting.id});setDeleting(null);await onRefresh()}catch(e){onError(e)}finally{setBusy(false)}};
  return <div className="assets-panel"><div className="panel-heading"><h3>素材库</h3><span className="tag">{assets.length}</span></div>
    <input ref={input} hidden type="file" accept=".png,.jpg,.jpeg,.webp,.gif,.mp4,.webm,.mp3,.wav,.ogg,.flac" onChange={e=>{const file=e.target.files?.[0];if(file)void importFile(file);e.target.value=""}}/>
    <button className="button secondary full" disabled={busy} onClick={()=>input.current?.click()}><Upload size={15}/>{busy?"处理中…":"导入本地素材"}</button>
    <p className="muted small">图片、视频或音频，单个文件最大 64 MB。</p>
    {assets.length===0?<div className="panel-empty"><FileImage size={28} strokeWidth={1.1}/><p>你的素材，留在本机。</p></div>:null}
    <div className="asset-grid">{assets.map(a=><article className="asset-card" key={a.id}>
      <div className="asset-preview">{a.mimeType.startsWith("image/")?<img loading="lazy" src={assetURL(a.id)} alt={a.name}/>:<FileImage size={25} strokeWidth={1.2}/>}</div>
      <strong title={a.name}>{a.name}</strong><small>{a.mimeType.split("/")[1]} · {(a.bytes/1024/1024).toFixed(1)} MB</small>
      <div className="asset-actions"><button className="icon-button" aria-label={"附加 "+a.name} title="附加到所选节点" disabled={!selected||!a.mimeType.startsWith(selected.kind+"/")} onClick={()=>{try{if(selected)controller.attachAsset({nodeId:selected.id,asset:a})}catch(e){onError(e)}}><Link2 size={14}/></button>
        <button className="icon-button danger" aria-label={"删除 "+a.name} disabled={busy} onClick={()=>setDeleting(a)}><Trash2 size={14}/></button></div>
    </article>)}</div>
    {deleting?<div className="confirm-inline"><p>确认删除「{deleting.name}」？被项目或任务引用的素材无法删除。</p><button className="button danger" disabled={busy} onClick={()=>void remove()}>确认删除</button><button className="text-button" onClick={()=>setDeleting(null)}>取消</button></div>:null}
  </div>;
}
