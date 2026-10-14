package generation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"sync"
	"time"
	"github.com/google/uuid"
	"github.com/mohua-hub/mohua-canvas/internal/application/ports"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)
type SubmitInput struct { RequestID,ProjectID,NodeID,ChannelID,Model string; Kind domain.Kind; Prompt string }
type View struct { ID,ProjectID,NodeID string; Kind domain.Kind; Status,Text,AssetID,ErrorCode,CreatedAt,UpdatedAt string }
type Service struct {
	store ports.JobStore; projects ports.ProjectStore; channels ports.ChannelStore
	assets ports.GeneratedAssets; providers map[string]ports.ModelProvider
	ctx context.Context; stop context.CancelFunc
	mu sync.Mutex; cancels map[string]context.CancelFunc; pending int
	slots chan struct{}; wg sync.WaitGroup
}
func New(store ports.JobStore,p ports.ProjectStore,c ports.ChannelStore,a ports.GeneratedAssets,models map[string]ports.ModelProvider) *Service {
	ctx,stop:=context.WithCancel(context.Background())
	return &Service{store:store,projects:p,channels:c,assets:a,providers:models,ctx:ctx,stop:stop,cancels:map[string]context.CancelFunc{},slots:make(chan struct{},2)}
}
func view(j domain.Job) View { return View{j.ID,j.ProjectID,j.NodeID,j.Kind,j.Status,j.Text,j.AssetID,j.ErrorCode,j.CreatedAt,j.UpdatedAt} }
func (s *Service) Submit(ctx context.Context,in SubmitInput) (View,error) {
	if !domain.ValidID(in.RequestID) || !domain.ValidID(in.ProjectID) || !domain.ValidID(in.NodeID) ||
		!domain.ValidID(in.ChannelID) || !in.Kind.Valid() || strings.TrimSpace(in.Model)=="" || len(in.Model)>200 || strings.TrimSpace(in.Prompt)=="" || len(in.Prompt)>32000 {
		return View{},domain.Fail("INVALID_INPUT")
	}
	payload,_:=json.Marshal(in); hash:=sha256.Sum256(payload); inputHash:=hex.EncodeToString(hash[:])
	if old,err:=s.store.GetJobByRequest(ctx,in.RequestID); err==nil {
		if old.InputHash!=inputHash { return View{},domain.Fail("CONFLICT") }; return view(old),nil
	} else if domain.Code(err)!="NOT_FOUND" { return View{},err }
	p,err:=s.projects.OpenProject(ctx,in.ProjectID); if err!=nil { return View{},err }
	found:=false; for _,n:=range p.Document.Nodes { if n.ID==in.NodeID && n.Kind==in.Kind { found=true } }
	if !found { return View{},domain.Fail("INVALID_INPUT") }
	c,err:=s.channels.GetChannel(ctx,in.ChannelID); if err!=nil { return View{},err }
	provider,ok:=s.providers[c.PluginID]
	if !ok || !c.Supports(in.Model,in.Kind) { return View{},domain.Fail("UNSUPPORTED") }
	now:=time.Now().UTC().Format(time.RFC3339Nano)
	j:=domain.Job{ID:uuid.NewString(),RequestID:in.RequestID,InputHash:inputHash,ProjectID:in.ProjectID,NodeID:in.NodeID,ChannelID:in.ChannelID,Model:in.Model,Kind:in.Kind,Prompt:in.Prompt,Status:"queued",CreatedAt:now,UpdatedAt:now}
	s.mu.Lock(); defer s.mu.Unlock()
	if s.ctx.Err()!=nil { return View{},domain.Fail("BUSY") }
	// The store returns an existing receipt before applying the queue bound.
	j,created,err:=s.store.CreateIdempotent(ctx,j,c,s.pending<34); if err!=nil { return View{},err }
	if !created { return view(j),nil }
	taskCtx,cancel:=context.WithTimeout(s.ctx,10*time.Minute)
	s.cancels[j.ID]=cancel; s.pending++; s.wg.Add(1)
	go s.run(taskCtx,j,c,provider)
	return view(j),nil
}
func (s *Service) run(ctx context.Context,j domain.Job,c domain.Channel,p ports.ModelProvider) {
	defer s.wg.Done()
	defer func(){ s.mu.Lock(); if cancel:=s.cancels[j.ID]; cancel!=nil { cancel() }; delete(s.cancels,j.ID); s.pending--; s.mu.Unlock() }()
	select { case s.slots<-struct{}{}: defer func(){<-s.slots}(); case <-ctx.Done(): s.finish(j,ctx.Err()); return }
	j.Status="running"; j.UpdatedAt=time.Now().UTC().Format(time.RFC3339Nano)
	changed,err:=s.store.SetJobStatus(ctx,j,[]string{"queued"}); if err!=nil {s.finish(j,err);return}; if !changed { return }
	result,err:=p.Generate(ctx,domain.GenerationRequest{Model:j.Model,Kind:j.Kind,Prompt:j.Prompt},c)
	if ctx.Err()!=nil { s.finish(j,ctx.Err()); return }
	if err!=nil { s.finish(j,err); return }
	if j.Kind==domain.Text {
		if result.Text=="" || len(result.Text)>32000 { s.finish(j,domain.Fail("UPSTREAM_FAILED")); return }; j.Text=result.Text
	} else {
		if len(result.Media)==0 { s.finish(j,domain.Fail("UPSTREAM_FAILED")); return }
		a,err:=s.assets.ImportGenerated(ctx,result.MediaName,result.Media)
		if err!=nil { s.finish(j,err); return }; j.AssetID=a.ID
		match:=strings.HasPrefix(a.MIMEType,string(j.Kind)+"/")
		if !match { _=s.assets.Delete(context.Background(),a.ID); s.finish(j,domain.Fail("UPSTREAM_FAILED")); return }
	}
	j.Status="succeeded"; j.UpdatedAt=time.Now().UTC().Format(time.RFC3339Nano)
	persistCtx,cancel:=context.WithTimeout(context.Background(),5*time.Second); defer cancel()
	changed,err=s.store.SetJobStatus(persistCtx,j,[]string{"running"})
	if j.AssetID!="" && (!changed || err!=nil) { _=s.assets.Delete(persistCtx,j.AssetID) }
	if err!=nil { j.Text,j.AssetID="","";s.finish(j,err) }
}
func (s *Service) finish(j domain.Job,err error) {
	j.Text,j.AssetID="",""
	j.Status="failed"; j.ErrorCode=domain.Code(err)
	if err==context.DeadlineExceeded { j.ErrorCode="TIMEOUT" }
	if err==context.Canceled { j.Status="cancelled"; j.ErrorCode="" }
	j.UpdatedAt=time.Now().UTC().Format(time.RFC3339Nano)
	ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second); defer cancel()
	_,_=s.store.SetJobStatus(ctx,j,[]string{"queued","running"})
}
func (s *Service) Get(ctx context.Context,id string) (View,error) {
	if !domain.ValidID(id) { return View{},domain.Fail("INVALID_INPUT") }
	j,err:=s.store.GetJob(ctx,id); return view(j),err
}
func (s *Service) List(ctx context.Context,projectID string) ([]View,error) {
	if !domain.ValidID(projectID) { return nil,domain.Fail("INVALID_INPUT") }
	js,err:=s.store.ListJobs(ctx,projectID); if err!=nil { return nil,err }
	out:=make([]View,0,len(js)); for _,j:=range js { out=append(out,view(j)) }; return out,nil
}
func (s *Service) Cancel(ctx context.Context,id string) (View,error) {
	if !domain.ValidID(id) { return View{},domain.Fail("INVALID_INPUT") }
	s.mu.Lock(); defer s.mu.Unlock()
	j,err:=s.store.GetJob(ctx,id); if err!=nil { return View{},err }
	if !j.Pending() { return view(j),nil }
	j.Status,j.ErrorCode="cancelled",""; j.UpdatedAt=time.Now().UTC().Format(time.RFC3339Nano)
	changed,err:=s.store.SetJobStatus(ctx,j,[]string{"queued","running"}); if err!=nil { return View{},err }
	if changed { if cancel:=s.cancels[id]; cancel!=nil { cancel() }; return view(j),nil }
	j,err=s.store.GetJob(ctx,id); return view(j),err
}
func (s *Service) Close() { s.mu.Lock(); s.stop(); s.mu.Unlock(); s.wg.Wait() }
