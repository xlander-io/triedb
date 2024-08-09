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

### 
