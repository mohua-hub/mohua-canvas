package gemini

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"github.com/mohua-hub/mohua-canvas/internal/application/ports"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
	"github.com/mohua-hub/mohua-canvas/internal/platform/httpclient"
)
type Provider struct { vault ports.CredentialVault; client *http.Client }
func New(v ports.CredentialVault) *Provider { return &Provider{v,httpclient.New()} }
func (p *Provider) ID() string { return "provider.gemini" }
func (p *Provider) Capabilities() []domain.Kind { return []domain.Kind{domain.Text,domain.Image} }
func (p *Provider) Close() { p.client.CloseIdleConnections() }
type part struct {
	Text string `json:"text,omitempty"`
	Inline *inlineData `json:"inlineData,omitempty"`
}
type inlineData struct { MIMEType string `json:"mimeType"`; Data string `json:"data"` }
type content struct { Parts []part `json:"parts"` }
type request struct {
	Contents []content `json:"contents"`
	Config *generationConfig `json:"generationConfig,omitempty"`
}
type generationConfig struct { Modalities []string `json:"responseModalities"` }
type response struct { Candidates []struct { Content content `json:"content"` } `json:"candidates"` }
func (p *Provider) Generate(ctx context.Context,in domain.GenerationRequest,c domain.Channel) (domain.GenerationResult,error) {
	if in.Kind!=domain.Text && in.Kind!=domain.Image { return domain.GenerationResult{},domain.Fail("UNSUPPORTED") }
	key,err:=p.vault.GetCredential(ctx,c.CredentialRef); if err!=nil { return domain.GenerationResult{},err }
	req:=request{Contents:[]content{{Parts:[]part{{Text:in.Prompt}}}}}
	if in.Kind==domain.Image { req.Config=&generationConfig{Modalities:[]string{"TEXT","IMAGE"}} }
	target:=strings.TrimRight(c.BaseURL,"/")+"/models/"+url.PathEscape(strings.TrimPrefix(in.Model,"models/"))+":generateContent"
	var out response
	// Image responses may contain base64 media; use the media-sized bound.
	raw,err:=json.Marshal(req); if err!=nil { return domain.GenerationResult{},err }
	data,err:=httpclient.Request(ctx,p.client,"POST",target,map[string]string{"x-goog-api-key":key,"Content-Type":"application/json"},raw,90<<20)
	if err!=nil { return domain.GenerationResult{},err }
	if err=json.Unmarshal(data,&out); err!=nil || len(out.Candidates)==0 { return domain.GenerationResult{},domain.Fail("UPSTREAM_FAILED") }
	result:=domain.GenerationResult{}
	for _,part:=range out.Candidates[0].Content.Parts {
		if in.Kind==domain.Text { result.Text+=part.Text }
		if in.Kind==domain.Image && part.Inline!=nil {
			if !strings.HasPrefix(part.Inline.MIMEType,"image/") { continue }
			result.Media,err=base64.StdEncoding.DecodeString(part.Inline.Data); if err!=nil { return domain.GenerationResult{},domain.Fail("UPSTREAM_FAILED") }
			result.MediaName="生成图片"; return result,nil
		}
	}
	if in.Kind==domain.Text && result.Text!="" { return result,nil }; return result,domain.Fail("UPSTREAM_FAILED")
}
