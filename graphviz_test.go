package triedb

import (
	"fmt"
	"os"
	"testing"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv_leveldb"
)

const _DATABASE_PATH = "./kv_leveldb_test.db"

func TestBasic(t *testing.T) {
	os.RemoveAll(_DATABASE_PATH)
	kvdb, _ := kv_leveldb.NewDB(_DATABASE_PATH)
	//
	c, _ := cache.New(nil)
	tdb, _ := NewTrieDB(kvdb, c, nil)

	tdb.Update([]byte("12"), []byte("val12"))
	tdb.Update([]byte("1a"), []byte("val1a"))
	tdb.Update([]byte("1b"), []byte("val1b"))
	tdb.Update([]byte("123"), []byte("val123"))
	tdb.Update([]byte("12a"), []byte("val12a"))
	tdb.Update([]byte("12b"), []byte("val12b"))
	tdb.Update([]byte("1234"), []byte("val1234"))
	tdb.Update([]byte("123a"), []byte("val123a"))

	fmt.Println(tdb.GenDotString())
	tdb.GenDotFile("./test_basic.dot")
}
