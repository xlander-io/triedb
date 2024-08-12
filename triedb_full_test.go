package triedb

// import (
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"os"
// 	"reflect"
// 	"strings"
// 	"testing"
// 	"text/template"

// 	"github.com/xlander-io/cache"
// 	"github.com/xlander-io/hash"
// 	"github.com/xlander-io/kv"
// 	"github.com/xlander-io/kv_leveldb"
// )

// func testPrepareTrieDB(dataPath string, rootHash *hash.Hash) (*TrieDB, error) {
// 	kvdb, err := kv_leveldb.NewDB(dataPath)
// 	if nil != err {
// 		return nil, err
// 	}
// 	//
// 	c, err := cache.New(nil)
// 	if nil != err {
// 		return nil, err
// 	}
// 	tdb, err := NewTrieDB(kvdb, c, &TrieDBConfig{
// 		Root_hash:           rootHash,
// 		Commit_thread_limit: 1,
// 	})

// 	return tdb, err
// }

// func testCloseTrieDB(tdb *TrieDB) {
// 	tdb.kvdb.Close()
// }

// func (tdb *TrieDB) testCommit() (*hash.Hash, error) {
// 	rootHash, toUpdate, toDel, err := tdb.Commit()
// 	if err != nil {
// 		return nil, err
// 	}

// 	b := kv.NewBatch()
// 	for hex_string, update_v := range toUpdate {
// 		b.Put([]byte(hex_string), update_v)
// 	}

// 	for _, del_v := range toDel {
// 		b.Delete(del_v.Bytes())
// 	}

// 	err = tdb.kvdb.WriteBatch(b, true)

// 	if nil != err {
// 		return nil, err
// 	}

// 	return rootHash, err
// }

// type TEST struct {
// 	label    string
// 	expected interface{}
// 	actual   interface{}
// 	ok       bool
// }

// type doCheck interface {
// 	isOk(o interface{}) bool
// }

// type isNotNil struct{}

// type isEqualTo struct {
// 	object interface{}
// }

// func (x isNotNil) isOk(o any) bool {

// 	if O, ok := o.(*hash.Hash); ok {
// 		return !hash.IsNilHash(O)
// 	}

// 	if nil != o {
// 		return !reflect.ValueOf(o).IsNil()
// 	} else {
// 		return false
// 	}
// }

// func (x isEqualTo) isOk(o any) bool {
// 	if O, ok := o.([]byte); ok {
// 		if X, ok := x.object.([]byte); ok {
// 			return bytes.Equal(X, O)
// 		} else {
// 			return false
// 		}
// 	}
// 	return x.object == o
// }

// type Expect struct {
// 	prefix       []byte
// 	dirty        bool
// 	parent_nodes doCheck

// 	node_bytes doCheck
// 	node_hash  doCheck
// 	//node_hash_recovered bool

// 	val                []byte
// 	val_hash           doCheck
// 	val_hash_recovered bool
// 	val_dirty          bool

// 	prefix_child_nodes                doCheck
// 	prefix_child_nodes_hash           doCheck
// 	prefix_child_nodes_hash_recovered bool

// 	folder_child_nodes                doCheck
// 	folder_child_nodes_hash           doCheck
// 	folder_child_nodes_hash_recovered bool

// 	prefix_child_nodes_len   int
// 	prefix_child_nodes_dirty bool

// 	folder_child_nodes_len   int
// 	folder_child_nodes_dirty bool
// }

// func (ex Expect) makeTests(o *Node) []TEST {
// 	tests := []TEST{
// 		{"prefix", ex.prefix, o.prefix, bytes.Equal(ex.prefix, o.prefix)},
// 		{"dirty", ex.dirty, o.dirty, ex.dirty == o.dirty},

// 		{"val_dirty", ex.val_dirty, o.val_dirty, ex.val_dirty == o.val_dirty},
// 		{"val_hash_recovered", ex.val_hash_recovered, o.val_hash_recovered, ex.val_hash_recovered == o.val_hash_recovered},
// 		{"prefix_child_nodes_hash_recovered", ex.prefix_child_nodes_hash_recovered, o.prefix_child_nodes_hash_recovered, ex.prefix_child_nodes_hash_recovered == o.prefix_child_nodes_hash_recovered},
// 		{"folder_child_nodes_hash_recovered", ex.folder_child_nodes_hash_recovered, o.folder_child_nodes_hash_recovered, ex.folder_child_nodes_hash_recovered == o.folder_child_nodes_hash_recovered},
// 	}

// 	if nil == ex.parent_nodes {
// 		tests = append(tests, TEST{"parent_nodes", ex.parent_nodes, o.parent_nodes, nil == o.parent_nodes})
// 	} else {
// 		tests = append(tests, TEST{"parent_nodes", ex.parent_nodes, o.parent_nodes, ex.parent_nodes.isOk(o.parent_nodes)})
// 	}

// 	if nil == ex.prefix_child_nodes {
// 		tests = append(tests, TEST{"prefix_child_nodes", ex.prefix_child_nodes, o.prefix_child_nodes, nil == o.prefix_child_nodes})
// 	} else {
// 		b := ex.prefix_child_nodes.isOk(o.prefix_child_nodes)
// 		tests = append(tests, TEST{"prefix_child_nodes", ex.prefix_child_nodes, o.prefix_child_nodes, b})
// 	}

// 	if nil == ex.folder_child_nodes {
// 		tests = append(tests, TEST{"folder_child_nodes", ex.folder_child_nodes, o.folder_child_nodes, nil == o.folder_child_nodes})
// 	} else {
// 		b := ex.folder_child_nodes.isOk(o.folder_child_nodes)
// 		tests = append(tests, TEST{"folder_child_nodes", ex.folder_child_nodes, o.folder_child_nodes, b})
// 	}

// 	if nil == ex.node_bytes {
// 		tests = append(tests, TEST{"node_bytes", ex.node_bytes, o.node_bytes, nil == o.node_bytes})
// 	} else {
// 		tests = append(tests, TEST{"node_bytes", ex.node_bytes, o.node_bytes, ex.node_bytes.isOk(o.node_bytes)})
// 	}

// 	if nil == ex.val {
// 		tests = append(tests, TEST{"val", ex.val, o.val, nil == o.val})
// 	} else {
// 		tests = append(tests, TEST{"val", ex.val, o.val, bytes.Equal(ex.val, o.val)})
// 	}

// 	if nil == ex.node_hash {
// 		tests = append(tests, TEST{"node_hash", ex.node_hash, o.node_hash, hash.IsNilHash(o.node_hash)})
// 	} else {
// 		tests = append(tests, TEST{"node_hash", ex.node_hash, o.node_hash, ex.node_hash.isOk(o.node_hash)})
// 	}

// 	if nil == ex.val_hash {
// 		tests = append(tests, TEST{"val_hash", ex.val_hash, o.val_hash, hash.IsNilHash(o.val_hash)})
// 	} else {
// 		tests = append(tests, TEST{"val_hash", ex.val_hash, o.val_hash, ex.val_hash.isOk(o.val_hash)})
// 	}

// 	if nil == ex.prefix_child_nodes_hash {
// 		tests = append(tests, TEST{"prefix_child_nodes_hash", ex.prefix_child_nodes_hash, o.prefix_child_nodes_hash, hash.IsNilHash(o.prefix_child_nodes_hash)})
// 	} else {
// 		tests = append(tests, TEST{"prefix_child_nodes_hash", ex.prefix_child_nodes_hash, o.prefix_child_nodes_hash, ex.prefix_child_nodes_hash.isOk(o.prefix_child_nodes_hash)})
// 	}

// 	if nil == ex.folder_child_nodes_hash {
// 		tests = append(tests, TEST{"folder_child_nodes_hash", ex.folder_child_nodes_hash, o.folder_child_nodes_hash, hash.IsNilHash(o.folder_child_nodes_hash)})
// 	} else {
// 		tests = append(tests, TEST{"folder_child_nodes_hash", ex.folder_child_nodes_hash, o.folder_child_nodes_hash, ex.folder_child_nodes_hash.isOk(o.folder_child_nodes_hash)})
// 	}

// 	if nil != o.prefix_child_nodes {
// 		tests = append(tests, []TEST{
// 			{"len(prefix_child_nodes)=n", ex.prefix_child_nodes_len, o.prefix_child_nodes.btree.Len(), ex.prefix_child_nodes_len == o.prefix_child_nodes.btree.Len()},
// 			{"len(prefix_child_nodes)>0", ex.prefix_child_nodes_len, o.prefix_child_nodes.btree.Len(), o.prefix_child_nodes.btree.Len() > 0},
// 			{"prefix_child_nodes.dirty", ex.prefix_child_nodes_dirty, o.prefix_child_nodes.dirty, ex.prefix_child_nodes_dirty == o.prefix_child_nodes.dirty},
// 			{"prefix_child_nodes.parent_node", o, o.prefix_child_nodes.parent_node, o == o.prefix_child_nodes.parent_node},
// 		}...)
// 	}

// 	if nil != o.folder_child_nodes {
// 		tests = append(tests, []TEST{
// 			{"len(folder_child_nodes)=n", ex.folder_child_nodes_len, o.folder_child_nodes.btree.Len(), ex.folder_child_nodes_len == o.folder_child_nodes.btree.Len()},
// 			{"len(folder_child_nodes)>0", ex.folder_child_nodes_len, o.folder_child_nodes.btree.Len(), o.folder_child_nodes.btree.Len() > 0},
// 			{"folder_child_nodes.dirty", ex.folder_child_nodes_dirty, o.folder_child_nodes.dirty, ex.folder_child_nodes_dirty == o.folder_child_nodes.dirty},
// 			{"folder_child_nodes.parent_node", o, o.folder_child_nodes.parent_node, o == o.folder_child_nodes.parent_node},
// 		}...)
// 	}

// 	return tests
// }

// func string2Bytes(x string) []byte {
// 	b := bytes.NewBufferString(x)
// 	return bytes.Clone(b.Bytes())
// }
// func bytes2String(x []byte) string {
// 	b := bytes.NewBuffer(x)
// 	return b.String()
// }

// func (n *Node) genCheckStatementsString() string {
// 	var buf bytes.Buffer
// 	n.genCheckStatements(&buf)
// 	return buf.String()
// }

// func (n *Node) genCheckStatements(b io.Writer) {

// 	funcMaps := template.FuncMap{
// 		"Name": func(n *Node) string {
// 			if nil == n.prefix {
// 				return "root"
// 			} else {
// 				return n.node_path_flat_str()
// 			}
// 		},
// 		"Bytes": func(n *Node, field string) string {
// 			if bytes.Equal([]byte("prefix"), string2Bytes(field)) {
// 				if nil != n.prefix {
// 					return fmt.Sprintf("[]byte(\"%s\")", n.prefix)
// 				}
// 			} else if bytes.Equal([]byte("val"), string2Bytes(field)) {
// 				if nil != n.val {
// 					return fmt.Sprintf("[]byte(\"%s\")", n.val)
// 				}
// 			} else {
// 				fmt.Printf("ERROR: Bytes unexpected field [%#v]!", field)
// 			}

