package triedb

import (
	"fmt"
	"os"
	"testing"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv_leveldb"
)

const _DATABASE_PATH = "./triedb_graphviz_test.db"

func TestGraphviz(t *testing.T) {
	os.RemoveAll(_DATABASE_PATH)
	kvdb, _ := kv_leveldb.NewDB(_DATABASE_PATH)
	//
	c, _ := cache.New(nil)
	tdb, _ := NewTrieDB(kvdb, c, nil)

	tdb.Update(Path([]byte("12")), []byte("val12"), true)
	tdb.Update(Path([]byte("1a")), []byte("val1a"), true)
	tdb.Update(Path([]byte("1b")), []byte("val1b"), true)
	tdb.Update(Path([]byte("1ab")), []byte("val1ab"), true)
	tdb.Update(Path([]byte("123")), []byte("val123"), true)
	tdb.Update(Path([]byte("12a")), []byte("val12a"), true)
	tdb.Update(Path([]byte("12b")), []byte("val12b"), true)
	tdb.Update(Path([]byte("1234")), []byte("val1234"), true)
	tdb.Update(Path([]byte("123a")), []byte("val123a"), true)

	tdb.Update(Path([]byte("a"), []byte("a"), []byte("a")), []byte("valaaa"), true)
	tdb.Update(Path([]byte("ab"), []byte("cd")), []byte("valabcd"), true)

	//tdb.CalHash()

	fmt.Println(tdb.GenDotString(true))
	// dot -Tpdf -O test_graphviz.dot && open test_graphviz.dot.pdf
	tdb.GenDotFile("./test_graphviz.dot", false)
}
