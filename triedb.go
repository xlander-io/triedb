package triedb

import (
	"errors"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv"
	"github.com/xlander-io/triedb/util"
)

//const PATH_LEN_LIMIT = 64 * 1024         // in bytes uint16 limit
//const VAL_LEN_LIMIT = 4096 * 1024 * 1024 // in bytes uint32 limit

// ///////////////////////////
type trie_cache_item struct {
	val []byte
}

func (item *trie_cache_item) CacheBytes() int {
	return len(item.val)
}

/////////////////////////////

type TrieDB struct {
	root_node              *Node
	kvdb                   kv.KVDB
	cache                  *cache.Cache
	cache_default_ttl_secs int64
}

func NewTrieDB(kvdb_ kv.KVDB, cache_ *cache.Cache, cache_default_ttl_secs_ int64, root_hash *util.Hash) (*TrieDB, error) {
	if kvdb_ == nil {
		return nil, errors.New("NewTrieDB kvdb is nil")
	}

	if cache_ == nil {
		return nil, errors.New("NewTrieDB cache is nil")
	}

	if root_hash == nil {

		return &TrieDB{
			cache:                  cache_,
			cache_default_ttl_secs: cache_default_ttl_secs_,
			kvdb:                   kvdb_,
			root_node:              &Node{},
		}, nil

	} else {
		trie_db := TrieDB{
			cache:                  cache_,
			cache_default_ttl_secs: cache_default_ttl_secs_,
			kvdb:                   kvdb_,
			root_node:              &Node{node_hash: root_hash},
		}

		root_node_bytes, root_node_err := trie_db.getFromCacheKVDB(root_hash[:])
		if root_node_err != nil {
			return nil, errors.New("getBytesFromKVDB err in NewTrieDB, err: " + root_node_err.Error())
		}

		trie_db.root_node.node_bytes = root_node_bytes
		trie_db.root_node.deserialize()

		return &trie_db, nil
	}
}

// get first from cache then from kvdb
func (trie_db *TrieDB) getFromCacheKVDB(key []byte) ([]byte, error) {

	key_str := string(key)
	//from cache
	c_item, _ := trie_db.cache.Get(key_str)
	if c_item != nil {
		return c_item.(*trie_cache_item).val, nil
	}

	//from kvdb
	node_val, get_err := trie_db.kvdb.Get(key)
	if get_err != nil {
		return nil, get_err
	} else {
		//set to cache
		trie_db.cache.Set(key_str, &trie_cache_item{
			val: node_val,
		}, trie_db.cache_default_ttl_secs)

		return node_val, nil
	}
}

// 1.get from internal nodes which is the lastest val(dirty or not dirty val)
// 2.get from cache
// 3.get from kvdb
func (trie_db *TrieDB) Get(key []byte) ([]byte, error) {
	return nil, nil
}

//to think , any lock over Get ? any lock over update?

// return error may be caused by kvdb io as get reading may happen inside update
func (trie_db *TrieDB) Update(key []byte, val []byte) error {
	return nil
}
