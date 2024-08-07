package triedb

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"text/template"

	crand "crypto/rand"
	mrand "math/rand"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/hash"
	"github.com/xlander-io/kv"
	"github.com/xlander-io/kv_leveldb"
)

func testPrepareTrieDB(dataPath string, rootHash *hash.Hash) (*TrieDB, error) {
	kvdb, err := kv_leveldb.NewDB(dataPath)
	if nil != err {
		return nil, err
	}
	//
	c, err := cache.New(nil)
	if nil != err {
		return nil, err
	}
	tdb, err := NewTrieDB(kvdb, c, &TrieDBConfig{
		Root_hash:           rootHash,
		Commit_thread_limit: 1,
	})

	return tdb, err
}

func testCloseTrieDB(tdb *TrieDB) {
	tdb.kvdb.Close()
}

func (tdb *TrieDB) testCommit() (*hash.Hash, error) {
	rootHash, toUpdate, toDel, err := tdb.Commit()
	if err != nil {
		return nil, err
	}

	b := kv.NewBatch()
	for hex_string, update_v := range toUpdate {
		b.Put([]byte(hex_string), update_v)
	}

	for _, del_v := range toDel {
		b.Delete(del_v.Bytes())
	}

	err = tdb.kvdb.WriteBatch(b, true)

	if nil != err {
		return nil, err
	}

	return rootHash, err
}

type TEST struct {
	label    string
	expected interface{}
	actual   interface{}
	ok       bool
}

type doCheck interface {
	isOk(o interface{}) bool
}

type isNotNil struct{}

type isEqualTo struct {
	object interface{}
}

func (x isNotNil) isOk(o any) bool {

	if O, ok := o.(*hash.Hash); ok {
		return !hash.IsNilHash(O)
	}

	if nil != o {
		return !reflect.ValueOf(o).IsNil()
	} else {
		return false
	}
}

func (x isEqualTo) isOk(o any) bool {
	if O, ok := o.([]byte); ok {
		if X, ok := x.object.([]byte); ok {
			return bytes.Equal(X, O)
		} else {
			return false
		}
	}
	return x.object == o
}

type Expect struct {
	prefix       []byte
	dirty        bool
	parent_nodes doCheck

	node_bytes doCheck
	node_hash  doCheck
	//node_hash_recovered bool

	val                []byte
	val_hash           doCheck
	val_hash_recovered bool
	val_dirty          bool

	prefix_child_nodes                doCheck
	prefix_child_nodes_hash           doCheck
	prefix_child_nodes_hash_recovered bool

	folder_child_nodes                doCheck
	folder_child_nodes_hash           doCheck
	folder_child_nodes_hash_recovered bool

	prefix_child_nodes_len   int
	prefix_child_nodes_dirty bool

	folder_child_nodes_len   int
	folder_child_nodes_dirty bool
}

