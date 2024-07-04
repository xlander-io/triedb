package triedb

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/xlander-io/hash"
)

type makeNameFunc interface {
	makeName() string
}

type makeTableFunc interface {
	makeTable() string
}

type applyFullModeFunc interface {
	applyFullMode(fullMode bool)
}

type vizGraph struct {
	Root     *vizNode
	nextID   int
	fullMode bool
}

type vizNode struct {
	ID                int
	Path              string
	Value             string
	HashNode          string
	HashChildren      string
	HashValue         string
	RecoveredNode     bool
	RecoveredChildren bool
	RecoveredValue    bool
	Bytes             []byte
	Children          *vizNodes
	Dirty             bool
}

type vizNodes struct {
	ID        int
	Recovered bool
	Bytes     []byte
	Index     map[string]*vizNode
	Dirty     bool
}

func newVizGraphFromTrieDB(tdb *TrieDB, fullMode bool) *vizGraph {
	var vg vizGraph
	vg.fullMode = fullMode
	vg.fromTrieDB(tdb)
	vg.recursiveUpdateID(vg.Root)
	vg.recursiveApplyFullMode(vg.Root)
	return &vg
}

func shortenText(text string, startLen, endLen int, sep string) string {
	if len(text) > startLen+len(sep)+endLen {
		s := bytes.NewBufferString(text).Bytes()
		var buf []byte
		buf = append(buf, s[:startLen]...)
		buf = append(buf, sep...)
		buf = append(buf, s[len(s)-endLen:]...)
		return string(buf)
	}
	return text
}

func shortenBytes(text []byte, startLen, endLen int, sep []byte) []byte {
	if len(text) > 5+1+3 {
		s := text
		var buf []byte
		buf = append(buf, s[:startLen]...)
		buf = append(buf, sep...)
		buf = append(buf, s[len(s)-endLen:]...)
		return buf
	}
	return text
}

func (vg *vizGraph) fromTrieDB(tdb *TrieDB) {
	vg.Root = &vizNode{}
	vg.Root.fromTrieNode(tdb.root_node)
}

func (vn *vizNode) fromTrieNode(n *Node) {
	vn.Path = string(bytes.Clone(n.path))
	vn.Value = string(bytes.Clone(n.val))
	vn.Bytes = bytes.Clone(n.node_bytes)
	vn.RecoveredNode = n.node_hash_recovered
	vn.RecoveredChildren = n.child_nodes_hash_recovered
	vn.RecoveredValue = n.val_hash_recovered
	vn.Dirty = n.dirty

	if !hash.IsNilHash(n.node_hash) {
		vn.HashNode = hex.EncodeToString(n.node_hash.Bytes())
	}
	if !hash.IsNilHash(n.child_nodes_hash) {
		vn.HashChildren = hex.EncodeToString(n.child_nodes_hash.Bytes())
	}
	if !hash.IsNilHash(n.val_hash) {
		vn.HashValue = hex.EncodeToString(n.val_hash.Bytes())
	}

	if nil != n.child_nodes {
		vn.Children = &vizNodes{}
		vn.Children.fromTrieNodes(n.child_nodes)
	}
}