// 			return "nil"
// 		},
// 		"Boolean": func(n *Node, field string) string {
// 			if bytes.Equal([]byte("dirty"), string2Bytes(field)) {
// 				if false != n.dirty {
// 					return "true"
// 				}
// 			} else if bytes.Equal([]byte("val_hash_recovered"), string2Bytes(field)) {
// 				if false != n.val_hash_recovered {
// 					return "true"
// 				}
// 			} else if bytes.Equal([]byte("val_dirty"), string2Bytes(field)) {
// 				if false != n.val_dirty {
// 					return "true"
// 				}
// 			} else if bytes.Equal([]byte("prefix_child_nodes_hash_recovered"), string2Bytes(field)) {
// 				if false != n.prefix_child_nodes_hash_recovered {
// 					return "true"
// 				}
// 			} else if bytes.Equal([]byte("folder_child_nodes_hash_recovered"), string2Bytes(field)) {
// 				if false != n.folder_child_nodes_hash_recovered {
// 					return "true"
// 				}
// 			} else if bytes.Equal([]byte("prefix_child_nodes_dirty"), string2Bytes(field)) {
// 				if nil != n.prefix_child_nodes {
// 					if false != n.prefix_child_nodes.dirty {
// 						return "true"
// 					}
// 				}
// 			} else if bytes.Equal([]byte("folder_child_nodes_dirty"), string2Bytes(field)) {
// 				if nil != n.folder_child_nodes {
// 					if false != n.folder_child_nodes.dirty {
// 						return "true"
// 					}
// 				}
// 			} else {
// 				fmt.Printf("ERROR: Boolean unexpected field [%#v]!", field)
// 			}
// 			return string("false")
// 		},
// 		"Length": func(n *Node, field string) string {
// 			if bytes.Equal([]byte("prefix_child_nodes_len"), string2Bytes(field)) {
// 				if nil != n.prefix_child_nodes {
// 					return fmt.Sprintf("%d", n.prefix_child_nodes.btree.Len())
// 				}
// 			} else if bytes.Equal([]byte("folder_child_nodes_len"), string2Bytes(field)) {
// 				if nil != n.folder_child_nodes {
// 					return fmt.Sprintf("%d", n.folder_child_nodes.btree.Len())
// 				}
// 			} else {
// 				fmt.Printf("ERROR: Length unexpected field [%#v]!", field)
// 			}

// 			return "0"
// 		},
// 		"isNotNil": func(n *Node, field string) string {
// 			if bytes.Equal([]byte("prefix_child_nodes"), string2Bytes(field)) {
// 				if nil != n.prefix_child_nodes {
// 					return "isNotNil{}"
// 				}
// 			} else if bytes.Equal([]byte("folder_child_nodes"), string2Bytes(field)) {
// 				if nil != n.folder_child_nodes {
// 					return "isNotNil{}"
// 				}
// 			} else if bytes.Equal([]byte("node_hash"), string2Bytes(field)) {
// 				if !hash.IsNilHash(n.node_hash) {
// 					return "isNotNil{}"
// 				}
// 			} else if bytes.Equal([]byte("val_hash"), string2Bytes(field)) {
// 				if !hash.IsNilHash(n.val_hash) {
// 					return "isNotNil{}"
// 				}
// 			} else if bytes.Equal([]byte("prefix_child_nodes_hash"), string2Bytes(field)) {
// 				if !hash.IsNilHash(n.prefix_child_nodes_hash) {
// 					return "isNotNil{}"
// 				}
// 			} else if bytes.Equal([]byte("folder_child_nodes_hash"), string2Bytes(field)) {
// 				if !hash.IsNilHash(n.folder_child_nodes_hash) {
// 					return "isNotNil{}"
// 				}
// 			} else if bytes.Equal([]byte("node_bytes"), string2Bytes(field)) {
// 				if nil != n.node_bytes {
// 					return "isNotNil{}"
// 				}
// 			} else {
// 				fmt.Printf("ERROR: isNotNil unexpected field [%#v]!", field)
// 			}

// 			return "nil"
// 		},
// 		"isEqualTo": func(n *Node, field string) string {
// 			if bytes.Equal([]byte("parent_nodes"), string2Bytes(field)) {
// 				if nil != n.parent_nodes {
// 					if len(n.parent_nodes.parent_node.node_path_flat()) > 0 {
// 						if n.parent_nodes.is_folder_child_nodes {
// 							return fmt.Sprintf("isEqualTo{_%s.folder_child_nodes}", n.parent_nodes.parent_node.node_path_flat_str())
// 						} else {
// 							return fmt.Sprintf("isEqualTo{_%s.prefix_child_nodes}", n.parent_nodes.parent_node.node_path_flat_str())
// 						}
// 					} else {
// 						if n.parent_nodes.is_folder_child_nodes {
// 							return "isEqualTo{_root.folder_child_nodes}"
// 						} else {
// 							return "isEqualTo{_root.prefix_child_nodes}"
// 						}
// 					}
// 				}
// 			} else {
// 				fmt.Printf("ERROR: isEqualTo unexpected field [%#v]!", field)
// 			}

// 			return "nil"
// 		},
// 		"Children": func(n *Node) string {
// 			if nil == n.prefix_child_nodes && nil == n.folder_child_nodes {
// 				return ""
// 			}
// 			var children []string
// 			children = append(children, "{")
// 			extract := func(ns *nodes, field string) {
// 				iter := ns.btree.Before(uint8(0))
// 				for iter.Next() {
// 					k := iter.Key.(uint8)
// 					n := iter.Value.(*Node)

// 					V := n.node_path_flat_str()
// 					PV := "root"
// 					if nil != n.parent_nodes {
// 						if nil != n.parent_nodes.parent_node {
// 							if x := n.parent_nodes.parent_node.node_path_flat(); len(x) > 0 {
// 								PV = n.parent_nodes.parent_node.node_path_flat_str()
// 							}
// 						}
// 					}

// 					children = append(children, fmt.Sprintf("_%s_ := _%s.%s.btree.Get(uint8('%s'))", V, PV, field, string([]byte{k})))
// 					children = append(children, fmt.Sprintf("_%s := _%s_.(*Node)", V, V))
// 					children = append(children, n.genCheckStatementsString())
// 				}
// 			}
// 			if nil != n.prefix_child_nodes {
// 				extract(n.prefix_child_nodes, "prefix_child_nodes")
// 			}
// 			if nil != n.folder_child_nodes {
// 				extract(n.folder_child_nodes, "folder_child_nodes")
// 			}

// 			children = append(children, "}")
// 			return strings.Join(children, "\n")
// 		},
// 	}
// 	tmpl := template.New("Node")
// 	tmpl.Funcs(funcMaps)

// 	tmpl = template.Must(tmpl.Parse(`
// _{{Name .}}Tests := Expect{
// 	prefix:       {{Bytes . "prefix"}},
// 	dirty:        {{Boolean . "dirty"}},
// 	parent_nodes: {{isEqualTo . "parent_nodes"}},

// 	node_bytes: {{isNotNil . "node_bytes"}},
// 	node_hash:  {{isNotNil . "node_hash"}},

// 	val:                {{Bytes . "val"}},
// 	val_hash:           {{isNotNil . "val_hash"}},
// 	val_hash_recovered: {{Boolean . "val_hash_recovered"}},
// 	val_dirty:          {{Boolean . "val_dirty"}},

// 	prefix_child_nodes:                {{isNotNil . "prefix_child_nodes"}},
// 	prefix_child_nodes_hash:           {{isNotNil . "prefix_child_nodes_hash"}},
// 	prefix_child_nodes_hash_recovered: {{Boolean . "prefix_child_nodes_hash_recovered"}},

// 	folder_child_nodes:                {{isNotNil . "folder_child_nodes"}},
// 	folder_child_nodes_hash:           {{isNotNil . "folder_child_nodes_hash"}},
// 	folder_child_nodes_hash_recovered: {{Boolean . "folder_child_nodes_hash_recovered"}},

// 	prefix_child_nodes_len:   {{Length . "prefix_child_nodes_len"}},
// 	prefix_child_nodes_dirty: {{Boolean . "prefix_child_nodes_dirty"}},

// 	folder_child_nodes_len:   {{Length . "folder_child_nodes_len"}},
// 	folder_child_nodes_dirty: {{Boolean . "folder_child_nodes_dirty"}},
// }.makeTests(_{{Name .}})

// for _, tt := range _{{Name .}}Tests {
// 	if !tt.ok {
// 		t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 	}
// }

// {{Children .}}

// `))

// 	err := tmpl.Execute(b, n)
// 	if err != nil {
// 		fmt.Println(err)
// 		//panic(err)
// 	}
// }

// func TestMainWorkflow(t *testing.T) {

// 	// dot -Tpdf -O *.dot && open *.dot.pdf
// 	// tdb.GenDotFile("./test_mainworkflow.dot", false)

// 	const db_path = "./triedb_mainworkflow_test.db"
// 	defer os.RemoveAll(db_path)

// 	var rootHash *hash.Hash = nil

// 	// first:  create many trie data
// 	{
// 		tdb, err := testPrepareTrieDB(db_path, rootHash)

// 		if nil != err {
// 			t.Fatal(err)
// 		}

// 		_root := tdb.root_node

// 		tdb.GenDotFile("./test_mainworkflow_1_1.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: nil,
// 				node_hash:  nil,

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: true,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: true,

// 				folder_child_nodes:                nil,
// 				folder_child_nodes_hash:           nil,
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   0,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}
// 		}

// 		tdb.Put(Path([]byte("A")), []byte("val_A"), true)

// 		tdb.GenDotFile("./test_mainworkflow_1_2.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        true,
// 				parent_nodes: nil,

// 				node_bytes: nil,
// 				node_hash:  nil,

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: true,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: true,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           nil,
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   1,
// 				folder_child_nodes_dirty: true,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: nil,
// 					node_hash:  nil,

// 					val:                []byte("val_A"),
// 					val_hash:           nil,
// 					val_hash_recovered: true,
// 					val_dirty:          true,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           nil,
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}
// 			}
// 		}

// 		tdb.Put(Path([]byte("AB")), []byte("val_AB"), true)
// 		tdb.Put(Path([]byte("AC")), []byte("val_AC"), true)
// 		tdb.Put(Path([]byte("AD")), []byte("val_AD"), true)
// 		tdb.Put(Path([]byte("ABC")), []byte("val_ABC"), true)
// 		tdb.Put(Path([]byte("ABCD")), []byte("val_ABCD"), true)

// 		tdb.GenDotFile("./test_mainworkflow_1_3.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        true,
// 				parent_nodes: nil,

// 				node_bytes: nil,
// 				node_hash:  nil,

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: true,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: true,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           nil,
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   1,
// 				folder_child_nodes_dirty: true,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: nil,
// 					node_hash:  nil,

