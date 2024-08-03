package triedb

import (
	"fmt"
	"testing"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv_leveldb"
)

func TestUpdateTrieDB(t *testing.T) {

	kvdb, err := kv_leveldb.NewDB("./test.db")
	if err != nil {
		t.Error(err)
	}

	c, err := cache.New(nil)
	if err != nil {
		t.Error(err)
	}
	tdb, err := NewTrieDB(kvdb, c, &TrieDBConfig{
		Root_hash:           nil,
		Commit_thread_limit: 1,
	})

	if err != nil {
		t.Error(err)
	}

	n, update_err := tdb.Put(Path([]byte("abc")), []byte("valabc"), true)
	if update_err != nil {
		t.Error(update_err)
	}
	fmt.Println("flat path:", n.node_path_flat_str())

	////////////

	n, update_err = tdb.Put(Path([]byte("ab")), []byte("valab"), true)
	if update_err != nil {
		t.Error(update_err)
	}
	fmt.Println("flat path:", n.node_path_flat_str())

	/////
	n, update_err = tdb.Put(Path([]byte("a")), []byte("vala"), true)
	if update_err != nil {
		t.Error(update_err)
	}
	fmt.Println("flat path:", n.node_path_flat_str())

	/////
	n, update_err = tdb.Put(Path([]byte("a"), []byte("a"), []byte("a")), []byte("valaaa"), true)
	if update_err != nil {
		t.Error(update_err)
	}
	fmt.Println("flat path:", n.node_path_flat_str())

	/////
	n, update_err = tdb.Put(Path([]byte("ab"), []byte("cd")), []byte("valabcd"), true)
	if update_err != nil {
		t.Error(update_err)
	}
	fmt.Println("flat path:", n.node_path_flat_str())

	///
	fmt.Println(tdb.Del(Path([]byte("ab"), []byte("cd"))))

	///
	fmt.Println(tdb.Get(Path([]byte("a"), []byte("a"), []byte("a"))))

	root_hash, updated, deleted, _ := tdb.Commit()

	fmt.Println("commit root hash:", root_hash)
	fmt.Println("commit updated:", updated)
	fmt.Println("commit deleted:", deleted)

	fmt.Println("end")

}
