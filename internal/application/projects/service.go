package projects

import (
	"context"
	"strings"
	"time"
	"github.com/google/uuid"
	"github.com/mohua-hub/mohua-canvas/internal/application/ports"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)

type View struct { ID, Name string; Revision int64; Document domain.Document; CreatedAt, UpdatedAt string }
type SaveInput struct { ID, Name string; ExpectedRevision int64; Document domain.Document }
type Service struct { store ports.ProjectStore; channels ports.ChannelStore; assets ports.AssetStore; jobs ports.JobStore }
func New(s ports.ProjectStore, c ports.ChannelStore, a ports.AssetStore, j ports.JobStore) *Service { return &Service{s,c,a,j} }
func view(p domain.Project) View { return View{p.ID,p.Name,p.Revision,p.Document,p.CreatedAt,p.UpdatedAt} }
func (s *Service) List(ctx context.Context) ([]View,error) {
	items, err := s.store.ListProjects(ctx); if err != nil { return nil,err }
	result := make([]View,0,len(items)); for _, p := range items { result = append(result,view(p)) }; return result,nil
}
func (s *Service) Open(ctx context.Context, id string) (View,error) {
	if !domain.ValidID(id) { return View{},domain.Fail("INVALID_INPUT") }
	p,err := s.store.OpenProject(ctx,id); return view(p),err
}
func (s *Service) Create(ctx context.Context, name string) (View,error) {
	name = strings.TrimSpace(name); if !domain.ValidName(name) { return View{},domain.Fail("INVALID_INPUT") }
	now := time.Now().UTC().Format(time.RFC3339Nano)
	p := domain.Project{ID:uuid.NewString(),Name:name,Revision:1,Document:domain.EmptyDocument(),CreatedAt:now,UpdatedAt:now}
	return view(p),s.store.CreateProject(ctx,p)
}
func (s *Service) Save(ctx context.Context, in SaveInput) (View,error) {
	if !domain.ValidID(in.ID) || !domain.ValidName(in.Name) || in.ExpectedRevision < 1 { return View{},domain.Fail("INVALID_INPUT") }
	if err := in.Document.Validate(); err != nil { return View{},err }
	p,err := s.store.OpenProject(ctx,in.ID); if err != nil { return View{},err }
	for _, n := range in.Document.Nodes {
		if n.ChannelID != "" { c,err := s.channels.GetChannel(ctx,n.ChannelID); if err != nil || !c.Supports(n.Model,n.Kind) { return View{},domain.Fail("INVALID_INPUT") } }
		if n.AssetID != "" {
			a,err := s.assets.GetAsset(ctx,n.AssetID)
			if err != nil || a.Kind() != n.Kind { return View{},domain.Fail("INVALID_INPUT") }
		}
		if n.JobID != "" {
			j,err := s.jobs.GetJob(ctx,n.JobID)
			if err != nil || j.ProjectID != p.ID || j.NodeID != n.ID || j.Kind != n.Kind { return View{},domain.Fail("INVALID_INPUT") }
		}
	}
	p.Name,p.Document,p.Revision,p.UpdatedAt = strings.TrimSpace(in.Name),in.Document,in.ExpectedRevision+1,time.Now().UTC().Format(time.RFC3339Nano)
	if err := s.store.SaveProject(ctx,p,in.ExpectedRevision); err != nil { return View{},err }; return view(p),nil
}
func (s *Service) Delete(ctx context.Context,id string) error {
	if !domain.ValidID(id) { return domain.Fail("INVALID_INPUT") }; return s.store.DeleteProject(ctx,id)
}
