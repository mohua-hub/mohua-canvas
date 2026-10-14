package channels

import (
	"context"
	"strings"
	"sync"
	"github.com/google/uuid"
	"github.com/mohua-hub/mohua-canvas/internal/application/ports"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)
type View struct { ID, Name, PluginID, BaseURL string; Models []domain.Model; HasCredential bool }
type ConfigureInput struct { ID, Name, PluginID, BaseURL, APIKey string; Models []domain.Model }
type Service struct { store ports.ChannelStore; vault ports.CredentialVault; providers map[string]ports.ModelProvider; mu sync.Mutex }
func New(s ports.ChannelStore,v ports.CredentialVault,p map[string]ports.ModelProvider) *Service { return &Service{store:s,vault:v,providers:p} }
func view(c domain.Channel) View { return View{c.ID,c.Name,c.PluginID,c.BaseURL,c.Models,c.CredentialRef!=""} }
func (s *Service) List(ctx context.Context) ([]View,error) {
	cs,err := s.store.ListChannels(ctx); if err != nil { return nil,err }
	out := make([]View,0,len(cs)); for _,c := range cs { out=append(out,view(c)) }; return out,nil
}
func (s *Service) Configure(ctx context.Context,in ConfigureInput) (View,error) {
	s.mu.Lock();defer s.mu.Unlock()
	c := domain.Channel{ID:in.ID,Name:strings.TrimSpace(in.Name),PluginID:in.PluginID,BaseURL:strings.TrimRight(in.BaseURL,"/"),Models:in.Models}
	oldRef := ""
	if c.ID == "" { c.ID=uuid.NewString() } else {
		if !domain.ValidID(c.ID) { return View{},domain.Fail("INVALID_INPUT") }
		old,err := s.store.GetChannel(ctx,c.ID); if err != nil { return View{},err }; oldRef=old.CredentialRef
	}
	if err := c.Validate(); err != nil { return View{},err }
	provider,ok := s.providers[c.PluginID]; if !ok { return View{},domain.Fail("UNSUPPORTED") }
	for _,m := range c.Models {
		supported:=false; for _,k:=range provider.Capabilities() { if k==m.Kind { supported=true } }
		if !supported { return View{},domain.Fail("UNSUPPORTED") }
	}
	if len(in.APIKey)>8192 || (in.APIKey=="" && oldRef=="") { return View{},domain.Fail("INVALID_INPUT") }
	c.CredentialRef=oldRef
	if in.APIKey!="" {
		c.CredentialRef=uuid.NewString()
		if err:=s.vault.PutCredential(ctx,c.CredentialRef,in.APIKey); err!=nil { return View{},err }
	}
	if err:=s.store.SaveChannel(ctx,c); err!=nil {
		if c.CredentialRef!=oldRef { _=s.vault.DeleteCredential(ctx,c.CredentialRef) }
		return View{},err
	}
	if oldRef!="" && c.CredentialRef!=oldRef { _=s.vault.DeleteCredential(ctx,oldRef) }
	return view(c),nil
}
func (s *Service) Delete(ctx context.Context,id string) error {
	s.mu.Lock();defer s.mu.Unlock()
	if !domain.ValidID(id) { return domain.Fail("INVALID_INPUT") }
	c,err:=s.store.GetChannel(ctx,id); if err!=nil { return err }
	if err=s.store.DeleteChannel(ctx,id); err!=nil { return err }
	return s.vault.DeleteCredential(ctx,c.CredentialRef)
}
