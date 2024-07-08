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

	//var rootHash *hash.Hash

	{
		kvdb, _ := kv_leveldb.NewDB(db_path)
		//
		c, _ := cache.New(nil)
		tdb, _ := triedb.NewTrieDB(kvdb, c, &triedb.TrieDBConfig{
			Root_hash:           nil,
			Commit_thread_limit: 1,
		})

		tdb.Update([]byte("/"), []byte("/"))
		test_n, _ := tdb.Update([]byte("/abc/"), []byte("/abc/"))
		tdb.Update([]byte("/abc/def"), []byte("/abc/def"))
		tdb.Update([]byte("/abc/defgde"), []byte("/abc/defgde"))
		tdb.Update([]byte("/abc/123"), []byte("/abc/123"))
		tdb.Update([]byte("/abcddddd"), []byte("/abcddddd"))
		pn, _ := test_n.ParentNode()
		fmt.Println("parent node path:", string(pn.FullPath()))

		iter, err := triedb.NewIterator(pn, test_n)
		if err != nil {
			panic("new iter err" + err.Error())
		}

		for {
			n, err := iter.GetNode()
			if err != nil {
				panic(err)
			}
			fmt.Println("path:", string(n.FullPath()))
			fmt.Println("val:", string(n.Val()))

			next_ok, _ := iter.SkipNext()
			if !next_ok {
				break
			}
		}

		//	tdb.Update([]byte("14"), []byte("val_14"))

		//	tdb.Update([]byte("1234"), []byte("val_1234"))

		tdb.GenDotFile("./test_pre.dot", false)

		root_hash, to_update, to_del, cal_herr := tdb.CalHash()
		if cal_herr != nil {
			fmt.Println("tdb.CalHash() err:", cal_herr.Error())
			return
		}

		tdb.GenDotFile("./test_after_hash.dot", false)

		fmt.Println("root_hash:", fmt.Sprintf("%x", root_hash.Bytes()))

		// fmt.Println("len(to_update):", len(to_update))

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

		//rootHash = root_hash
		kvdb.Close()
	}

	// {
	// 	kvdb, _ := kv_leveldb.NewDB(db_path)
	// 	//
	// 	c, _ := cache.New(nil)

	// 	tdb2, _ := triedb.NewTrieDB(kvdb, c, &triedb.TrieDBConfig{
	// 		Root_hash:           rootHash,
	// 		Commit_thread_limit: 1,
	// 	})

	// 	tdb2.Update([]byte("2"), []byte("val_2"))
	// 	tdb2.Update([]byte("1a"), []byte([]byte("val_1a")))
	// 	tdb2.Update([]byte("12a"), []byte([]byte("val_12a")))
	// 	updated_node, _ := tdb2.Update([]byte("123a"), []byte([]byte("val_123a")))

	// 	tdb2.CalHash()

	// 	fmt.Println("updated  fullpath:", string(updated_node.FullPath()))
	// 	fmt.Println("updated  path:", string(updated_node.Path()))
	// 	fmt.Println("updated  hash:", fmt.Sprintf("%x", string(updated_node.Hash().Bytes())))
	// 	fmt.Println("updated  val:", string(updated_node.Val()))

	// 	tdb2.GenDotFile("./read_from_kvdb.dot", false)

	// 	tnode, _ := tdb2.Get([]byte("123a"))

	// 	fmt.Println("tnode, fullpath", string(tnode.FullPath()))
	// 	fmt.Println("tnode, path", string(tnode.Path()))
	// 	fmt.Println("tnode, val", string(tnode.Val()))

	// 	kvdb.Close()

	// }

}
