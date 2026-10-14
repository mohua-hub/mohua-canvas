package bootstrap

import "testing"
func TestOnlyDeclaredFactoriesActivate(t *testing.T) {
	manifests,err:=LoadManifests("../../config/plugins.json");if err!=nil{t.Fatal(err)}
	position:=map[string]int{};for i,m:=range manifests{position[m.ID]=i}
	if position["storage.sqlite"]>=position["provider.openai"]{t.Fatal("dependency activation order")}
	cases:=map[string]func([]Manifest)[]Manifest{
		"unknown factory":func(ms []Manifest)[]Manifest{ms[0].Entry="arbitrary/file";return ms},
		"version":func(ms []Manifest)[]Manifest{ms[0].APIVersion="2";return ms},
		"duplicate":func(ms []Manifest)[]Manifest{return append(ms,ms[0])},
		"missing storage":func(ms []Manifest)[]Manifest{return ms[1:]},
		"undeclared dependency":func(ms []Manifest)[]Manifest{ms[len(ms)-1].Dependencies=[]string{"missing"};return ms},
	}
	for name,mutate:=range cases{t.Run(name,func(t *testing.T){copy:=append([]Manifest{},manifests...);if _,err:=OrderManifests(mutate(copy));err==nil{t.Fatal("invalid manifest accepted")}})}
}
