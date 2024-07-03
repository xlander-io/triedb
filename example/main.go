package main

import (
	"fmt"
	"os"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv"
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
	tdb.Update([]byte("1a"), []byte("val1a"))
	tdb.Update([]byte("1b"), []byte("val1b"))
	tdb.Update([]byte("123"), []byte("val123"))
	tdb.Update([]byte("12a"), []byte("val12a"))
	tdb.Update([]byte("12b"), []byte("val12b"))
	tdb.Update([]byte("1234"), []byte("val1234"))
	tdb.Update([]byte("123a"), []byte("val123a"))
	tdb.Update([]byte("1234"), nil)

	root_hash, to_update, to_del := tdb.CalHash()

	fmt.Println("hex:" + fmt.Sprintf("%x", root_hash))

	// fmt.Println("================to update len:", len(to_update), "====================")
	// for hex_string, update_v := range to_update {
	// 	fmt.Println("hex:" + fmt.Sprintf("%x", hex_string))
	// 	fmt.Println("val:" + string(update_v))
	// }

	// fmt.Println("================to del len:", len(to_del), "====================")
	// for hex_string, del_v := range to_del {
	// 	fmt.Println("hex:" + fmt.Sprintf("%x", hex_string))
	// 	fmt.Println("val:" + string(del_v.Bytes()))
	// }

	///excute the code

	b := kv.NewBatch()
	for hex_string, update_v := range to_update {
		b.Put([]byte(hex_string), update_v)
	}

	for _, del_v := range to_del {
		b.Delete(del_v.Bytes())
	}

	write_err := kvdb.WriteBatch(b, true)

	if write_err != nil {
		fmt.Println("write_err:", write_err)
	}

	///////
	c2, _ := cache.New(nil)
	tdb2, _ := triedb.NewTrieDB(kvdb, c2, &triedb.TrieDBConfig{
		Commit_thread_limit: 1,
		Root_hash:           root_hash,
	})

	key_12_val, key_12_val_err := tdb2.Get([]byte("12"))
	if key_12_val_err != nil {
		fmt.Println("get err key 12:", key_12_val_err)
	} else {
		fmt.Println("key_12_val val:", string(key_12_val))
	}

	key_1234_val, key_1234_val_err := tdb2.Get([]byte("1234"))
	if key_1234_val_err != nil {
		fmt.Println("get err key 1234:", key_1234_val_err)
	} else {
		fmt.Println("key_1234_val val:", string(key_1234_val))
	}

}
