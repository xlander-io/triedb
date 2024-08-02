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
		Root_hash: nil,
	})

	if err != nil {
		t.Error(err)
	}

	n, update_err := tdb.Update(Path([]byte("abc")), []byte("val"), true)
	if update_err != nil {
		t.Error(update_err)
	}

	fmt.Println("flat path:", n.node_path_flat_str())

	update_n, update_err := tdb.Update(Path([]byte("abc"), []byte("de")), []byte("val2"), true)
	if update_err != nil {
		t.Error(update_err)
	}

	fmt.Println("flat path:", update_n.node_path_flat_str())

	update_n, update_err = tdb.Update(Path([]byte("a"), []byte("de")), []byte("val2"), true)
	if update_err != nil {
		t.Error(update_err)
	}
	fmt.Println("flat path:", update_n.node_path_flat_str())

	fmt.Println("flat path:", tdb.root_node.node_path_flat_str())

	fmt.Println(tdb.Del(Path([]byte("abc"))))
	fmt.Println(tdb.Del(Path([]byte("adsf"), []byte("adsf"))))
	fmt.Println(tdb.Del(Path([]byte("abc"), []byte("de"))))
	fmt.Println(tdb.Del(Path([]byte("a"))))
	fmt.Println(tdb.Del(Path([]byte("a"), []byte("de"))))

}
