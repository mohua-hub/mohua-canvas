import { useState } from "react";
import { Plus, Trash2, X } from "lucide-react";
import type { Channel, Kind, Model, Plugin } from "../../../contracts/desktop";
import type { APIClient } from "../../services/api";
const labels:Record<Kind,string>={text:"文本",image:"图片",video:"视频",audio:"音频"};

export function Settings({api,channels,plugins,onRefresh,onClose,onError}:{
  api:APIClient;channels:Channel[];plugins:Plugin[];onRefresh:()=>Promise<void>;onClose:()=>void;onError:(e:unknown)=>void;
}){
  const [selected,setSelected]=useState<Channel|null>(null);
  const [name,setName]=useState(""),[pluginId,setPlugin]=useState(plugins.find(p=>p.type==="provider")?.id??"");
  const [baseUrl,setURL]=useState(""),[apiKey,setKey]=useState(""),[models,setModels]=useState<Model[]>([{name:"",kind:"text"}]);
  const [busy,setBusy]=useState(false),[removing,setRemoving]=useState(false);
  const provider=plugins.find(p=>p.id===pluginId);
  const select=(c:Channel|null)=>{setSelected(c);setName(c?.name??"");setPlugin(c?.pluginId??plugins.find(p=>p.type==="provider")?.id??"");setURL(c?.baseUrl??"");setKey("");setModels(c?.models.map(m=>({...m}))??[{name:"",kind:"text"}]);setRemoving(false)};
  const save=async()=>{setBusy(true);try{
    const c=await api.call("configureChannel",{...(selected?{id:selected.id}:{}),name,pluginId,baseUrl,models,...(apiKey?{apiKey}:{})});
    setKey("");await onRefresh();select(c);
  }catch(e){onError(e)}finally{setBusy(false)}};
  const remove=async()=>{if(!selected)return;setBusy(true);try{await api.call("deleteChannel",{id:selected.id});await onRefresh();select(null)}catch(e){onError(e)}finally{setBusy(false)}};
  return <div className="modal-backdrop"><section className="settings-modal" role="dialog" aria-modal="true" aria-labelledby="settings-title">
    <header className="modal-heading"><div><span className="eyebrow">WORKSPACE SETTINGS</span><h2 id="settings-title">渠道与插件</h2></div><button className="icon-button" aria-label="关闭设置" disabled={busy} onClick={onClose}><X size={20}/></button></header>
    <div className="settings-body"><aside className="settings-list">
      <button className="button secondary full" disabled={busy} onClick={()=>select(null)}><Plus size={14}/>新建渠道</button>
      {channels.map(c=><button key={c.id} className={"channel-item "+(selected?.id===c.id?"active":"")} disabled={busy} onClick={()=>select(c)}><strong>{c.name}</strong><span>{c.pluginId}</span></button>)}
      <div className="panel-divider"/><span className="eyebrow">ACTIVE PLUGINS</span>
      {plugins.map(p=><div className="plugin-item" key={p.id}><span className="status-dot"/><div><strong>{p.id}</strong><small>v{p.version} · {p.status}</small></div></div>)}
    </aside><form className="channel-form" onSubmit={e=>{e.preventDefault();void save()}}>
      <h3>{selected?"编辑渠道":"添加一个模型渠道"}</h3>
      <label>渠道名称<input value={name} onChange={e=>setName(e.target.value)} required placeholder="例如：我的创作模型"/></label>
      <label>供应商插件<select value={pluginId} onChange={e=>{setPlugin(e.target.value);setModels([{name:"",kind:"text"}])}}>{plugins.filter(p=>p.type==="provider").map(p=><option value={p.id} key={p.id}>{p.id}</option>)}</select></label>
      <label>API 基础地址<input type="url" value={baseUrl} onChange={e=>setURL(e.target.value)} required placeholder="https://api.example.com/v1"/></label>
      <label>API Key<input type="password" autoComplete="off" value={apiKey} onChange={e=>setKey(e.target.value)} required={!selected} placeholder={selected?"留空保留已保存密钥":"输入供应商密钥"}/></label>
      <p className="muted small">密钥通过系统安全存储保护，保存后不回显。</p>
      <span className="field-caption">模型清单</span>{models.map((m,i)=><div className="model-row" key={i}>
        <select aria-label="模型类型" value={m.kind} onChange={e=>setModels(ms=>ms.map((v,n)=>n===i?{...v,kind:e.target.value as Kind}:v))}>{provider?.capabilities.map(k=><option value={k} key={k}>{labels[k]}</option>)}</select>
        <input aria-label="模型标识" value={m.name} placeholder="供应商模型标识" required onChange={e=>setModels(ms=>ms.map((v,n)=>n===i?{...v,name:e.target.value}:v))}/>
        <button className="icon-button" type="button" aria-label="移除模型" disabled={models.length===1} onClick={()=>setModels(ms=>ms.filter((_,n)=>n!==i))}><X size={14}/></button>
      </div>)}
      <button className="text-button" type="button" disabled={models.length>=50} onClick={()=>setModels(ms=>[...ms,{name:"",kind:provider?.capabilities[0]??"text"}])}><Plus size={14}/>添加模型</button>
      <div className="form-actions"><button className="button primary" disabled={busy} type="submit">{busy?"保存中…":"保存渠道"}</button>
        {selected?<button type="button" className="text-button danger" disabled={busy} onClick={()=>setRemoving(true)}><Trash2 size={14}/>删除渠道</button>:null}</div>
      {removing?<div className="confirm-inline"><p>删除「{selected?.name}」及其密钥？被画布或任务引用的渠道无法删除。</p><button type="button" className="button danger" disabled={busy} onClick={()=>void remove()}>确认删除</button><button type="button" className="text-button" onClick={()=>setRemoving(false)}>取消</button></div>:null}
    </form></div>
  </section></div>;
}