func (ex Expect) makeTests(o *Node) []TEST {
	tests := []TEST{
		{"prefix", ex.prefix, o.prefix, bytes.Equal(ex.prefix, o.prefix)},
		{"dirty", ex.dirty, o.dirty, ex.dirty == o.dirty},

		{"val_dirty", ex.val_dirty, o.val_dirty, ex.val_dirty == o.val_dirty},
		{"val_hash_recovered", ex.val_hash_recovered, o.val_hash_recovered, ex.val_hash_recovered == o.val_hash_recovered},
		{"prefix_child_nodes_hash_recovered", ex.prefix_child_nodes_hash_recovered, o.prefix_child_nodes_hash_recovered, ex.prefix_child_nodes_hash_recovered == o.prefix_child_nodes_hash_recovered},
		{"folder_child_nodes_hash_recovered", ex.folder_child_nodes_hash_recovered, o.folder_child_nodes_hash_recovered, ex.folder_child_nodes_hash_recovered == o.folder_child_nodes_hash_recovered},
	}

	if nil == ex.parent_nodes {
		tests = append(tests, TEST{"parent_nodes", ex.parent_nodes, o.parent_nodes, nil == o.parent_nodes})
	} else {
		tests = append(tests, TEST{"parent_nodes", ex.parent_nodes, o.parent_nodes, ex.parent_nodes.isOk(o.parent_nodes)})
	}

	if nil == ex.prefix_child_nodes {
		tests = append(tests, TEST{"prefix_child_nodes", ex.prefix_child_nodes, o.prefix_child_nodes, nil == o.prefix_child_nodes})
	} else {
		b := ex.prefix_child_nodes.isOk(o.prefix_child_nodes)
		tests = append(tests, TEST{"prefix_child_nodes", ex.prefix_child_nodes, o.prefix_child_nodes, b})
	}

	if nil == ex.folder_child_nodes {
		tests = append(tests, TEST{"folder_child_nodes", ex.folder_child_nodes, o.folder_child_nodes, nil == o.folder_child_nodes})
	} else {
		b := ex.folder_child_nodes.isOk(o.folder_child_nodes)
		tests = append(tests, TEST{"folder_child_nodes", ex.folder_child_nodes, o.folder_child_nodes, b})
	}

	if nil == ex.node_bytes {
		tests = append(tests, TEST{"node_bytes", ex.node_bytes, o.node_bytes, nil == o.node_bytes})
	} else {
		tests = append(tests, TEST{"node_bytes", ex.node_bytes, o.node_bytes, ex.node_bytes.isOk(o.node_bytes)})
	}

	if nil == ex.val {
		tests = append(tests, TEST{"val", ex.val, o.val, nil == o.val})
	} else {
		tests = append(tests, TEST{"val", ex.val, o.val, bytes.Equal(ex.val, o.val)})
	}

	if nil == ex.node_hash {
		tests = append(tests, TEST{"node_hash", ex.node_hash, o.node_hash, hash.IsNilHash(o.node_hash)})
	} else {
		tests = append(tests, TEST{"node_hash", ex.node_hash, o.node_hash, ex.node_hash.isOk(o.node_hash)})
	}

	if nil == ex.val_hash {
		tests = append(tests, TEST{"val_hash", ex.val_hash, o.val_hash, hash.IsNilHash(o.val_hash)})
	} else {
		tests = append(tests, TEST{"val_hash", ex.val_hash, o.val_hash, ex.val_hash.isOk(o.val_hash)})
	}

	if nil == ex.prefix_child_nodes_hash {
		tests = append(tests, TEST{"prefix_child_nodes_hash", ex.prefix_child_nodes_hash, o.prefix_child_nodes_hash, hash.IsNilHash(o.prefix_child_nodes_hash)})
	} else {
		tests = append(tests, TEST{"prefix_child_nodes_hash", ex.prefix_child_nodes_hash, o.prefix_child_nodes_hash, ex.prefix_child_nodes_hash.isOk(o.prefix_child_nodes_hash)})
	}

	if nil == ex.folder_child_nodes_hash {
		tests = append(tests, TEST{"folder_child_nodes_hash", ex.folder_child_nodes_hash, o.folder_child_nodes_hash, hash.IsNilHash(o.folder_child_nodes_hash)})
	} else {
		tests = append(tests, TEST{"folder_child_nodes_hash", ex.folder_child_nodes_hash, o.folder_child_nodes_hash, ex.folder_child_nodes_hash.isOk(o.folder_child_nodes_hash)})
	}

	if nil != o.prefix_child_nodes {
		tests = append(tests, []TEST{
			{"len(prefix_child_nodes)", ex.prefix_child_nodes_len, o.prefix_child_nodes.btree.Len(), ex.prefix_child_nodes_len == o.prefix_child_nodes.btree.Len()},
			{"prefix_child_nodes.dirty", ex.prefix_child_nodes_dirty, o.prefix_child_nodes.dirty, ex.prefix_child_nodes_dirty == o.prefix_child_nodes.dirty},
			{"prefix_child_nodes.parent_node", o, o.prefix_child_nodes.parent_node, o == o.prefix_child_nodes.parent_node},
		}...)
	}

	if nil != o.folder_child_nodes {
		tests = append(tests, []TEST{
			{"len(folder_child_nodes)", ex.folder_child_nodes_len, o.folder_child_nodes.btree.Len(), ex.folder_child_nodes_len == o.folder_child_nodes.btree.Len()},
			{"folder_child_nodes.dirty", ex.folder_child_nodes_dirty, o.folder_child_nodes.dirty, ex.folder_child_nodes_dirty == o.folder_child_nodes.dirty},
			{"folder_child_nodes.parent_node", o, o.folder_child_nodes.parent_node, o == o.folder_child_nodes.parent_node},
		}...)
	}

	return tests
}