// 					val:                []byte("val_A"),
// 					val_hash:           nil,
// 					val_hash_recovered: true,
// 					val_dirty:          true,

// 					prefix_child_nodes:                isNotNil{},
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           nil,
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   3,
// 					prefix_child_nodes_dirty: true,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_2AB_ := _1A.prefix_child_nodes.btree.Get(uint8('B'))
// 					_2AB := _2AB_.(*Node)

// 					_2ABTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: nil,
// 						node_hash:  nil,

// 						val:                []byte("val_AB"),
// 						val_hash:           nil,
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                isNotNil{},
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   1,
// 						prefix_child_nodes_dirty: true,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AB)

// 					for _, tt := range _2ABTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_3ABC_ := _2AB.prefix_child_nodes.btree.Get(uint8('C'))
// 						_3ABC := _3ABC_.(*Node)

// 						_3ABCTests := Expect{
// 							prefix:       []byte("C"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: nil,
// 							node_hash:  nil,

// 							val:                []byte("val_ABC"),
// 							val_hash:           nil,
// 							val_hash_recovered: true,
// 							val_dirty:          true,

// 							prefix_child_nodes:                isNotNil{},
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: true,

// 							prefix_child_nodes_len:   1,
// 							prefix_child_nodes_dirty: true,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABC)

// 						for _, tt := range _3ABCTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 						{
// 							_4ABCD_ := _3ABC.prefix_child_nodes.btree.Get(uint8('D'))
// 							_4ABCD := _4ABCD_.(*Node)

// 							_4ABCDTests := Expect{
// 								prefix:       []byte("D"),
// 								dirty:        true,
// 								parent_nodes: isEqualTo{_3ABC.prefix_child_nodes},

// 								node_bytes: nil,
// 								node_hash:  nil,

// 								val:                []byte("val_ABCD"),
// 								val_hash:           nil,
// 								val_hash_recovered: true,
// 								val_dirty:          true,

// 								prefix_child_nodes:                nil,
// 								prefix_child_nodes_hash:           nil,
// 								prefix_child_nodes_hash_recovered: true,

// 								folder_child_nodes:                nil,
// 								folder_child_nodes_hash:           nil,
// 								folder_child_nodes_hash_recovered: true,

// 								prefix_child_nodes_len:   0,
// 								prefix_child_nodes_dirty: false,

// 								folder_child_nodes_len:   0,
// 								folder_child_nodes_dirty: false,
// 							}.makeTests(_4ABCD)

// 							for _, tt := range _4ABCDTests {
// 								if !tt.ok {
// 									t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 								}
// 							}

// 						}

// 					}

// 					_2AC_ := _1A.prefix_child_nodes.btree.Get(uint8('C'))
// 					_2AC := _2AC_.(*Node)

// 					_2ACTests := Expect{
// 						prefix:       []byte("C"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: nil,
// 						node_hash:  nil,

// 						val:                []byte("val_AC"),
// 						val_hash:           nil,
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AC)

// 					for _, tt := range _2ACTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AD_ := _1A.prefix_child_nodes.btree.Get(uint8('D'))
// 					_2AD := _2AD_.(*Node)

// 					_2ADTests := Expect{
// 						prefix:       []byte("D"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: nil,
// 						node_hash:  nil,

// 						val:                []byte("val_AD"),
// 						val_hash:           nil,
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AD)

// 					for _, tt := range _2ADTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 				}
// 			}
// 		}

// 		tdb.Put(Path([]byte("A"), []byte("A"), []byte("A")), []byte("val_A_A_A"), true)
// 		tdb.Put(Path([]byte("AB"), []byte("CD")), []byte("val_AB_CD"), true)

// 		_rootHash, err := tdb.testCommit()

// 		if nil != err {
// 			t.Fatal(err)
// 		}

// 		rootHash = _rootHash

// 		tdb.GenDotFile("./test_mainworkflow_1_4.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        true,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: true,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: true,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   1,
// 				folder_child_nodes_dirty: true,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                []byte("val_A"),
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: true,
// 					val_dirty:          true,

// 					prefix_child_nodes:                isNotNil{},
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                isNotNil{},
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   3,
// 					prefix_child_nodes_dirty: true,

// 					folder_child_nodes_len:   1,
// 					folder_child_nodes_dirty: true,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_2AB_ := _1A.prefix_child_nodes.btree.Get(uint8('B'))
// 					_2AB := _2AB_.(*Node)

// 					_2ABTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                []byte("val_AB"),
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                isNotNil{},
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                isNotNil{},
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   1,
// 						prefix_child_nodes_dirty: true,

// 						folder_child_nodes_len:   1,
// 						folder_child_nodes_dirty: true,
// 					}.makeTests(_2AB)

// 					for _, tt := range _2ABTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_3ABC_ := _2AB.prefix_child_nodes.btree.Get(uint8('C'))
// 						_3ABC := _3ABC_.(*Node)

// 						_3ABCTests := Expect{
// 							prefix:       []byte("C"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_ABC"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          true,

// 							prefix_child_nodes:                isNotNil{},
// 							prefix_child_nodes_hash:           isNotNil{},
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: true,

// 							prefix_child_nodes_len:   1,
// 							prefix_child_nodes_dirty: true,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABC)

// 						for _, tt := range _3ABCTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 						{
// 							_4ABCD_ := _3ABC.prefix_child_nodes.btree.Get(uint8('D'))
// 							_4ABCD := _4ABCD_.(*Node)

// 							_4ABCDTests := Expect{
// 								prefix:       []byte("D"),
// 								dirty:        true,
// 								parent_nodes: isEqualTo{_3ABC.prefix_child_nodes},

// 								node_bytes: isNotNil{},
// 								node_hash:  isNotNil{},

// 								val:                []byte("val_ABCD"),
// 								val_hash:           isNotNil{},
// 								val_hash_recovered: true,
// 								val_dirty:          true,

// 								prefix_child_nodes:                nil,
// 								prefix_child_nodes_hash:           nil,
// 								prefix_child_nodes_hash_recovered: true,

// 								folder_child_nodes:                nil,
// 								folder_child_nodes_hash:           nil,
// 								folder_child_nodes_hash_recovered: true,

// 								prefix_child_nodes_len:   0,
// 								prefix_child_nodes_dirty: false,

// 								folder_child_nodes_len:   0,
// 								folder_child_nodes_dirty: false,
// 							}.makeTests(_4ABCD)

// 							for _, tt := range _4ABCDTests {
// 								if !tt.ok {
// 									t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 								}
// 							}

// 						}

// 						_2AB2CD_ := _2AB.folder_child_nodes.btree.Get(uint8('C'))
// 						_2AB2CD := _2AB2CD_.(*Node)

// 						_2AB2CDTests := Expect{
// 							prefix:       []byte("CD"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_2AB.folder_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_AB_CD"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          true,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: true,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_2AB2CD)

// 						for _, tt := range _2AB2CDTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 					_2AC_ := _1A.prefix_child_nodes.btree.Get(uint8('C'))
// 					_2AC := _2AC_.(*Node)

// 					_2ACTests := Expect{
// 						prefix:       []byte("C"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                []byte("val_AC"),
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AC)

// 					for _, tt := range _2ACTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AD_ := _1A.prefix_child_nodes.btree.Get(uint8('D'))
// 					_2AD := _2AD_.(*Node)

// 					_2ADTests := Expect{
// 						prefix:       []byte("D"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                []byte("val_AD"),
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AD)

// 					for _, tt := range _2ADTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_1A1A_ := _1A.folder_child_nodes.btree.Get(uint8('A'))
// 					_1A1A := _1A1A_.(*Node)

// 					_1A1ATests := Expect{
// 						prefix:       []byte("A"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.folder_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           nil,
// 						val_hash_recovered: true,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                isNotNil{},
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   1,
// 						folder_child_nodes_dirty: true,
// 					}.makeTests(_1A1A)

// 					for _, tt := range _1A1ATests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_1A1A1A_ := _1A1A.folder_child_nodes.btree.Get(uint8('A'))
// 						_1A1A1A := _1A1A1A_.(*Node)

// 						_1A1A1ATests := Expect{
// 							prefix:       []byte("A"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_1A1A.folder_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_A_A_A"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          true,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: true,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_1A1A1A)

// 						for _, tt := range _1A1A1ATests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 		testCloseTrieDB(tdb)
// 	}

// 	// second: load the previous triedb in disk, and then create more trie data
// 	{
// 		tdb, err := testPrepareTrieDB(db_path, rootHash)

// 		if nil != err {
// 			t.Fatal(err)
// 		}

// 		_root := tdb.root_node

// 		tdb.GenDotFile("./test_mainworkflow_2_1.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                nil,
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: false,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   0,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}
// 		}

// 		tdb.Put(Path([]byte("B")), []byte("val_B"), true)
// 		tdb.Put(Path([]byte("Aa")), []byte("val_Aa"), true)
// 		tdb.Put(Path([]byte("ABa")), []byte("val_ABa"), true)
// 		tdb.Put(Path([]byte("ABCa")), []byte("val_ABCa"), true)
// 		tdb.GenDotFile("./test_mainworkflow_2_2.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        true,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: true,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                isNotNil{},
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   4,
// 					prefix_child_nodes_dirty: true,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_2AB_ := _1A.prefix_child_nodes.btree.Get(uint8('B'))
// 					_2AB := _2AB_.(*Node)

// 					_2ABTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                isNotNil{},
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   2,
// 						prefix_child_nodes_dirty: true,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AB)

// 					for _, tt := range _2ABTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_3ABC_ := _2AB.prefix_child_nodes.btree.Get(uint8('C'))
// 						_3ABC := _3ABC_.(*Node)

// 						_3ABCTests := Expect{
// 							prefix:       []byte("C"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                nil,
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: false,
// 							val_dirty:          false,

// 							prefix_child_nodes:                isNotNil{},
// 							prefix_child_nodes_hash:           isNotNil{},
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   2,
// 							prefix_child_nodes_dirty: true,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABC)

// 						for _, tt := range _3ABCTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 						{
// 							_4ABCD_ := _3ABC.prefix_child_nodes.btree.Get(uint8('D'))
// 							_4ABCD := _4ABCD_.(*Node)

// 							_4ABCDTests := Expect{
// 								prefix:       []byte("D"),
// 								dirty:        false,
// 								parent_nodes: isEqualTo{_3ABC.prefix_child_nodes},

// 								node_bytes: isNotNil{},
// 								node_hash:  isNotNil{},

// 								val:                nil,
// 								val_hash:           isNotNil{},
// 								val_hash_recovered: false,
// 								val_dirty:          false,

// 								prefix_child_nodes:                nil,
// 								prefix_child_nodes_hash:           nil,
// 								prefix_child_nodes_hash_recovered: false,

// 								folder_child_nodes:                nil,
// 								folder_child_nodes_hash:           nil,
// 								folder_child_nodes_hash_recovered: false,

// 								prefix_child_nodes_len:   0,
// 								prefix_child_nodes_dirty: false,

