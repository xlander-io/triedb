package triedb

import (
	"bytes"
	"os"
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

func TestMainWorkflow(t *testing.T) {

	// dot -Tpdf -O *.dot && open *.dot.pdf
	// tdb.GenDotFile("./test_mainworkflow.dot", false)

	const db_path = "./triedb_mainworkflow_test.db"
	os.RemoveAll(db_path)

	type TEST struct {
		label    string
		expected interface{}
		actual   interface{}
		ok       bool
	}

	var rootHash *hash.Hash

	// test delete data in a empty triedb
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

	// first:  create many trie data
	{
		tdb, err := testPrepareTrieDB(db_path, nil)

		if nil != err {
			t.Fatal(err)
		}

		{
			root := tdb.root_node

			var rootTests = []TEST{
				{"path", nil, root.path, nil == root.path},
				{"dirty", false, root.dirty, false == root.dirty},
				{"parent_nodes", nil, root.parent_nodes, nil == root.parent_nodes},

				{"node_bytes", nil, root.node_bytes, nil == root.node_bytes},
				{"node_hash", "hash.IsNilHash", root.node_hash, hash.IsNilHash(root.node_hash)},
				{"node_hash_recovered", true, root.node_hash_recovered, true == root.node_hash_recovered},

				{"val", nil, root.val, nil == root.val},
				{"val_hash", nil, root.val_hash, nil == root.val_hash},
				{"val_hash_recovered", true, root.val_hash_recovered, true == root.val_hash_recovered},

				{"child_nodes", nil, root.child_nodes, nil == root.child_nodes},
				{"child_nodes_hash", nil, root.child_nodes_hash, nil == root.child_nodes_hash},
				{"child_nodes_hash_recovered", true, root.child_nodes_hash_recovered, true == root.child_nodes_hash_recovered},
			}

			for _, tt := range rootTests {
				if !tt.ok {
					t.Errorf("root %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
				}
			}
		}

		tdb.Update([]byte("1"), []byte("val_1"))
		{
			root := tdb.root_node

			rootTests := []TEST{
				{"path", nil, root.path, nil == root.path},
				{"dirty", true, root.dirty, true == root.dirty},
				{"parent_nodes", nil, root.parent_nodes, nil == root.parent_nodes},

				{"node_bytes", nil, root.node_bytes, nil == root.node_bytes},
				{"node_hash", "hash.IsNilHash", root.node_hash, hash.IsNilHash(root.node_hash)},
				{"node_hash_recovered", true, root.node_hash_recovered, true == root.node_hash_recovered},

				{"val", nil, root.val, nil == root.val},
				{"val_hash", nil, root.val_hash, nil == root.val_hash},
				{"val_hash_recovered", true, root.val_hash_recovered, true == root.val_hash_recovered},

				{"child_nodes", "!nil", root.child_nodes, nil != root.child_nodes},
				{"child_nodes_hash", nil, root.child_nodes_hash, nil == root.child_nodes_hash},
				{"child_nodes_hash_recovered", true, root.child_nodes_hash_recovered, true == root.child_nodes_hash_recovered},

				{"child_nodes.path_btree.Len()", 1, root.child_nodes.path_btree.Len(), int(1) == root.child_nodes.path_btree.Len()},
				{"child_nodes.dirty", true, root.child_nodes.dirty, true == root.child_nodes.dirty},
				{"child_nodes.parent_node", root, root.child_nodes.parent_node, root == root.child_nodes.parent_node},
			}

			for _, tt := range rootTests {
				if !tt.ok {
					t.Errorf("root %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
				}
			}

			{
				_1 := root.child_nodes.path_btree.Get(byte('1')).(*Node)
				if nil == _1 {
					t.Fatalf("unexpect nil pointer for node %v", '1')
				}
				_1Tests := []TEST{
					{"path", []byte{'1'}, _1.path, bytes.Equal([]byte{'1'}, _1.path)},
					{"dirty", true, _1.dirty, true == _1.dirty},
					{"parent_nodes", root.child_nodes, _1.parent_nodes, root.child_nodes == _1.parent_nodes},

					{"node_bytes", nil, _1.node_bytes, nil == _1.node_bytes},
					{"node_hash", "hash.IsNilHash", _1.node_hash, hash.IsNilHash(_1.node_hash)},
					{"node_hash_recovered", true, _1.node_hash_recovered, true == _1.node_hash_recovered},

					{"val", []byte("val_1"), _1.val, bytes.Equal([]byte("val_1"), _1.val)},
					{"val_hash", nil, _1.val_hash, nil == _1.val_hash},
					{"val_hash_recovered", true, _1.val_hash_recovered, true == _1.val_hash_recovered},

					{"child_nodes", nil, _1.child_nodes, nil == _1.child_nodes},
					{"child_nodes_hash", nil, _1.child_nodes_hash, nil == _1.child_nodes_hash},
					{"child_nodes_hash_recovered", true, _1.child_nodes_hash_recovered, true == _1.child_nodes_hash_recovered},
				}

				for _, tt := range _1Tests {
					if !tt.ok {
						t.Errorf("_1 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
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

			rootTests := []TEST{
				{"path", nil, root.path, nil == root.path},
				{"dirty", true, root.dirty, true == root.dirty},
				{"parent_nodes", nil, root.parent_nodes, nil == root.parent_nodes},

				{"node_bytes", nil, root.node_bytes, nil == root.node_bytes},
				{"node_hash", "hash.IsNilHash", root.node_hash, hash.IsNilHash(root.node_hash)},
				{"node_hash_recovered", true, root.node_hash_recovered, true == root.node_hash_recovered},

				{"val", nil, root.val, nil == root.val},
				{"val_hash", nil, root.val_hash, nil == root.val_hash},
				{"val_hash_recovered", true, root.val_hash_recovered, true == root.val_hash_recovered},

				{"child_nodes", "!nil", root.child_nodes, nil != root.child_nodes},
				{"child_nodes_hash", nil, root.child_nodes_hash, nil == root.child_nodes_hash},
				{"child_nodes_hash_recovered", true, root.child_nodes_hash_recovered, true == root.child_nodes_hash_recovered},
			}
			if nil != root.child_nodes {
				rootTests = append(rootTests, []TEST{
					{"child_nodes.path_btree.Len()", 1, root.child_nodes.path_btree.Len(), int(1) == root.child_nodes.path_btree.Len()},
					{"child_nodes.dirty", true, root.child_nodes.dirty, true == root.child_nodes.dirty},
					{"child_nodes.parent_node", root, root.child_nodes.parent_node, root == root.child_nodes.parent_node},
				}...)
			}

			for _, tt := range rootTests {
				if !tt.ok {
					t.Errorf("root %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
				}
			}

			{
				_1 := root.child_nodes.path_btree.Get(byte('1')).(*Node)
				if nil == _1 {
					t.Fatalf("unexpect nil pointer for node full path %v", "1")
				}
				_1Tests := []TEST{
					{"path", []byte{'1'}, _1.path, bytes.Equal([]byte{'1'}, _1.path)},
					{"dirty", true, _1.dirty, true == _1.dirty},
					{"parent_nodes", root.child_nodes, _1.parent_nodes, root.child_nodes == _1.parent_nodes},

					{"node_bytes", nil, _1.node_bytes, nil == _1.node_bytes},
					{"node_hash", "hash.IsNilHash", _1.node_hash, hash.IsNilHash(_1.node_hash)},
					{"node_hash_recovered", true, _1.node_hash_recovered, true == _1.node_hash_recovered},

					{"val", []byte("val_1"), _1.val, bytes.Equal([]byte("val_1"), _1.val)},
					{"val_hash", nil, _1.val_hash, nil == _1.val_hash},
					{"val_hash_recovered", true, _1.val_hash_recovered, true == _1.val_hash_recovered},

					{"child_nodes", "!nil", _1.child_nodes, nil != _1.child_nodes},
					{"child_nodes_hash", nil, _1.child_nodes_hash, nil == _1.child_nodes_hash},
					{"child_nodes_hash_recovered", true, _1.child_nodes_hash_recovered, true == _1.child_nodes_hash_recovered},
				}
				if nil != _1.child_nodes {
					_1Tests = append(_1Tests, []TEST{
						{"child_nodes.path_btree.Len()", 3, _1.child_nodes.path_btree.Len(), int(3) == _1.child_nodes.path_btree.Len()},
						{"child_nodes.dirty", true, _1.child_nodes.dirty, true == _1.child_nodes.dirty},
						{"child_nodes.parent_node", _1, _1.child_nodes.parent_node, _1 == _1.child_nodes.parent_node},
					}...)
				}

				for _, tt := range _1Tests {
					if !tt.ok {
						t.Errorf("_1 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
					}
				}

				{
					_12 := _1.child_nodes.path_btree.Get(byte('2')).(*Node)
					_13 := _1.child_nodes.path_btree.Get(byte('3')).(*Node)
					_14 := _1.child_nodes.path_btree.Get(byte('4')).(*Node)
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
						_12Tests := []TEST{
							{"path", []byte{'2'}, _12.path, bytes.Equal([]byte{'2'}, _12.path)},
							{"dirty", true, _12.dirty, true == _12.dirty},
							{"parent_nodes", _1.child_nodes, _12.parent_nodes, _1.child_nodes == _12.parent_nodes},

							{"node_bytes", nil, _12.node_bytes, nil == _12.node_bytes},
							{"node_hash", "hash.IsNilHash", _12.node_hash, hash.IsNilHash(_12.node_hash)},
							{"node_hash_recovered", true, _12.node_hash_recovered, true == _12.node_hash_recovered},

							{"val", []byte("val_12"), _12.val, bytes.Equal([]byte("val_12"), _12.val)},
							{"val_hash", nil, _12.val_hash, nil == _12.val_hash},
							{"val_hash_recovered", true, _12.val_hash_recovered, true == _12.val_hash_recovered},

							{"child_nodes", "!nil", _12.child_nodes, nil != _12.child_nodes},
							{"child_nodes_hash", nil, _12.child_nodes_hash, nil == _12.child_nodes_hash},
							{"child_nodes_hash_recovered", true, _12.child_nodes_hash_recovered, true == _12.child_nodes_hash_recovered},
						}
						if nil != _12.child_nodes {
							_12Tests = append(_12Tests, []TEST{
								{"child_nodes.path_btree.Len()", 1, _12.child_nodes.path_btree.Len(), int(1) == _12.child_nodes.path_btree.Len()},
								{"child_nodes.dirty", true, _12.child_nodes.dirty, true == _12.child_nodes.dirty},
								{"child_nodes.parent_node", _12, _12.child_nodes.parent_node, _12 == _12.child_nodes.parent_node},
							}...)
						}

						for _, tt := range _12Tests {
							if !tt.ok {
								t.Errorf("_12 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
							}
						}

						{
							_123 := _12.child_nodes.path_btree.Get(byte('3')).(*Node)
							if nil == _123 {
								t.Fatalf("unexpect nil pointer for node full path %v", "123")
							}
							_123Tests := []TEST{
								{"path", []byte{'3'}, _123.path, bytes.Equal([]byte{'3'}, _123.path)},
								{"dirty", true, _123.dirty, true == _123.dirty},
								{"parent_nodes", _12.child_nodes, _123.parent_nodes, _12.child_nodes == _123.parent_nodes},

								{"node_bytes", nil, _123.node_bytes, nil == _123.node_bytes},
								{"node_hash", "hash.IsNilHash", _123.node_hash, hash.IsNilHash(_123.node_hash)},
								{"node_hash_recovered", true, _123.node_hash_recovered, true == _123.node_hash_recovered},

								{"val", []byte("val_123"), _123.val, bytes.Equal([]byte("val_123"), _123.val)},
								{"val_hash", nil, _123.val_hash, nil == _123.val_hash},
								{"val_hash_recovered", true, _123.val_hash_recovered, true == _123.val_hash_recovered},

								{"child_nodes", "!nil", _123.child_nodes, nil != _123.child_nodes},
								{"child_nodes_hash", nil, _123.child_nodes_hash, nil == _123.child_nodes_hash},
								{"child_nodes_hash_recovered", true, _123.child_nodes_hash_recovered, true == _123.child_nodes_hash_recovered},
							}

							if nil != _123.child_nodes {
								_123Tests = append(_123Tests, []TEST{
									{"child_nodes.path_btree.Len()", 1, _123.child_nodes.path_btree.Len(), int(1) == _123.child_nodes.path_btree.Len()},
									{"child_nodes.dirty", true, _123.child_nodes.dirty, true == _123.child_nodes.dirty},
									{"child_nodes.parent_node", _123, _123.child_nodes.parent_node, _123 == _123.child_nodes.parent_node},
								}...)
							}

							for _, tt := range _123Tests {
								if !tt.ok {
									t.Errorf("_123 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
								}
							}

							{
								_1234 := _123.child_nodes.path_btree.Get(byte('4')).(*Node)
								if nil == _1234 {
									t.Fatalf("unexpect nil pointer for node full path %v", "1234")
								}
								_1234Tests := []TEST{
									{"path", []byte{'4'}, _1234.path, bytes.Equal([]byte{'4'}, _1234.path)},
									{"dirty", true, _1234.dirty, true == _1234.dirty},
									{"parent_nodes", _123.child_nodes, _1234.parent_nodes, _123.child_nodes == _1234.parent_nodes},

									{"node_bytes", nil, _1234.node_bytes, nil == _1234.node_bytes},
									{"node_hash", "hash.IsNilHash", _1234.node_hash, hash.IsNilHash(_1234.node_hash)},
									{"node_hash_recovered", true, _1234.node_hash_recovered, true == _1234.node_hash_recovered},

									{"val", []byte("val_1234"), _1234.val, bytes.Equal([]byte("val_1234"), _1234.val)},
									{"val_hash", nil, _1234.val_hash, nil == _1234.val_hash},
									{"val_hash_recovered", true, _1234.val_hash_recovered, true == _1234.val_hash_recovered},

									{"child_nodes", nil, _1234.child_nodes, nil == _1234.child_nodes},
									{"child_nodes_hash", nil, _1234.child_nodes_hash, nil == _1234.child_nodes_hash},
									{"child_nodes_hash_recovered", true, _1234.child_nodes_hash_recovered, true == _1234.child_nodes_hash_recovered},
								}
								if nil != _1234.child_nodes {
									_1234Tests = append(_1234Tests, []TEST{
										{"child_nodes.path_btree.Len()", 1, _1234.child_nodes.path_btree.Len(), int(1) == _1234.child_nodes.path_btree.Len()},
										{"child_nodes.dirty", true, _1234.child_nodes.dirty, true == _1234.child_nodes.dirty},
										{"child_nodes.parent_node", _1234, _1234.child_nodes.parent_node, _1234 == _1234.child_nodes.parent_node},
									}...)
								}

								for _, tt := range _1234Tests {
									if !tt.ok {
										t.Errorf("_1234 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
									}
								}
							}
						}
					}
					{
						_13Tests := []TEST{
							{"path", []byte{'3'}, _13.path, bytes.Equal([]byte{'3'}, _13.path)},
							{"dirty", true, _13.dirty, true == _13.dirty},
							{"parent_nodes", _1.child_nodes, _13.parent_nodes, _1.child_nodes == _13.parent_nodes},

							{"node_bytes", nil, _13.node_bytes, nil == _13.node_bytes},
							{"node_hash", "hash.IsNilHash", _13.node_hash, hash.IsNilHash(_13.node_hash)},
							{"node_hash_recovered", true, _13.node_hash_recovered, true == _13.node_hash_recovered},

							{"val", []byte("val_13"), _13.val, bytes.Equal([]byte("val_13"), _13.val)},
							{"val_hash", nil, _13.val_hash, nil == _13.val_hash},
							{"val_hash_recovered", true, _13.val_hash_recovered, true == _13.val_hash_recovered},

							{"child_nodes", nil, _13.child_nodes, nil == _13.child_nodes},
							{"child_nodes_hash", nil, _13.child_nodes_hash, nil == _13.child_nodes_hash},
							{"child_nodes_hash_recovered", true, _13.child_nodes_hash_recovered, true == _13.child_nodes_hash_recovered},
						}
						if nil != _13.child_nodes {
							_13Tests = append(_13Tests, []TEST{
								{"child_nodes.path_btree.Len()", 3, _13.child_nodes.path_btree.Len(), int(3) == _13.child_nodes.path_btree.Len()},
								{"child_nodes.dirty", true, _13.child_nodes.dirty, true == _13.child_nodes.dirty},
								{"child_nodes.parent_node", _13, _13.child_nodes.parent_node, _13 == _13.child_nodes.parent_node},
							}...)
						}

						for _, tt := range _13Tests {
							if !tt.ok {
								t.Errorf("_13 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
							}
						}
					}
					{
						_14Tests := []TEST{
							{"path", []byte{'4'}, _14.path, bytes.Equal([]byte{'4'}, _14.path)},
							{"dirty", true, _14.dirty, true == _14.dirty},
							{"parent_nodes", _1.child_nodes, _14.parent_nodes, _1.child_nodes == _14.parent_nodes},

							{"node_bytes", nil, _14.node_bytes, nil == _14.node_bytes},
							{"node_hash", "hash.IsNilHash", _14.node_hash, hash.IsNilHash(_14.node_hash)},
							{"node_hash_recovered", true, _14.node_hash_recovered, true == _14.node_hash_recovered},

							{"val", []byte("val_14"), _14.val, bytes.Equal([]byte("val_14"), _14.val)},
							{"val_hash", nil, _14.val_hash, nil == _14.val_hash},
							{"val_hash_recovered", true, _14.val_hash_recovered, true == _14.val_hash_recovered},

							{"child_nodes", nil, _14.child_nodes, nil == _14.child_nodes},
							{"child_nodes_hash", nil, _14.child_nodes_hash, nil == _14.child_nodes_hash},
							{"child_nodes_hash_recovered", true, _14.child_nodes_hash_recovered, true == _14.child_nodes_hash_recovered},
						}
						if nil != _14.child_nodes {
							_14Tests = append(_14Tests, []TEST{
								{"child_nodes.path_btree.Len()", 3, _14.child_nodes.path_btree.Len(), int(3) == _14.child_nodes.path_btree.Len()},
								{"child_nodes.dirty", true, _14.child_nodes.dirty, true == _14.child_nodes.dirty},
								{"child_nodes.parent_node", _14, _14.child_nodes.parent_node, _14 == _14.child_nodes.parent_node},
							}...)
						}

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
				rootTests := []TEST{
					{"path", nil, root.path, nil == root.path},
					{"dirty", true, root.dirty, true == root.dirty},
					{"parent_nodes", nil, root.parent_nodes, nil == root.parent_nodes},

					{"node_bytes", "!nil", root.node_bytes, nil != root.node_bytes},
					{"node_hash", "!hash.IsNilHash", root.node_hash, !hash.IsNilHash(root.node_hash)},
					{"node_hash_recovered", true, root.node_hash_recovered, true == root.node_hash_recovered},

					{"val", nil, root.val, nil == root.val},
					{"val_hash", nil, root.val_hash, nil == root.val_hash},
					{"val_hash_recovered", true, root.val_hash_recovered, true == root.val_hash_recovered},

					{"child_nodes", "!nil", root.child_nodes, nil != root.child_nodes},
					{"child_nodes_hash", "!hash.IsNilHash", root.child_nodes_hash, !hash.IsNilHash(root.child_nodes_hash)},
					{"child_nodes_hash_recovered", true, root.child_nodes_hash_recovered, true == root.child_nodes_hash_recovered},
				}
				if nil != root.child_nodes {
					rootTests = append(rootTests, []TEST{
						{"child_nodes.path_btree.Len()", 2, root.child_nodes.path_btree.Len(), int(2) == root.child_nodes.path_btree.Len()},
						{"child_nodes.dirty", true, root.child_nodes.dirty, true == root.child_nodes.dirty},
						{"child_nodes.parent_node", root, root.child_nodes.parent_node, root == root.child_nodes.parent_node},
					}...)
				}

				for _, tt := range rootTests {
					if !tt.ok {
						t.Errorf("root %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
					}
				}
			}
			{
				_1 := root.child_nodes.path_btree.Get(byte('1')).(*Node)
				_2 := root.child_nodes.path_btree.Get(byte('2')).(*Node)
				if nil == _1 {
					t.Fatalf("unexpect nil pointer for node full path %v", "1")
				}
				if nil == _2 {
					t.Fatalf("unexpect nil pointer for node full path %v", "2")
				}

				{
					_1Tests := []TEST{
						{"path", []byte("1"), _1.path, bytes.Equal([]byte("1"), _1.path)},
						{"dirty", true, _1.dirty, true == _1.dirty},
						{"parent_nodes", root, _1.parent_nodes, root.child_nodes == _1.parent_nodes},

						{"node_bytes", "!nil", _1.node_bytes, nil != _1.node_bytes},
						{"node_hash", "!hash.IsNilHash", _1.node_hash, !hash.IsNilHash(_1.node_hash)},
						{"node_hash_recovered", true, _1.node_hash_recovered, true == _1.node_hash_recovered},

						{"val", []byte("val_1"), _1.val, bytes.Equal([]byte("val_1"), _1.val)},
						{"val_hash", "!hash.IsNilHash", _1.val_hash, !hash.IsNilHash(_1.val_hash)},
						{"val_hash_recovered", true, _1.val_hash_recovered, true == _1.val_hash_recovered},

						{"child_nodes", "!nil", _1.child_nodes, nil != _1.child_nodes},
						{"child_nodes_hash", "!hash.IsNilHash", _1.child_nodes_hash, !hash.IsNilHash(_1.child_nodes_hash)},
						{"child_nodes_hash_recovered", true, _1.child_nodes_hash_recovered, true == _1.child_nodes_hash_recovered},
					}
					if nil != _1.child_nodes {
						_1Tests = append(_1Tests, []TEST{
							{"child_nodes.path_btree.Len()", 4, _1.child_nodes.path_btree.Len(), int(4) == _1.child_nodes.path_btree.Len()},
							{"child_nodes.dirty", true, _1.child_nodes.dirty, true == _1.child_nodes.dirty},
							{"child_nodes.parent_node", _1, _1.child_nodes.parent_node, _1 == _1.child_nodes.parent_node},
						}...)
					}

					for _, tt := range _1Tests {
						if !tt.ok {
							t.Errorf("_1 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
						}
					}
				}
				{
					_2Tests := []TEST{
						{"path", []byte("2"), _2.path, bytes.Equal([]byte("2"), _2.path)},
						{"dirty", true, _2.dirty, true == _2.dirty},
						{"parent_nodes", root, _2.parent_nodes, root.child_nodes == _2.parent_nodes},

						{"node_bytes", "!nil", _2.node_bytes, nil != _2.node_bytes},
						{"node_hash", "!hash.IsNilHash", _2.node_hash, !hash.IsNilHash(_2.node_hash)},
						{"node_hash_recovered", true, _2.node_hash_recovered, true == _2.node_hash_recovered},

						{"val", []byte("val_2"), _2.val, bytes.Equal([]byte("val_2"), _2.val)},
						{"val_hash", "!hash.IsNilHash", _2.val_hash, !hash.IsNilHash(_2.val_hash)},
						{"val_hash_recovered", true, _2.val_hash_recovered, true == _2.val_hash_recovered},

						{"child_nodes", nil, _2.child_nodes, nil == _2.child_nodes},
						{"child_nodes_hash", "hash.IsNilHash", _2.child_nodes_hash, hash.IsNilHash(_2.child_nodes_hash)},
						{"child_nodes_hash_recovered", true, _2.child_nodes_hash_recovered, true == _2.child_nodes_hash_recovered},
					}
					if nil != _2.child_nodes {
						_2Tests = append(_2Tests, []TEST{
							{"child_nodes.path_btree.Len()", 2, _2.child_nodes.path_btree.Len(), int(2) == _2.child_nodes.path_btree.Len()},
							{"child_nodes.dirty", true, _2.child_nodes.dirty, true == _2.child_nodes.dirty},
							{"child_nodes.parent_node", _2, _2.child_nodes.parent_node, _2 == _2.child_nodes.parent_node},
						}...)
					}

					for _, tt := range _2Tests {
						if !tt.ok {
							t.Errorf("_2 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
						}
					}
				}

				if nil != _1.child_nodes {
					_12 := _1.child_nodes.path_btree.Get(byte('2')).(*Node)
					_13 := _1.child_nodes.path_btree.Get(byte('3')).(*Node)
					_14 := _1.child_nodes.path_btree.Get(byte('4')).(*Node)
					_1a := _1.child_nodes.path_btree.Get(byte('a')).(*Node)
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
						_12Tests := []TEST{
							{"path", []byte("2"), _12.path, bytes.Equal([]byte("2"), _12.path)},
							{"dirty", true, _12.dirty, true == _12.dirty},
							{"parent_nodes", _1, _12.parent_nodes, _1.child_nodes == _12.parent_nodes},

							{"node_bytes", "!nil", _12.node_bytes, nil != _12.node_bytes},
							{"node_hash", "!hash.IsNilHash", _12.node_hash, !hash.IsNilHash(_12.node_hash)},
							{"node_hash_recovered", true, _12.node_hash_recovered, true == _12.node_hash_recovered},

							{"val", []byte("val_12"), _12.val, bytes.Equal([]byte("val_12"), _12.val)},
							{"val_hash", "!hash.IsNilHash", _12.val_hash, !hash.IsNilHash(_12.val_hash)},
							{"val_hash_recovered", true, _12.val_hash_recovered, true == _12.val_hash_recovered},

							{"child_nodes", "!nil", _12.child_nodes, nil != _12.child_nodes},
							{"child_nodes_hash", "!hash.IsNilHash", _12.child_nodes_hash, !hash.IsNilHash(_12.child_nodes_hash)},
							{"child_nodes_hash_recovered", true, _12.child_nodes_hash_recovered, true == _12.child_nodes_hash_recovered},
						}
						if nil != _12.child_nodes {
							_12Tests = append(_12Tests, []TEST{
								{"child_nodes.path_btree.Len()", 2, _12.child_nodes.path_btree.Len(), int(2) == _12.child_nodes.path_btree.Len()},
								{"child_nodes.dirty", true, _12.child_nodes.dirty, true == _12.child_nodes.dirty},
								{"child_nodes.parent_node", _12, _12.child_nodes.parent_node, _12 == _12.child_nodes.parent_node},
							}...)
						}

						for _, tt := range _12Tests {
							if !tt.ok {
								t.Errorf("_12 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
							}
						}
					}
					{
						_13Tests := []TEST{
							{"path", []byte("3"), _13.path, bytes.Equal([]byte("3"), _13.path)},
							{"dirty", false, _13.dirty, false == _13.dirty},
							{"parent_nodes", _1, _13.parent_nodes, _1.child_nodes == _13.parent_nodes},

							{"node_bytes", nil, _13.node_bytes, nil == _13.node_bytes},
							{"node_hash", "!hash.IsNilHash", _13.node_hash, !hash.IsNilHash(_13.node_hash)},
							{"node_hash_recovered", false, _13.node_hash_recovered, false == _13.node_hash_recovered},

							{"val", nil, _13.val, nil == _13.val},
							{"val_hash", "hash.IsNilHash", _13.val_hash, hash.IsNilHash(_13.val_hash)},
							{"val_hash_recovered", false, _13.val_hash_recovered, false == _13.val_hash_recovered},

							{"child_nodes", nil, _13.child_nodes, nil == _13.child_nodes},
							{"child_nodes_hash", "hash.IsNilHash", _13.child_nodes_hash, hash.IsNilHash(_13.child_nodes_hash)},
							{"child_nodes_hash_recovered", false, _13.child_nodes_hash_recovered, false == _13.child_nodes_hash_recovered},
						}
						if nil != _13.child_nodes {
							_13Tests = append(_13Tests, []TEST{
								{"child_nodes.path_btree.Len()", 2, _13.child_nodes.path_btree.Len(), int(2) == _13.child_nodes.path_btree.Len()},
								{"child_nodes.dirty", true, _13.child_nodes.dirty, true == _13.child_nodes.dirty},
								{"child_nodes.parent_node", _13, _13.child_nodes.parent_node, _13 == _13.child_nodes.parent_node},
							}...)
						}

						for _, tt := range _13Tests {
							if !tt.ok {
								t.Errorf("_13 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
							}
						}
					}
					{
						_14Tests := []TEST{
							{"path", []byte("4"), _14.path, bytes.Equal([]byte("4"), _14.path)},
							{"dirty", false, _14.dirty, false == _14.dirty},
							{"parent_nodes", _1, _14.parent_nodes, _1.child_nodes == _14.parent_nodes},

							{"node_bytes", nil, _14.node_bytes, nil == _14.node_bytes},
							{"node_hash", "!hash.IsNilHash", _14.node_hash, !hash.IsNilHash(_14.node_hash)},
							{"node_hash_recovered", false, _14.node_hash_recovered, false == _14.node_hash_recovered},

							{"val", nil, _14.val, nil == _14.val},
							{"val_hash", "hash.IsNilHash", _14.val_hash, hash.IsNilHash(_14.val_hash)},
							{"val_hash_recovered", false, _14.val_hash_recovered, false == _14.val_hash_recovered},

							{"child_nodes", nil, _14.child_nodes, nil == _14.child_nodes},
							{"child_nodes_hash", "hash.IsNilHash", _14.child_nodes_hash, hash.IsNilHash(_14.child_nodes_hash)},
							{"child_nodes_hash_recovered", false, _14.child_nodes_hash_recovered, false == _14.child_nodes_hash_recovered},
						}
						if nil != _14.child_nodes {
							_14Tests = append(_14Tests, []TEST{
								{"child_nodes.path_btree.Len()", 2, _14.child_nodes.path_btree.Len(), int(2) == _14.child_nodes.path_btree.Len()},
								{"child_nodes.dirty", true, _14.child_nodes.dirty, true == _14.child_nodes.dirty},
								{"child_nodes.parent_node", _14, _14.child_nodes.parent_node, _14 == _14.child_nodes.parent_node},
							}...)
						}

						for _, tt := range _14Tests {
							if !tt.ok {
								t.Errorf("_14 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
							}
						}
					}
					{
						_1aTests := []TEST{
							{"path", []byte("a"), _1a.path, bytes.Equal([]byte("a"), _1a.path)},
							{"dirty", true, _1a.dirty, true == _1a.dirty},
							{"parent_nodes", _1, _1a.parent_nodes, _1.child_nodes == _1a.parent_nodes},

							{"node_bytes", "!nil", _1a.node_bytes, nil != _1a.node_bytes},
							{"node_hash", "!hash.IsNilHash", _1a.node_hash, !hash.IsNilHash(_1a.node_hash)},
							{"node_hash_recovered", true, _1a.node_hash_recovered, true == _1a.node_hash_recovered},

							{"val", []byte("val_1a"), _1a.val, bytes.Equal([]byte("val_1a"), _1a.val)},
							{"val_hash", "!hash.IsNilHash", _1a.val_hash, !hash.IsNilHash(_1a.val_hash)},
							{"val_hash_recovered", true, _1a.val_hash_recovered, true == _1a.val_hash_recovered},

							{"child_nodes", nil, _1a.child_nodes, nil == _1a.child_nodes},
							{"child_nodes_hash", "hash.IsNilHash", _1a.child_nodes_hash, hash.IsNilHash(_1a.child_nodes_hash)},
							{"child_nodes_hash_recovered", true, _1a.child_nodes_hash_recovered, true == _1a.child_nodes_hash_recovered},
						}
						if nil != _1a.child_nodes {
							_1aTests = append(_1aTests, []TEST{
								{"child_nodes.path_btree.Len()", 2, _1a.child_nodes.path_btree.Len(), int(2) == _1a.child_nodes.path_btree.Len()},
								{"child_nodes.dirty", true, _1a.child_nodes.dirty, true == _1a.child_nodes.dirty},
								{"child_nodes.parent_node", _1a, _1a.child_nodes.parent_node, _1a == _1a.child_nodes.parent_node},
							}...)
						}

						for _, tt := range _1aTests {
							if !tt.ok {
								t.Errorf("_1a node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
							}
						}
					}

					if nil != _12.child_nodes {
						_123 := _12.child_nodes.path_btree.Get(byte('3')).(*Node)
						_12a := _12.child_nodes.path_btree.Get(byte('a')).(*Node)
						if nil == _123 {
							t.Fatalf("unexpect nil pointer for node full path %v", "123")
						}
						if nil == _12a {
							t.Fatalf("unexpect nil pointer for node full path %v", "12a")
						}

						{
							_123Tests := []TEST{
								{"path", []byte("3"), _123.path, bytes.Equal([]byte("3"), _123.path)},
								{"dirty", true, _123.dirty, true == _123.dirty},
								{"parent_nodes", _12, _123.parent_nodes, _12.child_nodes == _123.parent_nodes},

								{"node_bytes", "!nil", _123.node_bytes, nil != _123.node_bytes},
								{"node_hash", "!hash.IsNilHash", _123.node_hash, !hash.IsNilHash(_123.node_hash)},
								{"node_hash_recovered", true, _123.node_hash_recovered, true == _123.node_hash_recovered},

								{"val", []byte("val_123"), _123.val, bytes.Equal([]byte("val_123"), _123.val)},
								{"val_hash", "!hash.IsNilHash", _123.val_hash, !hash.IsNilHash(_123.val_hash)},
								{"val_hash_recovered", true, _123.val_hash_recovered, true == _123.val_hash_recovered},

								{"child_nodes", "!nil", _123.child_nodes, nil != _123.child_nodes},
								{"child_nodes_hash", "!hash.IsNilHash", _123.child_nodes_hash, !hash.IsNilHash(_123.child_nodes_hash)},
								{"child_nodes_hash_recovered", true, _123.child_nodes_hash_recovered, true == _123.child_nodes_hash_recovered},
							}
							if nil != _123.child_nodes {
								_123Tests = append(_123Tests, []TEST{
									{"child_nodes.path_btree.Len()", 2, _123.child_nodes.path_btree.Len(), int(2) == _123.child_nodes.path_btree.Len()},
									{"child_nodes.dirty", true, _123.child_nodes.dirty, true == _123.child_nodes.dirty},
									{"child_nodes.parent_node", _123, _123.child_nodes.parent_node, _123 == _123.child_nodes.parent_node},
								}...)
							}

							for _, tt := range _123Tests {
								if !tt.ok {
									t.Errorf("_123 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
								}
							}
						}
						{
							_12aTests := []TEST{
								{"path", []byte("a"), _12a.path, bytes.Equal([]byte("a"), _12a.path)},
								{"dirty", true, _12a.dirty, true == _12a.dirty},
								{"parent_nodes", _12, _12a.parent_nodes, _12.child_nodes == _12a.parent_nodes},

								{"node_bytes", "!nil", _12a.node_bytes, nil != _12a.node_bytes},
								{"node_hash", "!hash.IsNilHash", _12a.node_hash, !hash.IsNilHash(_12a.node_hash)},
								{"node_hash_recovered", true, _12a.node_hash_recovered, true == _12a.node_hash_recovered},

								{"val", []byte("val_12a"), _12a.val, bytes.Equal([]byte("val_12a"), _12a.val)},
								{"val_hash", "!hash.IsNilHash", _12a.val_hash, !hash.IsNilHash(_12a.val_hash)},
								{"val_hash_recovered", true, _12a.val_hash_recovered, true == _12a.val_hash_recovered},

								{"child_nodes", nil, _12a.child_nodes, nil == _12a.child_nodes},
								{"child_nodes_hash", "hash.IsNilHash", _12a.child_nodes_hash, hash.IsNilHash(_12a.child_nodes_hash)},
								{"child_nodes_hash_recovered", true, _12a.child_nodes_hash_recovered, true == _12a.child_nodes_hash_recovered},
							}
							if nil != _12a.child_nodes {
								_12aTests = append(_12aTests, []TEST{
									{"child_nodes.path_btree.Len()", 2, _12a.child_nodes.path_btree.Len(), int(2) == _12a.child_nodes.path_btree.Len()},
									{"child_nodes.dirty", true, _12a.child_nodes.dirty, true == _12a.child_nodes.dirty},
									{"child_nodes.parent_node", _12a, _12a.child_nodes.parent_node, _12a == _12a.child_nodes.parent_node},
								}...)
							}

							for _, tt := range _12aTests {
								if !tt.ok {
									t.Errorf("_12a node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
								}
							}
						}

						if nil != _123.child_nodes {
							_1234 := _123.child_nodes.path_btree.Get(byte('4')).(*Node)
							_123a := _123.child_nodes.path_btree.Get(byte('a')).(*Node)
							if nil == _1234 {
								t.Fatalf("unexpect nil pointer for node full path %v", "1234")
							}
							if nil == _123a {
								t.Fatalf("unexpect nil pointer for node full path %v", "123a")
							}

							{
								_1234Tests := []TEST{
									{"path", []byte("4"), _1234.path, bytes.Equal([]byte("4"), _1234.path)},
									{"dirty", false, _1234.dirty, false == _1234.dirty},
									{"parent_nodes", _123, _1234.parent_nodes, _123.child_nodes == _1234.parent_nodes},

									{"node_bytes", nil, _1234.node_bytes, nil == _1234.node_bytes},
									{"node_hash", "!hash.IsNilHash", _1234.node_hash, !hash.IsNilHash(_1234.node_hash)},
									{"node_hash_recovered", false, _1234.node_hash_recovered, false == _1234.node_hash_recovered},

									{"val", nil, _1234.val, nil == _1234.val},
									{"val_hash", "hash.IsNilHash", _1234.val_hash, hash.IsNilHash(_1234.val_hash)},
									{"val_hash_recovered", false, _1234.val_hash_recovered, false == _1234.val_hash_recovered},

									{"child_nodes", nil, _1234.child_nodes, nil == _1234.child_nodes},
									{"child_nodes_hash", "hash.IsNilHash", _1234.child_nodes_hash, hash.IsNilHash(_1234.child_nodes_hash)},
									{"child_nodes_hash_recovered", false, _1234.child_nodes_hash_recovered, false == _1234.child_nodes_hash_recovered},
								}
								if nil != _1234.child_nodes {
									_1234Tests = append(_1234Tests, []TEST{
										{"child_nodes.path_btree.Len()", 2, _1234.child_nodes.path_btree.Len(), int(2) == _1234.child_nodes.path_btree.Len()},
										{"child_nodes.dirty", true, _1234.child_nodes.dirty, true == _1234.child_nodes.dirty},
										{"child_nodes.parent_node", _1234, _1234.child_nodes.parent_node, _1234 == _1234.child_nodes.parent_node},
									}...)
								}

								for _, tt := range _1234Tests {
									if !tt.ok {
										t.Errorf("_1234 node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
									}
								}
							}
							{
								_123aTests := []TEST{
									{"path", []byte("a"), _123a.path, bytes.Equal([]byte("a"), _123a.path)},
									{"dirty", true, _123a.dirty, true == _123a.dirty},
									{"parent_nodes", _123, _123a.parent_nodes, _123.child_nodes == _123a.parent_nodes},

									{"node_bytes", "!nil", _123a.node_bytes, nil != _123a.node_bytes},
									{"node_hash", "!hash.IsNilHash", _123a.node_hash, !hash.IsNilHash(_123a.node_hash)},
									{"node_hash_recovered", true, _123a.node_hash_recovered, true == _123a.node_hash_recovered},

									{"val", []byte("val_123a"), _123a.val, bytes.Equal([]byte("val_123a"), _123a.val)},
									{"val_hash", "!hash.IsNilHash", _123a.val_hash, !hash.IsNilHash(_123a.val_hash)},
									{"val_hash_recovered", true, _123a.val_hash_recovered, true == _123a.val_hash_recovered},

									{"child_nodes", nil, _123a.child_nodes, nil == _123a.child_nodes},
									{"child_nodes_hash", "hash.IsNilHash", _123a.child_nodes_hash, hash.IsNilHash(_123a.child_nodes_hash)},
									{"child_nodes_hash_recovered", true, _123a.child_nodes_hash_recovered, true == _123a.child_nodes_hash_recovered},
								}
								if nil != _123a.child_nodes {
									_123aTests = append(_123aTests, []TEST{
										{"child_nodes.path_btree.Len()", 2, _123a.child_nodes.path_btree.Len(), int(2) == _123a.child_nodes.path_btree.Len()},
										{"child_nodes.dirty", true, _123a.child_nodes.dirty, true == _123a.child_nodes.dirty},
										{"child_nodes.parent_node", _123a, _123a.child_nodes.parent_node, _123a == _123a.child_nodes.parent_node},
									}...)
								}

								for _, tt := range _123aTests {
									if !tt.ok {
										t.Errorf("_123a node %s expect: %v, but: %v", tt.label, tt.expected, tt.actual)
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

	// test Get operation to influence the lazy status
	{
		tdb, err := testPrepareTrieDB(db_path, rootHash)

		if nil != err {
			t.Fatal(err)
		}

		_, err = tdb.Get([]byte("13"))
		if nil != err {
			t.Fatal(`Get item "13" should receive node_bytes of that kv item!`)
		}

		err = tdb.Delete([]byte("12"))
		if nil != err {
			t.Fatal(`Delete item "12" should work as expected!`)
		}

		tdb.GenDotFile("./test_mainworkflow_3.dot", false)
		testCloseTrieDB(tdb)
	}

	// test delete data which not exist in a nonempty triedb
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