func string2Bytes(x string) []byte {
	b := bytes.NewBufferString(x)
	return bytes.Clone(b.Bytes())
}
func bytes2String(x []byte) string {
	b := bytes.NewBuffer(x)
	return b.String()
}

func (n *Node) genCheckStatementsString() string {
	var buf bytes.Buffer
	n.genCheckStatements(&buf)
	return buf.String()
}

func (n *Node) genCheckStatements(b io.Writer) {

	funcMaps := template.FuncMap{
		"Name": func(n *Node) string {
			if nil == n.prefix {
				return "root"
			} else {
				return n.node_path_flat_str()
			}
		},
		"Bytes": func(n *Node, field string) string {
			if bytes.Equal([]byte("prefix"), string2Bytes(field)) {
				if nil != n.prefix {
					return fmt.Sprintf("[]byte(\"%s\")", n.prefix)
				}
			} else if bytes.Equal([]byte("val"), string2Bytes(field)) {
				if nil != n.val {
					return fmt.Sprintf("[]byte(\"%s\")", n.val)
				}
			} else if bytes.Equal([]byte("node_bytes"), string2Bytes(field)) {
				if nil != n.node_bytes {
					return fmt.Sprintf("[]byte(\"%s\")", n.node_bytes)
				}
			} else {
				fmt.Printf("ERROR: Bytes unexpected field [%#v]!", field)
			}

			return "nil"
		},
		"Boolean": func(n *Node, field string) string {
			if bytes.Equal([]byte("dirty"), string2Bytes(field)) {
				if false != n.dirty {
					return "true"
				}
			} else if bytes.Equal([]byte("val_hash_recovered"), string2Bytes(field)) {
				if false != n.val_hash_recovered {
					return "true"
				}
			} else if bytes.Equal([]byte("val_dirty"), string2Bytes(field)) {
				if false != n.val_dirty {
					return "true"
				}
			} else if bytes.Equal([]byte("prefix_child_nodes_hash_recovered"), string2Bytes(field)) {
				if false != n.prefix_child_nodes_hash_recovered {
					return "true"
				}
			} else if bytes.Equal([]byte("folder_child_nodes_hash_recovered"), string2Bytes(field)) {
				if false != n.folder_child_nodes_hash_recovered {
					return "true"
				}
			} else if bytes.Equal([]byte("prefix_child_nodes_dirty"), string2Bytes(field)) {
				if nil != n.prefix_child_nodes {
					if false != n.prefix_child_nodes.dirty {
						return "true"
					}
				}
			} else if bytes.Equal([]byte("folder_child_nodes_dirty"), string2Bytes(field)) {
				if nil != n.folder_child_nodes {
					if false != n.folder_child_nodes.dirty {
						return "true"
					}
				}
			} else {
				fmt.Printf("ERROR: Boolean unexpected field [%#v]!", field)
			}
			return string("false")
		},
		"Length": func(n *Node, field string) string {
			if bytes.Equal([]byte("prefix_child_nodes_len"), string2Bytes(field)) {
				if nil != n.prefix_child_nodes {
					return fmt.Sprintf("%d", n.prefix_child_nodes.btree.Len())
				}
			} else if bytes.Equal([]byte("folder_child_nodes_len"), string2Bytes(field)) {
				if nil != n.folder_child_nodes {
					return fmt.Sprintf("%d", n.folder_child_nodes.btree.Len())
				}
			} else {
				fmt.Printf("ERROR: Length unexpected field [%#v]!", field)
			}

			return "0"
		},
		"isNotNil": func(n *Node, field string) string {
			if bytes.Equal([]byte("prefix_child_nodes"), string2Bytes(field)) {
				if nil != n.prefix_child_nodes {
					return "isNotNil{}"
				}
			} else if bytes.Equal([]byte("folder_child_nodes"), string2Bytes(field)) {
				if nil != n.folder_child_nodes {
					return "isNotNil{}"
				}
			} else if bytes.Equal([]byte("node_hash"), string2Bytes(field)) {
				if !hash.IsNilHash(n.node_hash) {
					return "isNotNil{}"
				}
			} else if bytes.Equal([]byte("val_hash"), string2Bytes(field)) {
				if !hash.IsNilHash(n.val_hash) {
					return "isNotNil{}"
				}
			} else if bytes.Equal([]byte("prefix_child_nodes_hash"), string2Bytes(field)) {
				if !hash.IsNilHash(n.prefix_child_nodes_hash) {
					return "isNotNil{}"
				}
			} else if bytes.Equal([]byte("folder_child_nodes_hash"), string2Bytes(field)) {
				if !hash.IsNilHash(n.folder_child_nodes_hash) {
					return "isNotNil{}"
				}
			} else {
				fmt.Printf("ERROR: isNotNil unexpected field [%#v]!", field)
			}

			return "nil"
		},
		"isEqualTo": func(n *Node, field string) string {
			if bytes.Equal([]byte("parent_nodes"), string2Bytes(field)) {
				if nil != n.parent_nodes {
					if len(n.parent_nodes.parent_node.node_path_flat()) > 0 {
						if n.parent_nodes.is_folder_child_nodes {
							return fmt.Sprintf("isEqualTo{_%s.folder_child_nodes}", n.parent_nodes.parent_node.node_path_flat_str())
						} else {
							return fmt.Sprintf("isEqualTo{_%s.prefix_child_nodes}", n.parent_nodes.parent_node.node_path_flat_str())
						}
					} else {
						if n.parent_nodes.is_folder_child_nodes {
							return "isEqualTo{_root.folder_child_nodes}"
						} else {
							return "isEqualTo{_root.prefix_child_nodes}"
						}
					}
				}
			} else {
				fmt.Printf("ERROR: isEqualTo unexpected field [%#v]!", field)
			}

			return "nil"
		},
		"Children": func(n *Node) string {
			if nil == n.prefix_child_nodes && nil == n.folder_child_nodes {
				return ""
			}
			var children []string
			children = append(children, "{")
			extract := func(ns *nodes, field string) {
				iter := ns.btree.Before(uint8(0))
				for iter.Next() {
					k := iter.Key.(uint8)
					n := iter.Value.(*Node)

					V := n.node_path_flat_str()
					PV := "root"
					if nil != n.parent_nodes {
						if nil != n.parent_nodes.parent_node {
							if x := n.parent_nodes.parent_node.node_path_flat(); len(x) > 0 {
								PV = n.parent_nodes.parent_node.node_path_flat_str()
							}
						}
					}

					children = append(children, fmt.Sprintf("_%s_ := _%s.%s.btree.Get(uint8('%s'))", V, PV, field, string([]byte{k})))
					children = append(children, fmt.Sprintf("_%s := _%s_.(*Node)", V, V))
					children = append(children, n.genCheckStatementsString())
				}
			}
			if nil != n.prefix_child_nodes {
				extract(n.prefix_child_nodes, "prefix_child_nodes")
			}
			if nil != n.folder_child_nodes {
				extract(n.folder_child_nodes, "folder_child_nodes")
			}

			children = append(children, "}")
			return strings.Join(children, "\n")
		},
	}
	tmpl := template.New("Node")
	tmpl.Funcs(funcMaps)

	tmpl = template.Must(tmpl.Parse(`
_{{Name .}}Tests := Expect{
	prefix:       {{Bytes . "prefix"}},
	dirty:        {{Boolean . "dirty"}},
	parent_nodes: {{isEqualTo . "parent_nodes"}},

	node_bytes: {{Bytes . "node_bytes"}},
	node_hash:  {{isNotNil . "node_hash"}},

	val:                {{Bytes . "val"}},
	val_hash:           {{isNotNil . "val_hash"}},
	val_hash_recovered: {{Boolean . "val_hash_recovered"}},
	val_dirty:          {{Boolean . "val_dirty"}},

	prefix_child_nodes:                {{isNotNil . "prefix_child_nodes"}},
	prefix_child_nodes_hash:           {{isNotNil . "prefix_child_nodes_hash"}},
	prefix_child_nodes_hash_recovered: {{Boolean . "prefix_child_nodes_hash_recovered"}},

	folder_child_nodes:                {{isNotNil . "folder_child_nodes"}},
	folder_child_nodes_hash:           {{isNotNil . "folder_child_nodes_hash"}},
	folder_child_nodes_hash_recovered: {{Boolean . "folder_child_nodes_hash_recovered"}},

	prefix_child_nodes_len:   {{Length . "prefix_child_nodes_len"}},
	prefix_child_nodes_dirty: {{Boolean . "prefix_child_nodes_dirty"}},

	folder_child_nodes_len:   {{Length . "folder_child_nodes_len"}},
	folder_child_nodes_dirty: {{Boolean . "folder_child_nodes_dirty"}},
}.makeTests(_{{Name .}})

for _, tt := range _{{Name .}}Tests {
	if !tt.ok {
		t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
	}
}

{{Children .}}

`))

	err := tmpl.Execute(b, n)
	if err != nil {
		fmt.Println(err)
		//panic(err)
	}
}

