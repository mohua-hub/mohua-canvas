package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"time"
	"github.com/mohua-hub/mohua-canvas/internal/domain"
)
const MaxMedia = 64 << 20
func New() *http.Client {
	transport:=http.DefaultTransport.(*http.Transport).Clone()
	return &http.Client{Transport:transport,Timeout:3*time.Minute,CheckRedirect:func(*http.Request,[]*http.Request) error { return http.ErrUseLastResponse }}
}
func Request(ctx context.Context,client *http.Client,method,target string,headers map[string]string,body []byte,limit int64) ([]byte,error) {
	req,err:=http.NewRequestWithContext(ctx,method,target,bytes.NewReader(body)); if err!=nil { return nil,domain.Fail("INVALID_INPUT") }
	for k,v:=range headers { req.Header.Set(k,v) }
	res,err:=client.Do(req); if err!=nil {
		if ctx.Err()!=nil { return nil,ctx.Err() }
		var nerr net.Error; if e,ok:=err.(net.Error); ok { nerr=e }; if nerr!=nil && nerr.Timeout() { return nil,domain.Fail("TIMEOUT") }
		return nil,domain.Fail("UPSTREAM_FAILED")
	}
	defer res.Body.Close()
	if res.StatusCode<200 || res.StatusCode>=300 {
		switch res.StatusCode { case 401,403:return nil,domain.Fail("UPSTREAM_AUTH"); case 429:return nil,domain.Fail("RATE_LIMITED"); case 408,504:return nil,domain.Fail("TIMEOUT") }
		return nil,domain.Fail("UPSTREAM_FAILED")
	}
	data,err:=io.ReadAll(io.LimitReader(res.Body,limit+1)); if err!=nil || int64(len(data))>limit { return nil,domain.Fail("UPSTREAM_FAILED") }; return data,nil
}
func JSON(ctx context.Context,client *http.Client,method,target string,headers map[string]string,input any,output any) error {
	var body []byte
	if input!=nil { var err error; body,err=json.Marshal(input); if err!=nil { return domain.Fail("INVALID_INPUT") } }
	h:=make(map[string]string,len(headers)+1); for k,v:=range headers { h[k]=v }; h["Content-Type"]="application/json"
	data,err:=Request(ctx,client,method,target,h,body,4<<20); if err!=nil { return err }
	if err=json.Unmarshal(data,output); err!=nil { return domain.Fail("UPSTREAM_FAILED") }; return nil
}
var reserved = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),netip.MustParsePrefix("10.0.0.0/8"),netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("127.0.0.0/8"),netip.MustParsePrefix("169.254.0.0/16"),netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.0.0.0/24"),netip.MustParsePrefix("192.0.2.0/24"),netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("198.18.0.0/15"),netip.MustParsePrefix("198.51.100.0/24"),netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("224.0.0.0/4"),netip.MustParsePrefix("240.0.0.0/4"),netip.MustParsePrefix("::/96"),
	netip.MustParsePrefix("64:ff9b::/96"),netip.MustParsePrefix("100::/64"),netip.MustParsePrefix("2001::/32"),
	netip.MustParsePrefix("2001:db8::/32"),netip.MustParsePrefix("2002::/16"),netip.MustParsePrefix("fc00::/7"),netip.MustParsePrefix("fe80::/10"),
}
func PublicIP(ip netip.Addr) bool {
	ip=ip.Unmap(); if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() { return false }
	for _,p:=range reserved { if p.Contains(ip) { return false } }; return true
}
func Download(ctx context.Context,target string) ([]byte,error) {
	u,err:=url.Parse(target); if err!=nil || u.Scheme!="https" || u.Host=="" || u.User!=nil { return nil,domain.Fail("UPSTREAM_FAILED") }
	tr:=http.DefaultTransport.(*http.Transport).Clone(); tr.Proxy=nil
	tr.DialContext=func(ctx context.Context,network,address string) (net.Conn,error) {
		host,port,err:=net.SplitHostPort(address); if err!=nil { return nil,err }
		ips,err:=net.DefaultResolver.LookupNetIP(ctx,"ip",host); if err!=nil || len(ips)==0 { return nil,domain.Fail("UPSTREAM_FAILED") }
		for _,ip:=range ips { if !PublicIP(ip) { return nil,domain.Fail("UPSTREAM_FAILED") } }
		dialer:=net.Dialer{Timeout:20*time.Second}
		return dialer.DialContext(ctx,network,net.JoinHostPort(ips[0].String(),port))
	}
	defer tr.CloseIdleConnections()
	client:=&http.Client{Transport:tr,Timeout:2*time.Minute,CheckRedirect:func(*http.Request,[]*http.Request) error{return http.ErrUseLastResponse}}
	return Request(ctx,client,http.MethodGet,target,nil,nil,MaxMedia)
}
