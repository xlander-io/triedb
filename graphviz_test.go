package triedb

import (
	"fmt"
	"os"
	"testing"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv_leveldb"
)

const _DATABASE_PATH = "./kv_leveldb_test.db"

func TestGraphviz(t *testing.T) {
	os.RemoveAll(_DATABASE_PATH)
	kvdb, _ := kv_leveldb.NewDB(_DATABASE_PATH)
	//
	c, _ := cache.New(nil)
	tdb, _ := NewTrieDB(kvdb, c, nil)

	tdb.Update(Path([]byte("12")), []byte("val12"), false)
	tdb.Update(Path([]byte("1a")), []byte("val1a"), false)
	tdb.Update(Path([]byte("1b")), []byte("val1b"), false)
	tdb.Update(Path([]byte("1ab")), []byte("val1ab"), false)
	tdb.Update(Path([]byte("123")), []byte("val123"), false)
	tdb.Update(Path([]byte("12a")), []byte("val12a"), false)
	tdb.Update(Path([]byte("12b")), []byte("val12b"), false)
	tdb.Update(Path([]byte("1234")), []byte("val1234"), false)
	tdb.Update(Path([]byte("123a")), []byte("val123a"), false)

	//tdb.CalHash()

	fmt.Println(tdb.GenDotString(true))
	// dot -Tpdf -O test_graphviz.dot && open test_graphviz.dot.pdf
	tdb.GenDotFile("./test_graphviz.dot", false)
}
