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

	// tdb.Put(Path([]byte("AB")), []byte("valAB"), true)
	// tdb.Put(Path([]byte("Aa")), []byte("valAa"), true)
	// tdb.Put(Path([]byte("Ab")), []byte("valAb"), true)
	// tdb.Put(Path([]byte("Aab")), []byte("valAab"), true)
	// tdb.Put(Path([]byte("ABC")), []byte("valABC"), true)
	// tdb.Put(Path([]byte("ABa")), []byte("valABa"), true)
	// tdb.Put(Path([]byte("ABb")), []byte("valABb"), true)
	// tdb.Put(Path([]byte("ABCD")), []byte("valABCD"), true)
	// tdb.Put(Path([]byte("ABCa")), []byte("valABCa"), true)
	tdb.Put(Path([]byte("a"), []byte("a"), []byte("a")), []byte("valaaa"), true)
	tdb.Put(Path([]byte("ab"), []byte("cd")), []byte("valabcd"), true)

	// tdb.Put(Path([]byte("a"), []byte("a"), []byte("a")), []byte("valaaa"), false)

	// fmt.Println(tdb.Del(Path([]byte("a"), []byte("a"), []byte("a"))))
	// fmt.Println(tdb.Del(Path([]byte("a"), []byte("a"), []byte("a"))))

	root_hash, _, _, _ := tdb.Commit()
	fmt.Println(root_hash.Hex())

	tdb.GenDotString(true)
	// dot -Tpdf -O test_graphviz.dot && open test_graphviz.dot.pdf
	tdb.GenDotFile("./test_graphviz.dot", false)
}