// 								folder_child_nodes_len:   0,
// 								folder_child_nodes_dirty: false,
// 							}.makeTests(_4ABCD)

// 							for _, tt := range _4ABCDTests {
// 								if !tt.ok {
// 									t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 								}
// 							}

// 							_4ABCa_ := _3ABC.prefix_child_nodes.btree.Get(uint8('a'))
// 							_4ABCa := _4ABCa_.(*Node)

// 							_4ABCaTests := Expect{
// 								prefix:       []byte("a"),
// 								dirty:        true,
// 								parent_nodes: isEqualTo{_3ABC.prefix_child_nodes},

// 								node_bytes: nil,
// 								node_hash:  nil,

// 								val:                []byte("val_ABCa"),
// 								val_hash:           nil,
// 								val_hash_recovered: true,
// 								val_dirty:          true,

// 								prefix_child_nodes:                nil,
// 								prefix_child_nodes_hash:           nil,
// 								prefix_child_nodes_hash_recovered: true,

// 								folder_child_nodes:                nil,
// 								folder_child_nodes_hash:           nil,
// 								folder_child_nodes_hash_recovered: true,

// 								prefix_child_nodes_len:   0,
// 								prefix_child_nodes_dirty: false,

// 								folder_child_nodes_len:   0,
// 								folder_child_nodes_dirty: false,
// 							}.makeTests(_4ABCa)

// 							for _, tt := range _4ABCaTests {
// 								if !tt.ok {
// 									t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 								}
// 							}

// 						}

// 						_3ABa_ := _2AB.prefix_child_nodes.btree.Get(uint8('a'))
// 						_3ABa := _3ABa_.(*Node)

// 						_3ABaTests := Expect{
// 							prefix:       []byte("a"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: nil,
// 							node_hash:  nil,

// 							val:                []byte("val_ABa"),
// 							val_hash:           nil,
// 							val_hash_recovered: true,
// 							val_dirty:          true,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: true,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABa)

// 						for _, tt := range _3ABaTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 					_2AC_ := _1A.prefix_child_nodes.btree.Get(uint8('C'))
// 					_2AC := _2AC_.(*Node)

// 					_2ACTests := Expect{
// 						prefix:       []byte("C"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AC)

// 					for _, tt := range _2ACTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AD_ := _1A.prefix_child_nodes.btree.Get(uint8('D'))
// 					_2AD := _2AD_.(*Node)

// 					_2ADTests := Expect{
// 						prefix:       []byte("D"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AD)

// 					for _, tt := range _2ADTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2Aa_ := _1A.prefix_child_nodes.btree.Get(uint8('a'))
// 					_2Aa := _2Aa_.(*Node)

// 					_2AaTests := Expect{
// 						prefix:       []byte("a"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: nil,
// 						node_hash:  nil,

// 						val:                []byte("val_Aa"),
// 						val_hash:           nil,
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2Aa)

// 					for _, tt := range _2AaTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: nil,
// 					node_hash:  nil,

// 					val:                []byte("val_B"),
// 					val_hash:           nil,
// 					val_hash_recovered: true,
// 					val_dirty:          true,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           nil,
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 			}
// 		}

// 		tdb.Put(Path([]byte("B"), []byte("B"), []byte("B")), []byte("val_B_B_B"), true)
// 		tdb.Put(Path([]byte("B"), []byte("B")), []byte("val_B_B"), true)
// 		tdb.GenDotFile("./test_mainworkflow_2_3.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        true,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: true,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                isNotNil{},
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   4,
// 					prefix_child_nodes_dirty: true,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_2AB_ := _1A.prefix_child_nodes.btree.Get(uint8('B'))
// 					_2AB := _2AB_.(*Node)

// 					_2ABTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                isNotNil{},
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   2,
// 						prefix_child_nodes_dirty: true,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AB)

// 					for _, tt := range _2ABTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_3ABC_ := _2AB.prefix_child_nodes.btree.Get(uint8('C'))
// 						_3ABC := _3ABC_.(*Node)

// 						_3ABCTests := Expect{
// 							prefix:       []byte("C"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                nil,
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: false,
// 							val_dirty:          false,

// 							prefix_child_nodes:                isNotNil{},
// 							prefix_child_nodes_hash:           isNotNil{},
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   2,
// 							prefix_child_nodes_dirty: true,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABC)

// 						for _, tt := range _3ABCTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 						{
// 							_4ABCD_ := _3ABC.prefix_child_nodes.btree.Get(uint8('D'))
// 							_4ABCD := _4ABCD_.(*Node)

// 							_4ABCDTests := Expect{
// 								prefix:       []byte("D"),
// 								dirty:        false,
// 								parent_nodes: isEqualTo{_3ABC.prefix_child_nodes},

// 								node_bytes: isNotNil{},
// 								node_hash:  isNotNil{},

// 								val:                nil,
// 								val_hash:           isNotNil{},
// 								val_hash_recovered: false,
// 								val_dirty:          false,

// 								prefix_child_nodes:                nil,
// 								prefix_child_nodes_hash:           nil,
// 								prefix_child_nodes_hash_recovered: false,

// 								folder_child_nodes:                nil,
// 								folder_child_nodes_hash:           nil,
// 								folder_child_nodes_hash_recovered: false,

// 								prefix_child_nodes_len:   0,
// 								prefix_child_nodes_dirty: false,

// 								folder_child_nodes_len:   0,
// 								folder_child_nodes_dirty: false,
// 							}.makeTests(_4ABCD)

// 							for _, tt := range _4ABCDTests {
// 								if !tt.ok {
// 									t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 								}
// 							}

// 							_4ABCa_ := _3ABC.prefix_child_nodes.btree.Get(uint8('a'))
// 							_4ABCa := _4ABCa_.(*Node)

// 							_4ABCaTests := Expect{
// 								prefix:       []byte("a"),
// 								dirty:        true,
// 								parent_nodes: isEqualTo{_3ABC.prefix_child_nodes},

// 								node_bytes: nil,
// 								node_hash:  nil,

// 								val:                []byte("val_ABCa"),
// 								val_hash:           nil,
// 								val_hash_recovered: true,
// 								val_dirty:          true,

// 								prefix_child_nodes:                nil,
// 								prefix_child_nodes_hash:           nil,
// 								prefix_child_nodes_hash_recovered: true,

// 								folder_child_nodes:                nil,
// 								folder_child_nodes_hash:           nil,
// 								folder_child_nodes_hash_recovered: true,

// 								prefix_child_nodes_len:   0,
// 								prefix_child_nodes_dirty: false,

// 								folder_child_nodes_len:   0,
// 								folder_child_nodes_dirty: false,
// 							}.makeTests(_4ABCa)

// 							for _, tt := range _4ABCaTests {
// 								if !tt.ok {
// 									t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 								}
// 							}

// 						}

// 						_3ABa_ := _2AB.prefix_child_nodes.btree.Get(uint8('a'))
// 						_3ABa := _3ABa_.(*Node)

// 						_3ABaTests := Expect{
// 							prefix:       []byte("a"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: nil,
// 							node_hash:  nil,

// 							val:                []byte("val_ABa"),
// 							val_hash:           nil,
// 							val_hash_recovered: true,
// 							val_dirty:          true,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: true,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABa)

// 						for _, tt := range _3ABaTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 					_2AC_ := _1A.prefix_child_nodes.btree.Get(uint8('C'))
// 					_2AC := _2AC_.(*Node)

// 					_2ACTests := Expect{
// 						prefix:       []byte("C"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AC)

// 					for _, tt := range _2ACTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AD_ := _1A.prefix_child_nodes.btree.Get(uint8('D'))
// 					_2AD := _2AD_.(*Node)

// 					_2ADTests := Expect{
// 						prefix:       []byte("D"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AD)

// 					for _, tt := range _2ADTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2Aa_ := _1A.prefix_child_nodes.btree.Get(uint8('a'))
// 					_2Aa := _2Aa_.(*Node)

// 					_2AaTests := Expect{
// 						prefix:       []byte("a"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: nil,
// 						node_hash:  nil,

// 						val:                []byte("val_Aa"),
// 						val_hash:           nil,
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2Aa)

// 					for _, tt := range _2AaTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: nil,
// 					node_hash:  nil,

// 					val:                []byte("val_B"),
// 					val_hash:           nil,
// 					val_hash_recovered: true,
// 					val_dirty:          true,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                isNotNil{},
// 					folder_child_nodes_hash:           nil,
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   1,
// 					folder_child_nodes_dirty: true,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_1B1B_ := _1B.folder_child_nodes.btree.Get(uint8('B'))
// 					_1B1B := _1B1B_.(*Node)

// 					_1B1BTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1B.folder_child_nodes},

// 						node_bytes: nil,
// 						node_hash:  nil,

// 						val:                []byte("val_B_B"),
// 						val_hash:           nil,
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                isNotNil{},
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   1,
// 						folder_child_nodes_dirty: true,
// 					}.makeTests(_1B1B)

// 					for _, tt := range _1B1BTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_1B1B1B_ := _1B1B.folder_child_nodes.btree.Get(uint8('B'))
// 						_1B1B1B := _1B1B1B_.(*Node)

// 						_1B1B1BTests := Expect{
// 							prefix:       []byte("B"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_1B1B.folder_child_nodes},

// 							node_bytes: nil,
// 							node_hash:  nil,

// 							val:                []byte("val_B_B_B"),
// 							val_hash:           nil,
// 							val_hash_recovered: true,
// 							val_dirty:          true,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: true,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_1B1B1B)

// 						for _, tt := range _1B1B1BTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 				}

// 			}
// 		}

// 		_rootHash, err := tdb.testCommit()

// 		if nil != err {
// 			t.Fatal(err)
// 		}

// 		rootHash = _rootHash

// 		tdb.GenDotFile("./test_mainworkflow_2_4.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        true,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: true,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                isNotNil{},
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   4,
// 					prefix_child_nodes_dirty: true,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_2AB_ := _1A.prefix_child_nodes.btree.Get(uint8('B'))
// 					_2AB := _2AB_.(*Node)

// 					_2ABTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                isNotNil{},
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   2,
// 						prefix_child_nodes_dirty: true,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AB)

// 					for _, tt := range _2ABTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_3ABC_ := _2AB.prefix_child_nodes.btree.Get(uint8('C'))
// 						_3ABC := _3ABC_.(*Node)

// 						_3ABCTests := Expect{
// 							prefix:       []byte("C"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                nil,
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: false,
// 							val_dirty:          false,

// 							prefix_child_nodes:                isNotNil{},
// 							prefix_child_nodes_hash:           isNotNil{},
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   2,
// 							prefix_child_nodes_dirty: true,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABC)

// 						for _, tt := range _3ABCTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 						{
// 							_4ABCD_ := _3ABC.prefix_child_nodes.btree.Get(uint8('D'))
// 							_4ABCD := _4ABCD_.(*Node)

