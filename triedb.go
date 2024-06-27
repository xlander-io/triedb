package triedb

import (
	"errors"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv"
	"github.com/xlander-io/triedb/util"
)

//const PATH_LEN_LIMIT = 64 * 1024         // in bytes uint16 limit
//const VAL_LEN_LIMIT = 4096 * 1024 * 1024 // in bytes uint32 limit

type TrieDB struct {
	root_node *Node
	kvdb      kv.KVDB
	cache     *cache.Cache
}

func NewTrieDB(kvdb_ kv.KVDB, cache_ *cache.Cache, root_hash *util.Hash) (*TrieDB, error) {
	if kvdb_ == nil {
		return nil, errors.New("NewTrieDB kvdb is nil")
	}

	if cache_ == nil {
		return nil, errors.New("NewTrieDB cache is nil")
	}

	if root_hash == nil {

		return &TrieDB{
			cache:     cache_,
			kvdb:      kvdb_,
			root_node: &Node{},
		}, nil

	} else {
		trie_db := TrieDB{
			cache:     cache_,
			kvdb:      kvdb_,
			root_node: &Node{node_hash: root_hash},
		}

		root_node_bytes, root_node_err := trie_db.getBytesFromKVDB(root_hash)
		if root_node_err != nil {
			return nil, errors.New("getBytesFromKVDB err in NewTrieDB, err: " + root_node_err.Error())
		}

		trie_db.root_node.node_bytes = root_node_bytes
		trie_db.root_node.deserialize()

		return &trie_db, nil
	}
}

func (trie_db *TrieDB) getBytesFromKVDB(node_hash *util.Hash) ([]byte, error) {

	if node_hash == nil {
		return nil, nil
	}

	node_val, get_err := trie_db.kvdb.Get((*node_hash)[:])
	if get_err != nil {
		return nil, get_err
	} else {
		return node_val, nil
	}

}
