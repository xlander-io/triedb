package main

import (
	"os"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv_leveldb"
	"github.com/xlander-io/triedb"
)

func main() {

	const db_path = "./kv_leveldb_test.db"
	os.RemoveAll(db_path)

	{
		kvdb, _ := kv_leveldb.NewDB(db_path)
		//
		c, _ := cache.New(nil)
		tdb, _ := triedb.NewTrieDB(kvdb, c, &triedb.TrieDBConfig{
			Root_hash:           nil,
			Commit_thread_limit: 1,
		})

		tdb.Update([]byte("/"), []byte("/"))
		tdb.Update([]byte("/123"), []byte("/123"))
		tdb.Update([]byte("/abcc/"), []byte("/abcc/"))
		tdb.Update([]byte("/abc/"), []byte("/abc/"))
		tdb.Update([]byte("/abc/def"), []byte("/abc/def"))
		tdb.Update([]byte("/abc/fdef"), []byte("/abc/fdef"))
		tdb.Update([]byte("/fff"), []byte("/fff"))
		tdb.Update([]byte("/ggg"), []byte("/ggg"))

		tdb.CalHash()

		tdb.GenDotFile("./1.dot", false)

	}

}
