package httpapi

import (
	"net/http/httptest"
	"strings"
	"testing"
)
func TestEveryRouteRequiresBearerSession(t *testing.T) {
	token:=strings.Repeat("a",64);handler:=New(Services{},token)
	for _,path:=range []string{"/v1/health","/v1/assets/00000000-0000-4000-8000-000000000001/content","/v1/projects"}{
		for _,authorization:=range []string{"",token,"Basic "+token,"Bearer wrong"}{
			r:=httptest.NewRequest("GET",path,nil);r.Header.Set("Authorization",authorization);w:=httptest.NewRecorder();handler.ServeHTTP(w,r)
			if w.Code!=401||strings.Contains(w.Body.String(),token){t.Fatalf("%s status=%d",path,w.Code)}
		}
	}
	r:=httptest.NewRequest("GET","/v1/health",nil);r.Header.Set("Authorization","Bearer "+token)
	w:=httptest.NewRecorder();handler.ServeHTTP(w,r);if w.Code!=200||!strings.Contains(w.Body.String(),"\"version\":1"){t.Fatal(w.Body.String())}
}
func TestHTTPRejectsExtraFieldsAndTrailingJSON(t *testing.T) {
	token:=strings.Repeat("b",64);handler:=New(Services{},token)
	for _,body:=range []string{`{"name":"项目","userId":"unexpected"}`,`{"name":"项目"}{}`,`null{}`}{
		r:=httptest.NewRequest("POST","/v1/projects",strings.NewReader(body));r.Header.Set("Authorization","Bearer "+token)
		w:=httptest.NewRecorder();handler.ServeHTTP(w,r)
		if w.Code!=400||!strings.Contains(w.Body.String(),"INVALID_INPUT"){t.Fatal(w.Body.String())}
	}
}