// 							_4ABCDTests := Expect{
// 								prefix:       []byte("D"),
// 								dirty:        false,
// 								parent_nodes: isEqualTo{_3ABC.prefix_child_nodes},

// 								node_bytes: isNotNil{},
// 								node_hash:  isNotNil{},

// 								val:                nil,
// 								val_hash:           isNotNil{},
// 								val_hash_recovered: false,
// 								val_dirty:          false,

// 								prefix_child_nodes:                nil,
// 								prefix_child_nodes_hash:           nil,
// 								prefix_child_nodes_hash_recovered: false,

// 								folder_child_nodes:                nil,
// 								folder_child_nodes_hash:           nil,
// 								folder_child_nodes_hash_recovered: false,

// 								prefix_child_nodes_len:   0,
// 								prefix_child_nodes_dirty: false,

// 								folder_child_nodes_len:   0,
// 								folder_child_nodes_dirty: false,
// 							}.makeTests(_4ABCD)

// 							for _, tt := range _4ABCDTests {
// 								if !tt.ok {
// 									t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 								}
// 							}

// 							_4ABCa_ := _3ABC.prefix_child_nodes.btree.Get(uint8('a'))
// 							_4ABCa := _4ABCa_.(*Node)

// 							_4ABCaTests := Expect{
// 								prefix:       []byte("a"),
// 								dirty:        true,
// 								parent_nodes: isEqualTo{_3ABC.prefix_child_nodes},

// 								node_bytes: isNotNil{},
// 								node_hash:  isNotNil{},

// 								val:                []byte("val_ABCa"),
// 								val_hash:           isNotNil{},
// 								val_hash_recovered: true,
// 								val_dirty:          true,

// 								prefix_child_nodes:                nil,
// 								prefix_child_nodes_hash:           nil,
// 								prefix_child_nodes_hash_recovered: true,

// 								folder_child_nodes:                nil,
// 								folder_child_nodes_hash:           nil,
// 								folder_child_nodes_hash_recovered: true,

// 								prefix_child_nodes_len:   0,
// 								prefix_child_nodes_dirty: false,

// 								folder_child_nodes_len:   0,
// 								folder_child_nodes_dirty: false,
// 							}.makeTests(_4ABCa)

// 							for _, tt := range _4ABCaTests {
// 								if !tt.ok {
// 									t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 								}
// 							}

// 						}

// 						_3ABa_ := _2AB.prefix_child_nodes.btree.Get(uint8('a'))
// 						_3ABa := _3ABa_.(*Node)

// 						_3ABaTests := Expect{
// 							prefix:       []byte("a"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_ABa"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          true,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: true,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABa)

// 						for _, tt := range _3ABaTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 					_2AC_ := _1A.prefix_child_nodes.btree.Get(uint8('C'))
// 					_2AC := _2AC_.(*Node)

// 					_2ACTests := Expect{
// 						prefix:       []byte("C"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AC)

// 					for _, tt := range _2ACTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AD_ := _1A.prefix_child_nodes.btree.Get(uint8('D'))
// 					_2AD := _2AD_.(*Node)

// 					_2ADTests := Expect{
// 						prefix:       []byte("D"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AD)

// 					for _, tt := range _2ADTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2Aa_ := _1A.prefix_child_nodes.btree.Get(uint8('a'))
// 					_2Aa := _2Aa_.(*Node)

// 					_2AaTests := Expect{
// 						prefix:       []byte("a"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                []byte("val_Aa"),
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2Aa)

// 					for _, tt := range _2AaTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                []byte("val_B"),
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: true,
// 					val_dirty:          true,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                isNotNil{},
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   1,
// 					folder_child_nodes_dirty: true,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_1B1B_ := _1B.folder_child_nodes.btree.Get(uint8('B'))
// 					_1B1B := _1B1B_.(*Node)

// 					_1B1BTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1B.folder_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                []byte("val_B_B"),
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                isNotNil{},
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   1,
// 						folder_child_nodes_dirty: true,
// 					}.makeTests(_1B1B)

// 					for _, tt := range _1B1BTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_1B1B1B_ := _1B1B.folder_child_nodes.btree.Get(uint8('B'))
// 						_1B1B1B := _1B1B1B_.(*Node)

// 						_1B1B1BTests := Expect{
// 							prefix:       []byte("B"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_1B1B.folder_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_B_B_B"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          true,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: true,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_1B1B1B)

// 						for _, tt := range _1B1B1BTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 				}

// 			}
// 		}

// 		testCloseTrieDB(tdb)
// 	}

// 	// test Get and Delete operations to influence the lazy status
// 	{
// 		tdb, err := testPrepareTrieDB(db_path, rootHash)

// 		if nil != err {
// 			t.Fatal(err)
// 		}

// 		_root := tdb.root_node

// 		tdb.GenDotFile("./test_mainworkflow_3_1.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                nil,
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: false,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   0,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}
// 		}

// 		{
// 			_, _, err := tdb.Get(Path([]byte("ABC")))
// 			if nil != err {
// 				t.Fatal(`Get item "ABC" should receive node_bytes of that kv item!`)
// 			}
// 		}

// 		tdb.GenDotFile("./test_mainworkflow_3_2.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                isNotNil{},
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   4,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_2AB_ := _1A.prefix_child_nodes.btree.Get(uint8('B'))
// 					_2AB := _2AB_.(*Node)

// 					_2ABTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                isNotNil{},
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   2,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AB)

// 					for _, tt := range _2ABTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_3ABC_ := _2AB.prefix_child_nodes.btree.Get(uint8('C'))
// 						_3ABC := _3ABC_.(*Node)

// 						_3ABCTests := Expect{
// 							prefix:       []byte("C"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_ABC"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           isNotNil{},
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABC)

// 						for _, tt := range _3ABCTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 						_3ABa_ := _2AB.prefix_child_nodes.btree.Get(uint8('a'))
// 						_3ABa := _3ABa_.(*Node)

// 						_3ABaTests := Expect{
// 							prefix:       []byte("a"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                nil,
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: false,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABa)

// 						for _, tt := range _3ABaTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 					_2AC_ := _1A.prefix_child_nodes.btree.Get(uint8('C'))
// 					_2AC := _2AC_.(*Node)

// 					_2ACTests := Expect{
// 						prefix:       []byte("C"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AC)

// 					for _, tt := range _2ACTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AD_ := _1A.prefix_child_nodes.btree.Get(uint8('D'))
// 					_2AD := _2AD_.(*Node)

// 					_2ADTests := Expect{
// 						prefix:       []byte("D"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AD)

// 					for _, tt := range _2ADTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2Aa_ := _1A.prefix_child_nodes.btree.Get(uint8('a'))
// 					_2Aa := _2Aa_.(*Node)

// 					_2AaTests := Expect{
// 						prefix:       []byte("a"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2Aa)

// 					for _, tt := range _2AaTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 			}
// 		}

// 		{
// 			_, _, err := tdb.Get(Path([]byte("AB")))
// 			if nil != err {
// 				t.Fatal(`Delete item "AB" should work as expected!`)
// 			}
// 		}

// 		tdb.GenDotFile("./test_mainworkflow_3_3.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                isNotNil{},
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   4,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_2AB_ := _1A.prefix_child_nodes.btree.Get(uint8('B'))
// 					_2AB := _2AB_.(*Node)

// 					_2ABTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                []byte("val_AB"),
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: true,
// 						val_dirty:          false,

// 						prefix_child_nodes:                isNotNil{},
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   2,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AB)

// 					for _, tt := range _2ABTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_3ABC_ := _2AB.prefix_child_nodes.btree.Get(uint8('C'))
// 						_3ABC := _3ABC_.(*Node)

// 						_3ABCTests := Expect{
// 							prefix:       []byte("C"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_ABC"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           isNotNil{},
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABC)

// 						for _, tt := range _3ABCTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 						_3ABa_ := _2AB.prefix_child_nodes.btree.Get(uint8('a'))
// 						_3ABa := _3ABa_.(*Node)

// 						_3ABaTests := Expect{
// 							prefix:       []byte("a"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                nil,
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: false,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABa)

// 						for _, tt := range _3ABaTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 					_2AC_ := _1A.prefix_child_nodes.btree.Get(uint8('C'))
// 					_2AC := _2AC_.(*Node)

// 					_2ACTests := Expect{
// 						prefix:       []byte("C"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AC)

// 					for _, tt := range _2ACTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AD_ := _1A.prefix_child_nodes.btree.Get(uint8('D'))
// 					_2AD := _2AD_.(*Node)

// 					_2ADTests := Expect{
// 						prefix:       []byte("D"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AD)

// 					for _, tt := range _2ADTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2Aa_ := _1A.prefix_child_nodes.btree.Get(uint8('a'))
// 					_2Aa := _2Aa_.(*Node)

// 					_2AaTests := Expect{
// 						prefix:       []byte("a"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2Aa)

// 					for _, tt := range _2AaTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 			}
// 		}

// 		{
// 			_, err := tdb.Del(Path([]byte("AB")))
// 			if nil != err {
// 				t.Fatal(`Delete item "AB" should work as expected!`)
// 			}
// 		}

// 		tdb.GenDotFile("./test_mainworkflow_3_4.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        true,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: true,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                isNotNil{},
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   4,
// 					prefix_child_nodes_dirty: true,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_2AB_ := _1A.prefix_child_nodes.btree.Get(uint8('B'))
// 					_2AB := _2AB_.(*Node)

// 					_2ABTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           nil,
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                isNotNil{},
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   2,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AB)

// 					for _, tt := range _2ABTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_3ABC_ := _2AB.prefix_child_nodes.btree.Get(uint8('C'))
// 						_3ABC := _3ABC_.(*Node)

// 						_3ABCTests := Expect{
// 							prefix:       []byte("C"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_ABC"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           isNotNil{},
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABC)

// 						for _, tt := range _3ABCTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 						_3ABa_ := _2AB.prefix_child_nodes.btree.Get(uint8('a'))
// 						_3ABa := _3ABa_.(*Node)

// 						_3ABaTests := Expect{
// 							prefix:       []byte("a"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                nil,
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: false,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABa)

// 						for _, tt := range _3ABaTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 					_2AC_ := _1A.prefix_child_nodes.btree.Get(uint8('C'))
// 					_2AC := _2AC_.(*Node)

// 					_2ACTests := Expect{
// 						prefix:       []byte("C"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AC)

// 					for _, tt := range _2ACTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AD_ := _1A.prefix_child_nodes.btree.Get(uint8('D'))
// 					_2AD := _2AD_.(*Node)

// 					_2ADTests := Expect{
// 						prefix:       []byte("D"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AD)

// 					for _, tt := range _2ADTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2Aa_ := _1A.prefix_child_nodes.btree.Get(uint8('a'))
// 					_2Aa := _2Aa_.(*Node)

// 					_2AaTests := Expect{
// 						prefix:       []byte("a"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2Aa)

// 					for _, tt := range _2AaTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 			}
// 		}