func (vns *vizNodes) fromTrieNodes(ns *Nodes) {
	vns.Index = make(map[string]*vizNode)
	vns.Bytes = bytes.Clone(ns.nodes_bytes)
	vns.Dirty = ns.dirty

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

func (vg *vizGraph) recursiveApplyFullMode(o applyFullModeFunc) {
	o.applyFullMode(vg.fullMode)

	if vn, ok := o.(*vizNode); ok {
		if nil != vn.Children {
			vn.Children.applyFullMode(vg.fullMode)
			for _, v := range vn.Children.Index {
				vg.recursiveApplyFullMode(v)
			}
		}
	}
}

func (vn *vizNode) makeName() string {
	return fmt.Sprintf("ND%d", vn.ID)
}

func (vns *vizNodes) makeName() string {
	return fmt.Sprintf("NS%d", vns.ID)
}

func (vn *vizNode) applyFullMode(fullMode bool) {
	if !fullMode {
		vn.HashNode = shortenText(vn.HashNode, 8, 5, "...")
		vn.HashChildren = shortenText(vn.HashChildren, 8, 5, "...")
		vn.HashValue = shortenText(vn.HashValue, 8, 5, "...")

		vn.Bytes = shortenBytes(vn.Bytes, 5, 0, []byte(nil))
	}
}

func (vns *vizNodes) applyFullMode(fullMode bool) {
	if !fullMode {
		vns.Bytes = shortenBytes(vns.Bytes, 5, 0, []byte(nil))
	}
}

func (vn *vizNode) makeTable() string {
	BORDER := func(n int) string { return fmt.Sprintf(`BORDER="%d"`, n) }
	CELLBORDER := func(n int) string { return fmt.Sprintf(`CELLBORDER="%d"`, n) }
	CELLSPACING := func(n int) string { return fmt.Sprintf(`CELLSPACING="%d"`, n) }
	CELLPADDING := func(n int) string { return fmt.Sprintf(`CELLPADDING="%d"`, n) }
	ALIGN := func(n string) string { return fmt.Sprintf(`ALIGN="%s"`, n) }
	FONT := func(text string) string { return fmt.Sprintf(`<FONT COLOR="gray40">%s</FONT>`, text) }
	COLOR := `COLOR="gray"`

	TR := func(style1, value1, style2, value2 string) string {
		return fmt.Sprintf(`<TR><TD %s>%s</TD><TD %s>%v</TD></TR>`, style1, value1, style2, value2)
	}

	path_ := TR(ALIGN("RIGHT"), FONT("path"), ALIGN("LEFT"), vn.Path)
	value := TR(ALIGN("RIGHT"), FONT("value"), ALIGN("LEFT"), vn.Value)

	hashNode := TR(ALIGN("RIGHT"), FONT("node hash"), ALIGN("LEFT"), vn.HashNode)
	hashChildren := TR(ALIGN("RIGHT"), FONT("children hash"), ALIGN("LEFT"), vn.HashChildren)
	hashValue := TR(ALIGN("RIGHT"), FONT("value hash"), ALIGN("LEFT"), vn.HashValue)

	recoveredNode := TR(ALIGN("RIGHT"), FONT("recovered node hash"), ALIGN("LEFT"), strconv.FormatBool(vn.RecoveredNode))
	recoveredChildren := TR(ALIGN("RIGHT"), FONT("recovered children hash"), ALIGN("LEFT"), strconv.FormatBool(vn.RecoveredChildren))
	recoveredValue := TR(ALIGN("RIGHT"), FONT("recovered value hash"), ALIGN("LEFT"), strconv.FormatBool(vn.RecoveredValue))

	bytes := TR(ALIGN("RIGHT"), FONT("node bytes"), ALIGN("LEFT"), fmt.Sprint(vn.Bytes))
	dirty := TR(ALIGN("RIGHT"), FONT("dirty"), ALIGN("LEFT"), strconv.FormatBool(vn.Dirty))

	STYLEs := strings.Join([]string{BORDER(0), CELLBORDER(1), CELLSPACING(0), CELLPADDING(1), COLOR}, " ")
	VALUEs := strings.Join([]string{
		path_,
		value,
		hashNode,
		recoveredNode,
		hashChildren,
		recoveredChildren,
		hashValue,
		recoveredValue,
		bytes,
		dirty}, "")
	return fmt.Sprintf(`<TABLE %v>%v</TABLE>`, STYLEs, VALUEs)
}

func (vns *vizNodes) makeTable() string {
	BORDER := func(n int) string { return fmt.Sprintf(`BORDER="%d"`, n) }
	CELLBORDER := func(n int) string { return fmt.Sprintf(`CELLBORDER="%d"`, n) }
	CELLSPACING := func(n int) string { return fmt.Sprintf(`CELLSPACING="%d"`, n) }
	CELLPADDING := func(n int) string { return fmt.Sprintf(`CELLPADDING="%d"`, n) }
	ALIGN := func(n string) string { return fmt.Sprintf(`ALIGN="%s"`, n) }
	FONT := func(text string) string { return fmt.Sprintf(`<FONT COLOR="gray40">%s</FONT>`, text) }
	COLOR := `COLOR="gray"`
	STYLE := `STYLE="rounded"`

	TR := func(style1, value1, style2, value2 string) string {
		return fmt.Sprintf(`<TR><TD %s>%s</TD><TD %s>%v</TD></TR>`, style1, value1, style2, value2)
	}

	ports := func() string {
		var b bytes.Buffer
		for k, v := range vns.Index {
			b.WriteString(fmt.Sprintf("<TD PORT=\"%s\">%s</TD>", k, v.makeTable()))
		}

		return fmt.Sprintf("<TR>%s</TR>", b.String())
	}

	bytes := TR(ALIGN("RIGHT"), FONT("nodes bytes"), ALIGN("LEFT"), fmt.Sprint(vns.Bytes))
	dirty := TR(ALIGN("RIGHT"), FONT("dirty"), ALIGN("LEFT"), strconv.FormatBool(vns.Dirty))
	styles := strings.Join([]string{BORDER(0), CELLBORDER(1), CELLSPACING(0), CELLPADDING(2), COLOR, STYLE}, " ")
	values := strings.Join([]string{bytes, dirty}, "")
	header := fmt.Sprintf(`<TR><TD COLSPAN="%d"><TABLE %v>%v</TABLE></TD></TR>`, len(vns.Index), styles, values)

	STYLEs := strings.Join([]string{BORDER(1), CELLBORDER(0), CELLSPACING(10), CELLPADDING(5), COLOR, STYLE}, " ")
	VALUEs := strings.Join([]string{header, ports()}, "")
	return fmt.Sprintf(`<TABLE %v>%v</TABLE>`, STYLEs, VALUEs)
}

func (tdb *TrieDB) WriteDot(b io.Writer, fullMode bool) {

	funcMaps := template.FuncMap{
		"Name":  func(o makeNameFunc) string { return o.makeName() },
		"Table": func(o makeTableFunc) string { return o.makeTable() },
		"Edge": func(vnS, vnD *vizNode, port string) string {
			return fmt.Sprintf(`[tailport="%s" color=gray]`, port)
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
		{{- if and .Children (eq .ID 0)}} {{- Name $O}} -> {{Name $O.Children}} [color=gray] {{end}}
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

	vg := newVizGraphFromTrieDB(tdb, fullMode)
	err := tmpl.Execute(b, &vg)
	if err != nil {
		panic(err)
	}
}

func (tdb *TrieDB) GenDotString(fullMode bool) string {
	var b bytes.Buffer
	tdb.WriteDot(&b, fullMode)
	return b.String()
}

func (tdb *TrieDB) GenDotFile(filepath string, fullMode bool) {
	dot, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if nil != err {
		panic(err)
	}
	defer dot.Close()
	tdb.WriteDot(dot, fullMode)
}
