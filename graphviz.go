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

const _KIND_PREFIX = "prefix"
const _KIND_FOLDER = "folder"

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
	ID       int
	Prefix   string
	Value    string
	PathFlat string

	HashIndex string
	HashNode  string

	HashChildrenPrefix      string
	HashChildrenFolder      string
	HashValue               string
	RecoveredChildrenPrefix bool
	RecoveredChildrenFolder bool
	RecoveredValue          bool

	Bytes []byte
	Dirty bool

	ChildrenPrefix *vizNodes
	ChildrenFolder *vizNodes
}

type vizNodes struct {
	ID        int
	Recovered bool
	Bytes     []byte
	Index     []*vizIndex
	Dirty     bool
	Kind      string // prefix or folder

	ParentNode *vizNode
}

type vizIndex struct {
	Key  string
	Node *vizNode
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

// HTML label of graphviz does not distinguish lower and uppper case port name
// so we should distinguish them by ourself: hash it
func makePortName(portName string) string {
	b := bytes.NewBufferString(portName)
	x := hash.CalHash(b.Bytes())
	return x.Hex()
}

func (vg *vizGraph) fromTrieDB(tdb *TrieDB) {
	vg.Root = &vizNode{}
	vg.Root.fromTrieNode(tdb.root_node)
}

func (vn *vizNode) fromTrieNode(n *Node) {
	vn.Prefix = string(bytes.Clone(n.prefix))
	vn.PathFlat = n.node_path_flat_str()
	if nil == n.val {
		vn.Value = "&lt;nil&gt;"
	} else {
		vn.Value = string(bytes.Clone(n.val))
	}
	vn.Bytes = bytes.Clone(n.node_bytes)

	vn.RecoveredChildrenPrefix = n.prefix_child_nodes_hash_recovered
	vn.RecoveredChildrenFolder = n.folder_child_nodes_hash_recovered
	vn.RecoveredValue = n.val_hash_recovered
	vn.Dirty = n.dirty

	if !hash.IsNilHash(n.index_hash) {
		vn.HashIndex = hex.EncodeToString(n.index_hash.Bytes())
	}
	if !hash.IsNilHash(n.node_hash) {
		vn.HashNode = hex.EncodeToString(n.node_hash.Bytes())
	}
	if !hash.IsNilHash(n.prefix_child_nodes_hash) {
		vn.HashChildrenPrefix = hex.EncodeToString(n.prefix_child_nodes_hash.Bytes())
	}
	if !hash.IsNilHash(n.folder_child_nodes_hash) {
		vn.HashChildrenFolder = hex.EncodeToString(n.folder_child_nodes_hash.Bytes())
	}
	if !hash.IsNilHash(n.val_hash) {
		vn.HashValue = hex.EncodeToString(n.val_hash.Bytes())
	}

	if nil != n.prefix_child_nodes {
		vn.ChildrenPrefix = &vizNodes{ParentNode: vn}
		vn.ChildrenPrefix.fromTrieNodes(n.prefix_child_nodes)
	}

	if nil != n.folder_child_nodes {
		vn.ChildrenFolder = &vizNodes{ParentNode: vn}
		vn.ChildrenFolder.fromTrieNodes(n.folder_child_nodes)
	}
}

func (vns *vizNodes) fromTrieNodes(ns *nodes) {

	vns.Bytes = bytes.Clone(ns.nodes_bytes)
	vns.Dirty = vns.Dirty || ns.dirty

	if ns.is_folder_child_nodes {
		vns.Kind = _KIND_FOLDER
	} else {
		vns.Kind = _KIND_PREFIX
	}

	iter := ns.btree.Before(uint8(0))
	for iter.Next() {
		n := iter.Value.(*Node)
		var vn vizNode
		vn.fromTrieNode(n)
		vi := vizIndex{Key: string(n.prefix), Node: &vn}
		vns.Index = append(vns.Index, &vi)
	}
}

func (vg *vizGraph) recursiveUpdateID(vn *vizNode) {
	vn.ID = vg.nextID
	vg.nextID++
	if nil != vn.ChildrenPrefix {
		vn.ChildrenPrefix.ID = vg.nextID
		vg.nextID++
		for _, vi := range vn.ChildrenPrefix.Index {
			vg.recursiveUpdateID(vi.Node)
		}
	}
	if nil != vn.ChildrenFolder {
		vn.ChildrenFolder.ID = vg.nextID
		vg.nextID++
		for _, vi := range vn.ChildrenFolder.Index {
			vg.recursiveUpdateID(vi.Node)
		}
	}
}

func (vg *vizGraph) recursiveApplyFullMode(o applyFullModeFunc) {
	o.applyFullMode(vg.fullMode)

	if vn, ok := o.(*vizNode); ok {
		if nil != vn.ChildrenPrefix {
			vn.ChildrenPrefix.applyFullMode(vg.fullMode)
			for _, vi := range vn.ChildrenPrefix.Index {
				vg.recursiveApplyFullMode(vi.Node)
			}
		}
		if nil != vn.ChildrenFolder {
			vn.ChildrenFolder.applyFullMode(vg.fullMode)
			for _, vi := range vn.ChildrenFolder.Index {
				vg.recursiveApplyFullMode(vi.Node)
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
		vn.HashIndex = shortenText(vn.HashIndex, 8, 5, "...")
		vn.HashChildrenPrefix = shortenText(vn.HashChildrenPrefix, 8, 5, "...")
		vn.HashChildrenFolder = shortenText(vn.HashChildrenFolder, 8, 5, "...")
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
	COLOR := func() string {
		if vn.Dirty {
			return `COLOR="red"`
		} else {
			return `COLOR="gray"`
		}
	}

	TR := func(style1, value1, style2, value2 string) string {
		return fmt.Sprintf(`<TR><TD %s>%s</TD><TD %s>%v</TD></TR>`, style1, value1, style2, value2)
	}

	pathFlat := TR(ALIGN("RIGHT"), FONT("flat path"), ALIGN("LEFT"), vn.PathFlat)
	prefix := TR(ALIGN("RIGHT"), FONT("prefix"), ALIGN("LEFT"), vn.Prefix)
	value := TR(ALIGN("RIGHT"), FONT("value"), ALIGN("LEFT"), vn.Value)

	hashNode := TR(ALIGN("RIGHT"), FONT("node hash"), ALIGN("LEFT"), vn.HashNode)
	hashIndex := TR(ALIGN("RIGHT"), FONT("index hash"), ALIGN("LEFT"), vn.HashIndex)
	hashChildrenPrefix := TR(ALIGN("RIGHT"), FONT("prefix children hash"), ALIGN("LEFT"), vn.HashChildrenPrefix)
	hashChildrenFolder := TR(ALIGN("RIGHT"), FONT("folder children hash"), ALIGN("LEFT"), vn.HashChildrenFolder)
	hashValue := TR(ALIGN("RIGHT"), FONT("value hash"), ALIGN("LEFT"), vn.HashValue)

	recoveredChildrenPrefix := TR(ALIGN("RIGHT"), FONT("recovered prefix children hash"), ALIGN("LEFT"), strconv.FormatBool(vn.RecoveredChildrenPrefix))
	recoveredChildrenFolder := TR(ALIGN("RIGHT"), FONT("recovered folder children hash"), ALIGN("LEFT"), strconv.FormatBool(vn.RecoveredChildrenFolder))
	recoveredValue := TR(ALIGN("RIGHT"), FONT("recovered value hash"), ALIGN("LEFT"), strconv.FormatBool(vn.RecoveredValue))

	bytes := TR(ALIGN("RIGHT"), FONT("node bytes"), ALIGN("LEFT"), fmt.Sprint(vn.Bytes))
	dirty := TR(ALIGN("RIGHT"), FONT("dirty"), ALIGN("LEFT"), strconv.FormatBool(vn.Dirty))

	STYLEs := strings.Join([]string{BORDER(0), CELLBORDER(1), CELLSPACING(0), CELLPADDING(1), COLOR()}, " ")
	VALUEs := strings.Join([]string{
		pathFlat,
		prefix,
		value,
		hashNode,
		hashIndex,
		hashChildrenPrefix,
		recoveredChildrenPrefix,
		hashChildrenFolder,
		recoveredChildrenFolder,
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
	COLOR := func() string {
		if vns.Dirty {
			return `COLOR="red"`
		} else {
			return `COLOR="gray"`
		}
	}
	STYLE := `STYLE="rounded"`

	TR := func(style1, value1, style2, value2 string) string {
		return fmt.Sprintf(`<TR><TD %s>%s</TD><TD %s>%v</TD></TR>`, style1, value1, style2, value2)
	}

	ports := func() string {
		var b bytes.Buffer
		for _, vi := range vns.Index {
			b.WriteString(fmt.Sprintf("<TD PORT=\"%s\">%s</TD>", makePortName(vi.Key), vi.Node.makeTable()))
		}

		return fmt.Sprintf("<TR>%s</TR>", b.String())
	}

	kind := TR(ALIGN("RIGHT"), FONT("kind"), ALIGN("LEFT"), fmt.Sprint(vns.Kind))
	bytes := TR(ALIGN("RIGHT"), FONT("nodes bytes"), ALIGN("LEFT"), fmt.Sprint(vns.Bytes))
	dirty := TR(ALIGN("RIGHT"), FONT("dirty"), ALIGN("LEFT"), strconv.FormatBool(vns.Dirty))
	styles := strings.Join([]string{BORDER(0), CELLBORDER(1), CELLSPACING(0), CELLPADDING(2), COLOR(), STYLE}, " ")
	values := strings.Join([]string{kind, bytes, dirty}, "")
	header := fmt.Sprintf(`<TR><TD COLSPAN="%d"><TABLE %v>%v</TABLE></TD></TR>`, len(vns.Index), styles, values)

	STYLEs := strings.Join([]string{BORDER(1), CELLBORDER(0), CELLSPACING(10), CELLPADDING(5), COLOR(), STYLE}, " ")
	VALUEs := strings.Join([]string{header, ports()}, "")
	return fmt.Sprintf(`<TABLE %v>%v</TABLE>`, STYLEs, VALUEs)
}

func (tdb *TrieDB) WriteDot(b io.Writer, fullMode bool) {

	funcMaps := template.FuncMap{
		"Name":  func(o makeNameFunc) string { return o.makeName() },
		"Table": func(o makeTableFunc) string { return o.makeTable() },
		"Edge": func(vnS, vnD *vizNode, port string) string {
			return fmt.Sprintf(`[tailport="%s" color=gray]`, makePortName(port))
		},
	}
	tmpl := template.New("trie")
	tmpl.Funcs(funcMaps)

	tmpl = template.Must(tmpl.Parse(`
	digraph G {
		edge [dir=both arrowtail=dot]
		{{template "vertices" .Root}}
		{{template "edges" .Root}}
	}

	{{- define "vertices"}}
		{{if eq .ID 0}}{{Name .}} [shape=plain label=<{{Table .}}>]{{end}}
		{{if .ChildrenPrefix}}{{Name .ChildrenPrefix}} [shape=plain label=<{{Table .ChildrenPrefix}}>]{{end}}
		{{if .ChildrenFolder}}{{Name .ChildrenFolder}} [shape=plain label=<{{Table .ChildrenFolder}}>]{{end}}
		{{- if .ChildrenPrefix}}
			{{- range $_, $vi := .ChildrenPrefix.Index}}
				{{- template "vertices" $vi.Node}}
			{{- end}}
		{{- end}}
		{{- if .ChildrenFolder}}
			{{- range $_, $vi := .ChildrenFolder.Index}}
				{{- template "vertices" $vi.Node}}
			{{- end}}
		{{- end}}
	{{- end}}

	{{- define "edges"}}
		{{- $O := .}}
		{{- if and .ChildrenPrefix (eq .ID 0)}} {{- Name $O}} -> {{Name $O.ChildrenPrefix}} [color=gray] {{end}}
		{{- if and .ChildrenFolder (eq .ID 0)}} {{- Name $O}} -> {{Name $O.ChildrenFolder}} [color=gray] {{end}}
		{{- template "children" .ChildrenPrefix}}
		{{- template "children" .ChildrenFolder}}
	{{- end}}

	{{- define "children"}}
		{{- $children := .}}
		{{- if $children}}
			{{- $parent := $children.ParentNode}}
			{{- range $_, $vi := $children.Index}}
				{{- $k := $vi.Key}}
				{{- $n := $vi.Node}}
				{{- if $n.ChildrenPrefix}}
					{{Name $children}} -> {{Name $n.ChildrenPrefix}} {{Edge $parent $n $k}}
				{{- end}}
				{{- if $n.ChildrenFolder}}
					{{Name $children}} -> {{Name $n.ChildrenFolder}} {{Edge $parent $n $k}}
				{{- end}}
				{{- template "edges" $n}}
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
