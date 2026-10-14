package bootstrap

import (
	"context"
	"path/filepath"
	"github.com/mohua-hub/mohua-canvas/internal/adapters/outbound/media"
	"github.com/mohua-hub/mohua-canvas/internal/adapters/outbound/providers/gemini"
	"github.com/mohua-hub/mohua-canvas/internal/adapters/outbound/providers/openai"
	"github.com/mohua-hub/mohua-canvas/internal/adapters/outbound/sqlite"
	"github.com/mohua-hub/mohua-canvas/internal/application/assets"
	"github.com/mohua-hub/mohua-canvas/internal/application/channels"
	"github.com/mohua-hub/mohua-canvas/internal/application/generation"
	"github.com/mohua-hub/mohua-canvas/internal/application/library"
	"github.com/mohua-hub/mohua-canvas/internal/application/plugins"
	"github.com/mohua-hub/mohua-canvas/internal/application/ports"
	"github.com/mohua-hub/mohua-canvas/internal/application/projects"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)
type Runtime struct {
	Projects *projects.Service; Channels *channels.Service; Assets *assets.Service
	Jobs *generation.Service; Library *library.Service; Plugins []plugins.View
	closers []func()
}
func New(dataDir,configPath string,key []byte) (*Runtime,error) {
	manifests,err:=LoadManifests(configPath); if err!=nil { return nil,err }
	r:=&Runtime{Plugins:[]plugins.View{}}
	var db *sqlite.Store; var files *media.Local
	providers:=map[string]ports.ModelProvider{}
	fail:=func(err error)(*Runtime,error){ r.Close(); return nil,err }
	for _,m:=range manifests {
		switch m.Entry {
		case "sqlite":
			db,err=sqlite.Open(dataDir,key); if err!=nil { return fail(err) }; r.closers=append(r.closers,func(){_=db.Close()})
		case "local-media":
			files,err=media.New(filepath.Join(dataDir,"media")); if err!=nil { return fail(err) }
		case "openai":
			if db==nil { return fail(domain.Fail("INVALID_INPUT")) }; p:=openai.New(db); providers[p.ID()]=p; r.closers=append(r.closers,p.Close)
		case "gemini":
			if db==nil { return fail(domain.Fail("INVALID_INPUT")) }; p:=gemini.New(db); providers[p.ID()]=p; r.closers=append(r.closers,p.Close)
		default: return fail(domain.Fail("UNSUPPORTED"))
		}
		if m.Type=="provider" {
			p:=providers[m.ID]; caps:=p.Capabilities()
			if len(caps)!=len(m.Capabilities) { return fail(domain.Fail("INVALID_INPUT")) }
			for i,k:=range caps { if string(k)!=m.Capabilities[i] { return fail(domain.Fail("INVALID_INPUT")) } }
		}
		r.Plugins=append(r.Plugins,plugins.View{ID:m.ID,Type:m.Type,Version:m.Version,APIVersion:m.APIVersion,Status:"active",Capabilities:append([]string{},m.Capabilities...)})
	}
	if db==nil || files==nil { return fail(domain.Fail("INVALID_INPUT")) }
	if err=db.InterruptPending(context.Background()); err!=nil { return fail(err) }
	r.Projects=projects.New(db,db,db,db); r.Channels=channels.New(db,db,providers)
	r.Assets=assets.New(db,files); r.Library=library.New(db)
	r.Jobs=generation.New(db,db,db,r.Assets,providers)
	return r,nil
}
func (r *Runtime) Close() {
	if r.Jobs!=nil { r.Jobs.Close(); r.Jobs=nil }
	for i:=len(r.closers)-1;i>=0;i-- { r.closers[i]() }; r.closers=nil
}
