package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/xlander-io/cache"
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

	{
		kvdb, _ := kv_leveldb.NewDB(db_path)
		//
		c, _ := cache.New(nil)
		tdb, _ := triedb.NewTrieDB(kvdb, c, &triedb.TrieDBConfig{
			Root_hash:           nil,
			Commit_thread_limit: 1,
		})

		root_folder_n, _ := tdb.Update([]byte("/"), []byte("/"))
		tdb.Update([]byte("/123"), []byte("/123"))
		tdb.Update([]byte("/abc/"), []byte("/abc/"))
		tdb.Update([]byte("/abc/def"), []byte("/abc/def"))
		tdb.Update([]byte("/abc/fdef"), []byte("/abc/fdef"))
		tdb.Update([]byte("/fff"), []byte("/fff"))
		tdb.Update([]byte("/ggg"), []byte("/ggg"))

		iter, _ := triedb.NewChildIterator(root_folder_n, nil)

		next_n, _ := iter.SkipNext(true)
		fmt.Println(string(next_n.FullPath()))

		//
		next_n, _ = iter.SkipNext(true)
		fmt.Println(string(next_n.FullPath()))

		//
		next_n, _ = iter.Next(true)
		fmt.Println(string(next_n.FullPath()))

		//
		next_n, _ = iter.SkipNext(true)
		fmt.Println(string(next_n.FullPath()))

		//
		next_n, _ = iter.SkipNext(true)
		fmt.Println(string(next_n.FullPath()))

		// next_n, _ = iter.Next(true)
		// fmt.Println(string(next_n.FullPath()))

		// next_n, _ = iter.SkipNext(true)
		// fmt.Println(string(next_n.FullPath()))

		//next_n, _ := iter.SkipNext(false)

		//fmt.Println(string(next_n.FullPath()))

		// tdb.Update([]byte("13"), []byte("val_13"))
		// tdb.Update([]byte("14"), []byte("val_14"))
		// tdb.Update([]byte("123"), []byte("val_123"))
		// tdb.Update([]byte("1234"), []byte("val_1234"))

		// root_hash_, to_update, to_del, cal_herr := tdb.CalHash()
		// if cal_herr != nil {
		// 	fmt.Println("tdb.CalHash() err:", cal_herr.Error())
		// 	return
		// }

		// tdb.GenDotFile("./1.dot", false)

		// fmt.Println("root_hash:", fmt.Sprintf("%x", root_hash_.Bytes()))

		// b := kv.NewBatch()
		// for hex_string, update_v := range to_update {
		// 	b.Put([]byte(hex_string), update_v)
		// 	fmt.Println("hex_string ", fmt.Sprintf("%x", hex_string))
		// }

		// for _, del_v := range to_del {
		// 	b.Delete(del_v.Bytes())
		// }

		// write_err := kvdb.WriteBatch(b, true)

		// if write_err != nil {
		// 	fmt.Println("write_err:", write_err)
		// }

		// root_hash = root_hash_
		// kvdb.Close()
	}

}
