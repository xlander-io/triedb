package triedb

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/xlander-io/hash"
)

type NameFunc interface {
	makeName() string
}

type TableFunc interface {
	makeTable() string
}

type vizGraph struct {
	Root   *vizNode
	nextID int
}

type vizNode struct {
	ID       int
	Path     string
	Value    string
	Hash     string
	Dirty    bool
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
	if !hash.IsNilHash(n.node_hash) {
		vn.Hash = hex.EncodeToString(n.node_hash.Bytes())
	}
	vn.Dirty = n.dirty

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

func (vns *vizNodes) makeName() string {
	return fmt.Sprintf("NS%d", vns.ID)
}

func (vn *vizNode) makeTable() string {
	BORDER := func(n int) string { return fmt.Sprintf(`BORDER="%d"`, n) }
	CELLBORDER := func(n int) string { return fmt.Sprintf(`CELLBORDER="%d"`, n) }
	CELLSPACING := func(n int) string { return fmt.Sprintf(`CELLSPACING="%d"`, n) }
	CELLPADDING := func(n int) string { return fmt.Sprintf(`CELLPADDING="%d"`, n) }
	COLOR := `COLOR="gray"`
	//title := fmt.Sprintf("<TR>%s<TD>%d</TD></TR>", "<TD>ID</TD>", vn.ID)
	path_ := fmt.Sprintf("<TR>%s<TD>%v</TD></TR>", "<TD>path</TD>", vn.Path)
	value := fmt.Sprintf("<TR>%s<TD>%v</TD></TR>", "<TD>value</TD>", vn.Value)
	hash_ := fmt.Sprintf("<TR>%s<TD>%v</TD></TR>", "<TD>hash</TD>", vn.Hash)
	dirty := fmt.Sprintf("<TR>%s<TD>%v</TD></TR>", "<TD>dirty</TD>", vn.Dirty)

	STYLEs := fmt.Sprintf("%v %v %v %v %v", BORDER(0), CELLBORDER(1), CELLSPACING(0), CELLPADDING(1), COLOR)
	VALUEs := fmt.Sprintf("%v%v%v%v", path_, value, hash_, dirty)
	return fmt.Sprintf(`<TABLE %v>%v</TABLE>`, STYLEs, VALUEs)
}

func (vns *vizNodes) makeTable() string {
	BORDER := func(n int) string { return fmt.Sprintf(`BORDER="%d"`, n) }
	CELLBORDER := func(n int) string { return fmt.Sprintf(`CELLBORDER="%d"`, n) }
	CELLSPACING := func(n int) string { return fmt.Sprintf(`CELLSPACING="%d"`, n) }
	CELLPADDING := func(n int) string { return fmt.Sprintf(`CELLPADDING="%d"`, n) }
	COLOR := `COLOR="gray"`
	STYLE := `STYLE="rounded"`
	var b bytes.Buffer
	for k, v := range vns.Index {
		b.WriteString(fmt.Sprintf("<TD PORT=\"%s\">%s</TD>", k, v.makeTable()))
	}
	//id___ := fmt.Sprintf("<TR>%s<TD>%d</TD></TR>", "<TD>ID</TD>", vns.ID)
	ports := fmt.Sprintf("<TR>%s</TR>", b.String())

	STYLEs := fmt.Sprintf("%v %v %v %v %v %v", BORDER(1), CELLBORDER(0), CELLSPACING(15), CELLPADDING(0), COLOR, STYLE)
	VALUEs := fmt.Sprintf("%v", ports)
	return fmt.Sprintf(`<TABLE %v>%v</TABLE>`, STYLEs, VALUEs)
}

func (tdb *TrieDB) WriteDot(b io.Writer) {

	funcMaps := template.FuncMap{
		"Name":  func(o NameFunc) string { return o.makeName() },
		"Table": func(o TableFunc) string { return o.makeTable() },
		"Edge": func(vnS, vnD *vizNode, port string) string {
			return fmt.Sprintf(`[tailport="%s"]`, port)
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
		{{if eq .ID 0}}{{Name .}} [shape=plain label=<{{Table .}}>]{{end}}
		{{if .Children}}{{Name .Children}} [shape=plain label=<{{Table .Children}}>]{{end}}
		{{- if .Children}}
			{{- range $k,$v := .Children.Index}}
				{{- template "vertices" $v}}
			{{- end}}
		{{- end}}
	{{- end}}

	{{- define "edges"}}
		{{- $O := .}}
		{{- if and .Children (eq .ID 0)}} {{- Name $O}} -> {{Name $O.Children}} {{end}}
		{{- if .Children}}
			{{- range $k,$v := .Children.Index}}
				{{- if $v.Children}}
					{{- Name $O.Children}} -> {{Name $v.Children}} {{Edge $O $v $k}}
				{{- end}}
				{{- template "edges" $v}}
			{{- end}}
		{{- end}}
	{{- end}}
	`))

	vg := newVizGraphFromTrieDB(tdb)
	err := tmpl.Execute(b, &vg)
	if err != nil {
		panic(err)
	}
}

func (tdb *TrieDB) GenDotString() string {
	var b bytes.Buffer
	tdb.WriteDot(&b)
	return b.String()
}

func (tdb *TrieDB) GenDotFile(filepath string) {
	dot, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if nil != err {
		panic(err)
	}
	defer dot.Close()
	tdb.WriteDot(dot)
}
