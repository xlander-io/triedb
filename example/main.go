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
		Root_hash:           hash.NewHashFromString("0x15a85c01dd6a14b645daa953e0452ac2089071ba5d87a5b299d91cced8dc7be9"),
		Commit_thread_limit: 10,
		Read_only:           true,
	})

	if err != nil {
		panic(err)
	}

	iter, iter_err := tdb.NewIterator(triedb.Path())
	if iter_err != nil {
		panic(iter_err)
	}

	for {

		val, val_err := iter.Val()
		if val_err != nil {
			panic(val_err)
		} else {
			fmt.Println("val:", string(val))
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
