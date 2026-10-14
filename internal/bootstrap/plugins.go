package bootstrap

import (
	"encoding/json"
	"io"
	"os"
	"reflect"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)
type Manifest struct {
	ID string `json:"id"`; Type string `json:"type"`; Version string `json:"version"`
	APIVersion string `json:"apiVersion"`; Entry string `json:"entry"`
	Provides []string `json:"provides"`; Dependencies []string `json:"dependencies"`; Capabilities []string `json:"capabilities"`
}
type pluginConfig struct { Version int `json:"version"`; Plugins []Manifest `json:"plugins"` }
type definition struct { id,kind string; provides,dependencies,capabilities []string }
func definitions() map[string]definition { return map[string]definition{
	"sqlite":{"storage.sqlite","storage",[]string{"storage","credentials"},[]string{},[]string{}},
	"local-media":{"media.local","media",[]string{"media"},[]string{},[]string{}},
	"openai":{"provider.openai","provider",[]string{"model"},[]string{"storage.sqlite"},[]string{"text","image","video","audio"}},
	"gemini":{"provider.gemini","provider",[]string{"model"},[]string{"storage.sqlite"},[]string{"text","image"}},
} }
func LoadManifests(path string) ([]Manifest,error) {
	f,err:=os.Open(path); if err!=nil { return nil,err }; defer f.Close()
	decoder:=json.NewDecoder(io.LimitReader(f,1<<20)); decoder.DisallowUnknownFields()
	var config pluginConfig
	if err=decoder.Decode(&config); err!=nil { return nil,err }
	if decoder.Decode(&struct{}{})!=io.EOF || config.Version!=1 { return nil,domain.Fail("INVALID_INPUT") }
	return OrderManifests(config.Plugins)
}
func OrderManifests(manifests []Manifest) ([]Manifest,error) {
	definitions:=definitions()
	byID:=map[string]Manifest{}; storage,media:=0,0
	for _,m:=range manifests {
		d,ok:=definitions[m.Entry]
		if !ok || byID[m.ID].ID!="" || m.ID!=d.id || m.Type!=d.kind || m.Version!="1.0.0" || m.APIVersion!="1" ||
			!reflect.DeepEqual(m.Provides,d.provides) || !reflect.DeepEqual(m.Dependencies,d.dependencies) || !reflect.DeepEqual(m.Capabilities,d.capabilities) {
			return nil,domain.Fail("INVALID_INPUT")
		}
		byID[m.ID]=m; if m.Type=="storage" { storage++ }; if m.Type=="media" { media++ }
	}
	if storage!=1 || media!=1 { return nil,domain.Fail("INVALID_INPUT") }
	state:=map[string]int{}; out:=[]Manifest{}
	var visit func(string) error
	visit=func(id string) error {
		if state[id]==2 { return nil }; if state[id]==1 { return domain.Fail("INVALID_INPUT") }
		m,ok:=byID[id]; if !ok { return domain.Fail("INVALID_INPUT") }; state[id]=1
		for _,dep:=range m.Dependencies { if err:=visit(dep); err!=nil { return err } }
		state[id]=2; out=append(out,m); return nil
	}
	for _,m:=range manifests { if err:=visit(m.ID); err!=nil { return nil,err } }; return out,nil
}
