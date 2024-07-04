package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv"
	"github.com/xlander-io/kv_leveldb"
	"github.com/xlander-io/triedb"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func main() {

	const db_path = "./kv_leveldb_test.db"
	os.RemoveAll(db_path)
	kvdb, _ := kv_leveldb.NewDB(db_path)
	//
	c, _ := cache.New(nil)
	tdb, _ := triedb.NewTrieDB(kvdb, c, &triedb.TrieDBConfig{
		Commit_thread_limit: 1,
	})

	tdb.Update([]byte("1"), []byte([]byte("val_1")))
	tdb.Update([]byte("12"), []byte([]byte("val_12")))
	tdb.Update([]byte("13"), []byte([]byte("val_13")))
	tdb.Update([]byte("14"), []byte([]byte("val_14")))
	tdb.Update([]byte("123"), []byte([]byte("val_123")))
	tdb.Update([]byte("1234"), []byte([]byte("val_1234")))

	tdb.GenDotFile("./test_pre.dot", false)

	root_hash, to_update, to_del, cal_herr := tdb.CalHash()
	if cal_herr != nil {
		fmt.Println("tdb.CalHash() err:", cal_herr.Error())
		return
	}
	fmt.Println("root_hash:", fmt.Sprintf("%x", root_hash.Bytes()))

	fmt.Println("len(to_update):", len(to_update))

	tdb.GenDotFile("./test_after_hash.dot", false)

	b := kv.NewBatch()
	for hex_string, update_v := range to_update {
		b.Put([]byte(hex_string), update_v)
		fmt.Println("hex_string ", fmt.Sprintf("%x", hex_string))
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

	///////

	del_err := tdb2.Delete([]byte("1"))
	fmt.Println(del_err)

	//////

	tdb2.GenDotFile("./from kvdb.dot", false)

	////here we go
	fmt.Println("//////////////////////////////")

	root_hash2, to_update2, to_del2, cal_herr2 := tdb2.CalHash()
	if cal_herr2 != nil {
		fmt.Println("tdb2.CalHash() err:", cal_herr2.Error())
		return
	}

	fmt.Println("to_update2", len(to_update2))
	fmt.Println("to_del2", len(to_del2))

	fmt.Println("root_hash2", fmt.Sprintf("%x", root_hash2))

	tdb2.GenDotFile("./afterdel.dot", false)

	///
	// println("//////// to_update2 ///////////")
	// for hash_str, _ := range to_update2 {
	// 	fmt.Println(fmt.Sprintf("%x", hash_str))
	// }

	// println("//////// to_del2 ///////////")
	// for _, hash_ := range to_del2 {
	// 	fmt.Println(fmt.Sprintf("%x", hash_.Bytes()))
	// }

}

func getMe(tb *triedb.TrieDB, key string) {
	val, val_err := tb.Get([]byte(key))
	if val_err != nil {
		fmt.Println("get err key :", key, val)
	} else {
		fmt.Println(key, " val:", string(val))
	}
}
