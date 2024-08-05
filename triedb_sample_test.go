package triedb

import (
	"fmt"
	"testing"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv_leveldb"
)

func TestReadTrieDB(t *testing.T) {

	kvdb, err := kv_leveldb.NewDB("./test2.db")
	if err != nil {
		t.Error(err)
		return
	}

	c, err := cache.New(nil)
	if err != nil {
		t.Error(err)
		return
	}
	tdb, err := NewTrieDB(kvdb, c, &TrieDBConfig{
		Root_hash:           nil,
		Commit_thread_limit: 10,
	})

	if err != nil {
		t.Error(err)
		return
	}

	tdb.Put(Path([]byte("a"), []byte("a"), []byte("a")), []byte("valaaa"), false)

	n, err := tdb.Get(Path([]byte("a"), []byte("a"), []byte("a")))
	if err != nil {
		t.Error(err)
		return
	}

	if n != nil {
		fmt.Println(string(n.val))
	}

	fmt.Println(tdb.Del(Path([]byte("a"), []byte("a"), []byte("a"))))
	fmt.Println(tdb.Del(Path([]byte("a"), []byte("a"), []byte("a"))))

}

// func TestUpdateTrieDB(t *testing.T) {

// 	kvdb, err := kv_leveldb.NewDB("./test.db")
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	c, err := cache.New(nil)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	tdb, err := NewTrieDB(kvdb, c, &TrieDBConfig{
// 		Root_hash:           nil,
// 		Commit_thread_limit: 10,
// 	})

// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}

// 	tdb.Put(Path([]byte("12")), []byte("val12"), true)
// 	tdb.Put(Path([]byte("1a")), []byte("val1a"), true)
// 	tdb.Put(Path([]byte("1b")), []byte("val1b"), true)
// 	tdb.Put(Path([]byte("1ab")), []byte("val1ab"), true)
// 	tdb.Put(Path([]byte("123")), []byte("val123"), true)
// 	tdb.Put(Path([]byte("12a")), []byte("val12a"), true)
// 	tdb.Put(Path([]byte("12b")), []byte("val12b"), true)
// 	tdb.Put(Path([]byte("1234")), []byte("val1234"), true)
// 	tdb.Put(Path([]byte("123a")), []byte("val123a"), true)
// 	tdb.Put(Path([]byte("a"), []byte("a"), []byte("a")), []byte("valaaa"), true)
// 	tdb.Put(Path([]byte("ab"), []byte("cd")), []byte("valabcd"), true)

// 	root_hash, updated, deleted, _ := tdb.Commit()

// 	fmt.Println("commit root hash:", root_hash.Hex())
// 	fmt.Println("commit updated len:", len(updated))
// 	fmt.Println("commit deleted len:", len(deleted))

// 	b := kv.NewBatch()

// 	for key, val := range updated {
// 		b.Put([]byte(key), val)
// 	}

// 	for key, _ := range deleted {
// 		b.Delete([]byte(key))
// 	}

// 	err = kvdb.WriteBatch(b, true)
// 	if err != nil {
// 		fmt.Println("del batch err", err)
// 	}

// }
