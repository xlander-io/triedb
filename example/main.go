package main

import (
	"fmt"
	"os"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/hash"
	"github.com/xlander-io/kv"
	"github.com/xlander-io/kv_leveldb"
	"github.com/xlander-io/triedb"
)

type update_item struct {
	path           [][]byte
	val            []byte
	gen_hash_index bool
}

func main() {
	//////////////
	update_items := []update_item{}

	update_items = append(update_items, update_item{
		path:           [][]byte{[]byte("1"), []byte("23")},
		val:            []byte("val123"),
		gen_hash_index: true,
	})

	update_items = append(update_items, update_item{
		path:           [][]byte{[]byte("12")},
		val:            []byte("val12"),
		gen_hash_index: true,
	})

	update_items = append(update_items, update_item{
		path:           [][]byte{[]byte("1a")},
		val:            []byte("val1a"),
		gen_hash_index: true,
	})

	update_items = append(update_items, update_item{
		path:           [][]byte{[]byte("1b")},
		val:            []byte("val1b"),
		gen_hash_index: true,
	})

	update_items = append(update_items, update_item{
		path:           [][]byte{[]byte("1ab")},
		val:            []byte("val1ab"),
		gen_hash_index: true,
	})

	update_items = append(update_items, update_item{
		path:           [][]byte{[]byte("123")},
		val:            []byte("val123"),
		gen_hash_index: true,
	})

	update_items = append(update_items, update_item{
		path:           [][]byte{[]byte("12a")},
		val:            []byte("val12a"),
		gen_hash_index: true,
	})

	update_items = append(update_items, update_item{
		path:           [][]byte{[]byte("12b")},
		val:            []byte("val12b"),
		gen_hash_index: true,
	})

	update_items = append(update_items, update_item{
		path:           [][]byte{[]byte("1234")},
		val:            []byte("val1234"),
		gen_hash_index: true,
	})

	update_items = append(update_items, update_item{
		path:           [][]byte{[]byte("123a")},
		val:            []byte("val123a"),
		gen_hash_index: true,
	})

	update_items = append(update_items, update_item{
		path:           [][]byte{[]byte("a"), []byte("a"), []byte("a")},
		val:            []byte("valaaa"),
		gen_hash_index: true,
	})

	update_items = append(update_items, update_item{
		path:           [][]byte{[]byte("ab"), []byte("cd")},
		val:            []byte("valabcd"),
		gen_hash_index: true,
	})

	////////////

	os.Remove("./test.db")

	kvdb, err := kv_leveldb.NewDB("./test.db")
	if err != nil {
		panic(err)
	}

	c, err := cache.New(nil)
	if err != nil {
		panic(err)
	}
	tdb, err := triedb.NewTrieDB(kvdb, c, &triedb.TrieDBConfig{
		Root_hash:           nil,
		Commit_thread_limit: 10,
	})

	if err != nil {
		panic(err)
	}

	for _, item := range update_items {
		tdb.Put(item.path, item.val, item.gen_hash_index)
	}

	root_hash, updated, deleted, _ := tdb.Commit()

	fmt.Println("commit root hash:", root_hash.Hex())
	fmt.Println("commit updated len:", len(updated))
	fmt.Println("commit deleted len:", len(deleted))

	b := kv.NewBatch()

	for key, val := range updated {
		fmt.Println("to update:", hash.NewHashFromBytes([]byte(key)).Hex())
		b.Put([]byte(key), val)
	}

	for key, val := range deleted {
		fmt.Println("to del:", key, val.Hex())
		b.Delete([]byte(key))
	}

	err = kvdb.WriteBatch(b, true)
	if err != nil {
		fmt.Println("del batch err", err)
	}

	kvdb.Close()

	///////////////////////////////////////////////////////////

	kvdb2, err := kv_leveldb.NewDB("./test.db")
	if err != nil {
		panic("new kvdb err:" + err.Error())
	}

	c2, err := cache.New(nil)
	if err != nil {
		panic(err)
	}

	tdb2, err := triedb.NewTrieDB(kvdb2, c2, &triedb.TrieDBConfig{
		Root_hash:           root_hash,
		Commit_thread_limit: 10,
	})

	if err != nil {
		panic(err)
	}

	// iter, iter_err := tdb2.NewIterator(triedb.Path())
	// if iter_err != nil {
	// 	panic(iter_err)
	// }

	// for {

	// 	val, val_err := iter.Val()
	// 	if val_err != nil {
	// 		panic(val_err)
	// 	} else {
	// 		fmt.Println("val:", string(val), "flat path:", iter.FullPathFlatStr(), " is folder:", iter.IsFolder())
	// 	}

	// 	has_next, next_err := iter.Next()
	// 	if next_err != nil {
	// 		panic(next_err)
	// 	}
	// 	if !has_next {
	// 		break
	// 	}
	// }
	// //////////
	// for {

	// 	val, val_err := iter.Val()
	// 	if val_err != nil {
	// 		panic(val_err)
	// 	} else {
	// 		fmt.Println("val:", string(val), "flat path:", iter.FullPathFlatStr(), " is folder:", iter.IsFolder())
	// 	}

	// 	has_prev, prev_err := iter.Previous()
	// 	if prev_err != nil {
	// 		panic(prev_err)
	// 	}
	// 	if !has_prev {
	// 		break
	// 	}
	// }

	///////////////////

	fmt.Println("///////////////////////")

	del_items := update_items[1:]
	for _, item := range del_items {
		tdb2.Del(item.path)
	}

	fmt.Println(tdb2.Put(triedb.Path([]byte("abc")), []byte("valabc"), true))

	root_hash2, updated2, deleted2, _ := tdb2.Commit()

	tdb2.GenDotString(true)
	tdb2.GenDotFile("./test_graphviz.dot", false)

	fmt.Println("commit root hash:", root_hash2.Hex())
	fmt.Println("commit updated len:", len(updated2))
	fmt.Println("commit deleted len:", len(deleted2))

	fmt.Println("update2 len:", len(updated2))

	b2 := kv.NewBatch()

	for key, val := range updated2 {
		fmt.Println("to update:", hash.NewHashFromBytes([]byte(key)).Hex())
		b2.Put([]byte(key), val)
	}

	for key, val := range deleted2 {
		fmt.Println("to del:", val.Hex())
		b2.Delete([]byte(key))
	}

	err = kvdb2.WriteBatch(b2, true)
	if err != nil {
		fmt.Println("del2 batch err", err)
	}

	kvdb2.Close()
	/////

	kvdb3, err := kv_leveldb.NewDB("./test.db")
	if err != nil {
		panic("new kvdb err:" + err.Error())
	}

	c3, err := cache.New(nil)
	if err != nil {
		panic(err)
	}

	tdb3, err := triedb.NewTrieDB(kvdb3, c3, &triedb.TrieDBConfig{
		Root_hash:           root_hash2,
		Commit_thread_limit: 10,
	})

	if err != nil {
		panic(err)
	}

	result, _, _ := tdb3.Get(triedb.Path([]byte("abc")))
	fmt.Println(string(result))

	result2, _, _ := tdb3.Get(triedb.Path([]byte("1"), []byte("23")))
	fmt.Println(string(result2))

}
