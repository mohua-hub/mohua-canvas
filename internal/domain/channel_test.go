package domain

import "testing"

func TestChannelRejectsPlaintextNonLoopbackEndpoints(t *testing.T) {
	base:=Channel{ID:"00000000-0000-4000-8000-000000000001",Name:"渠道",PluginID:"provider.openai",Models:[]Model{{Name:"model",Kind:Text}},CredentialRef:"00000000-0000-4000-8000-000000000002"}
	for _,endpoint:=range []string{"http://example.com/v1","http://10.0.0.2:8080/v1"}{base.BaseURL=endpoint;if Code(base.Validate())!="INVALID_INPUT"{t.Fatalf("accepted %s",endpoint)}}
	for _,endpoint:=range []string{"http://127.0.0.1:8080/v1","http://[::1]:8080/v1","https://example.com/v1"}{base.BaseURL=endpoint;if err:=base.Validate();err!=nil{t.Fatalf("rejected %s: %v",endpoint,err)}}
}
