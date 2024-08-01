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

	_, update_err := tdb.Update(Path([]byte("abc")), []byte("val"), true)
	if update_err != nil {
		t.Error(update_err)
	}

	update_n, update_err := tdb.Update(Path([]byte("abc"), []byte("de")), []byte("val2"), true)
	if update_err != nil {
		t.Error(update_err)
	}

	update_n, update_err = tdb.Update(Path([]byte("a"), []byte("de")), []byte("val2"), true)
	if update_err != nil {
		t.Error(update_err)
	}

	fmt.Println(update_n)
}
