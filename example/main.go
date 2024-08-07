package main

import (
	"fmt"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/hash"
	"github.com/xlander-io/kv_leveldb"
	"github.com/xlander-io/triedb"
)

func main() {

	kvdb, err := kv_leveldb.NewDB("/Users/leo/Desktop/repo/triedb/test.db")
	if err != nil {
		panic(err)
	}

	c, err := cache.New(nil)
	if err != nil {
		panic(err)
	}
	tdb, err := triedb.NewTrieDB(kvdb, c, &triedb.TrieDBConfig{
		Root_hash:           hash.NewHashFromString("0xba9b8eb5478fed6a3fa9c5d637b851290c4b35d446da617e40f95b887c051f0c"),
		Commit_thread_limit: 10,
		Read_only:           true,
	})

	if err != nil {
		panic(err)
	}

	iter, iter_err := tdb.NewIterator(triedb.Path([]byte("a")))
	if iter_err != nil {
		panic(iter_err)
	}

	fmt.Println(iter.SetCursorWithFullPath([][]byte{[]byte("a"), []byte("a")}))

	for {

		val, val_err := iter.Val()
		if val_err != nil {
			panic(val_err)
		} else {
			fmt.Println("val:", string(val), "flat path:", iter.FullPathFlatStr(), " is folder:", iter.IsFolder())
		}

		has_next, next_err := iter.Next()
		if next_err != nil {
			panic(next_err)
		}
		if !has_next {
			break
		}
	}

}
