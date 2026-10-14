package httpapi

import (
	"github.com/mohua-hub/mohua-canvas/internal/application/assets"
	"github.com/mohua-hub/mohua-canvas/internal/application/channels"
	"github.com/mohua-hub/mohua-canvas/internal/application/generation"
	"github.com/mohua-hub/mohua-canvas/internal/application/library"
	"github.com/mohua-hub/mohua-canvas/internal/application/plugins"
	"github.com/mohua-hub/mohua-canvas/internal/application/projects"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)
type nodeDTO struct {
	ID string `json:"id"`; Kind domain.Kind `json:"kind"`
	X float64 `json:"x"`; Y float64 `json:"y"`
	Title string `json:"title"`; Content string `json:"content"`; Prompt string `json:"prompt"`
	ChannelID string `json:"channelId"`; Model string `json:"model"`; AssetID string `json:"assetId"`; JobID string `json:"jobId"`
}
type edgeDTO struct { ID string `json:"id"`; Source string `json:"source"`; Target string `json:"target"` }
type viewportDTO struct { X float64 `json:"x"`; Y float64 `json:"y"`; Zoom float64 `json:"zoom"` }
type documentDTO struct { Version int `json:"version"`; Nodes []nodeDTO `json:"nodes"`; Edges []edgeDTO `json:"edges"`; Viewport viewportDTO `json:"viewport"` }
func (d documentDTO) domain() domain.Document {
	out:=domain.Document{Version:d.Version,Nodes:[]domain.Node{},Edges:[]domain.Edge{},Viewport:domain.Viewport{X:d.Viewport.X,Y:d.Viewport.Y,Zoom:d.Viewport.Zoom}}
	for _,n:=range d.Nodes { out.Nodes=append(out.Nodes,domain.Node{ID:n.ID,Kind:n.Kind,X:n.X,Y:n.Y,Title:n.Title,Content:n.Content,Prompt:n.Prompt,ChannelID:n.ChannelID,Model:n.Model,AssetID:n.AssetID,JobID:n.JobID}) }
	for _,e:=range d.Edges { out.Edges=append(out.Edges,domain.Edge{ID:e.ID,Source:e.Source,Target:e.Target}) }; return out
}
func document(d domain.Document) documentDTO {
	out:=documentDTO{Version:d.Version,Nodes:[]nodeDTO{},Edges:[]edgeDTO{},Viewport:viewportDTO{d.Viewport.X,d.Viewport.Y,d.Viewport.Zoom}}
	for _,n:=range d.Nodes { out.Nodes=append(out.Nodes,nodeDTO{n.ID,n.Kind,n.X,n.Y,n.Title,n.Content,n.Prompt,n.ChannelID,n.Model,n.AssetID,n.JobID}) }
	for _,e:=range d.Edges { out.Edges=append(out.Edges,edgeDTO{e.ID,e.Source,e.Target}) }; return out
}
type projectSummaryDTO struct {
	ID string `json:"id"`; Name string `json:"name"`; Revision int64 `json:"revision"`; CreatedAt string `json:"createdAt"`; UpdatedAt string `json:"updatedAt"`
}
type projectDTO struct { projectSummaryDTO; Document documentDTO `json:"document"` }
func summary(p projects.View) projectSummaryDTO { return projectSummaryDTO{p.ID,p.Name,p.Revision,p.CreatedAt,p.UpdatedAt} }
func project(p projects.View) projectDTO { return projectDTO{summary(p),document(p.Document)} }
type modelDTO struct { Name string `json:"name"`; Kind domain.Kind `json:"kind"` }
type channelDTO struct {
	ID string `json:"id"`; Name string `json:"name"`; PluginID string `json:"pluginId"`; BaseURL string `json:"baseUrl"`
	Models []modelDTO `json:"models"`; HasCredential bool `json:"hasCredential"`
}
func channel(c channels.View) channelDTO {
	models:=[]modelDTO{}; for _,m:=range c.Models { models=append(models,modelDTO{m.Name,m.Kind}) }; return channelDTO{c.ID,c.Name,c.PluginID,c.BaseURL,models,c.HasCredential}
}
type assetDTO struct { ID string `json:"id"`; Name string `json:"name"`; MIMEType string `json:"mimeType"`; Bytes int64 `json:"bytes"`; CreatedAt string `json:"createdAt"` }
func asset(a assets.View) assetDTO { return assetDTO{a.ID,a.Name,a.MIMEType,a.Bytes,a.CreatedAt} }
type jobDTO struct {
	ID string `json:"id"`; ProjectID string `json:"projectId"`; NodeID string `json:"nodeId"`; Kind domain.Kind `json:"kind"`
	Status string `json:"status"`; Text string `json:"text"`; AssetID string `json:"assetId"`; ErrorCode string `json:"errorCode"`
	CreatedAt string `json:"createdAt"`; UpdatedAt string `json:"updatedAt"`
}
func job(j generation.View) jobDTO { return jobDTO{j.ID,j.ProjectID,j.NodeID,j.Kind,j.Status,j.Text,j.AssetID,j.ErrorCode,j.CreatedAt,j.UpdatedAt} }
type libraryDTO struct { ID string `json:"id"`; Kind string `json:"kind"`; Name string `json:"name"`; Content string `json:"content"`; CreatedAt string `json:"createdAt"`; UpdatedAt string `json:"updatedAt"` }
func entry(e library.View) libraryDTO { return libraryDTO{e.ID,e.Kind,e.Name,e.Content,e.CreatedAt,e.UpdatedAt} }
type pluginDTO struct {
	ID string `json:"id"`; Type string `json:"type"`; Version string `json:"version"`; APIVersion string `json:"apiVersion"`
	Capabilities []string `json:"capabilities"`; Status string `json:"status"`
}
func plugin(p plugins.View) pluginDTO { return pluginDTO{p.ID,p.Type,p.Version,p.APIVersion,p.Capabilities,p.Status} }
func list[A,B any](items []A,mapItem func(A)B) []B { out:=make([]B,0,len(items)); for _,a:=range items { out=append(out,mapItem(a)) }; return out }
