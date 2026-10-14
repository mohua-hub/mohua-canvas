package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
	"github.com/mohua-hub/mohua-canvas/internal/application/ports"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
	"github.com/mohua-hub/mohua-canvas/internal/platform/httpclient"
)
type Provider struct { vault ports.CredentialVault; client *http.Client }
func New(v ports.CredentialVault) *Provider { return &Provider{v,httpclient.New()} }
func (p *Provider) ID() string { return "provider.openai" }
func (p *Provider) Capabilities() []domain.Kind { return []domain.Kind{domain.Text,domain.Image,domain.Video,domain.Audio} }
func (p *Provider) Close() { p.client.CloseIdleConnections() }
type chatRequest struct { Model string `json:"model"`; Messages []message `json:"messages"` }
type message struct { Role string `json:"role"`; Content string `json:"content"` }
type chatResponse struct { Choices []struct { Message message `json:"message"` } `json:"choices"` }
type imageRequest struct { Model string `json:"model"`; Prompt string `json:"prompt"`; N int `json:"n"` }
type imageResponse struct { Data []struct { B64 string `json:"b64_json"`; URL string `json:"url"` } `json:"data"` }
type speechRequest struct { Model string `json:"model"`; Input string `json:"input"`; Voice string `json:"voice"`; Format string `json:"response_format"` }
type videoResponse struct { ID string `json:"id"`; Status string `json:"status"` }
func (p *Provider) Generate(ctx context.Context,in domain.GenerationRequest,c domain.Channel) (domain.GenerationResult,error) {
	key,err:=p.vault.GetCredential(ctx,c.CredentialRef); if err!=nil { return domain.GenerationResult{},err }
	headers:=map[string]string{"Authorization":"Bearer "+key}
	base:=strings.TrimRight(c.BaseURL,"/")
	switch in.Kind {
	case domain.Text:
		var out chatResponse
		err:=httpclient.JSON(ctx,p.client,"POST",base+"/chat/completions",headers,chatRequest{in.Model,[]message{{"user",in.Prompt}}},&out)
		if err!=nil { return domain.GenerationResult{},err }; if len(out.Choices)==0 { return domain.GenerationResult{},domain.Fail("UPSTREAM_FAILED") }
		return domain.GenerationResult{Text:out.Choices[0].Message.Content},nil
	case domain.Image:
		var out imageResponse
		raw,err:=json.Marshal(imageRequest{in.Model,in.Prompt,1});if err!=nil{return domain.GenerationResult{},err}
		imageHeaders:=map[string]string{"Authorization":headers["Authorization"],"Content-Type":"application/json"}
		response,err:=httpclient.Request(ctx,p.client,"POST",base+"/images/generations",imageHeaders,raw,90<<20)
		if err==nil && json.Unmarshal(response,&out)!=nil {err=domain.Fail("UPSTREAM_FAILED")}
		if err!=nil { return domain.GenerationResult{},err }; if len(out.Data)==0 { return domain.GenerationResult{},domain.Fail("UPSTREAM_FAILED") }
		var data []byte
		if out.Data[0].B64!="" { data,err=base64.StdEncoding.DecodeString(out.Data[0].B64) } else { data,err=httpclient.Download(ctx,out.Data[0].URL) }
		if err!=nil { return domain.GenerationResult{},domain.Fail("UPSTREAM_FAILED") }
		return domain.GenerationResult{Media:data,MediaName:"生成图片"},nil
	case domain.Audio:
		input:=speechRequest{in.Model,in.Prompt,"alloy","mp3"}
		var body bytes.Buffer
		// JSON encoding remains inside this provider's protocol boundary.
		if err:=json.NewEncoder(&body).Encode(input); err!=nil { return domain.GenerationResult{},err }
		headers["Content-Type"]="application/json"
		data,err:=httpclient.Request(ctx,p.client,"POST",base+"/audio/speech",headers,body.Bytes(),httpclient.MaxMedia)
		return domain.GenerationResult{Media:data,MediaName:"生成语音.mp3"},err
	case domain.Video:
		var body bytes.Buffer; writer:=multipart.NewWriter(&body)
		if err:=writer.WriteField("model",in.Model); err!=nil { return domain.GenerationResult{},err }
		if err:=writer.WriteField("prompt",in.Prompt); err!=nil { return domain.GenerationResult{},err }
		if err:=writer.Close(); err!=nil { return domain.GenerationResult{},err }
		headers["Content-Type"]=writer.FormDataContentType()
		data,err:=httpclient.Request(ctx,p.client,"POST",base+"/videos",headers,body.Bytes(),4<<20); if err!=nil { return domain.GenerationResult{},err }
		var video videoResponse; if err=json.Unmarshal(data,&video); err!=nil || video.ID=="" || len(video.ID)>200 { return domain.GenerationResult{},domain.Fail("UPSTREAM_FAILED") }
		taskURL:=base+"/videos/"+url.PathEscape(video.ID)
		for {
			switch video.Status {
			case "completed":
				data,err:=httpclient.Request(ctx,p.client,"GET",taskURL+"/content",headers,nil,httpclient.MaxMedia)
				return domain.GenerationResult{Media:data,MediaName:"生成视频.mp4"},err
			case "failed","cancelled": return domain.GenerationResult{},domain.Fail("UPSTREAM_FAILED")
			case "queued","in_progress","processing","": // Poll with cancellation.
			default: return domain.GenerationResult{},domain.Fail("UPSTREAM_FAILED")
			}
			timer:=time.NewTimer(3*time.Second)
			select { case <-ctx.Done(): timer.Stop(); return domain.GenerationResult{},ctx.Err(); case <-timer.C: }
			if err=httpclient.JSON(ctx,p.client,"GET",taskURL,headers,nil,&video); err!=nil { return domain.GenerationResult{},err }
		}
	}
	return domain.GenerationResult{},domain.Fail("UNSUPPORTED")
}
