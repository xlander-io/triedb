package main

import (
	"fmt"
	"os"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv_leveldb"
	"github.com/xlander-io/triedb"
)

func main() {
	//
	const db_path = "./kv_leveldb_test.db"
	os.RemoveAll(db_path)
	kvdb, _ := kv_leveldb.NewDB(db_path)
	//
	c, _ := cache.New(nil)
	tdb, _ := triedb.NewTrieDB(kvdb, c, nil)

	tdb.Update([]byte("12"), []byte("val12"))
	tdb.Update([]byte("45"), []byte("val45"))
	tdb.Update([]byte("123"), []byte("val123"))
	tdb.Update([]byte("4"), nil)
	tdb.Update([]byte("45"), nil)
	tdb.Update([]byte("12"), nil)
	fmt.Println(tdb)
}
