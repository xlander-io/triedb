package triedb

import (
	"bytes"
	"os"
	"reflect"
	"testing"

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
	rootHash, toUpdate, toDel, err := tdb.CalHash()
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

func (x isNotNil) isOk(o interface{}) bool {

	if O, ok := o.(*hash.Hash); ok {
		return !hash.IsNilHash(O)
	}

	if nil != o {
		return !reflect.ValueOf(o).IsNil()
	} else {
		return false
	}
}

func (x isEqualTo) isOk(o interface{}) bool {
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
	path         []byte
	dirty        bool
	parent_nodes doCheck

	node_bytes          doCheck
	node_hash           doCheck
	node_hash_recovered bool

	val                []byte
	val_hash           doCheck
	val_hash_recovered bool

	child_nodes                doCheck
	child_nodes_hash           doCheck
	child_nodes_hash_recovered bool

	child_nodes_path_index_len int
	child_nodes_dirty          bool
}

func (ex Expect) makeTests(o *Node) []TEST {
	tests := []TEST{
		{"path", ex.path, o.path, bytes.Equal(ex.path, o.path)},
		{"dirty", ex.dirty, o.dirty, ex.dirty == o.dirty},

		{"node_hash_recovered", ex.node_hash_recovered, o.node_hash_recovered, ex.node_hash_recovered == o.node_hash_recovered},
		{"val_hash_recovered", ex.val_hash_recovered, o.val_hash_recovered, ex.val_hash_recovered == o.val_hash_recovered},
		{"child_nodes_hash_recovered", ex.child_nodes_hash_recovered, o.child_nodes_hash_recovered, ex.child_nodes_hash_recovered == o.child_nodes_hash_recovered},
	}

	if nil == ex.parent_nodes {
		tests = append(tests, TEST{"parent_nodes", ex.parent_nodes, o.parent_nodes, nil == o.parent_nodes})
	} else {
		tests = append(tests, TEST{"parent_nodes", ex.parent_nodes, o.parent_nodes, ex.parent_nodes.isOk(o.parent_nodes)})
	}

	if nil == ex.child_nodes {
		tests = append(tests, TEST{"child_nodes", ex.child_nodes, o.child_nodes, nil == o.child_nodes})
	} else {
		// b := !(nil == o.child_nodes)
		b := ex.child_nodes.isOk(o.child_nodes)
		tests = append(tests, TEST{"child_nodes", ex.child_nodes, o.child_nodes, b})
		// tests = append(tests, TEST{"child_nodes", ex.child_nodes, o.child_nodes, false})
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

	if nil == ex.child_nodes_hash {
		tests = append(tests, TEST{"child_nodes_hash", ex.child_nodes_hash, o.child_nodes_hash, hash.IsNilHash(o.child_nodes_hash)})
	} else {
		tests = append(tests, TEST{"child_nodes_hash", ex.child_nodes_hash, o.child_nodes_hash, ex.child_nodes_hash.isOk(o.child_nodes_hash)})
	}

	if nil != o.child_nodes {
		tests = append(tests, []TEST{
			{"len(child_nodes.path_index)", ex.child_nodes_path_index_len, len(o.child_nodes.path_index), ex.child_nodes_path_index_len == len(o.child_nodes.path_index)},
			{"child_nodes.dirty", ex.child_nodes_dirty, o.child_nodes.dirty, ex.child_nodes_dirty == o.child_nodes.dirty},
			{"child_nodes.parent_node", o, o.child_nodes.parent_node, o == o.child_nodes.parent_node},
		}...)
	}

	return tests
}

func TestMainWorkflow(t *testing.T) {

	// dot -Tpdf -O *.dot && open *.dot.pdf
	// tdb.GenDotFile("./test_mainworkflow.dot", false)

	const db_path = "./triedb_mainworkflow_test.db"
	os.RemoveAll(db_path)

	var rootHash *hash.Hash = nil

	// first:  create many trie data
	{
		tdb, err := testPrepareTrieDB(db_path, rootHash)

		if nil != err {
			t.Fatal(err)
		}

		{
			root := tdb.root_node

			rootTests := Expect{
				path:         nil,
				dirty:        false,
				parent_nodes: nil,

				node_bytes:          nil,
				node_hash:           nil,
				node_hash_recovered: true,

				val:                nil,
				val_hash:           nil,
				val_hash_recovered: true,

				child_nodes:                nil,
				child_nodes_hash:           nil,
				child_nodes_hash_recovered: true,

				// child_nodes_path_index_len: 0,
				// child_nodes_dirty:          false,
			}.makeTests(root)

			for _, tt := range rootTests {
				if !tt.ok {
					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
				}
			}
		}

		tdb.Update([]byte("1"), []byte("val_1"))
		{
			root := tdb.root_node

			rootTests := Expect{
				path:         nil,
				dirty:        true,
				parent_nodes: nil,

				node_bytes:          nil,
				node_hash:           nil,
				node_hash_recovered: true,

				val:                nil,
				val_hash:           nil,
				val_hash_recovered: true,

				child_nodes:                isNotNil{},
				child_nodes_hash:           nil,
				child_nodes_hash_recovered: true,

				child_nodes_path_index_len: 1,
				child_nodes_dirty:          true,
			}.makeTests(root)

			for _, tt := range rootTests {
				if !tt.ok {
					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
				}
			}

			{
				_1 := root.child_nodes.path_index[byte('1')]
				if nil == _1 {
					t.Fatalf("unexpect nil pointer for node %v", '1')
				}
				_1Tests := Expect{
					path:         []byte{'1'},
					dirty:        true,
					parent_nodes: isEqualTo{root.child_nodes},

					node_bytes:          nil,
					node_hash:           nil,
					node_hash_recovered: true,

					val:                []byte("val_1"),
					val_hash:           nil,
					val_hash_recovered: true,

					child_nodes:                nil,
					child_nodes_hash:           nil,
					child_nodes_hash_recovered: true,

					child_nodes_path_index_len: 1,
					child_nodes_dirty:          true,
				}.makeTests(_1)

				for _, tt := range _1Tests {
					if !tt.ok {
						t.Errorf("_1 node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
					}
				}
			}
		}

		tdb.Update([]byte("12"), []byte("val_12"))
		tdb.Update([]byte("13"), []byte("val_13"))
		tdb.Update([]byte("14"), []byte("val_14"))
		tdb.Update([]byte("123"), []byte("val_123"))
		tdb.Update([]byte("1234"), []byte("val_1234"))

		{
			root := tdb.root_node

			rootTests := Expect{
				path:         nil,
				dirty:        true,
				parent_nodes: nil,

				node_bytes:          nil,
				node_hash:           nil,
				node_hash_recovered: true,

				val:                nil,
				val_hash:           nil,
				val_hash_recovered: true,

				child_nodes:                isNotNil{},
				child_nodes_hash:           nil,
				child_nodes_hash_recovered: true,

				child_nodes_path_index_len: 1,
				child_nodes_dirty:          true,
			}.makeTests(root)

			for _, tt := range rootTests {
				if !tt.ok {
					t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
				}
			}

			{
				_1 := root.child_nodes.path_index[byte('1')]
				if nil == _1 {
					t.Fatalf("unexpect nil pointer for node full path %v", "1")
				}

				_1Tests := Expect{
					path:         []byte{'1'},
					dirty:        true,
					parent_nodes: isEqualTo{root.child_nodes},

					node_bytes:          nil,
					node_hash:           nil,
					node_hash_recovered: true,

					val:                []byte("val_1"),
					val_hash:           nil,
					val_hash_recovered: true,

					child_nodes:                isNotNil{},
					child_nodes_hash:           nil,
					child_nodes_hash_recovered: true,

					child_nodes_path_index_len: 3,
					child_nodes_dirty:          true,
				}.makeTests(_1)

				for _, tt := range _1Tests {
					if !tt.ok {
						t.Errorf("_1 node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
					}
				}

				{
					_12 := _1.child_nodes.path_index[byte('2')]
					_13 := _1.child_nodes.path_index[byte('3')]
					_14 := _1.child_nodes.path_index[byte('4')]
					if nil == _12 {
						t.Fatalf("unexpect nil pointer for node full path %v", "12")
					}
					if nil == _13 {
						t.Fatalf("unexpect nil pointer for node full path %v", "13")
					}
					if nil == _14 {
						t.Fatalf("unexpect nil pointer for node full path %v", "14")
					}

					{
						_12Tests := Expect{
							path:         []byte{'2'},
							dirty:        true,
							parent_nodes: isEqualTo{_1.child_nodes},

							node_bytes:          nil,
							node_hash:           nil,
							node_hash_recovered: true,

							val:                []byte("val_12"),
							val_hash:           nil,
							val_hash_recovered: true,

							child_nodes:                isNotNil{},
							child_nodes_hash:           nil,
							child_nodes_hash_recovered: true,

							child_nodes_path_index_len: 1,
							child_nodes_dirty:          true,
						}.makeTests(_12)

						for _, tt := range _12Tests {
							if !tt.ok {
								t.Errorf("_12 node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
							}
						}

						{
							_123 := _12.child_nodes.path_index[byte('3')]
							if nil == _123 {
								t.Fatalf("unexpect nil pointer for node full path %v", "123")
							}

							_123Tests := Expect{
								path:         []byte{'3'},
								dirty:        true,
								parent_nodes: isEqualTo{_12.child_nodes},

								node_bytes:          nil,
								node_hash:           nil,
								node_hash_recovered: true,

								val:                []byte("val_123"),
								val_hash:           nil,
								val_hash_recovered: true,

								child_nodes:                isNotNil{},
								child_nodes_hash:           nil,
								child_nodes_hash_recovered: true,

								child_nodes_path_index_len: 1,
								child_nodes_dirty:          true,
							}.makeTests(_123)

							for _, tt := range _123Tests {
								if !tt.ok {
									t.Errorf("_123 node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
								}
							}

							{
								_1234 := _123.child_nodes.path_index[byte('4')]
								if nil == _1234 {
									t.Fatalf("unexpect nil pointer for node full path %v", "1234")
								}

								_1234Tests := Expect{
									path:         []byte{'4'},
									dirty:        true,
									parent_nodes: isEqualTo{_123.child_nodes},

									node_bytes:          nil,
									node_hash:           nil,
									node_hash_recovered: true,

									val:                []byte("val_1234"),
									val_hash:           nil,
									val_hash_recovered: true,

									child_nodes:                nil,
									child_nodes_hash:           nil,
									child_nodes_hash_recovered: true,

									child_nodes_path_index_len: 1,
									child_nodes_dirty:          true,
								}.makeTests(_1234)

								for _, tt := range _1234Tests {
									if !tt.ok {
										t.Errorf("_1234 node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
									}
								}
							}
						}
					}
					{
						_13Tests := Expect{
							path:         []byte{'3'},
							dirty:        true,
							parent_nodes: isEqualTo{_1.child_nodes},

							node_bytes:          nil,
							node_hash:           nil,
							node_hash_recovered: true,

							val:                []byte("val_13"),
							val_hash:           nil,
							val_hash_recovered: true,

							child_nodes:                nil,
							child_nodes_hash:           nil,
							child_nodes_hash_recovered: true,

							child_nodes_path_index_len: 1,
							child_nodes_dirty:          true,
						}.makeTests(_13)

						for _, tt := range _13Tests {
							if !tt.ok {
								t.Errorf("_13 node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
							}
						}
					}
					{
						_14Tests := Expect{
							path:         []byte{'4'},
							dirty:        true,
							parent_nodes: isEqualTo{_1.child_nodes},

							node_bytes:          nil,
							node_hash:           nil,
							node_hash_recovered: true,

							val:                []byte("val_14"),
							val_hash:           nil,
							val_hash_recovered: true,

							child_nodes:                nil,
							child_nodes_hash:           nil,
							child_nodes_hash_recovered: true,

							// child_nodes_path_index_len: 1,
							// child_nodes_dirty:          false,
						}.makeTests(_14)

						for _, tt := range _14Tests {
							if !tt.ok {
								t.Errorf("_14 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
							}
						}
					}
				}
			}
		}

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

		tdb.Update([]byte("2"), []byte("val_2"))
		tdb.Update([]byte("1a"), []byte([]byte("val_1a")))
		tdb.Update([]byte("12a"), []byte([]byte("val_12a")))
		tdb.Update([]byte("123a"), []byte([]byte("val_123a")))

		_rootHash, err := tdb.testCommit()

		if nil != err {
			t.Fatal(err)
		}

		rootHash = _rootHash

		tdb.GenDotFile("./test_mainworkflow_2.dot", false)

		{
			root := tdb.root_node
			{
				rootTests := Expect{
					path:         nil,
					dirty:        true,
					parent_nodes: nil,

					node_bytes:          isNotNil{},
					node_hash:           isNotNil{},
					node_hash_recovered: true,

					val:                nil,
					val_hash:           nil,
					val_hash_recovered: true,

					child_nodes:                isNotNil{},
					child_nodes_hash:           isNotNil{},
					child_nodes_hash_recovered: true,

					child_nodes_path_index_len: 2,
					child_nodes_dirty:          true,
				}.makeTests(root)

				for _, tt := range rootTests {
					if !tt.ok {
						t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
					}
				}
			}
			{
				_1 := root.child_nodes.path_index[byte('1')]
				_2 := root.child_nodes.path_index[byte('2')]
				if nil == _1 {
					t.Fatalf("unexpect nil pointer for node full path %v", "1")
				}
				if nil == _2 {
					t.Fatalf("unexpect nil pointer for node full path %v", "2")
				}

				{
					_1Tests := Expect{
						path:         []byte{'1'},
						dirty:        true,
						parent_nodes: isEqualTo{root.child_nodes},

						node_bytes:          isNotNil{},
						node_hash:           isNotNil{},
						node_hash_recovered: true,

						val:                []byte("val_1"),
						val_hash:           isNotNil{},
						val_hash_recovered: true,

						child_nodes:                isNotNil{},
						child_nodes_hash:           isNotNil{},
						child_nodes_hash_recovered: true,

						child_nodes_path_index_len: 4,
						child_nodes_dirty:          true,
					}.makeTests(_1)

					for _, tt := range _1Tests {
						if !tt.ok {
							t.Errorf("_1 node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
						}
					}
				}
				{
					_2Tests := Expect{
						path:         []byte{'2'},
						dirty:        true,
						parent_nodes: isEqualTo{root.child_nodes},

						node_bytes:          isNotNil{},
						node_hash:           isNotNil{},
						node_hash_recovered: true,

						val:                []byte("val_2"),
						val_hash:           isNotNil{},
						val_hash_recovered: true,

						child_nodes:                nil,
						child_nodes_hash:           nil,
						child_nodes_hash_recovered: true,

						// child_nodes_path_index_len: 4,
						// child_nodes_dirty:          true,
					}.makeTests(_2)

					for _, tt := range _2Tests {
						if !tt.ok {
							t.Errorf("_2 node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
						}
					}
				}

				if nil != _1.child_nodes {
					_12 := _1.child_nodes.path_index[byte('2')]
					_13 := _1.child_nodes.path_index[byte('3')]
					_14 := _1.child_nodes.path_index[byte('4')]
					_1a := _1.child_nodes.path_index[byte('a')]
					if nil == _12 {
						t.Fatalf("unexpect nil pointer for node full path %v", "12")
					}
					if nil == _13 {
						t.Fatalf("unexpect nil pointer for node full path %v", "13")
					}
					if nil == _14 {
						t.Fatalf("unexpect nil pointer for node full path %v", "14")
					}
					if nil == _1a {
						t.Fatalf("unexpect nil pointer for node full path %v", "1a")
					}

					{
						_12Tests := Expect{
							path:         []byte{'2'},
							dirty:        true,
							parent_nodes: isEqualTo{_1.child_nodes},

							node_bytes:          isNotNil{},
							node_hash:           isNotNil{},
							node_hash_recovered: true,

							val:                []byte("val_12"),
							val_hash:           isNotNil{},
							val_hash_recovered: true,

							child_nodes:                isNotNil{},
							child_nodes_hash:           isNotNil{},
							child_nodes_hash_recovered: true,

							child_nodes_path_index_len: 2,
							child_nodes_dirty:          true,
						}.makeTests(_12)

						for _, tt := range _12Tests {
							if !tt.ok {
								t.Errorf("_12 node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
							}
						}
					}
					{
						_13Tests := Expect{
							path:         []byte{'3'},
							dirty:        false,
							parent_nodes: isEqualTo{_1.child_nodes},

							node_bytes:          nil,
							node_hash:           isNotNil{},
							node_hash_recovered: false,

							val:                nil,
							val_hash:           nil,
							val_hash_recovered: false,

							child_nodes:                nil,
							child_nodes_hash:           nil,
							child_nodes_hash_recovered: false,

							// child_nodes_path_index_len: 2,
							// child_nodes_dirty:          true,
						}.makeTests(_13)

						for _, tt := range _13Tests {
							if !tt.ok {
								t.Errorf("_13 node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
							}
						}
					}

					{
						_14Tests := Expect{
							path:         []byte{'4'},
							dirty:        false,
							parent_nodes: isEqualTo{_1.child_nodes},

							node_bytes:          nil,
							node_hash:           isNotNil{},
							node_hash_recovered: false,

							val:                nil,
							val_hash:           nil,
							val_hash_recovered: false,

							child_nodes:                nil,
							child_nodes_hash:           nil,
							child_nodes_hash_recovered: false,

							// child_nodes_path_index_len: 2,
							// child_nodes_dirty:          true,
						}.makeTests(_14)

						for _, tt := range _14Tests {
							if !tt.ok {
								t.Errorf("_14 node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
							}
						}
					}
					{
						_1aTests := Expect{
							path:         []byte{'a'},
							dirty:        true,
							parent_nodes: isEqualTo{_1.child_nodes},

							node_bytes:          isNotNil{},
							node_hash:           isNotNil{},
							node_hash_recovered: true,

							val:                []byte("val_1a"),
							val_hash:           isNotNil{},
							val_hash_recovered: true,

							child_nodes:                nil,
							child_nodes_hash:           nil,
							child_nodes_hash_recovered: true,

							// child_nodes_path_index_len: 2,
							// child_nodes_dirty:          true,
						}.makeTests(_1a)

						for _, tt := range _1aTests {
							if !tt.ok {
								t.Errorf("_1a node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
							}
						}
					}

					if nil != _12.child_nodes {
						_123 := _12.child_nodes.path_index[byte('3')]
						_12a := _12.child_nodes.path_index[byte('a')]
						if nil == _123 {
							t.Fatalf("unexpect nil pointer for node full path %v", "123")
						}
						if nil == _12a {
							t.Fatalf("unexpect nil pointer for node full path %v", "12a")
						}

						{
							_123Tests := Expect{
								path:         []byte{'3'},
								dirty:        true,
								parent_nodes: isEqualTo{_12.child_nodes},

								node_bytes:          isNotNil{},
								node_hash:           isNotNil{},
								node_hash_recovered: true,

								val:                []byte("val_123"),
								val_hash:           isNotNil{},
								val_hash_recovered: true,

								child_nodes:                isNotNil{},
								child_nodes_hash:           isNotNil{},
								child_nodes_hash_recovered: true,

								child_nodes_path_index_len: 2,
								child_nodes_dirty:          true,
							}.makeTests(_123)

							for _, tt := range _123Tests {
								if !tt.ok {
									t.Errorf("_123 node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
								}
							}
						}
						{
							_12aTests := Expect{
								path:         []byte{'a'},
								dirty:        true,
								parent_nodes: isEqualTo{_12.child_nodes},

								node_bytes:          isNotNil{},
								node_hash:           isNotNil{},
								node_hash_recovered: true,

								val:                []byte("val_12a"),
								val_hash:           isNotNil{},
								val_hash_recovered: true,

								child_nodes:                nil,
								child_nodes_hash:           nil,
								child_nodes_hash_recovered: true,

								// child_nodes_path_index_len: 2,
								// child_nodes_dirty:          true,
							}.makeTests(_12a)

							for _, tt := range _12aTests {
								if !tt.ok {
									t.Errorf("_12a node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
								}
							}
						}

						if nil != _123.child_nodes {
							_1234 := _123.child_nodes.path_index[byte('4')]
							_123a := _123.child_nodes.path_index[byte('a')]
							if nil == _1234 {
								t.Fatalf("unexpect nil pointer for node full path %v", "1234")
							}
							if nil == _123a {
								t.Fatalf("unexpect nil pointer for node full path %v", "123a")
							}

							{
								_1234Tests := Expect{
									path:         []byte{'4'},
									dirty:        false,
									parent_nodes: isEqualTo{_123.child_nodes},

									node_bytes:          nil,
									node_hash:           isNotNil{},
									node_hash_recovered: false,

									val:                nil,
									val_hash:           nil,
									val_hash_recovered: false,

									child_nodes:                nil,
									child_nodes_hash:           nil,
									child_nodes_hash_recovered: false,

									// child_nodes_path_index_len: 2,
									// child_nodes_dirty:          true,
								}.makeTests(_1234)

								for _, tt := range _1234Tests {
									if !tt.ok {
										t.Errorf("_1234 node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
									}
								}
							}
							{
								_123aTests := Expect{
									path:         []byte{'a'},
									dirty:        true,
									parent_nodes: isEqualTo{_123.child_nodes},

									node_bytes:          isNotNil{},
									node_hash:           isNotNil{},
									node_hash_recovered: true,

									val:                []byte("val_123a"),
									val_hash:           isNotNil{},
									val_hash_recovered: true,

									child_nodes:                nil,
									child_nodes_hash:           nil,
									child_nodes_hash_recovered: true,

									// child_nodes_path_index_len: 2,
									// child_nodes_dirty:          true,
								}.makeTests(_123a)

								for _, tt := range _123aTests {
									if !tt.ok {
										t.Errorf("_123a node %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
									}
								}
							}
						}
					}
				}
			}
		}

		testCloseTrieDB(tdb)
	}

	// test Get and Delete operations to influence the lazy status
	{
		tdb, err := testPrepareTrieDB(db_path, rootHash)

		if nil != err {
			t.Fatal(err)
		}

		tdb.GenDotFile("./test_mainworkflow_3.dot", false)

		_, err = tdb.Get([]byte("123"))
		if nil != err {
			t.Fatal(`Get item "123" should receive node_bytes of that kv item!`)
		}

		tdb.GenDotFile("./test_mainworkflow_4.dot", false)
		err = tdb.Delete([]byte("12"))
		if nil != err {
			t.Fatal(`Delete item "12" should work as expected!`)
		}

		tdb.GenDotFile("./test_mainworkflow_5.dot", false)

		tdb.Update([]byte("13a"), []byte([]byte("val_13a")))
		tdb.Update([]byte("14a"), []byte([]byte("val_14a")))

		_rootHash, err := tdb.testCommit()

		if nil != err {
			t.Fatal(err)
		}

		rootHash = _rootHash

		tdb.GenDotFile("./test_mainworkflow_6.dot", false)
		testCloseTrieDB(tdb)
	}

	// test Delete the child node which parent node has only one child
	{
		tdb, err := testPrepareTrieDB(db_path, rootHash)

		if nil != err {
			t.Fatal(err)
		}

		err = tdb.Delete([]byte("13a"))
		if nil != err {
			t.Fatal(`Delete item "13a" should work as expected!`, err)
		}

		_rootHash, err := tdb.testCommit()

		if nil != err {
			t.Fatal(err)
		}

		rootHash = _rootHash

		tdb.GenDotFile("./test_mainworkflow_7.dot", false)
		testCloseTrieDB(tdb)
	}

	// test Delete data which not exist in a nonempty triedb
	{
		tdb, err := testPrepareTrieDB(db_path, rootHash)

		if nil != err {
			t.Fatal(err)
		}

		err = tdb.Delete([]byte("hello"))

		if nil != err {
			t.Fatal("Delete path [hello] should NOT trigger error!")
		}

		testCloseTrieDB(tdb)
	}
}

// test delete data on empty triedb
func TestDeleteOnEmptyTrieDB(t *testing.T) {

	tdb, err := testPrepareTrieDB("triedb_deleteonempty_test.db", nil)

	if nil != err {
		t.Fatal(err)
	}

	err = tdb.Delete([]byte("hello"))

	if nil != err {
		t.Fatal("Delete path [hello] should NOT trigger error!")
	}

	_rootHash, err := tdb.testCommit()

	if nil != err {
		t.Fatal(err)
	}

	if !hash.IsNilHash(_rootHash) {
		t.Errorf("Delete on an empty triedb, then calc hash, root hash should be: %#v, but: %#v", nil, _rootHash)
	}

	testCloseTrieDB(tdb)
}

func TestLongPath(t *testing.T) {
	tdb, err := testPrepareTrieDB("triedb_longpath_test.db", nil)

	if nil != err {
		t.Fatal(err)
	}

	{
		root := tdb.root_node
		rootTests := Expect{
			path:         nil,
			dirty:        false,
			parent_nodes: nil,

			node_bytes:          nil,
			node_hash:           nil,
			node_hash_recovered: true,

			val:                nil,
			val_hash:           nil,
			val_hash_recovered: true,

			child_nodes:                nil,
			child_nodes_hash:           nil,
			child_nodes_hash_recovered: true,

			// child_nodes_path_index_len: 0,
			// child_nodes_dirty:          false,
		}.makeTests(root)

		for _, tt := range rootTests {
			if !tt.ok {
				t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
			}
		}
	}

	err = tdb.Update([]byte("hello"), []byte("val_hello"))
	if nil != err {
		t.Fatal(err)
	}

	{
		root := tdb.root_node
		rootTests := Expect{
			path:         nil,
			dirty:        true,
			parent_nodes: nil,

			node_bytes:          nil,
			node_hash:           nil,
			node_hash_recovered: true,

			val:                nil,
			val_hash:           nil,
			val_hash_recovered: true,

			child_nodes:                isNotNil{},
			child_nodes_hash:           nil,
			child_nodes_hash_recovered: true,

			child_nodes_path_index_len: 1,
			child_nodes_dirty:          true,
		}.makeTests(root)

		for _, tt := range rootTests {
			if !tt.ok {
				t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
			}
		}

		{
			hello := root.child_nodes.path_index['h']

			helloTests := Expect{
				path:         []byte("hello"),
				dirty:        true,
				parent_nodes: isEqualTo{root.child_nodes},

				node_bytes:          nil,
				node_hash:           nil,
				node_hash_recovered: true,

				val:                []byte("val_hello"),
				val_hash:           nil,
				val_hash_recovered: true,

				child_nodes:                nil,
				child_nodes_hash:           nil,
				child_nodes_hash_recovered: true,

				// child_nodes_path_index_len: 0,
				// child_nodes_dirty:          false,
			}.makeTests(hello)

			for _, tt := range helloTests {
				if !tt.ok {
					t.Errorf("hello %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
				}
			}
		}
	}

	tdb.GenDotFile("./test_longpath_1.dot", false)

	err = tdb.Update([]byte("hellO"), []byte("val_hellO"))
	if nil != err {
		t.Fatal(err)
	}

	{
		root := tdb.root_node
		rootTests := Expect{
			path:         nil,
			dirty:        true,
			parent_nodes: nil,

			node_bytes:          nil,
			node_hash:           nil,
			node_hash_recovered: true,

			val:                nil,
			val_hash:           nil,
			val_hash_recovered: true,

			child_nodes:                isNotNil{},
			child_nodes_hash:           nil,
			child_nodes_hash_recovered: true,

			child_nodes_path_index_len: 1,
			child_nodes_dirty:          true,
		}.makeTests(root)

		for _, tt := range rootTests {
			if !tt.ok {
				t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
			}
		}

		{
			hell := root.child_nodes.path_index['h']

			if nil == hell {
				t.Fatalf("unexpect nil pointer for node full path %v", "hell")
			}

			hellTests := Expect{
				path:         []byte("hell"),
				dirty:        true,
				parent_nodes: isEqualTo{root.child_nodes},

				node_bytes:          nil,
				node_hash:           nil,
				node_hash_recovered: true,

				val:                nil,
				val_hash:           nil,
				val_hash_recovered: true,

				child_nodes:                isNotNil{},
				child_nodes_hash:           nil,
				child_nodes_hash_recovered: true,

				child_nodes_path_index_len: 2,
				child_nodes_dirty:          true,
			}.makeTests(hell)

			for _, tt := range hellTests {
				if !tt.ok {
					t.Errorf("hell %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
				}
			}

			{
				hello := hell.child_nodes.path_index['o']
				if nil == hello {
					t.Fatalf("unexpect nil pointer for node full path %v", "hello")
				}

				helloTests := Expect{
					path:         []byte("o"),
					dirty:        true,
					parent_nodes: isEqualTo{hell.child_nodes},

					node_bytes:          nil,
					node_hash:           nil,
					node_hash_recovered: true,

					val:                []byte("val_hello"),
					val_hash:           nil,
					val_hash_recovered: true,

					child_nodes:                nil,
					child_nodes_hash:           nil,
					child_nodes_hash_recovered: true,

					child_nodes_path_index_len: 2,
					child_nodes_dirty:          true,
				}.makeTests(hello)

				for _, tt := range helloTests {
					if !tt.ok {
						t.Errorf("hello %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
					}
				}
			}
			{
				hellO := hell.child_nodes.path_index['O']

				if nil == hellO {
					t.Fatalf("unexpect nil pointer for node full path %v", "hellO")
				}

				hellOTests := Expect{
					path:         []byte("O"),
					dirty:        true,
					parent_nodes: isEqualTo{hell.child_nodes},

					node_bytes:          nil,
					node_hash:           nil,
					node_hash_recovered: true,

					val:                []byte("val_hellO"),
					val_hash:           nil,
					val_hash_recovered: true,

					child_nodes:                nil,
					child_nodes_hash:           nil,
					child_nodes_hash_recovered: true,

					// child_nodes_path_index_len: 2,
					// child_nodes_dirty:          true,
				}.makeTests(hellO)

				for _, tt := range hellOTests {
					if !tt.ok {
						t.Errorf("hellO %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
					}
				}
			}
		}
	}

	tdb.GenDotFile("./test_longpath_2.dot", false)

	err = tdb.Delete([]byte("hello"))
	if nil != err {
		t.Fatal(err)
	}

	{
		root := tdb.root_node
		rootTests := Expect{
			path:         nil,
			dirty:        true,
			parent_nodes: nil,

			node_bytes:          nil,
			node_hash:           nil,
			node_hash_recovered: true,

			val:                nil,
			val_hash:           nil,
			val_hash_recovered: true,

			child_nodes:                isNotNil{},
			child_nodes_hash:           nil,
			child_nodes_hash_recovered: true,

			child_nodes_path_index_len: 1,
			child_nodes_dirty:          true,
		}.makeTests(root)

		for _, tt := range rootTests {
			if !tt.ok {
				t.Errorf("root %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
			}
		}

		{
			hellO := root.child_nodes.path_index['h']

			if nil == hellO {
				t.Fatalf("unexpect nil pointer for node full path %v", "hellO")
			}

			hellOTests := Expect{
				path:         []byte("hellO"),
				dirty:        true,
				parent_nodes: isEqualTo{root.child_nodes},

				node_bytes:          nil,
				node_hash:           nil,
				node_hash_recovered: true,

				val:                []byte("val_hellO"),
				val_hash:           nil,
				val_hash_recovered: true,

				child_nodes:                nil,
				child_nodes_hash:           nil,
				child_nodes_hash_recovered: true,

				// child_nodes_path_index_len: 2,
				// child_nodes_dirty:          true,
			}.makeTests(hellO)

			for _, tt := range hellOTests {
				if !tt.ok {
					t.Errorf("hellO %s expect: %#v, but: %#v", tt.label, tt.expected, tt.actual)
				}
			}
		}
	}

	tdb.testCommit()
	tdb.GenDotFile("./test_longpath_3.dot", false)
	testCloseTrieDB(tdb)
}
