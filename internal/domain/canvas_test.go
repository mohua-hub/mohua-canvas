package domain

import (
	"math"
	"strings"
	"testing"
)
const a="00000000-0000-4000-8000-000000000001"
const b="00000000-0000-4000-8000-000000000002"
const e1="00000000-0000-4000-8000-000000000003"
const e2="00000000-0000-4000-8000-000000000004"
func TestCanvasRejectsInvalidGraphsAndBoundaries(t *testing.T) {
	base:=EmptyDocument();base.Nodes=[]Node{{ID:a,Kind:Text,Title:"甲"},{ID:b,Kind:Image,Title:"乙"}}
	base.Edges=[]Edge{{ID:e1,Source:a,Target:b}}
	if err:=base.Validate();err!=nil{t.Fatal(err)}
	cases:=map[string]func(*Document){
		"cycle":func(d *Document){d.Edges=append(d.Edges,Edge{ID:e2,Source:b,Target:a})},
		"dangling":func(d *Document){d.Edges[0].Target=e2},
		"duplicate":func(d *Document){d.Edges=append(d.Edges,Edge{ID:e2,Source:a,Target:b})},
		"nonfinite":func(d *Document){d.Nodes[0].X=math.Inf(1)},
		"bytes":func(d *Document){d.Nodes[0].Title=strings.Repeat("墨",41)},
		"version":func(d *Document){d.Version=2},
	}
	for name,mutate:=range cases{t.Run(name,func(t *testing.T){d:=base;d.Nodes=append([]Node{},base.Nodes...);d.Edges=append([]Edge{},base.Edges...);mutate(&d);if Code(d.Validate())!="INVALID_INPUT"{t.Fatal("invalid canvas accepted")}})}
}
