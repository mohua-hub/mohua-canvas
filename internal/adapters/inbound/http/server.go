package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"github.com/mohua-hub/mohua-canvas/internal/application/assets"
	"github.com/mohua-hub/mohua-canvas/internal/application/channels"
	"github.com/mohua-hub/mohua-canvas/internal/application/generation"
	"github.com/mohua-hub/mohua-canvas/internal/application/library"
	"github.com/mohua-hub/mohua-canvas/internal/application/plugins"
	"github.com/mohua-hub/mohua-canvas/internal/application/projects"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)
type Services struct { Projects *projects.Service; Channels *channels.Service; Assets *assets.Service; Jobs *generation.Service; Library *library.Service; Plugins []plugins.View }
type errorDTO struct { Code string `json:"code"`; Message string `json:"message"` }
type envelope struct { OK bool `json:"ok"`; Data any `json:"data,omitempty"`; Error *errorDTO `json:"error,omitempty"` }
var messages=map[string]string{
	"INVALID_INPUT":"输入内容无效","UNAUTHORIZED":"会话未授权","NOT_FOUND":"内容不存在",
	"CONFLICT":"内容已改变或仍在使用，请重新确认","UNSUPPORTED":"该能力暂不支持","BUSY":"任务队列已满",
	"RATE_LIMITED":"模型服务限流","UPSTREAM_AUTH":"渠道凭据被模型服务拒绝","UPSTREAM_FAILED":"模型服务未能完成请求",
	"TIMEOUT":"请求超时","INTERNAL":"操作未能完成","INTERRUPTED":"任务因服务重启中断",
}
func respond(w http.ResponseWriter,data any,err error) {
	w.Header().Set("Content-Type","application/json; charset=utf-8"); w.Header().Set("Cache-Control","no-store")
	if err==nil { _=json.NewEncoder(w).Encode(struct{OK bool `json:"ok"`;Data any `json:"data"`}{true,data}); return }
	code:=domain.Code(err); status:=500
	switch code { case "INVALID_INPUT":status=400;case "UNAUTHORIZED":status=401;case "NOT_FOUND":status=404;case "CONFLICT":status=409;case "UNSUPPORTED":status=422;case "BUSY","RATE_LIMITED":status=429;case "UPSTREAM_AUTH","UPSTREAM_FAILED":status=502;case "TIMEOUT":status=504 }
	message,ok:=messages[code]; if !ok { code,message="INTERNAL",messages["INTERNAL"] }
	w.WriteHeader(status); _=json.NewEncoder(w).Encode(envelope{OK:false,Error:&errorDTO{code,message}})
}
func decode(w http.ResponseWriter,r *http.Request,out any) error {
	r.Body=http.MaxBytesReader(w,r.Body,8<<20)
	decoder:=json.NewDecoder(r.Body); decoder.DisallowUnknownFields()
	if decoder.Decode(out)!=nil || decoder.Decode(&struct{}{})!=io.EOF { return domain.Fail("INVALID_INPUT") }; return nil
}
func handle(fn func(http.ResponseWriter,*http.Request)(any,error)) http.HandlerFunc {
	return func(w http.ResponseWriter,r *http.Request){ data,err:=fn(w,r); respond(w,data,err) }
}
func New(s Services,token string) http.Handler {
	mux:=http.NewServeMux()
	mux.HandleFunc("GET /v1/health",handle(func(w http.ResponseWriter,r *http.Request)(any,error){return struct{Version int `json:"version"`}{1},nil}))
	mux.HandleFunc("GET /v1/projects",handle(func(w http.ResponseWriter,r *http.Request)(any,error){items,err:=s.Projects.List(r.Context());return list(items,summary),err}))
	mux.HandleFunc("POST /v1/projects",handle(func(w http.ResponseWriter,r *http.Request)(any,error){
		var in struct{Name string `json:"name"`}; if err:=decode(w,r,&in);err!=nil{return nil,err};p,err:=s.Projects.Create(r.Context(),in.Name);return project(p),err
	}))
	mux.HandleFunc("GET /v1/projects/{id}",handle(func(w http.ResponseWriter,r *http.Request)(any,error){p,err:=s.Projects.Open(r.Context(),r.PathValue("id"));return project(p),err}))
	mux.HandleFunc("POST /v1/projects/{id}/save",handle(func(w http.ResponseWriter,r *http.Request)(any,error){
		var in struct{Name string `json:"name"`;ExpectedRevision int64 `json:"expectedRevision"`;Document documentDTO `json:"document"`}
		if err:=decode(w,r,&in);err!=nil{return nil,err};p,err:=s.Projects.Save(r.Context(),projects.SaveInput{ID:r.PathValue("id"),Name:in.Name,ExpectedRevision:in.ExpectedRevision,Document:in.Document.domain()});return project(p),err
	}))
	mux.HandleFunc("DELETE /v1/projects/{id}",handle(func(w http.ResponseWriter,r *http.Request)(any,error){return nil,s.Projects.Delete(r.Context(),r.PathValue("id"))}))
	mux.HandleFunc("GET /v1/channels",handle(func(w http.ResponseWriter,r *http.Request)(any,error){items,err:=s.Channels.List(r.Context());return list(items,channel),err}))
	mux.HandleFunc("POST /v1/channels",handle(func(w http.ResponseWriter,r *http.Request)(any,error){
		var in struct{ID string `json:"id"`;Name string `json:"name"`;PluginID string `json:"pluginId"`;BaseURL string `json:"baseUrl"`;Models []modelDTO `json:"models"`;APIKey string `json:"apiKey"`}
		if err:=decode(w,r,&in);err!=nil{return nil,err};models:=[]domain.Model{};for _,m:=range in.Models{models=append(models,domain.Model{Name:m.Name,Kind:m.Kind})}
		c,err:=s.Channels.Configure(r.Context(),channels.ConfigureInput{ID:in.ID,Name:in.Name,PluginID:in.PluginID,BaseURL:in.BaseURL,Models:models,APIKey:in.APIKey});in.APIKey="";return channel(c),err
	}))
	mux.HandleFunc("DELETE /v1/channels/{id}",handle(func(w http.ResponseWriter,r *http.Request)(any,error){return nil,s.Channels.Delete(r.Context(),r.PathValue("id"))}))
	mux.HandleFunc("GET /v1/plugins",handle(func(w http.ResponseWriter,r *http.Request)(any,error){return list(s.Plugins,plugin),nil}))
	mux.HandleFunc("GET /v1/assets",handle(func(w http.ResponseWriter,r *http.Request)(any,error){items,err:=s.Assets.List(r.Context());return list(items,asset),err}))
	mux.HandleFunc("POST /v1/assets/import",handle(func(w http.ResponseWriter,r *http.Request)(any,error){
		r.Body=http.MaxBytesReader(w,r.Body,assets.MaxBytes);data,err:=io.ReadAll(r.Body);if err!=nil{return nil,domain.Fail("INVALID_INPUT")}
		name,err:=url.PathUnescape(r.Header.Get("X-Asset-Name"));if err!=nil{return nil,domain.Fail("INVALID_INPUT")}
		a,err:=s.Assets.Import(r.Context(),name,data);return asset(a),err
	}))
	mux.HandleFunc("GET /v1/assets/{id}/content",func(w http.ResponseWriter,r *http.Request){
		a,f,err:=s.Assets.Open(r.Context(),r.PathValue("id"));if err!=nil{respond(w,nil,err);return};defer f.Close()
		w.Header().Set("Content-Type",a.MIMEType);w.Header().Set("X-Content-Type-Options","nosniff")
		t,_:=time.Parse(time.RFC3339Nano,a.CreatedAt);http.ServeContent(w,r,a.Name,t,f)
	})
	mux.HandleFunc("DELETE /v1/assets/{id}",handle(func(w http.ResponseWriter,r *http.Request)(any,error){return nil,s.Assets.Delete(r.Context(),r.PathValue("id"))}))
	mux.HandleFunc("POST /v1/jobs",handle(func(w http.ResponseWriter,r *http.Request)(any,error){
		var in struct{RequestID string `json:"requestId"`;ProjectID string `json:"projectId"`;NodeID string `json:"nodeId"`;ChannelID string `json:"channelId"`;Model string `json:"model"`;Kind domain.Kind `json:"kind"`;Prompt string `json:"prompt"`}
		if err:=decode(w,r,&in);err!=nil{return nil,err}
		j,err:=s.Jobs.Submit(r.Context(),generation.SubmitInput{RequestID:in.RequestID,ProjectID:in.ProjectID,NodeID:in.NodeID,ChannelID:in.ChannelID,Model:in.Model,Kind:in.Kind,Prompt:in.Prompt});return job(j),err
	}))
	mux.HandleFunc("GET /v1/jobs",handle(func(w http.ResponseWriter,r *http.Request)(any,error){items,err:=s.Jobs.List(r.Context(),r.URL.Query().Get("projectId"));return list(items,job),err}))
	mux.HandleFunc("GET /v1/jobs/{id}",handle(func(w http.ResponseWriter,r *http.Request)(any,error){j,err:=s.Jobs.Get(r.Context(),r.PathValue("id"));return job(j),err}))
	mux.HandleFunc("POST /v1/jobs/{id}/cancel",handle(func(w http.ResponseWriter,r *http.Request)(any,error){
		var in struct{};if err:=decode(w,r,&in);err!=nil{return nil,err};j,err:=s.Jobs.Cancel(r.Context(),r.PathValue("id"));return job(j),err
	}))
	mux.HandleFunc("GET /v1/library",handle(func(w http.ResponseWriter,r *http.Request)(any,error){items,err:=s.Library.List(r.Context());return list(items,entry),err}))
	mux.HandleFunc("POST /v1/library",handle(func(w http.ResponseWriter,r *http.Request)(any,error){
		var in struct{ID string `json:"id"`;Kind string `json:"kind"`;Name string `json:"name"`;Content string `json:"content"`};if err:=decode(w,r,&in);err!=nil{return nil,err}
		e,err:=s.Library.Save(r.Context(),library.SaveInput{ID:in.ID,Kind:in.Kind,Name:in.Name,Content:in.Content});return entry(e),err
	}))
	mux.HandleFunc("DELETE /v1/library/{id}",handle(func(w http.ResponseWriter,r *http.Request)(any,error){return nil,s.Library.Delete(r.Context(),r.PathValue("id"))}))
	mux.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){respond(w,nil,domain.Fail("NOT_FOUND"))})
	return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
		defer func(){if recover()!=nil {respond(w,nil,domain.Fail("INTERNAL"))}}()
		authorization:=r.Header.Get("Authorization")
		provided:=strings.TrimPrefix(authorization,"Bearer ")
		if token=="" || !strings.HasPrefix(authorization,"Bearer ") || subtle.ConstantTimeCompare([]byte(provided),[]byte(token))!=1 {respond(w,nil,domain.Fail("UNAUTHORIZED"));return}
		mux.ServeHTTP(w,r)
	})
}