func TestMainWorkflow(t *testing.T) {

	// dot -Tpdf -O *.dot && open *.dot.pdf
	// tdb.GenDotFile("./test_mainworkflow.dot", false)

	const db_path = "./triedb_mainworkflow_test.db"
	defer os.RemoveAll(db_path)

	var rootHash *hash.Hash = nil

	// first:  create many trie data
	{
		tdb, err := testPrepareTrieDB(db_path, rootHash)

		if nil != err {
			t.Fatal(err)
		}

		// _root := tdb.root_node
		//root.genCheckStatements(os.Stdout)

		tdb.Put(Path([]byte("A")), []byte("val_A"), true)
		// _root.genCheckStatements(os.Stdout)

		tdb.Put(Path([]byte("AB")), []byte("val_AB"), true)
		tdb.Put(Path([]byte("AC")), []byte("val_AC"), true)
		tdb.Put(Path([]byte("AD")), []byte("val_AD"), true)
		tdb.Put(Path([]byte("ABC")), []byte("val_ABC"), true)
		tdb.Put(Path([]byte("ABCD")), []byte("val_ABCD"), true)

		tdb.Put(Path([]byte("A"), []byte("A"), []byte("A")), []byte("val_A_A_A"), true)
		tdb.Put(Path([]byte("AB"), []byte("CD")), []byte("val_AB_CD"), true)

		_rootHash, err := tdb.testCommit()

		if nil != err {
			t.Fatal(err)
		}

		rootHash = _rootHash

		tdb.GenDotFile("./test_mainworkflow_1.dot", false)
		testCloseTrieDB(tdb)
	}

	// second: load the previous triedb in disk, and then create more trie data
	{
		tdb, err := testPrepareTrieDB(db_path, rootHash)

		if nil != err {
			t.Fatal(err)
		}

		_root := tdb.root_node

		tdb.Put(Path([]byte("B")), []byte("val_B"), true)
		tdb.Put(Path([]byte("Aa")), []byte("val_Aa"), true)
		tdb.Put(Path([]byte("ABa")), []byte("val_ABa"), true)
		tdb.Put(Path([]byte("ABCa")), []byte("val_ABCa"), true)

		tdb.Put(Path([]byte("B"), []byte("B"), []byte("B")), []byte("val_B_B_B"), true)
		tdb.Put(Path([]byte("B"), []byte("B")), []byte("val_B_B"), true)

		_root.genCheckStatements(os.Stdout)

		_rootHash, err := tdb.testCommit()

		if nil != err {
			t.Fatal(err)
		}

		rootHash = _rootHash

		tdb.GenDotFile("./test_mainworkflow_2.dot", false)

		testCloseTrieDB(tdb)
	}

	// test Get and Delete operations to influence the lazy status
	{
		tdb, err := testPrepareTrieDB(db_path, rootHash)

		if nil != err {
			t.Fatal(err)
		}

		tdb.GenDotFile("./test_mainworkflow_3.dot", false)

		{
			_, err := tdb.Get(Path([]byte("ABC")))
			if nil != err {
				t.Fatal(`Get item "ABC" should receive node_bytes of that kv item!`)
			}
		}

		tdb.GenDotFile("./test_mainworkflow_4.dot", false)
		{
			_, err := tdb.Del(Path([]byte("AB")))
			if nil != err {
				t.Fatal(`Delete item "AB" should work as expected!`)
			}
		}

		tdb.GenDotFile("./test_mainworkflow_5.dot", false)

		{
			tdb.Put(Path([]byte("ACa")), []byte("val_ACa"), true)
			tdb.Put(Path([]byte("ADa")), []byte("val_ADa"), true)

			_rootHash, err := tdb.testCommit()

			if nil != err {
				t.Fatal(err)
			}

			rootHash = _rootHash
		}

		tdb.GenDotFile("./test_mainworkflow_6.dot", false)
		testCloseTrieDB(tdb)
	}

	// test Delete the child node which parent node has only one child
	{
		tdb, err := testPrepareTrieDB(db_path, rootHash)

		if nil != err {
			t.Fatal(err)
		}

		{
			_, err := tdb.Del(Path([]byte("ACa")))
			if nil != err {
				t.Fatal(`Delete item "ACa" should work as expected!`, err)
			}
		}

		{
			_rootHash, err := tdb.testCommit()

			if nil != err {
				t.Fatal(err)
			}

			rootHash = _rootHash
		}

		tdb.GenDotFile("./test_mainworkflow_7.dot", false)
		testCloseTrieDB(tdb)
	}

	// test Delete intermediate node which has child but HAS NO value
	{
		tdb, err := testPrepareTrieDB(db_path, rootHash)

		if nil != err {
			t.Fatal(err)
		}

		{
			_, err := tdb.Get(Path([]byte("A"), []byte("A"), []byte("A")))

			if nil != err {
				t.Fatal("Get path [A,A,A] should NOT trigger error!")
			}
		}

		tdb.GenDotFile("./test_mainworkflow_8_1.dot", false)

		{
			_, err := tdb.Del(Path([]byte("A"), []byte("A")))

			if nil != err {
				t.Fatal("Delete path [A,A] should NOT trigger error!")
			}
		}

		{
			_rootHash, err := tdb.testCommit()

			if nil != err {
				t.Fatal(err)
			}

			rootHash = _rootHash
		}

		tdb.GenDotFile("./test_mainworkflow_8_2.dot", false)
		testCloseTrieDB(tdb)
	}

	// test Delete intermediate node which has child and HAS value
	{
		tdb, err := testPrepareTrieDB(db_path, rootHash)

		if nil != err {
			t.Fatal(err)
		}

		{
			_, err := tdb.Get(Path([]byte("B"), []byte("B"), []byte("B")))

			if nil != err {
				t.Fatal("Get path [B,B,B] should NOT trigger error!")
			}
		}

		tdb.GenDotFile("./test_mainworkflow_9_1.dot", false)

		{
			_, err := tdb.Del(Path([]byte("B"), []byte("B")))

			if nil != err {
				t.Fatal("Delete path [B,B] should NOT trigger error!")
			}
		}

		{
			_rootHash, err := tdb.testCommit()

			if nil != err {
				t.Fatal(err)
			}

			rootHash = _rootHash
		}

		tdb.GenDotFile("./test_mainworkflow_9_2.dot", false)
		testCloseTrieDB(tdb)
	}

	// test Delete data which not exist in a nonempty triedb
	{
		tdb, err := testPrepareTrieDB(db_path, rootHash)

		if nil != err {
			t.Fatal(err)
		}

		{
			_, err := tdb.Del(Path([]byte("hello")))

			if nil != err {
				t.Fatal("Delete path [hello] should NOT trigger error!")
			}
		}

		testCloseTrieDB(tdb)
	}
}

