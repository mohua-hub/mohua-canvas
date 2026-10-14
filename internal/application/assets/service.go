package assets

import (
	"context"
	"io"
	"strings"
	"time"
	"github.com/google/uuid"
	"github.com/gabriel-vasile/mimetype"
	"github.com/mohua-hub/mohua-canvas/internal/application/ports"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)
const MaxBytes = 64 << 20
type View struct { ID,Name,MIMEType string; Bytes int64; CreatedAt string }
type Service struct { store ports.AssetStore; media ports.MediaStore }
func New(s ports.AssetStore,m ports.MediaStore) *Service { return &Service{s,m} }
func view(a domain.Asset) View { return View{a.ID,a.Name,a.MIMEType,a.Bytes,a.CreatedAt} }
func (s *Service) List(ctx context.Context) ([]View,error) {
	as,err:=s.store.ListAssets(ctx); if err!=nil { return nil,err }
	out:=make([]View,0,len(as)); for _,a:=range as { out=append(out,view(a)) }; return out,nil
}
func MIME(data []byte) string {
	m:=strings.Split(mimetype.Detect(data).String(),";")[0]
	switch m {
	case "image/png","image/jpeg","image/webp","image/gif","video/mp4","video/webm","audio/mpeg","audio/wave","audio/wav","audio/x-wav","audio/ogg","audio/webm","application/ogg","audio/flac":
		if m=="application/ogg" { return "audio/ogg" }; if m=="audio/wave" || m=="audio/x-wav" { return "audio/wav" }; return m
	}
	return ""
}
func (s *Service) Import(ctx context.Context,name string,data []byte) (View,error) {
	if strings.TrimSpace(name)=="" || len(name)>240 || len(data)==0 || len(data)>MaxBytes { return View{},domain.Fail("INVALID_INPUT") }
	m:=MIME(data); if m=="" { return View{},domain.Fail("UNSUPPORTED") }
	a:=domain.Asset{ID:uuid.NewString(),Name:name,MIMEType:m,Bytes:int64(len(data)),CreatedAt:time.Now().UTC().Format(time.RFC3339Nano)}
	if err:=s.media.Put(ctx,a.ID,data); err!=nil { return View{},err }
	if err:=s.store.SaveAsset(ctx,a); err!=nil { _=s.media.Remove(ctx,a.ID); return View{},err }
	return view(a),nil
}
func (s *Service) ImportGenerated(ctx context.Context,name string,data []byte) (domain.Asset,error) {
	v,err:=s.Import(ctx,name,data)
	return domain.Asset{ID:v.ID,Name:v.Name,MIMEType:v.MIMEType,Bytes:v.Bytes,CreatedAt:v.CreatedAt},err
}
func (s *Service) Open(ctx context.Context,id string) (View,io.ReadSeekCloser,error) {
	if !domain.ValidID(id) { return View{},nil,domain.Fail("INVALID_INPUT") }
	a,err:=s.store.GetAsset(ctx,id); if err!=nil { return View{},nil,err }
	f,err:=s.media.Open(ctx,id); return view(a),f,err
}
func (s *Service) Delete(ctx context.Context,id string) error {
	if !domain.ValidID(id) { return domain.Fail("INVALID_INPUT") }
	if err:=s.store.DeleteAsset(ctx,id); err!=nil { return err }; return s.media.Remove(ctx,id)
}
