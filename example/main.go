package main

// import (
// 	"fmt"

// 	"github.com/xlander-io/cache"
// 	"github.com/xlander-io/hash"
// 	"github.com/xlander-io/kv_leveldb"
// 	"github.com/xlander-io/triedb"
// )

// func main() {

// 	kvdb, err := kv_leveldb.NewDB("./test.db")
// 	if err != nil {
// 		panic(err)
// 	}

// 	c, err := cache.New(nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	tdb, err := triedb.NewTrieDB(kvdb, c, &triedb.TrieDBConfig{
// 		Root_hash:           hash.NewHashFromString("0x55dc0c8d0a2a7b47419a0671216f047b99a91eb18507b5dfe7423234092b0c18"),
// 		Commit_thread_limit: 10,
// 	})

// 	if err != nil {
// 		panic(err)
// 	}

// 	n, err := tdb.Get(triedb.Path([]byte("a"), []byte("a"), []byte("a")))
// 	if err != nil {
// 		panic(err)
// 	}

// 	fmt.Println(string(n.Val()))

// }
