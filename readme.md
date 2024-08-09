# TriDB

## structure

internally tridb is implemented using prefix trie and btree, the triedb can be used as

1. a folder/file structure
2. btree search structure
3. a prefix compression structure

## usecase

### set get and delete
```
// save the value `valaaa` to the  path `/a/a/a`
tdb.Put(triedb.Path([]byte("a"), []byte("a"), []byte("a")), []byte("valaaa"), true)

// save the value  `valaaa` to the  path `/aaa`
tdb.Put(triedb.Path([]byte("aaa")), []byte("valaaa"), true)

// delete the path  `/aaa`
tdb.Del(triedb.Path([]byte("aaa")))

```

### iterator

```

iter, iter_err := tdb2.NewIterator(triedb.Path())
if iter_err != nil {
    panic(iter_err)
}

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

```
