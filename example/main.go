package main

import (
	"fmt"
	"os"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv_leveldb"
	"github.com/xlander-io/triedb"
)

func main() {

	const db_path = "./kv_leveldb_test.db"
	os.RemoveAll(db_path)
	kvdb, _ := kv_leveldb.NewDB(db_path)
	//
	c, _ := cache.New(nil)
	tdb, _ := triedb.NewTrieDB(kvdb, c, &triedb.TrieDBConfig{
		Commit_thread_limit: 1,
	})

	tdb.Update([]byte("12"), []byte("val12"))
	tdb.Update([]byte("123"), []byte("val123"))
	tdb.Update([]byte("1234"), []byte("val1234"))

	_, to_update, _ := tdb.CalHash()

	//fmt.Println(root_hash)

	fmt.Println(len(to_update))
	for hex_string, update_v := range to_update {
		fmt.Println("hex:" + fmt.Sprintf("%x", hex_string))
		fmt.Println("val:" + string(update_v))
	}

	fmt.Println(tdb)
}
