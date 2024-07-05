package triedb

import (
	"os"
	"testing"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/hash"
	"github.com/xlander-io/kv_leveldb"
)

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

		tdb.Update([]byte("1"), []byte([]byte("val_1")))
		tdb.Update([]byte("12"), []byte([]byte("val_12")))
		tdb.Update([]byte("13"), []byte([]byte("val_13")))
		tdb.Update([]byte("14"), []byte([]byte("val_14")))
		tdb.Update([]byte("123"), []byte([]byte("val_123")))
		tdb.Update([]byte("1234"), []byte([]byte("val_1234")))

		rootHash, _, _, _ = tdb.CalHash()
		//tdb.GenDotFile("./trie_test_simple.dot", false)

		kvdb.Close()
	}

	{
		kvdb, _ := kv_leveldb.NewDB(db_path)
		//
		c, _ := cache.New(nil)
		tdb, _ := NewTrieDB(kvdb, c, &TrieDBConfig{
			Root_hash:           rootHash.Clone(),
			Commit_thread_limit: 1,
		})

		err := tdb.Update([]byte("2"), []byte([]byte("val_1")))
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
