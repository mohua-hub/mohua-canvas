package sqlite

import (
	"bytes"
	"context"
	"sync"
	"testing"
	"time"
	"github.com/google/uuid"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)

func fixture(t *testing.T) (*Store,domain.Project,domain.Channel) {
	t.Helper()
	s,err:=Open(t.TempDir(),bytes.Repeat([]byte{7},32));if err!=nil{t.Fatal(err)}
	t.Cleanup(func(){_=s.Close()})
	ctx:=context.Background();now:=time.Now().UTC().Format(time.RFC3339Nano)
	c:=domain.Channel{ID:uuid.NewString(),Name:"测试模型",PluginID:"provider.openai",BaseURL:"https://example.com/v1",CredentialRef:uuid.NewString(),Models:[]domain.Model{{Name:"model",Kind:domain.Text}}}
	if err=s.PutCredential(ctx,c.CredentialRef,"secret");err!=nil{t.Fatal(err)}
	if err=s.SaveChannel(ctx,c);err!=nil{t.Fatal(err)}
	d:=domain.EmptyDocument();d.Nodes=[]domain.Node{{ID:uuid.NewString(),Kind:domain.Text,Title:"文本",ChannelID:c.ID,Model:"model"}}
	p:=domain.Project{ID:uuid.NewString(),Name:"项目",Revision:1,Document:d,CreatedAt:now,UpdatedAt:now}
	if err=s.CreateProject(ctx,p);err!=nil{t.Fatal(err)}
	return s,p,c
}
func newJob(p domain.Project,c domain.Channel) domain.Job {
	now:=time.Now().UTC().Format(time.RFC3339Nano)
	return domain.Job{ID:uuid.NewString(),RequestID:uuid.NewString(),InputHash:"hash",ProjectID:p.ID,NodeID:p.Document.Nodes[0].ID,ChannelID:c.ID,Model:"model",Kind:domain.Text,Prompt:"创作",Status:"queued",CreatedAt:now,UpdatedAt:now}
}
func TestConcurrentSavesUseRevisionCAS(t *testing.T) {
	s,p,_:=fixture(t);ctx:=context.Background();results:=make(chan error,2)
	var wg sync.WaitGroup
	for _,name:=range []string{"甲","乙"} {wg.Add(1);go func(name string){defer wg.Done();copy:=p;copy.Name=name;results<-s.SaveProject(ctx,copy,1)}(name)}
	wg.Wait();close(results);success,conflict:=0,0
	for err:=range results{if err==nil{success++}else if domain.Code(err)=="CONFLICT"{conflict++}else{t.Fatal(err)}}
	if success!=1||conflict!=1{t.Fatalf("success=%d conflict=%d",success,conflict)}
	saved,err:=s.OpenProject(ctx,p.ID);if err!=nil||saved.Revision!=2{t.Fatalf("saved=%+v err=%v",saved,err)}
}
func TestReceiptAdmissionAndChannelSnapshot(t *testing.T) {
	s,p,c:=fixture(t);ctx:=context.Background();j:=newJob(p,c)
	if _,_,err:=s.CreateIdempotent(ctx,j,c,false);domain.Code(err)!="BUSY"{t.Fatalf("want BUSY, got %v",err)}
	if _,err:=s.GetJobByRequest(ctx,j.RequestID);domain.Code(err)!="NOT_FOUND"{t.Fatal("BUSY must not persist a receipt")}
	stale:=c;stale.BaseURL="https://other.example/v1"
	if _,_,err:=s.CreateIdempotent(ctx,j,stale,true);domain.Code(err)!="CONFLICT"{t.Fatal("stale channel accepted")}
	stored,created,err:=s.CreateIdempotent(ctx,j,c,true);if err!=nil||!created{t.Fatalf("%v %v",created,err)}
	replayed,created,err:=s.CreateIdempotent(ctx,j,c,false)
	if err!=nil||created||replayed.ID!=stored.ID{t.Fatal("existing receipt must be returned even at capacity")}
	j.InputHash="different"
	if _,_,err=s.CreateIdempotent(ctx,j,c,true);domain.Code(err)!="CONFLICT"{t.Fatal("changed request reused")}
	c.Name="修改"
	if err=s.SaveChannel(ctx,c);domain.Code(err)!="CONFLICT"{t.Fatal("pending credentials can be replaced")}
	if err=s.DeleteProject(ctx,p.ID);domain.Code(err)!="CONFLICT"{t.Fatal("pending project can be deleted")}
	stored.Status="cancelled";if ok,err:=s.SetJobStatus(ctx,stored,[]string{"queued","running"});err!=nil||!ok{t.Fatal(err)}
	stored.Status="succeeded";stored.Text="迟到结果"
	if ok,err:=s.SetJobStatus(ctx,stored,[]string{"running"});err!=nil||ok{t.Fatal("late result overwrote cancellation")}
	actual,err:=s.GetJob(ctx,stored.ID);if err!=nil||actual.Status!="cancelled"{t.Fatal("cancellation not durable")}
}
func TestReferencesAndVaultAreNotPlaintext(t *testing.T) {
	s,p,c:=fixture(t);ctx:=context.Background()
	if err:=s.DeleteChannel(ctx,c.ID);domain.Code(err)!="CONFLICT"{t.Fatal("referenced channel deleted")}
	var ciphertext []byte
	if err:=s.db.QueryRow("SELECT ciphertext FROM credentials WHERE ref=?",c.CredentialRef).Scan(&ciphertext);err!=nil{t.Fatal(err)}
	if bytes.Contains(ciphertext,[]byte("secret")){t.Fatal("credential stored as plaintext")}
	key,err:=s.GetCredential(ctx,c.CredentialRef);if err!=nil||key!="secret"{t.Fatal("vault roundtrip failed")}
	p.Document.Nodes[0].AssetID=uuid.NewString()
	if err=s.SaveProject(ctx,p,1);domain.Code(err)!="INVALID_INPUT"{t.Fatal("missing media reference accepted")}
	saved,err:=s.OpenProject(ctx,p.ID);if err!=nil||saved.Revision!=1{t.Fatal("invalid save altered revision")}
}
func TestLibraryPreservesCreationTime(t *testing.T) {
	s,_,_:=fixture(t);ctx:=context.Background()
	e:=domain.LibraryEntry{ID:uuid.NewString(),Kind:"skill",Name:"方法",Content:"内容",CreatedAt:"original",UpdatedAt:"original"}
	if _,err:=s.SaveLibraryEntry(ctx,e);err!=nil{t.Fatal(err)}
	e.CreatedAt="replacement";e.UpdatedAt="updated";e.Content="新内容"
	stored,err:=s.SaveLibraryEntry(ctx,e)
	if err!=nil||stored.CreatedAt!="original"||stored.UpdatedAt!="updated"{t.Fatalf("%+v %v",stored,err)}
}