// 		{
// 			tdb.Put(Path([]byte("ACa")), []byte("val_ACa"), true)
// 			tdb.Put(Path([]byte("ADa")), []byte("val_ADa"), true)

// 			_rootHash, err := tdb.testCommit()

// 			if nil != err {
// 				t.Fatal(err)
// 			}

// 			rootHash = _rootHash
// 		}

// 		tdb.GenDotFile("./test_mainworkflow_3_5.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        true,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: true,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                isNotNil{},
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   4,
// 					prefix_child_nodes_dirty: true,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_2AB_ := _1A.prefix_child_nodes.btree.Get(uint8('B'))
// 					_2AB := _2AB_.(*Node)

// 					_2ABTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           nil,
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                isNotNil{},
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   2,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AB)

// 					for _, tt := range _2ABTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_3ABC_ := _2AB.prefix_child_nodes.btree.Get(uint8('C'))
// 						_3ABC := _3ABC_.(*Node)

// 						_3ABCTests := Expect{
// 							prefix:       []byte("C"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_ABC"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           isNotNil{},
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABC)

// 						for _, tt := range _3ABCTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 						_3ABa_ := _2AB.prefix_child_nodes.btree.Get(uint8('a'))
// 						_3ABa := _3ABa_.(*Node)

// 						_3ABaTests := Expect{
// 							prefix:       []byte("a"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_2AB.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                nil,
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: false,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ABa)

// 						for _, tt := range _3ABaTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 					_2AC_ := _1A.prefix_child_nodes.btree.Get(uint8('C'))
// 					_2AC := _2AC_.(*Node)

// 					_2ACTests := Expect{
// 						prefix:       []byte("C"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                isNotNil{},
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   1,
// 						prefix_child_nodes_dirty: true,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AC)

// 					for _, tt := range _2ACTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_3ACa_ := _2AC.prefix_child_nodes.btree.Get(uint8('a'))
// 						_3ACa := _3ACa_.(*Node)

// 						_3ACaTests := Expect{
// 							prefix:       []byte("a"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_2AC.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_ACa"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          true,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: true,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ACa)

// 						for _, tt := range _3ACaTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 					_2AD_ := _1A.prefix_child_nodes.btree.Get(uint8('D'))
// 					_2AD := _2AD_.(*Node)

// 					_2ADTests := Expect{
// 						prefix:       []byte("D"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                isNotNil{},
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   1,
// 						prefix_child_nodes_dirty: true,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AD)

// 					for _, tt := range _2ADTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_3ADa_ := _2AD.prefix_child_nodes.btree.Get(uint8('a'))
// 						_3ADa := _3ADa_.(*Node)

// 						_3ADaTests := Expect{
// 							prefix:       []byte("a"),
// 							dirty:        true,
// 							parent_nodes: isEqualTo{_2AD.prefix_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_ADa"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          true,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: true,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: true,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_3ADa)

// 						for _, tt := range _3ADaTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 					_2Aa_ := _1A.prefix_child_nodes.btree.Get(uint8('a'))
// 					_2Aa := _2Aa_.(*Node)

// 					_2AaTests := Expect{
// 						prefix:       []byte("a"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2Aa)

// 					for _, tt := range _2AaTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 			}
// 		}
// 		testCloseTrieDB(tdb)
// 	}

// 	// test Delete the child node which parent node has only one child
// 	{
// 		tdb, err := testPrepareTrieDB(db_path, rootHash)

// 		if nil != err {
// 			t.Fatal(err)
// 		}

// 		_root := tdb.root_node

// 		tdb.GenDotFile("./test_mainworkflow_4_1.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                nil,
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: false,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   0,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}
// 		}

// 		{
// 			_, err := tdb.Del(Path([]byte("ACa")))
// 			if nil != err {
// 				t.Fatal(`Delete item "ACa" should work as expected!`, err)
// 			}
// 		}
// 		tdb.GenDotFile("./test_mainworkflow_4_2.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        true,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: true,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                isNotNil{},
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   4,
// 					prefix_child_nodes_dirty: true,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_2AB_ := _1A.prefix_child_nodes.btree.Get(uint8('B'))
// 					_2AB := _2AB_.(*Node)

// 					_2ABTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           nil,
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AB)

// 					for _, tt := range _2ABTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AC_ := _1A.prefix_child_nodes.btree.Get(uint8('C'))
// 					_2AC := _2AC_.(*Node)

// 					_2ACTests := Expect{
// 						prefix:       []byte("C"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AC)

// 					for _, tt := range _2ACTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AD_ := _1A.prefix_child_nodes.btree.Get(uint8('D'))
// 					_2AD := _2AD_.(*Node)

// 					_2ADTests := Expect{
// 						prefix:       []byte("D"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AD)

// 					for _, tt := range _2ADTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2Aa_ := _1A.prefix_child_nodes.btree.Get(uint8('a'))
// 					_2Aa := _2Aa_.(*Node)

// 					_2AaTests := Expect{
// 						prefix:       []byte("a"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2Aa)

// 					for _, tt := range _2AaTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 			}
// 		}

// 		{
// 			_, _, err := tdb.Get(Path([]byte("ACa")))
// 			if nil == err {
// 				t.Fatal(`Get item "ACa" should report error!`, err)
// 			}
// 		}

// 		tdb.GenDotFile("./test_mainworkflow_4_3.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        true,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: true,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                isNotNil{},
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   4,
// 					prefix_child_nodes_dirty: true,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_2AB_ := _1A.prefix_child_nodes.btree.Get(uint8('B'))
// 					_2AB := _2AB_.(*Node)

// 					_2ABTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           nil,
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AB)

// 					for _, tt := range _2ABTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AC_ := _1A.prefix_child_nodes.btree.Get(uint8('C'))
// 					_2AC := _2AC_.(*Node)

// 					_2ACTests := Expect{
// 						prefix:       []byte("C"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AC)

// 					for _, tt := range _2ACTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AD_ := _1A.prefix_child_nodes.btree.Get(uint8('D'))
// 					_2AD := _2AD_.(*Node)

// 					_2ADTests := Expect{
// 						prefix:       []byte("D"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AD)

// 					for _, tt := range _2ADTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2Aa_ := _1A.prefix_child_nodes.btree.Get(uint8('a'))
// 					_2Aa := _2Aa_.(*Node)

// 					_2AaTests := Expect{
// 						prefix:       []byte("a"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2Aa)

// 					for _, tt := range _2AaTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 			}
// 		}

// 		{
// 			_rootHash, err := tdb.testCommit()

// 			if nil != err {
// 				t.Fatal(err)
// 			}

// 			rootHash = _rootHash
// 		}

// 		tdb.GenDotFile("./test_mainworkflow_4_4.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        true,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: true,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                isNotNil{},
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   4,
// 					prefix_child_nodes_dirty: true,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_2AB_ := _1A.prefix_child_nodes.btree.Get(uint8('B'))
// 					_2AB := _2AB_.(*Node)

// 					_2ABTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           nil,
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AB)

// 					for _, tt := range _2ABTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AC_ := _1A.prefix_child_nodes.btree.Get(uint8('C'))
// 					_2AC := _2AC_.(*Node)

// 					_2ACTests := Expect{
// 						prefix:       []byte("C"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: true,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AC)

// 					for _, tt := range _2ACTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2AD_ := _1A.prefix_child_nodes.btree.Get(uint8('D'))
// 					_2AD := _2AD_.(*Node)

// 					_2ADTests := Expect{
// 						prefix:       []byte("D"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           isNotNil{},
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2AD)

// 					for _, tt := range _2ADTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					_2Aa_ := _1A.prefix_child_nodes.btree.Get(uint8('a'))
// 					_2Aa := _2Aa_.(*Node)

// 					_2AaTests := Expect{
// 						prefix:       []byte("a"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.prefix_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                nil,
// 						folder_child_nodes_hash:           nil,
// 						folder_child_nodes_hash_recovered: false,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   0,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_2Aa)

// 					for _, tt := range _2AaTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 			}
// 		}
// 		testCloseTrieDB(tdb)
// 	}

// 	// test Delete intermediate node which has child but HAS NO value
// 	{
// 		tdb, err := testPrepareTrieDB(db_path, rootHash)

// 		if nil != err {
// 			t.Fatal(err)
// 		}

// 		_root := tdb.root_node

// 		tdb.GenDotFile("./test_mainworkflow_5_1.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                nil,
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: false,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   0,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}
// 		}

// 		{
// 			_, _, err := tdb.Get(Path([]byte("A"), []byte("A"), []byte("A")))

// 			if nil != err {
// 				t.Fatal("Get path [A,A,A] should NOT trigger error!")
// 			}
// 		}

// 		tdb.GenDotFile("./test_mainworkflow_5_2.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                isNotNil{},
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   1,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_1A1A_ := _1A.folder_child_nodes.btree.Get(uint8('A'))
// 					_1A1A := _1A1A_.(*Node)

// 					_1A1ATests := Expect{
// 						prefix:       []byte("A"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.folder_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           nil,
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                isNotNil{},
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   1,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_1A1A)

// 					for _, tt := range _1A1ATests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_1A1A1A_ := _1A1A.folder_child_nodes.btree.Get(uint8('A'))
// 						_1A1A1A := _1A1A1A_.(*Node)

// 						_1A1A1ATests := Expect{
// 							prefix:       []byte("A"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_1A1A.folder_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_A_A_A"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_1A1A1A)

// 						for _, tt := range _1A1A1ATests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 			}
// 		}

// 		{
// 			_, err := tdb.Del(Path([]byte("A"), []byte("A")))

// 			if nil != err {
// 				t.Fatal("Delete path [A,A] should NOT trigger error!")
// 			}
// 		}

// 		tdb.GenDotFile("./test_mainworkflow_5_3.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                isNotNil{},
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   1,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_1A1A_ := _1A.folder_child_nodes.btree.Get(uint8('A'))
// 					_1A1A := _1A1A_.(*Node)

// 					_1A1ATests := Expect{
// 						prefix:       []byte("A"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.folder_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           nil,
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                isNotNil{},
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   1,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_1A1A)

// 					for _, tt := range _1A1ATests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_1A1A1A_ := _1A1A.folder_child_nodes.btree.Get(uint8('A'))
// 						_1A1A1A := _1A1A1A_.(*Node)

// 						_1A1A1ATests := Expect{
// 							prefix:       []byte("A"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_1A1A.folder_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_A_A_A"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_1A1A1A)

// 						for _, tt := range _1A1A1ATests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 			}
// 		}

// 		{
// 			_rootHash, err := tdb.testCommit()

// 			if nil != err {
// 				t.Fatal(err)
// 			}

// 			rootHash = _rootHash
// 		}

// 		tdb.GenDotFile("./test_mainworkflow_5_4.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                isNotNil{},
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   1,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_1A1A_ := _1A.folder_child_nodes.btree.Get(uint8('A'))
// 					_1A1A := _1A1A_.(*Node)

// 					_1A1ATests := Expect{
// 						prefix:       []byte("A"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1A.folder_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           nil,
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                isNotNil{},
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   1,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_1A1A)

