package generation_test

import (
	"bytes"
	"context"
	"sync/atomic"
	"testing"
	"time"
	"github.com/google/uuid"
	"github.com/mohua-hub/mohua-canvas/internal/adapters/outbound/sqlite"
	"github.com/mohua-hub/mohua-canvas/internal/application/generation"
	"github.com/mohua-hub/mohua-canvas/internal/application/ports"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)
type blockingProvider struct{calls atomic.Int32;started chan struct{}}
func (p *blockingProvider) ID()string{return "provider.test"}
func (p *blockingProvider) Capabilities()[]domain.Kind{return []domain.Kind{domain.Text}}
func (p *blockingProvider) Generate(ctx context.Context,_ domain.GenerationRequest,_ domain.Channel)(domain.GenerationResult,error){
	p.calls.Add(1);select{case p.started<-struct{}{}:default:};<-ctx.Done()
	// Deliberately return a late success after cancellation.
	return domain.GenerationResult{Text:"late"},nil
}
func setup(t *testing.T)(*generation.Service,*sqlite.Store,generation.SubmitInput,*blockingProvider){
	t.Helper();ctx:=context.Background()
	store,err:=sqlite.Open(t.TempDir(),bytes.Repeat([]byte{4},32));if err!=nil{t.Fatal(err)}
	c:=domain.Channel{ID:uuid.NewString(),Name:"测试",PluginID:"provider.test",BaseURL:"https://example.com",CredentialRef:uuid.NewString(),Models:[]domain.Model{{Name:"text",Kind:domain.Text}}}
	if err=store.PutCredential(ctx,c.CredentialRef,"key");err!=nil{t.Fatal(err)}
	if err=store.SaveChannel(ctx,c);err!=nil{t.Fatal(err)}
	d:=domain.EmptyDocument();d.Nodes=[]domain.Node{{ID:uuid.NewString(),Kind:domain.Text,Title:"草稿"}}
	now:=time.Now().UTC().Format(time.RFC3339Nano)
	p:=domain.Project{ID:uuid.NewString(),Name:"项目",Revision:1,Document:d,CreatedAt:now,UpdatedAt:now}
	if err=store.CreateProject(ctx,p);err!=nil{t.Fatal(err)}
	provider:=&blockingProvider{started:make(chan struct{},2)}
	service:=generation.New(store,store,store,nil,map[string]ports.ModelProvider{provider.ID():provider})
	t.Cleanup(func(){service.Close();_=store.Close()})
	return service,store,generation.SubmitInput{RequestID:uuid.NewString(),ProjectID:p.ID,NodeID:d.Nodes[0].ID,ChannelID:c.ID,Model:"text",Kind:domain.Text,Prompt:"创作"},provider
}
func TestCancelIsDurableAndReceiptDoesNotResubmit(t *testing.T){
	s,store,in,provider:=setup(t);ctx:=context.Background()
	first,err:=s.Submit(ctx,in);if err!=nil{t.Fatal(err)}
	select{case <-provider.started:case <-time.After(3*time.Second):t.Fatal("worker did not start")}
	replayed,err:=s.Submit(ctx,in);if err!=nil||replayed.ID!=first.ID{t.Fatal("receipt changed")}
	cancelled,err:=s.Cancel(ctx,first.ID);if err!=nil||cancelled.Status!="cancelled"{t.Fatal("cancellation failed")}
	s.Close()
	saved,err:=store.GetJob(ctx,first.ID);if err!=nil||saved.Status!="cancelled"||saved.Text!=""{t.Fatalf("%+v %v",saved,err)}
	replayed,err=s.Submit(ctx,in);if err!=nil||replayed.ID!=first.ID||provider.calls.Load()!=1{t.Fatal("cancelled request was resubmitted")}
}
func TestQueueAdmissionDoesNotCreateFailedReceipts(t *testing.T){
	s,store,in,provider:=setup(t);ctx:=context.Background()
	for i:=0;i<34;i++{in.RequestID=uuid.NewString();if _,err:=s.Submit(ctx,in);err!=nil{t.Fatalf("request %d: %v",i,err)}}
	select{case <-provider.started:case <-time.After(3*time.Second):t.Fatal("worker did not start")}
	in.RequestID=uuid.NewString()
	if _,err:=s.Submit(ctx,in);domain.Code(err)!="BUSY"{t.Fatalf("want BUSY got %v",err)}
	if _,err:=store.GetJobByRequest(ctx,in.RequestID);domain.Code(err)!="NOT_FOUND"{t.Fatal("BUSY produced a persisted task")}
	if provider.calls.Load()>2{t.Fatal("concurrency exceeded")}
}