// test delete data on empty triedb
func TestDeleteOnEmptyTrieDB(t *testing.T) {

	const db_path = "triedb_deleteonempty_test.db"
	defer os.RemoveAll(db_path)

	tdb, err := testPrepareTrieDB(db_path, nil)

	if nil != err {
		t.Fatal(err)
	}

	{
		_, err := tdb.Del(Path([]byte("hello")))

		if nil != err {
			t.Fatal("Delete path [hello] should NOT trigger error!")
		}
	}

	{
		_rootHash, err := tdb.testCommit()

		if nil != err {
			t.Fatal(err)
		}

		if !hash.IsNilHash(_rootHash) {
			t.Errorf("Delete on an empty triedb, then calc hash, root hash should be: %#v, but: %#v", nil, _rootHash)
		}
	}

	testCloseTrieDB(tdb)
}

func TestLongPath(t *testing.T) {
	const db_path = "./triedb_longpath_test.db"
	defer os.RemoveAll(db_path)

	tdb, err := testPrepareTrieDB(db_path, nil)

	if nil != err {
		t.Fatal(err)
	}

	{
		root := tdb.root_node
		rootTests := Expect{
			prefix:       nil,
			dirty:        false,
			parent_nodes: nil,

			node_bytes: nil,
			node_hash:  nil,

			val:                nil,
			val_hash:           nil,
			val_hash_recovered: true,
			val_dirty:          false,

			prefix_child_nodes:                nil,
			prefix_child_nodes_hash:           nil,
			prefix_child_nodes_hash_recovered: true,

			folder_child_nodes:                nil,
			folder_child_nodes_hash:           nil,
			folder_child_nodes_hash_recovered: true,

			prefix_child_nodes_len:   0,
			prefix_child_nodes_dirty: false,

			folder_child_nodes_len:   0,
			folder_child_nodes_dirty: false,
		}.makeTests(root)

		for _, tt := range rootTests {
			if !tt.ok {
				t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
			}
		}
	}

	{
		err := tdb.Put(Path([]byte("hello")), []byte("val_hello"), true)
		if nil != err {
			t.Fatal(err)
		}
	}

	tdb.GenDotFile("./test_longpath_1.dot", false)

	{
		err := tdb.Put(Path([]byte("hellO")), []byte("val_hellO"), true)
		if nil != err {
			t.Fatal(err)
		}
	}

	tdb.GenDotFile("./test_longpath_2.dot", false)

	{
		_, err := tdb.Del(Path([]byte("hello")))
		if nil != err {
			t.Fatal(err)
		}
	}

	tdb.testCommit()
	tdb.GenDotFile("./test_longpath_3.dot", false)
	testCloseTrieDB(tdb)
}