// 					for _, tt := range _1A1ATests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_1A1A1A_ := _1A1A.folder_child_nodes.btree.Get(uint8('A'))
// 						_1A1A1A := _1A1A1A_.(*Node)

// 						_1A1A1ATests := Expect{
// 							prefix:       []byte("A"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_1A1A.folder_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_A_A_A"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_1A1A1A)

// 						for _, tt := range _1A1A1ATests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 			}
// 		}
// 		testCloseTrieDB(tdb)
// 	}

// 	// test Delete intermediate node which has child and HAS value
// 	{
// 		tdb, err := testPrepareTrieDB(db_path, rootHash)

// 		if nil != err {
// 			t.Fatal(err)
// 		}

// 		_root := tdb.root_node

// 		tdb.GenDotFile("./test_mainworkflow_6_1.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                nil,
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: false,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   0,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}
// 		}

// 		{
// 			_, _, err := tdb.Get(Path([]byte("B"), []byte("B"), []byte("B")))

// 			if nil != err {
// 				t.Fatal("Get path [B,B,B] should NOT trigger error!")
// 			}
// 		}

// 		tdb.GenDotFile("./test_mainworkflow_6_2.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                isNotNil{},
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   1,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_1B1B_ := _1B.folder_child_nodes.btree.Get(uint8('B'))
// 					_1B1B := _1B1B_.(*Node)

// 					_1B1BTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        false,
// 						parent_nodes: isEqualTo{_1B.folder_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           isNotNil{},
// 						val_hash_recovered: false,
// 						val_dirty:          false,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                isNotNil{},
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   1,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_1B1B)

// 					for _, tt := range _1B1BTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_1B1B1B_ := _1B1B.folder_child_nodes.btree.Get(uint8('B'))
// 						_1B1B1B := _1B1B1B_.(*Node)

// 						_1B1B1BTests := Expect{
// 							prefix:       []byte("B"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_1B1B.folder_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_B_B_B"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_1B1B1B)

// 						for _, tt := range _1B1B1BTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 				}

// 			}
// 		}

// 		{
// 			_, err := tdb.Del(Path([]byte("B"), []byte("B")))

// 			if nil != err {
// 				t.Fatal("Delete path [B,B] should NOT trigger error!")
// 			}
// 		}

// 		tdb.GenDotFile("./test_mainworkflow_6_3.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        true,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: true,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                isNotNil{},
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   1,
// 					folder_child_nodes_dirty: true,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_1B1B_ := _1B.folder_child_nodes.btree.Get(uint8('B'))
// 					_1B1B := _1B1B_.(*Node)

// 					_1B1BTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1B.folder_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           nil,
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                isNotNil{},
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   1,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_1B1B)

// 					for _, tt := range _1B1BTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_1B1B1B_ := _1B1B.folder_child_nodes.btree.Get(uint8('B'))
// 						_1B1B1B := _1B1B1B_.(*Node)

// 						_1B1B1BTests := Expect{
// 							prefix:       []byte("B"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_1B1B.folder_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_B_B_B"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_1B1B1B)

// 						for _, tt := range _1B1B1BTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 				}

// 			}
// 		}

// 		{
// 			_rootHash, err := tdb.testCommit()

// 			if nil != err {
// 				t.Fatal(err)
// 			}

// 			rootHash = _rootHash
// 		}

// 		tdb.GenDotFile("./test_mainworkflow_6_4.dot", false)
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        true,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: true,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                isNotNil{},
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   1,
// 					folder_child_nodes_dirty: true,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				{
// 					_1B1B_ := _1B.folder_child_nodes.btree.Get(uint8('B'))
// 					_1B1B := _1B1B_.(*Node)

// 					_1B1BTests := Expect{
// 						prefix:       []byte("B"),
// 						dirty:        true,
// 						parent_nodes: isEqualTo{_1B.folder_child_nodes},

// 						node_bytes: isNotNil{},
// 						node_hash:  isNotNil{},

// 						val:                nil,
// 						val_hash:           nil,
// 						val_hash_recovered: true,
// 						val_dirty:          true,

// 						prefix_child_nodes:                nil,
// 						prefix_child_nodes_hash:           nil,
// 						prefix_child_nodes_hash_recovered: false,

// 						folder_child_nodes:                isNotNil{},
// 						folder_child_nodes_hash:           isNotNil{},
// 						folder_child_nodes_hash_recovered: true,

// 						prefix_child_nodes_len:   0,
// 						prefix_child_nodes_dirty: false,

// 						folder_child_nodes_len:   1,
// 						folder_child_nodes_dirty: false,
// 					}.makeTests(_1B1B)

// 					for _, tt := range _1B1BTests {
// 						if !tt.ok {
// 							t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 						}
// 					}

// 					{
// 						_1B1B1B_ := _1B1B.folder_child_nodes.btree.Get(uint8('B'))
// 						_1B1B1B := _1B1B1B_.(*Node)

// 						_1B1B1BTests := Expect{
// 							prefix:       []byte("B"),
// 							dirty:        false,
// 							parent_nodes: isEqualTo{_1B1B.folder_child_nodes},

// 							node_bytes: isNotNil{},
// 							node_hash:  isNotNil{},

// 							val:                []byte("val_B_B_B"),
// 							val_hash:           isNotNil{},
// 							val_hash_recovered: true,
// 							val_dirty:          false,

// 							prefix_child_nodes:                nil,
// 							prefix_child_nodes_hash:           nil,
// 							prefix_child_nodes_hash_recovered: false,

// 							folder_child_nodes:                nil,
// 							folder_child_nodes_hash:           nil,
// 							folder_child_nodes_hash_recovered: false,

// 							prefix_child_nodes_len:   0,
// 							prefix_child_nodes_dirty: false,

// 							folder_child_nodes_len:   0,
// 							folder_child_nodes_dirty: false,
// 						}.makeTests(_1B1B1B)

// 						for _, tt := range _1B1B1BTests {
// 							if !tt.ok {
// 								t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 							}
// 						}

// 					}

// 				}

// 			}
// 		}
// 		testCloseTrieDB(tdb)
// 	}

// 	// test Delete data which not exist in a nonempty triedb
// 	{
// 		tdb, err := testPrepareTrieDB(db_path, rootHash)

// 		if nil != err {
// 			t.Fatal(err)
// 		}

// 		_root := tdb.root_node
// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                nil,
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: false,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   0,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}
// 		}

// 		{
// 			_, err := tdb.Del(Path([]byte("hello")))

// 			if nil != err {
// 				t.Fatal("Delete path [hello] should NOT trigger error!")
// 			}
// 		}

// 		// _root.genCheckStatements(os.Stdout)
// 		{
// 			_rootTests := Expect{
// 				prefix:       nil,
// 				dirty:        false,
// 				parent_nodes: nil,

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: false,
// 				val_dirty:          false,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: false,

// 				folder_child_nodes:                isNotNil{},
// 				folder_child_nodes_hash:           isNotNil{},
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   2,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_root)

// 			for _, tt := range _rootTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_1A_ := _root.folder_child_nodes.btree.Get(uint8('A'))
// 				_1A := _1A_.(*Node)

// 				_1ATests := Expect{
// 					prefix:       []byte("A"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           isNotNil{},
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1A)

// 				for _, tt := range _1ATests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				_1B_ := _root.folder_child_nodes.btree.Get(uint8('B'))
// 				_1B := _1B_.(*Node)

// 				_1BTests := Expect{
// 					prefix:       []byte("B"),
// 					dirty:        false,
// 					parent_nodes: isEqualTo{_root.folder_child_nodes},

// 					node_bytes: isNotNil{},
// 					node_hash:  isNotNil{},

// 					val:                nil,
// 					val_hash:           isNotNil{},
// 					val_hash_recovered: false,
// 					val_dirty:          false,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: false,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           isNotNil{},
// 					folder_child_nodes_hash_recovered: false,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_1B)

// 				for _, tt := range _1BTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 			}
// 		}

// 		testCloseTrieDB(tdb)
// 	}
// }

// // test delete data on empty triedb
// func TestDeleteOnEmptyTrieDB(t *testing.T) {

// 	const db_path = "triedb_deleteonempty_test.db"
// 	defer os.RemoveAll(db_path)

// 	tdb, err := testPrepareTrieDB(db_path, nil)

// 	if nil != err {
// 		t.Fatal(err)
// 	}

// 	_root := tdb.root_node
// 	// _root.genCheckStatements(os.Stdout)
// 	{
// 		_rootTests := Expect{
// 			prefix:       nil,
// 			dirty:        false,
// 			parent_nodes: nil,

// 			node_bytes: nil,
// 			node_hash:  nil,

// 			val:                nil,
// 			val_hash:           nil,
// 			val_hash_recovered: true,
// 			val_dirty:          false,

// 			prefix_child_nodes:                nil,
// 			prefix_child_nodes_hash:           nil,
// 			prefix_child_nodes_hash_recovered: true,

// 			folder_child_nodes:                nil,
// 			folder_child_nodes_hash:           nil,
// 			folder_child_nodes_hash_recovered: true,

// 			prefix_child_nodes_len:   0,
// 			prefix_child_nodes_dirty: false,

// 			folder_child_nodes_len:   0,
// 			folder_child_nodes_dirty: false,
// 		}.makeTests(_root)

// 		for _, tt := range _rootTests {
// 			if !tt.ok {
// 				t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 			}
// 		}
// 	}

// 	{
// 		_, err := tdb.Del(Path([]byte("hello")))

// 		if nil != err {
// 			t.Fatal("Delete path [hello] should NOT trigger error!")
// 		}
// 	}
// 	// _root.genCheckStatements(os.Stdout)
// 	{
// 		_rootTests := Expect{
// 			prefix:       nil,
// 			dirty:        false,
// 			parent_nodes: nil,

// 			node_bytes: nil,
// 			node_hash:  nil,

// 			val:                nil,
// 			val_hash:           nil,
// 			val_hash_recovered: true,
// 			val_dirty:          false,

// 			prefix_child_nodes:                nil,
// 			prefix_child_nodes_hash:           nil,
// 			prefix_child_nodes_hash_recovered: true,

// 			folder_child_nodes:                nil,
// 			folder_child_nodes_hash:           nil,
// 			folder_child_nodes_hash_recovered: true,

// 			prefix_child_nodes_len:   0,
// 			prefix_child_nodes_dirty: false,

// 			folder_child_nodes_len:   0,
// 			folder_child_nodes_dirty: false,
// 		}.makeTests(_root)

// 		for _, tt := range _rootTests {
// 			if !tt.ok {
// 				t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 			}
// 		}
// 	}

// 	{
// 		_rootHash, err := tdb.testCommit()

// 		if nil != err {
// 			t.Fatal(err)
// 		}

// 		if !hash.IsNilHash(_rootHash) {
// 			t.Errorf("Delete on an empty triedb, then calc hash, root hash should be: %#v, but: %#v", nil, _rootHash)
// 		}
// 	}

