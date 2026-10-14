package media

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)
type Local struct { dir string }
func New(dir string) (*Local,error) { if err:=os.MkdirAll(dir,0700); err!=nil { return nil,err }; return &Local{dir},nil }
func (s *Local) path(id string) (string,error) { if !domain.ValidID(id) { return "",domain.Fail("INVALID_INPUT") }; return filepath.Join(s.dir,id),nil }
func (s *Local) Put(ctx context.Context,id string,data []byte) error {
	if ctx.Err()!=nil { return ctx.Err() }; path,err:=s.path(id); if err!=nil { return err }
	f,err:=os.CreateTemp(s.dir,".import-"); if err!=nil { return err }
	temp:=f.Name(); defer os.Remove(temp)
	if _,err=f.Write(data); err!=nil { _=f.Close(); return err }
	if err=f.Sync(); err!=nil { _=f.Close(); return err }; if err=f.Close(); err!=nil { return err }
	if ctx.Err()!=nil { return ctx.Err() }; return os.Rename(temp,path)
}
func (s *Local) Open(ctx context.Context,id string) (io.ReadSeekCloser,error) {
	if ctx.Err()!=nil { return nil,ctx.Err() }; path,err:=s.path(id); if err!=nil { return nil,err }; return os.Open(path)
}
func (s *Local) Remove(ctx context.Context,id string) error {
	if ctx.Err()!=nil { return ctx.Err() }; path,err:=s.path(id); if err!=nil { return err }
	err=os.Remove(path); if os.IsNotExist(err) { return nil }; return err
}
