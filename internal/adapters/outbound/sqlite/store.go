package sqlite

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"time"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
	_ "modernc.org/sqlite"
)

type Store struct { db *sql.DB; aead cipher.AEAD }
func Open(dataDir string,key []byte) (*Store,error) {
	if len(key)!=32 { return nil,domain.Fail("INVALID_INPUT") }
	if err:=os.MkdirAll(dataDir,0700); err!=nil { return nil,err }
	block,err:=aes.NewCipher(key); if err!=nil { return nil,err }
	aead,err:=cipher.NewGCM(block); if err!=nil { return nil,err }
	db,err:=sql.Open("sqlite",filepath.Join(dataDir,"workspace.db")); if err!=nil { return nil,err }
	db.SetMaxOpenConns(1)
	s:=&Store{db:db,aead:aead}
	if err=s.initialize(); err!=nil { _=db.Close(); return nil,err }; return s,nil
}
func (s *Store) initialize() error {
	if _,err:=s.db.Exec("PRAGMA foreign_keys=ON; PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;"); err!=nil { return err }
	var version,tables int
	if err:=s.db.QueryRow("PRAGMA user_version").Scan(&version); err!=nil { return err }
	if version==1 { return nil }
	if err:=s.db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tables); err!=nil { return err }
	if version!=0 || tables!=0 { return domain.Fail("UNSUPPORTED") }
	_,err:=s.db.Exec(schema); return err
}
const schema = `
BEGIN;
CREATE TABLE projects(id TEXT PRIMARY KEY,name TEXT NOT NULL,revision INTEGER NOT NULL,document_json TEXT NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL);
CREATE TABLE credentials(ref TEXT PRIMARY KEY,nonce BLOB NOT NULL,ciphertext BLOB NOT NULL);
CREATE TABLE channels(id TEXT PRIMARY KEY,name TEXT NOT NULL,plugin_id TEXT NOT NULL,base_url TEXT NOT NULL,models_json TEXT NOT NULL,credential_ref TEXT NOT NULL REFERENCES credentials(ref));
CREATE TABLE assets(id TEXT PRIMARY KEY,name TEXT NOT NULL,mime_type TEXT NOT NULL,bytes INTEGER NOT NULL,created_at TEXT NOT NULL);
CREATE TABLE jobs(id TEXT PRIMARY KEY,request_id TEXT UNIQUE NOT NULL,input_hash TEXT NOT NULL,project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,node_id TEXT NOT NULL,channel_id TEXT NOT NULL,model TEXT NOT NULL,kind TEXT NOT NULL,prompt TEXT NOT NULL,status TEXT NOT NULL,text TEXT NOT NULL,asset_id TEXT NOT NULL,error_code TEXT NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL);
CREATE INDEX jobs_project ON jobs(project_id,created_at);
CREATE TABLE library_entries(id TEXT PRIMARY KEY,kind TEXT NOT NULL,name TEXT NOT NULL,content TEXT NOT NULL,created_at TEXT NOT NULL,updated_at TEXT NOT NULL);
PRAGMA user_version=1;
COMMIT;
`
func (s *Store) Close() error { return s.db.Close() }
func mapped(err error) error {
	if errors.Is(err,sql.ErrNoRows) { return domain.Fail("NOT_FOUND") }
	if err!=nil { return domain.Fail("INTERNAL") }; return nil
}
func changed(result sql.Result,err error) error {
	if err!=nil { return mapped(err) }; n,err:=result.RowsAffected(); if err!=nil { return mapped(err) }
	if n==0 { return domain.Fail("NOT_FOUND") }; return nil
}
type scanner interface { Scan(...any) error }
type nodeRecord struct {
	ID string `json:"id"`; Kind domain.Kind `json:"kind"`
	X float64 `json:"x"`; Y float64 `json:"y"`
	Title string `json:"title"`; Content string `json:"content"`; Prompt string `json:"prompt"`
	ChannelID string `json:"channelId"`; Model string `json:"model"`
	AssetID string `json:"assetId"`; JobID string `json:"jobId"`
}
type edgeRecord struct { ID string `json:"id"`; Source string `json:"source"`; Target string `json:"target"` }
type viewportRecord struct { X float64 `json:"x"`; Y float64 `json:"y"`; Zoom float64 `json:"zoom"` }
type documentRecord struct {
	Version int `json:"version"`; Nodes []nodeRecord `json:"nodes"`; Edges []edgeRecord `json:"edges"`; Viewport viewportRecord `json:"viewport"`
}
func encodeDocument(d domain.Document) (string,error) {
	r:=documentRecord{Version:d.Version,Nodes:[]nodeRecord{},Edges:[]edgeRecord{},Viewport:viewportRecord{d.Viewport.X,d.Viewport.Y,d.Viewport.Zoom}}
	for _,n:=range d.Nodes { r.Nodes=append(r.Nodes,nodeRecord{n.ID,n.Kind,n.X,n.Y,n.Title,n.Content,n.Prompt,n.ChannelID,n.Model,n.AssetID,n.JobID}) }
	for _,e:=range d.Edges { r.Edges=append(r.Edges,edgeRecord{e.ID,e.Source,e.Target}) }
	b,err:=json.Marshal(r); if err!=nil{return "",mapped(err)}; if len(b)>8<<20{return "",domain.Fail("INVALID_INPUT")}; return string(b),nil
}
func decodeDocument(raw string) (domain.Document,error) {
	var r documentRecord; if err:=json.Unmarshal([]byte(raw),&r); err!=nil { return domain.Document{},mapped(err) }
	d:=domain.Document{Version:r.Version,Nodes:[]domain.Node{},Edges:[]domain.Edge{},Viewport:domain.Viewport{X:r.Viewport.X,Y:r.Viewport.Y,Zoom:r.Viewport.Zoom}}
	for _,n:=range r.Nodes { d.Nodes=append(d.Nodes,domain.Node{ID:n.ID,Kind:n.Kind,X:n.X,Y:n.Y,Title:n.Title,Content:n.Content,Prompt:n.Prompt,ChannelID:n.ChannelID,Model:n.Model,AssetID:n.AssetID,JobID:n.JobID}) }
	for _,e:=range r.Edges { d.Edges=append(d.Edges,domain.Edge{ID:e.ID,Source:e.Source,Target:e.Target}) }; return d,nil
}
func scanProject(row scanner) (domain.Project,error) {
	var p domain.Project; var raw string
	if err:=row.Scan(&p.ID,&p.Name,&p.Revision,&raw,&p.CreatedAt,&p.UpdatedAt); err!=nil { return p,mapped(err) }
	d,err:=decodeDocument(raw); p.Document=d; return p,err
}
const projectColumns="id,name,revision,document_json,created_at,updated_at"
func (s *Store) ListProjects(ctx context.Context) ([]domain.Project,error) {
	rows,err:=s.db.QueryContext(ctx,"SELECT id,name,revision,created_at,updated_at FROM projects ORDER BY updated_at DESC"); if err!=nil { return nil,mapped(err) }; defer rows.Close()
	out:=[]domain.Project{}; for rows.Next() { var p domain.Project; if err:=rows.Scan(&p.ID,&p.Name,&p.Revision,&p.CreatedAt,&p.UpdatedAt);err!=nil{return nil,mapped(err)};out=append(out,p) }; return out,mapped(rows.Err())
}
func (s *Store) OpenProject(ctx context.Context,id string) (domain.Project,error) { return scanProject(s.db.QueryRowContext(ctx,"SELECT "+projectColumns+" FROM projects WHERE id=?",id)) }
func (s *Store) CreateProject(ctx context.Context,p domain.Project) error {
	raw,err:=encodeDocument(p.Document); if err!=nil { return err }
	_,err=s.db.ExecContext(ctx,"INSERT INTO projects("+projectColumns+") VALUES(?,?,?,?,?,?)",p.ID,p.Name,p.Revision,raw,p.CreatedAt,p.UpdatedAt); return mapped(err)
}
// Reference checks run inside the same transaction as the snapshot update.
func checkReferences(ctx context.Context,tx *sql.Tx,p domain.Project) error {
	for _,n:=range p.Document.Nodes {
		var count int
		if n.ChannelID!="" {
			c,err:=scanChannel(tx.QueryRowContext(ctx,"SELECT "+channelColumns+" FROM channels WHERE id=?",n.ChannelID)); if err!=nil || !c.Supports(n.Model,n.Kind) { return domain.Fail("INVALID_INPUT") }
		}
		if n.AssetID!="" {
			var mime string
			if err:=tx.QueryRowContext(ctx,"SELECT mime_type FROM assets WHERE id=?",n.AssetID).Scan(&mime); err!=nil { return domain.Fail("INVALID_INPUT") }
			if (domain.Asset{MIMEType:mime}).Kind()!=n.Kind { return domain.Fail("INVALID_INPUT") }
		}
		if n.JobID!="" {
			if err:=tx.QueryRowContext(ctx,"SELECT count(*) FROM jobs WHERE id=? AND project_id=? AND node_id=? AND kind=?",n.JobID,p.ID,n.ID,string(n.Kind)).Scan(&count); err!=nil { return mapped(err) }
			if count!=1 { return domain.Fail("INVALID_INPUT") }
		}
	}
	return nil
}
func (s *Store) SaveProject(ctx context.Context,p domain.Project,expected int64) error {
	raw,err:=encodeDocument(p.Document); if err!=nil { return err }
	tx,err:=s.db.BeginTx(ctx,nil); if err!=nil { return mapped(err) }; defer tx.Rollback()
	if err=checkReferences(ctx,tx,p); err!=nil { return err }
	result,err:=tx.ExecContext(ctx,"UPDATE projects SET name=?,revision=revision+1,document_json=?,updated_at=? WHERE id=? AND revision=?",p.Name,raw,p.UpdatedAt,p.ID,expected)
	if err!=nil { return mapped(err) }; count,_:=result.RowsAffected(); if count==0 { return domain.Fail("CONFLICT") }; return mapped(tx.Commit())
}
func (s *Store) DeleteProject(ctx context.Context,id string) error {
	tx,err:=s.db.BeginTx(ctx,nil); if err!=nil { return mapped(err) }; defer tx.Rollback()
	var count int
	if err=tx.QueryRowContext(ctx,"SELECT count(*) FROM jobs WHERE project_id=? AND status IN ('queued','running')",id).Scan(&count); err!=nil { return mapped(err) }
	if count>0 { return domain.Fail("CONFLICT") }
	if err=changed(tx.ExecContext(ctx,"DELETE FROM projects WHERE id=?",id)); err!=nil { return err }; return mapped(tx.Commit())
}
type modelRecord struct { Name string `json:"name"`; Kind domain.Kind `json:"kind"` }
func scanChannel(row scanner) (domain.Channel,error) {
	var c domain.Channel; var raw string
	if err:=row.Scan(&c.ID,&c.Name,&c.PluginID,&c.BaseURL,&raw,&c.CredentialRef); err!=nil { return c,mapped(err) }
	var records []modelRecord; if err:=json.Unmarshal([]byte(raw),&records); err!=nil { return c,mapped(err) }
	c.Models=[]domain.Model{}; for _,m:=range records { c.Models=append(c.Models,domain.Model{Name:m.Name,Kind:m.Kind}) }; return c,nil
}
const channelColumns="id,name,plugin_id,base_url,models_json,credential_ref"
func (s *Store) ListChannels(ctx context.Context) ([]domain.Channel,error) {
	rows,err:=s.db.QueryContext(ctx,"SELECT "+channelColumns+" FROM channels ORDER BY name"); if err!=nil { return nil,mapped(err) }; defer rows.Close()
	out:=[]domain.Channel{}; for rows.Next() { c,err:=scanChannel(rows); if err!=nil { return nil,err }; out=append(out,c) }; return out,mapped(rows.Err())
}
func (s *Store) GetChannel(ctx context.Context,id string) (domain.Channel,error) { return scanChannel(s.db.QueryRowContext(ctx,"SELECT "+channelColumns+" FROM channels WHERE id=?",id)) }
func (s *Store) SaveChannel(ctx context.Context,c domain.Channel) error {
	records:=[]modelRecord{}; for _,m:=range c.Models { records=append(records,modelRecord{m.Name,m.Kind}) }
	raw,err:=json.Marshal(records); if err!=nil { return mapped(err) }
	tx,err:=s.db.BeginTx(ctx,nil);if err!=nil{return mapped(err)};defer tx.Rollback()
	var count int
	if err=tx.QueryRowContext(ctx,"SELECT count(*) FROM jobs WHERE channel_id=? AND status IN ('queued','running')",c.ID).Scan(&count);err!=nil{return mapped(err)}
	if count>0{return domain.Fail("CONFLICT")}
	_,err=tx.ExecContext(ctx,"INSERT INTO channels("+channelColumns+") VALUES(?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET name=excluded.name,plugin_id=excluded.plugin_id,base_url=excluded.base_url,models_json=excluded.models_json,credential_ref=excluded.credential_ref",c.ID,c.Name,c.PluginID,c.BaseURL,string(raw),c.CredentialRef)
	if err!=nil{return mapped(err)};return mapped(tx.Commit())
}
func (s *Store) DeleteChannel(ctx context.Context,id string) error {
	tx,err:=s.db.BeginTx(ctx,nil); if err!=nil { return mapped(err) }; defer tx.Rollback()
	var count int
	query:="SELECT (SELECT count(*) FROM projects,json_each(projects.document_json,'$.nodes') n WHERE json_extract(n.value,'$.channelId')=?) + (SELECT count(*) FROM jobs WHERE channel_id=? AND status IN ('queued','running'))"
	if err=tx.QueryRowContext(ctx,query,id,id).Scan(&count); err!=nil { return mapped(err) }; if count>0 { return domain.Fail("CONFLICT") }
	if err=changed(tx.ExecContext(ctx,"DELETE FROM channels WHERE id=?",id)); err!=nil { return err }; return mapped(tx.Commit())
}
func (s *Store) PutCredential(ctx context.Context,ref,value string) error {
	nonce:=make([]byte,s.aead.NonceSize()); if _,err:=rand.Read(nonce); err!=nil { return mapped(err) }
	encrypted:=s.aead.Seal(nil,nonce,[]byte(value),[]byte(ref))
	_,err:=s.db.ExecContext(ctx,"INSERT INTO credentials(ref,nonce,ciphertext) VALUES(?,?,?)",ref,nonce,encrypted); return mapped(err)
}
func (s *Store) GetCredential(ctx context.Context,ref string) (string,error) {
	var nonce,data []byte
	if err:=s.db.QueryRowContext(ctx,"SELECT nonce,ciphertext FROM credentials WHERE ref=?",ref).Scan(&nonce,&data); err!=nil { return "",mapped(err) }
	if len(nonce)!=s.aead.NonceSize() { return "",domain.Fail("INTERNAL") }
	raw,err:=s.aead.Open(nil,nonce,data,[]byte(ref)); if err!=nil { return "",mapped(err) }; return string(raw),nil
}
func (s *Store) DeleteCredential(ctx context.Context,ref string) error { _,err:=s.db.ExecContext(ctx,"DELETE FROM credentials WHERE ref=?",ref); return mapped(err) }
const assetColumns="id,name,mime_type,bytes,created_at"
func scanAsset(row scanner) (domain.Asset,error) { var a domain.Asset; err:=row.Scan(&a.ID,&a.Name,&a.MIMEType,&a.Bytes,&a.CreatedAt); return a,mapped(err) }
func (s *Store) ListAssets(ctx context.Context) ([]domain.Asset,error) {
	rows,err:=s.db.QueryContext(ctx,"SELECT "+assetColumns+" FROM assets ORDER BY created_at DESC"); if err!=nil { return nil,mapped(err) }; defer rows.Close()
	out:=[]domain.Asset{}; for rows.Next() { a,err:=scanAsset(rows); if err!=nil { return nil,err }; out=append(out,a) }; return out,mapped(rows.Err())
}
func (s *Store) GetAsset(ctx context.Context,id string) (domain.Asset,error) { return scanAsset(s.db.QueryRowContext(ctx,"SELECT "+assetColumns+" FROM assets WHERE id=?",id)) }
func (s *Store) SaveAsset(ctx context.Context,a domain.Asset) error {
	_,err:=s.db.ExecContext(ctx,"INSERT INTO assets("+assetColumns+") VALUES(?,?,?,?,?)",a.ID,a.Name,a.MIMEType,a.Bytes,a.CreatedAt); return mapped(err)
}
func (s *Store) DeleteAsset(ctx context.Context,id string) error {
	tx,err:=s.db.BeginTx(ctx,nil); if err!=nil { return mapped(err) }; defer tx.Rollback()
	var count int
	query:="SELECT (SELECT count(*) FROM projects,json_each(projects.document_json,'$.nodes') n WHERE json_extract(n.value,'$.assetId')=?) + (SELECT count(*) FROM jobs WHERE asset_id=?)"
	if err=tx.QueryRowContext(ctx,query,id,id).Scan(&count); err!=nil { return mapped(err) }; if count>0 { return domain.Fail("CONFLICT") }
	// Idempotent removal lets a failed file deletion be retried.
	if _,err=tx.ExecContext(ctx,"DELETE FROM assets WHERE id=?",id); err!=nil { return mapped(err) }; return mapped(tx.Commit())
}
const jobColumns="id,request_id,input_hash,project_id,node_id,channel_id,model,kind,prompt,status,text,asset_id,error_code,created_at,updated_at"
func scanJob(row scanner) (domain.Job,error) {
	var j domain.Job; err:=row.Scan(&j.ID,&j.RequestID,&j.InputHash,&j.ProjectID,&j.NodeID,&j.ChannelID,&j.Model,&j.Kind,&j.Prompt,&j.Status,&j.Text,&j.AssetID,&j.ErrorCode,&j.CreatedAt,&j.UpdatedAt); return j,mapped(err)
}
func (s *Store) GetJobByRequest(ctx context.Context,id string) (domain.Job,error) { return scanJob(s.db.QueryRowContext(ctx,"SELECT "+jobColumns+" FROM jobs WHERE request_id=?",id)) }
func (s *Store) CreateIdempotent(ctx context.Context,j domain.Job,expectedChannel domain.Channel,admit bool) (domain.Job,bool,error) {
	tx,err:=s.db.BeginTx(ctx,nil); if err!=nil { return j,false,mapped(err) }; defer tx.Rollback()
	old,err:=scanJob(tx.QueryRowContext(ctx,"SELECT "+jobColumns+" FROM jobs WHERE request_id=?",j.RequestID))
	if err==nil { if old.InputHash!=j.InputHash { return old,false,domain.Fail("CONFLICT") }; return old,false,nil }
	if domain.Code(err)!="NOT_FOUND" { return j,false,err }
	if !admit {return j,false,domain.Fail("BUSY")}
	var raw string
	if err=tx.QueryRowContext(ctx,"SELECT document_json FROM projects WHERE id=?",j.ProjectID).Scan(&raw); err!=nil { return j,false,mapped(err) }
	doc,err:=decodeDocument(raw); if err!=nil { return j,false,err }
	found:=false; for _,n:=range doc.Nodes { if n.ID==j.NodeID && n.Kind==j.Kind { found=true } }; if !found { return j,false,domain.Fail("INVALID_INPUT") }
	channel,err:=scanChannel(tx.QueryRowContext(ctx,"SELECT "+channelColumns+" FROM channels WHERE id=?",j.ChannelID))
	if err!=nil{return j,false,err};if !reflect.DeepEqual(channel,expectedChannel) || !expectedChannel.Supports(j.Model,j.Kind){return j,false,domain.Fail("CONFLICT")}
	_,err=tx.ExecContext(ctx,"INSERT INTO jobs("+jobColumns+") VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",j.ID,j.RequestID,j.InputHash,j.ProjectID,j.NodeID,j.ChannelID,j.Model,j.Kind,j.Prompt,j.Status,j.Text,j.AssetID,j.ErrorCode,j.CreatedAt,j.UpdatedAt)
	if err!=nil { return j,false,mapped(err) }; if err=tx.Commit(); err!=nil { return j,false,mapped(err) }; return j,true,nil
}
func (s *Store) GetJob(ctx context.Context,id string) (domain.Job,error) { return scanJob(s.db.QueryRowContext(ctx,"SELECT "+jobColumns+" FROM jobs WHERE id=?",id)) }
func (s *Store) ListJobs(ctx context.Context,id string) ([]domain.Job,error) {
	query:="SELECT "+jobColumns+" FROM jobs WHERE project_id=? AND (status IN ('queued','running') OR id IN (SELECT id FROM jobs WHERE project_id=? ORDER BY created_at DESC LIMIT 100) OR id IN (SELECT json_extract(n.value,'$.jobId') FROM projects,json_each(projects.document_json,'$.nodes') n WHERE projects.id=?)) ORDER BY created_at DESC"
	rows,err:=s.db.QueryContext(ctx,query,id,id,id); if err!=nil { return nil,mapped(err) }; defer rows.Close()
	out:=[]domain.Job{}; for rows.Next() { j,err:=scanJob(rows); if err!=nil { return nil,err }; out=append(out,j) }; return out,mapped(rows.Err())
}
func (s *Store) SetJobStatus(ctx context.Context,j domain.Job,expected []string) (bool,error) {
	if len(expected)==0 { return false,domain.Fail("INVALID_INPUT") }
	query:="UPDATE jobs SET status=?,text=?,asset_id=?,error_code=?,updated_at=? WHERE id=? AND status IN ("
	args:=[]any{j.Status,j.Text,j.AssetID,j.ErrorCode,j.UpdatedAt,j.ID}
	for i,status:=range expected { if i>0 { query+="," }; query+="?"; args=append(args,status) }; query+=")"
	result,err:=s.db.ExecContext(ctx,query,args...); if err!=nil { return false,mapped(err) }; n,err:=result.RowsAffected(); return n>0,mapped(err)
}
func (s *Store) InterruptPending(ctx context.Context) error {
	_,err:=s.db.ExecContext(ctx,"UPDATE jobs SET status='failed',error_code='INTERRUPTED',updated_at=? WHERE status IN ('queued','running')",time.Now().UTC().Format(time.RFC3339Nano)); return mapped(err)
}
const libraryColumns="id,kind,name,content,created_at,updated_at"
func (s *Store) ListLibrary(ctx context.Context) ([]domain.LibraryEntry,error) {
	rows,err:=s.db.QueryContext(ctx,"SELECT "+libraryColumns+" FROM library_entries ORDER BY updated_at DESC"); if err!=nil { return nil,mapped(err) }; defer rows.Close()
	out:=[]domain.LibraryEntry{}; for rows.Next() { var e domain.LibraryEntry; if err=rows.Scan(&e.ID,&e.Kind,&e.Name,&e.Content,&e.CreatedAt,&e.UpdatedAt); err!=nil { return nil,mapped(err) }; out=append(out,e) }; return out,mapped(rows.Err())
}
func (s *Store) SaveLibraryEntry(ctx context.Context,e domain.LibraryEntry) (domain.LibraryEntry,error) {
	var stored domain.LibraryEntry
	err:=s.db.QueryRowContext(ctx,"INSERT INTO library_entries("+libraryColumns+") VALUES(?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET kind=excluded.kind,name=excluded.name,content=excluded.content,updated_at=excluded.updated_at RETURNING "+libraryColumns,e.ID,e.Kind,e.Name,e.Content,e.CreatedAt,e.UpdatedAt).Scan(&stored.ID,&stored.Kind,&stored.Name,&stored.Content,&stored.CreatedAt,&stored.UpdatedAt)
	return stored,mapped(err)
}
func (s *Store) DeleteLibraryEntry(ctx context.Context,id string) error { return changed(s.db.ExecContext(ctx,"DELETE FROM library_entries WHERE id=?",id)) }
