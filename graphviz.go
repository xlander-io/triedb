package triedb

import (
	"bytes"
	"fmt"
	"text/template"
)

type NameFunc interface {
	makeName() string
}

type vizGraph struct {
	Root   *vizNode
	nextID int
}

type vizNode struct {
	ID       int
	Path     string
	Value    string
	Children *vizNodes
}

type vizNodes struct {
	ID    int
	Index map[string]*vizNode
}

func newVizGraphFromTrieDB(tdb *TrieDB) *vizGraph {
	var vg vizGraph
	vg.fromTrieDB(tdb)
	vg.recursiveUpdateID(vg.Root)
	return &vg
}

func (vg *vizGraph) fromTrieDB(tdb *TrieDB) {
	vg.Root = &vizNode{}
	vg.Root.fromTrieNode(tdb.root_node)
}

func (vn *vizNode) fromTrieNode(n *Node) {
	vn.Path = string(bytes.Clone(n.path))
	vn.Value = string(bytes.Clone(n.val))

	if nil != n.child_nodes {
		vn.Children = &vizNodes{}
		vn.Children.fromTrieNodes(n.child_nodes)
	}
}

func (vns *vizNodes) fromTrieNodes(ns *Nodes) {
	vns.Index = make(map[string]*vizNode)
	for k, n := range ns.path_index {
		var vn vizNode
		vn.fromTrieNode(n)
		vns.Index[string(k)] = &vn
	}
}

func (vg *vizGraph) recursiveUpdateID(vn *vizNode) {
	vn.ID = vg.nextID
	vg.nextID++
	if nil != vn.Children {
		vn.Children.ID = vg.nextID
		vg.nextID++
		for _, v := range vn.Children.Index {
			vg.recursiveUpdateID(v)
		}
	}
}

func (vn *vizNode) makeName() string {
	return fmt.Sprintf("ND%d", vn.ID)
}

func (vn *vizNodes) makeName() string {
	return fmt.Sprintf("NS%d", vn.ID)
}

func (tdb *TrieDB) GenDot() string {

	var b bytes.Buffer

	funcMaps := template.FuncMap{
		"Name":  func(nf NameFunc) string { return nf.makeName() },
		"Vattr": func(vn *vizNode) string { return fmt.Sprintf(`[label="path:'%s', val:'%s'"]`, vn.Path, vn.Value) },
		"Eattr": func(vnS, vnD *vizNode, port string) string {
			return fmt.Sprintf(`[port='%s' label="path='%s'"]`, port, vnS.Path)
		},
	}
	tmpl := template.New("trie")
	tmpl.Funcs(funcMaps)

	tmpl = template.Must(tmpl.Parse(`
	digraph G {
		{{template "vertices" .Root}}
		{{template "edges" .Root}}
	}

	{{- define "vertices"}}
		{{Name .}} {{Vattr .}}
		{{if .Children}}{{Name .Children}} [label=<>]{{end}}
		{{- if .Children}}
			{{- range $k,$v := .Children.Index}}
				{{- template "vertices" $v}}
			{{- end}}
		{{- end}}
	{{- end}}

	{{- define "edges"}}
		{{- $O := .}}
		{{if .Children}}
			{{- range $k,$v := .Children.Index}}
				{{- Name $O.Children}} -> {{Name $v}} {{Eattr $O $v $k}}
				{{- template "edges" $v}}
			{{- end}}
		{{- end}}
	{{- end}}
	`))

	vg := newVizGraphFromTrieDB(tdb)
	err := tmpl.Execute(&b, &vg)
	if err != nil {
		panic(err)
	}

	return b.String()
}