type withFatal interface {
	Fatal(args ...any)
	Fatalf(string, ...any)
}

func prepareSampleDatabase(t withFatal, tdb *TrieDB) [][]byte {
	existingKeys := make(map[string]struct{}, 0)

	// about 218 seconds
	for i := 0; i < 10000*100*1; i++ {
		percent := mrand.Intn(100)
		if percent < 55 { // 55% Update
			tokenKey := make([]byte, 8)
			tokenVal := make([]byte, 8)
			{
				n, err := crand.Read(tokenKey)
				if nil != err {
					t.Fatal("unexpected random token error: ", err)
				}
				if n < 8 {
					t.Fatal("unexpected random token length: ", n)
				}
			}
			{
				n, err := crand.Read(tokenVal)
				if nil != err {
					t.Fatal("unexpected random token error: ", err)
				}
				if n < 8 {
					t.Fatal("unexpected random token length: ", n)
				}
			}
			// fmt.Println(n, tokenKey, tokenVal)
			err := tdb.Put(Path(tokenKey), tokenVal, true)
			if nil != err {
				t.Fatal("unexpected Update error: ", err)
			}
			existingKeys[bytes2String(tokenKey)] = struct{}{}
		} else if percent < 90 { // 35% Get
			for k := range existingKeys {
				v, err := tdb.Get(Path(string2Bytes(k)))
				if nil != err {
					t.Fatal("unexpected Get error: ", err)
				}
				if nil == v {
					t.Fatalf("value for key [%#v] must NOT be nil, but: %#v", k, v)
				}
				break
			}
		} else { // 10% Delete
			var deletingKey []byte
			if len(existingKeys) > 0 {
				index := mrand.Intn(len(existingKeys))
				for k := range existingKeys {
					if index <= 0 {
						deletingKey = string2Bytes(k)
						break
					}
					index--
				}
			}

			if nil != deletingKey {
				_, err := tdb.Del(Path(deletingKey))
				if nil != err {
					t.Fatal("unexpected Delete error: ", err)
				}
				delete(existingKeys, bytes2String(deletingKey))
			}
		}
	}

	bytesForExistingKeys := make([][]byte, 0)
	for k, _ := range existingKeys {
		bytesForExistingKeys = append(bytesForExistingKeys, string2Bytes(k))
	}
	return bytesForExistingKeys
}

