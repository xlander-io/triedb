package triedb

import (
	"os"
	"testing"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/hash"
	"github.com/xlander-io/kv"
	"github.com/xlander-io/kv_leveldb"
)

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

func TestSimple(t *testing.T) {

	const db_path = "./trie_test_simple.db"
	os.RemoveAll(db_path)
	var rootHash *hash.Hash
	{
		kvdb, _ := kv_leveldb.NewDB(db_path)
		//
		c, _ := cache.New(nil)
		tdb, _ := NewTrieDB(kvdb, c, &TrieDBConfig{
			Commit_thread_limit: 1,
		})

		tdb.Update([]byte("1"), []byte("val_1"))
		tdb.Update([]byte("12"), []byte("val_12"))
		tdb.Update([]byte("13"), []byte("val_13"))
		tdb.Update([]byte("14"), []byte("val_14"))
		tdb.Update([]byte("123"), []byte("val_123"))
		tdb.Update([]byte("1234"), []byte("val_1234"))

		_rootHash, err := tdb.testCommit()

		if nil != err {
			t.Fatal(err)
		}

		rootHash = _rootHash

		kvdb.Close()
	}

	{
		kvdb, err := kv_leveldb.NewDB(db_path)
		if nil != err {
			t.Fatal(err)
		}
		//
		c, err := cache.New(nil)
		if nil != err {
			t.Fatal(err)
		}
		tdb, err := NewTrieDB(kvdb, c, &TrieDBConfig{
			Root_hash:           rootHash.Clone(),
			Commit_thread_limit: 1,
		})
		if nil != err {
			t.Fatal(err)
		}

		err = tdb.Update([]byte("2"), []byte("val_1"))
		if nil != err {
			t.Fatal(err)
		}
		//tdb.Update([]byte("1a"), []byte([]byte("val_1a")))
		//tdb.Update([]byte("12a"), []byte([]byte("val_12a")))
		//tdb.Update([]byte("123a"), []byte([]byte("val_123a")))

		tdb.CalHash()
		tdb.GenDotFile("./trie_test_simple.dot", false)
		kvdb.Close()
	}
}