// 	// _root.genCheckStatements(os.Stdout)
// 	{
// 		_rootTests := Expect{
// 			prefix:       nil,
// 			dirty:        false,
// 			parent_nodes: nil,

// 			node_bytes: nil,
// 			node_hash:  nil,

// 			val:                nil,
// 			val_hash:           nil,
// 			val_hash_recovered: true,
// 			val_dirty:          false,

// 			prefix_child_nodes:                nil,
// 			prefix_child_nodes_hash:           nil,
// 			prefix_child_nodes_hash_recovered: true,

// 			folder_child_nodes:                nil,
// 			folder_child_nodes_hash:           nil,
// 			folder_child_nodes_hash_recovered: true,

// 			prefix_child_nodes_len:   0,
// 			prefix_child_nodes_dirty: false,

// 			folder_child_nodes_len:   0,
// 			folder_child_nodes_dirty: false,
// 		}.makeTests(_root)

// 		for _, tt := range _rootTests {
// 			if !tt.ok {
// 				t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 			}
// 		}
// 	}

// 	testCloseTrieDB(tdb)
// }

// func TestLongPath(t *testing.T) {
// 	const db_path = "./triedb_longpath_test.db"
// 	defer os.RemoveAll(db_path)

// 	tdb, err := testPrepareTrieDB(db_path, nil)

// 	if nil != err {
// 		t.Fatal(err)
// 	}
// 	_root := tdb.root_node

// 	// _root.genCheckStatements(os.Stdout)
// 	{
// 		_rootTests := Expect{
// 			prefix:       nil,
// 			dirty:        false,
// 			parent_nodes: nil,

// 			node_bytes: nil,
// 			node_hash:  nil,

// 			val:                nil,
// 			val_hash:           nil,
// 			val_hash_recovered: true,
// 			val_dirty:          false,

// 			prefix_child_nodes:                nil,
// 			prefix_child_nodes_hash:           nil,
// 			prefix_child_nodes_hash_recovered: true,

// 			folder_child_nodes:                nil,
// 			folder_child_nodes_hash:           nil,
// 			folder_child_nodes_hash_recovered: true,

// 			prefix_child_nodes_len:   0,
// 			prefix_child_nodes_dirty: false,

// 			folder_child_nodes_len:   0,
// 			folder_child_nodes_dirty: false,
// 		}.makeTests(_root)

// 		for _, tt := range _rootTests {
// 			if !tt.ok {
// 				t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 			}
// 		}
// 	}

// 	{
// 		err := tdb.Put(Path([]byte("hello")), []byte("val_hello"), true)
// 		if nil != err {
// 			t.Fatal(err)
// 		}
// 	}

// 	tdb.GenDotFile("./test_longpath_1.dot", false)
// 	// _root.genCheckStatements(os.Stdout)
// 	{
// 		_rootTests := Expect{
// 			prefix:       nil,
// 			dirty:        true,
// 			parent_nodes: nil,

// 			node_bytes: nil,
// 			node_hash:  nil,

// 			val:                nil,
// 			val_hash:           nil,
// 			val_hash_recovered: true,
// 			val_dirty:          false,

// 			prefix_child_nodes:                nil,
// 			prefix_child_nodes_hash:           nil,
// 			prefix_child_nodes_hash_recovered: true,

// 			folder_child_nodes:                isNotNil{},
// 			folder_child_nodes_hash:           nil,
// 			folder_child_nodes_hash_recovered: true,

// 			prefix_child_nodes_len:   0,
// 			prefix_child_nodes_dirty: false,

// 			folder_child_nodes_len:   1,
// 			folder_child_nodes_dirty: true,
// 		}.makeTests(_root)

// 		for _, tt := range _rootTests {
// 			if !tt.ok {
// 				t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 			}
// 		}

// 		{
// 			_5hello_ := _root.folder_child_nodes.btree.Get(uint8('h'))
// 			_5hello := _5hello_.(*Node)

// 			_5helloTests := Expect{
// 				prefix:       []byte("hello"),
// 				dirty:        true,
// 				parent_nodes: isEqualTo{_root.folder_child_nodes},

// 				node_bytes: nil,
// 				node_hash:  nil,

// 				val:                []byte("val_hello"),
// 				val_hash:           nil,
// 				val_hash_recovered: true,
// 				val_dirty:          true,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: true,

// 				folder_child_nodes:                nil,
// 				folder_child_nodes_hash:           nil,
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   0,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_5hello)

// 			for _, tt := range _5helloTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 		}
// 	}

// 	{
// 		err := tdb.Put(Path([]byte("hellO")), []byte("val_hellO"), true)
// 		if nil != err {
// 			t.Fatal(err)
// 		}
// 	}

// 	tdb.GenDotFile("./test_longpath_2.dot", false)
// 	// _root.genCheckStatements(os.Stdout)
// 	{
// 		_rootTests := Expect{
// 			prefix:       nil,
// 			dirty:        true,
// 			parent_nodes: nil,

// 			node_bytes: nil,
// 			node_hash:  nil,

// 			val:                nil,
// 			val_hash:           nil,
// 			val_hash_recovered: true,
// 			val_dirty:          false,

// 			prefix_child_nodes:                nil,
// 			prefix_child_nodes_hash:           nil,
// 			prefix_child_nodes_hash_recovered: true,

// 			folder_child_nodes:                isNotNil{},
// 			folder_child_nodes_hash:           nil,
// 			folder_child_nodes_hash_recovered: true,

// 			prefix_child_nodes_len:   0,
// 			prefix_child_nodes_dirty: false,

// 			folder_child_nodes_len:   1,
// 			folder_child_nodes_dirty: true,
// 		}.makeTests(_root)

// 		for _, tt := range _rootTests {
// 			if !tt.ok {
// 				t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 			}
// 		}

// 		{
// 			_4hell_ := _root.folder_child_nodes.btree.Get(uint8('h'))
// 			_4hell := _4hell_.(*Node)

// 			_4hellTests := Expect{
// 				prefix:       []byte("hell"),
// 				dirty:        true,
// 				parent_nodes: isEqualTo{_root.folder_child_nodes},

// 				node_bytes: nil,
// 				node_hash:  nil,

// 				val:                nil,
// 				val_hash:           nil,
// 				val_hash_recovered: true,
// 				val_dirty:          false,

// 				prefix_child_nodes:                isNotNil{},
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: true,

// 				folder_child_nodes:                nil,
// 				folder_child_nodes_hash:           nil,
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   2,
// 				prefix_child_nodes_dirty: true,

// 				folder_child_nodes_len:   0,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_4hell)

// 			for _, tt := range _4hellTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 			{
// 				_5hellO_ := _4hell.prefix_child_nodes.btree.Get(uint8('O'))
// 				_5hellO := _5hellO_.(*Node)

// 				_5hellOTests := Expect{
// 					prefix:       []byte("O"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_4hell.prefix_child_nodes},

// 					node_bytes: nil,
// 					node_hash:  nil,

// 					val:                []byte("val_hellO"),
// 					val_hash:           nil,
// 					val_hash_recovered: true,
// 					val_dirty:          true,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           nil,
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_5hellO)

// 				for _, tt := range _5hellOTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 				_5hello_ := _4hell.prefix_child_nodes.btree.Get(uint8('o'))
// 				_5hello := _5hello_.(*Node)

// 				_5helloTests := Expect{
// 					prefix:       []byte("o"),
// 					dirty:        true,
// 					parent_nodes: isEqualTo{_4hell.prefix_child_nodes},

// 					node_bytes: nil,
// 					node_hash:  nil,

// 					val:                []byte("val_hello"),
// 					val_hash:           nil,
// 					val_hash_recovered: true,
// 					val_dirty:          true,

// 					prefix_child_nodes:                nil,
// 					prefix_child_nodes_hash:           nil,
// 					prefix_child_nodes_hash_recovered: true,

// 					folder_child_nodes:                nil,
// 					folder_child_nodes_hash:           nil,
// 					folder_child_nodes_hash_recovered: true,

// 					prefix_child_nodes_len:   0,
// 					prefix_child_nodes_dirty: false,

// 					folder_child_nodes_len:   0,
// 					folder_child_nodes_dirty: false,
// 				}.makeTests(_5hello)

// 				for _, tt := range _5helloTests {
// 					if !tt.ok {
// 						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 					}
// 				}

// 			}

// 		}
// 	}

// 	{
// 		_, err := tdb.Del(Path([]byte("hello")))
// 		if nil != err {
// 			t.Fatal(err)
// 		}
// 	}

// 	tdb.testCommit()
// 	tdb.GenDotFile("./test_longpath_3.dot", false)
// 	// _root.genCheckStatements(os.Stdout)
// 	{
// 		_rootTests := Expect{
// 			prefix:       nil,
// 			dirty:        true,
// 			parent_nodes: nil,

// 			node_bytes: isNotNil{},
// 			node_hash:  isNotNil{},

// 			val:                nil,
// 			val_hash:           nil,
// 			val_hash_recovered: true,
// 			val_dirty:          false,

// 			prefix_child_nodes:                nil,
// 			prefix_child_nodes_hash:           nil,
// 			prefix_child_nodes_hash_recovered: true,

// 			folder_child_nodes:                isNotNil{},
// 			folder_child_nodes_hash:           isNotNil{},
// 			folder_child_nodes_hash_recovered: true,

// 			prefix_child_nodes_len:   0,
// 			prefix_child_nodes_dirty: false,

// 			folder_child_nodes_len:   1,
// 			folder_child_nodes_dirty: true,
// 		}.makeTests(_root)

// 		for _, tt := range _rootTests {
// 			if !tt.ok {
// 				t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 			}
// 		}

// 		{
// 			_5hellO_ := _root.folder_child_nodes.btree.Get(uint8('h'))
// 			_5hellO := _5hellO_.(*Node)

// 			_5hellOTests := Expect{
// 				prefix:       []byte("hellO"),
// 				dirty:        true,
// 				parent_nodes: isEqualTo{_root.folder_child_nodes},

// 				node_bytes: isNotNil{},
// 				node_hash:  isNotNil{},

// 				val:                []byte("val_hellO"),
// 				val_hash:           isNotNil{},
// 				val_hash_recovered: true,
// 				val_dirty:          true,

// 				prefix_child_nodes:                nil,
// 				prefix_child_nodes_hash:           nil,
// 				prefix_child_nodes_hash_recovered: true,

// 				folder_child_nodes:                nil,
// 				folder_child_nodes_hash:           nil,
// 				folder_child_nodes_hash_recovered: true,

// 				prefix_child_nodes_len:   0,
// 				prefix_child_nodes_dirty: false,

// 				folder_child_nodes_len:   0,
// 				folder_child_nodes_dirty: false,
// 			}.makeTests(_5hellO)

// 			for _, tt := range _5hellOTests {
// 				if !tt.ok {
// 					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
// 				}
// 			}

// 		}
// 	}
// 	testCloseTrieDB(tdb)
// }