func TestMorePressure(t *testing.T) {
	const db_path = "./triedb_morepressure_test.db"
	os.RemoveAll(db_path)

	tdb, err := testPrepareTrieDB(db_path, nil)

	if nil != err {
		t.Fatal(err)
	}

	prepareSampleDatabase(t, tdb)

	tdb.testCommit()
	// tdb.GenDotFile("./test_withmorepressure.dot", false)
	testCloseTrieDB(tdb)
}

// give more pressure to do more operations
func BenchmarkMoreOperations(b *testing.B) {
	const db_path = "./triedb_moreoperations_test.db"
	defer os.RemoveAll(db_path)

	tdb, err := testPrepareTrieDB(db_path, nil)

	if nil != err {
		b.Fatal(err)
	}
	existingKeys := prepareSampleDatabase(b, tdb)

	const COUNT = 10000 * 500

	randomUpdateKeys := make([][]byte, 0, COUNT)
	randomUpdateVals := make([][]byte, 0, COUNT)

	for i := 0; i < COUNT; i++ {
		tokenKey := make([]byte, mrand.Intn(16)+1)
		tokenVal := make([]byte, mrand.Intn(16)+1)
		crand.Read(tokenKey)
		crand.Read(tokenVal)
		randomUpdateKeys = append(randomUpdateKeys, tokenKey)
		randomUpdateVals = append(randomUpdateVals, tokenVal)
	}

	randomGetKeys := make([][]byte, 0, COUNT)
	for i := 0; i < COUNT; i++ {
		j := mrand.Intn(len(existingKeys))
		randomGetKeys = append(randomGetKeys, existingKeys[j])
	}

	randomDeleteKeys := make([][]byte, 0, COUNT)
	for i := 0; i < COUNT; i++ {
		j := mrand.Intn(len(existingKeys))
		randomDeleteKeys = append(randomDeleteKeys, existingKeys[j])
	}

	b.ResetTimer()
	b.Run("Update", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tokenKey := randomUpdateKeys[i]
			tokenVal := randomUpdateVals[i]
			err := tdb.Put(Path(tokenKey), tokenVal, true)
			if nil != err {
				b.Fatal(err)
			}
		}
	})
	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if len(existingKeys) > 0 {
				tokenKey := randomGetKeys[i]
				val, err := tdb.Get(Path(tokenKey))
				if nil != err {
					b.Fatal(err)
				}
				if nil == val {
					b.Fatal("unexpected nil value for key: ", tokenKey)
				}
			}
		}
	})
	b.Run("Delete", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if len(existingKeys) > 0 {
				tokenKey := randomDeleteKeys[i]
				_, err := tdb.Del(Path(tokenKey))
				if nil != err {
					b.Fatal(err)
				}
			} else {
				break
			}
		}
	})

	tdb.testCommit()
	// tdb.GenDotFile("./test_domoreoperations.dot", false)
	testCloseTrieDB(tdb)
}
