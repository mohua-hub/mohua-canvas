package library

import (
	"context"
	"strings"
	"time"
	"github.com/google/uuid"
	"github.com/mohua-hub/mohua-canvas/internal/application/ports"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)
type View struct { ID,Kind,Name,Content,CreatedAt,UpdatedAt string }
type SaveInput struct { ID,Kind,Name,Content string }
type Service struct { store ports.LibraryStore }
func New(s ports.LibraryStore) *Service { return &Service{s} }
func view(e domain.LibraryEntry) View { return View{e.ID,e.Kind,e.Name,e.Content,e.CreatedAt,e.UpdatedAt} }
func (s *Service) List(ctx context.Context) ([]View,error) {
	es,err:=s.store.ListLibrary(ctx); if err!=nil { return nil,err }
	out:=make([]View,0,len(es)); for _,e:=range es { out=append(out,view(e)) }; return out,nil
}
func (s *Service) Save(ctx context.Context,in SaveInput) (View,error) {
	if (in.ID!="" && !domain.ValidID(in.ID)) || (in.Kind!="prompt" && in.Kind!="skill") ||
		!domain.ValidName(in.Name) || strings.TrimSpace(in.Content)=="" || len(in.Content)>32000 { return View{},domain.Fail("INVALID_INPUT") }
	now:=time.Now().UTC().Format(time.RFC3339Nano)
	e:=domain.LibraryEntry{ID:in.ID,Kind:in.Kind,Name:strings.TrimSpace(in.Name),Content:in.Content,CreatedAt:now,UpdatedAt:now}
	if e.ID=="" { e.ID=uuid.NewString() }
	stored,err:=s.store.SaveLibraryEntry(ctx,e); if err!=nil { return View{},err }
	return view(stored),nil
}
func (s *Service) Delete(ctx context.Context,id string) error {
	if !domain.ValidID(id) { return domain.Fail("INVALID_INPUT") }; return s.store.DeleteLibraryEntry(ctx,id)
}
