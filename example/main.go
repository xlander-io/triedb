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

	tdb.Update([]byte("1a"), []byte("val1a"))

	root_hash, to_update, to_del := tdb.CalHash()

	tdb.GenDotFile("./check.dot")

	// fmt.Println("hex:" + fmt.Sprintf("%x", root_hash))

	fmt.Println("================to update len:", len(to_update), "====================")
	for hex_string, update_v := range to_update {
		fmt.Println("hex:" + fmt.Sprintf("%x", hex_string))
		fmt.Println("val:" + string(update_v))
	}

	fmt.Println("================to del len:", len(to_del), "====================")
	for hex_string, del_v := range to_del {
		fmt.Println("hex:" + fmt.Sprintf("%x", hex_string))
		fmt.Println("val:" + string(del_v.Bytes()))
	}

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

	//////
	tdb2.Update([]byte("1a"), nil)

	root_hash2, to_update2, to_del2 := tdb2.CalHash()

	fmt.Println("root_hash2 hex:" + fmt.Sprintf("%x", root_hash2))

	fmt.Println("================to update2 len:", len(to_update2), "====================")
	for hex_string, update_v := range to_update2 {
		fmt.Println("hex:" + fmt.Sprintf("%x", hex_string))
		fmt.Println("val:" + string(update_v))
	}

	fmt.Println("================to del2 len:", len(to_del2), "====================")
	for hex_string, _ := range to_del2 {
		fmt.Println("hex:" + fmt.Sprintf("%x", hex_string))
	}

	b2 := kv.NewBatch()
	for hex_string, update_v := range to_update2 {
		b2.Put([]byte(hex_string), update_v)
	}

	for _, del_v := range to_del2 {
		fmt.Println("to_del2 to del in batch :", fmt.Sprintf("%x", del_v.Bytes()))
		b2.Delete(del_v.Bytes())
	}

	write_err2 := kvdb.WriteBatch(b2, true)

	if write_err2 != nil {
		fmt.Println("write_err2:", write_err2)
	}

	//////////////////

	c3, _ := cache.New(nil)
	tdb3, trie_err := triedb.NewTrieDB(kvdb, c3, &triedb.TrieDBConfig{
		Commit_thread_limit: 1,
		Root_hash:           root_hash2,
	})

	if trie_err != nil {
		fmt.Println("trie_err ", trie_err)
		return
	}

	key_1a_val, key_1a_val_err := tdb3.Get([]byte("1a"))
	if key_1a_val_err != nil {
		fmt.Println("get err key 1a:", key_1a_val_err)
	} else {
		fmt.Println("key_1a_val val:", string(key_1a_val))
	}

}
