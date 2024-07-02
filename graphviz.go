package triedb

import (
	"bytes"
	"fmt"
	"text/template"
)

type NameFunc interface {
	makeName() string
}

type LabelFunc interface {
	makeLabel() string
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

func (vns *vizNodes) makeName() string {
	return fmt.Sprintf("NS%d", vns.ID)
}

func (vn *vizNode) makeLabel() string {
	BORDER := func(n int) string { return fmt.Sprintf(`BORDER="%d"`, n) }
	CELLBORDER := func(n int) string { return fmt.Sprintf(`CELLBORDER="%d"`, n) }
	CELLSPACING := func(n int) string { return fmt.Sprintf(`CELLSPACING="%d"`, n) }
	CELLPADDING := func(n int) string { return fmt.Sprintf(`CELLPADDING="%d"`, n) }
	COLOR := `COLOR="gray"`
	title := fmt.Sprintf("<TR>%s<TD>%d</TD></TR>", "<TD>ID</TD>", vn.ID)
	path_ := fmt.Sprintf("<TR>%s<TD>%v</TD></TR>", "<TD>Path</TD>", vn.Path)
	value := fmt.Sprintf("<TR>%s<TD>%v</TD></TR>", "<TD>Value</TD>", vn.Value)
	return fmt.Sprintf(`label=<<TABLE %v %v %v %v %v>%v%v%v</TABLE>>`, BORDER(0), CELLBORDER(1), CELLSPACING(0), CELLPADDING(1), COLOR, title, path_, value)
}

func (vns *vizNodes) makeLabel() string {
	BORDER := func(n int) string { return fmt.Sprintf(`BORDER="%d"`, n) }
	CELLBORDER := func(n int) string { return fmt.Sprintf(`CELLBORDER="%d"`, n) }
	CELLSPACING := func(n int) string { return fmt.Sprintf(`CELLSPACING="%d"`, n) }
	CELLPADDING := func(n int) string { return fmt.Sprintf(`CELLPADDING="%d"`, n) }
	COLOR := `COLOR="gray"`
	var b bytes.Buffer
	for k, _ := range vns.Index {
		b.WriteString(fmt.Sprintf("<TD PORT=\"%s\">%s</TD>", k, k))
	}
	TR_id___ := fmt.Sprintf("<TR>%s<TD>%d</TD></TR>", "<TD>ID</TD>", vns.ID)
	TR_ports := fmt.Sprintf("<TR>%s</TR>", b.String())
	return fmt.Sprintf(`label=<<TABLE %v %v %v %v %v>%s%s</TABLE>>`, BORDER(0), CELLBORDER(1), CELLSPACING(0), CELLPADDING(1), COLOR, TR_id___, TR_ports)
}

func (tdb *TrieDB) GenDot() string {

	var b bytes.Buffer

	funcMaps := template.FuncMap{
		"Name":  func(o NameFunc) string { return o.makeName() },
		"Label": func(o LabelFunc) string { return o.makeLabel() },
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
		{{Name .}} [shape=box style=rounded {{Label .}}]
		{{if .Children}}{{Name .Children}} [shape=plain {{Label .Children}}]{{end}}
		{{- if .Children}}
			{{- range $k,$v := .Children.Index}}
				{{- template "vertices" $v}}
			{{- end}}
		{{- end}}
	{{- end}}

	{{- define "edges"}}
		{{- $O := .}}
		{{if .Children}} {{Name $O}} -> {{Name $O.Children}} {{end}}
		{{if .Children}}
			{{- range $k,$v := .Children.Index}}
				{{- Name $O.Children}} -> {{Name $v}} {{Edge $O $v $k}}
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
