package domain

import (
	"math"
	"regexp"
	"strings"
)

type Kind string
const (
	Text Kind = "text"
	Image Kind = "image"
	Video Kind = "video"
	Audio Kind = "audio"
)
func (k Kind) Valid() bool { return k == Text || k == Image || k == Video || k == Audio }
var idPattern = regexp.MustCompile("^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$")
func ValidID(id string) bool { return idPattern.MatchString(id) }
func ValidName(s string) bool { return strings.TrimSpace(s) != "" && len(s) <= 120 }
func ValidPosition(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) && math.Abs(v) <= 1e6 }

type Node struct {
	ID string
	Kind Kind
	X, Y float64
	Title, Content, Prompt string
	ChannelID, Model, AssetID, JobID string
}
type Edge struct { ID, Source, Target string }
type Viewport struct { X, Y, Zoom float64 }
type Document struct {
	Version int
	Nodes []Node
	Edges []Edge
	Viewport Viewport
}
type Project struct {
	ID, Name string
	Revision int64
	Document Document
	CreatedAt, UpdatedAt string
}
func EmptyDocument() Document {
	return Document{Version: 1, Nodes: []Node{}, Edges: []Edge{}, Viewport: Viewport{Zoom: 1}}
}
func (d Document) Validate() error {
	if d.Version != 1 || len(d.Nodes) > 2000 || len(d.Edges) > 4000 ||
		!ValidPosition(d.Viewport.X) || !ValidPosition(d.Viewport.Y) ||
		math.IsNaN(d.Viewport.Zoom) || d.Viewport.Zoom < 0.1 || d.Viewport.Zoom > 4 {
		return Fail("INVALID_INPUT")
	}
	nodes := make(map[string]bool, len(d.Nodes))
	for _, n := range d.Nodes {
		if !ValidID(n.ID) || nodes[n.ID] || !n.Kind.Valid() || !ValidName(n.Title) ||
			!ValidPosition(n.X) || !ValidPosition(n.Y) || len(n.Content) > 32000 || len(n.Prompt) > 32000 || len(n.Model) > 200 {
			return Fail("INVALID_INPUT")
		}
		for _, id := range []string{n.ChannelID, n.AssetID, n.JobID} {
			if id != "" && !ValidID(id) { return Fail("INVALID_INPUT") }
		}
		nodes[n.ID] = true
	}
	ids, pairs := map[string]bool{}, map[string]bool{}
	graph := map[string][]string{}
	degree := map[string]int{}
	for id := range nodes { degree[id] = 0 }
	for _, e := range d.Edges {
		pair := e.Source + "/" + e.Target
		if !ValidID(e.ID) || ids[e.ID] || !nodes[e.Source] || !nodes[e.Target] || e.Source == e.Target || pairs[pair] {
			return Fail("INVALID_INPUT")
		}
		ids[e.ID], pairs[pair] = true, true
		graph[e.Source] = append(graph[e.Source], e.Target)
		degree[e.Target]++
	}
	queue := []string{}
	for id, count := range degree { if count == 0 { queue = append(queue, id) } }
	visited := 0
	for len(queue) > 0 {
		id := queue[0]; queue = queue[1:]; visited++
		for _, next := range graph[id] { degree[next]--; if degree[next] == 0 { queue = append(queue, next) } }
	}
	if visited != len(nodes) { return Fail("INVALID_INPUT") }
	return nil
}
