package triedb

import (
	"bytes"
	"errors"
	"sync"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv"
	"github.com/xlander-io/triedb/util"
)

const NODE_PREFIX = "node_prefix"
const NODE_VAL_PREFIX = "node_val_prefix"
const NODES_PREFIX = "nodes_prefix"

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
	root_hash *util.Hash
	root_node *Node

	node_prefix     []byte
	node_val_prefix []byte
	nodes_prefix    []byte

	kvdb                   kv.KVDB
	cache                  *cache.Cache
	cache_default_ttl_secs int64
	lock                   sync.Mutex

	attached_hash map[string]struct{} //hash => struct{}{} , all related hash in the trie
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
			root_hash:              nil,
			root_node:              &Node{node_hash: nil},
			node_prefix:            []byte(NODE_PREFIX),
			node_val_prefix:        []byte(NODE_VAL_PREFIX),
			nodes_prefix:           []byte(NODES_PREFIX),
		}, nil

	} else {

		root_hash_copy := util.NewHashFromBytes((*root_hash)[:])
		var root_hash_bytes []byte = (*root_hash)[:]

		trie_db := TrieDB{
			cache:                  cache_,
			cache_default_ttl_secs: cache_default_ttl_secs_,
			kvdb:                   kvdb_,
			root_hash:              root_hash_copy,
			root_node:              &Node{node_hash: root_hash_copy},
			node_prefix:            append([]byte(NODE_PREFIX), root_hash_bytes...),
			node_val_prefix:        append([]byte(NODE_VAL_PREFIX), root_hash_bytes...),
			nodes_prefix:           append([]byte(NODES_PREFIX), root_hash_bytes...),
		}

		root_node_bytes, root_node_err := trie_db.getFromCacheKVDB(root_hash_copy[:])
		if root_node_err != nil {
			return nil, errors.New("getBytesFromKVDB err in NewTrieDB, err: " + root_node_err.Error())
		}

		trie_db.root_node.node_bytes = root_node_bytes
		trie_db.root_node.deserialize()

		trie_db.attachHash(root_hash_copy)
		return &trie_db, nil
	}
}

func (trie_db *TrieDB) attachHash(hash *util.Hash) {
	trie_db.attached_hash[string((*hash)[:])] = struct{}{}
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
	trie_db.lock.Lock()
	defer trie_db.lock.Unlock()

	return nil, nil
}

func (trie_db *TrieDB) recover_child_nodes(node *Node) error {
	if node.child_nodes == nil && node.child_nodes_hash != nil {

		nodes_bytes, err := trie_db.getFromCacheKVDB(node.child_nodes_hash[:])
		if err != nil {
			return errors.New("recover_child_nodes err :" + err.Error())
		}

		child_nodes_ := Nodes{
			nodes_bytes: nodes_bytes,
			nodes_hash:  node.child_nodes_hash,
		}

		child_nodes_.deserialize()
		//
		for _, c_n := range child_nodes_.path_index {
			trie_db.attachHash(c_n.node_hash)
		}
		trie_db.attachHash(node.child_nodes_hash)
		//
		return nil

	} else {
		//
		return nil
	}
}

func (trie_db *TrieDB) recursive_mark_dirty(node *Node) {
	if node == nil {
		return
	}
	//
	node.node_bytes = nil
	node.node_hash = nil
	if node.parent_nodes != nil {
		node.parent_nodes.nodes_bytes = nil
		node.parent_nodes.nodes_hash = nil
		trie_db.recursive_mark_dirty(node.parent_nodes.parent_node)
	}
}

// full_path len !=0 is required
// val == nil stands for del
// return error may be caused by kvdb io as get reading may happen inside update
func (trie_db *TrieDB) Update(full_path []byte, val []byte) error {

	if len(full_path) == 0 {
		return errors.New("full_path len err")
	}

	trie_db.lock.Lock()
	defer trie_db.lock.Unlock()

	//
	recover_err := trie_db.recover_child_nodes(trie_db.root_node)
	if recover_err != nil {
		return recover_err
	}

	//nothing todo
	if trie_db.root_node.child_nodes == nil && val == nil {
		return nil
	}

	// new child_nodes if not exist
	if trie_db.root_node.child_nodes == nil {
		trie_db.root_node.child_nodes = &Nodes{
			path_index:  make(map[byte]*Node),
			parent_node: trie_db.root_node,
		}
	}

	return trie_db.target_nodes(trie_db.root_node.child_nodes, full_path, val)

}

// len(left_path) is >0
func (trie_db *TrieDB) target_nodes(target_nodes *Nodes, left_path []byte, val []byte) error {

	/////// target the next node
	if target_nodes.path_index[left_path[0]] != nil {
		return trie_db.target_node(target_nodes.path_index[left_path[0]], left_path, val)
	}

	////// no common first byte
	if val == nil {
		//nothing todo
		return nil
	} else {
		//simply add a new node
		target_nodes.path_index[left_path[0]] = &Node{
			path:         left_path,
			parent_nodes: target_nodes,
			val:          val,
		}
		//mark dirty
		target_nodes.nodes_bytes = nil
		target_nodes.nodes_hash = nil
		trie_db.recursive_mark_dirty(target_nodes.parent_node)
	}

	return nil
}

// left_path has at least one byte same compared with target_node
func (trie_db *TrieDB) target_node(target_node *Node, left_path []byte, val []byte) error {

	//target exactly
	if bytes.Equal(target_node.path, left_path) {
		if val == nil {
			//del this node
			target_node.val = nil
			target_node.val_hash = nil
			//should remove this node when any child_node exist
			if target_node.child_nodes == nil {
				delete(target_node.parent_nodes.path_index, target_node.path[0])
			}
		} else {
			//update this node
			target_node.val = val
			target_node.val_hash = nil
		}
		trie_db.recursive_mark_dirty(target_node)
		return nil

	} else if (len(left_path) > len(target_node.path)) && bytes.Equal(target_node.path, left_path[0:len(target_node.path)]) {
		// left_path start with target_node.path
		if target_node.child_nodes != nil {
			///
			return trie_db.target_nodes(target_node.child_nodes, left_path[len(target_node.path):], val)
		} else {
			///
			if target_node.child_nodes_hash != nil {
				recover_err := trie_db.recover_child_nodes(target_node)
				if recover_err != nil {
					return errors.New("target_node recover_child_nodes err:" + recover_err.Error())
				}
			}
			// nothing to do with nil (del action)
			if target_node.child_nodes == nil && val == nil {
				return nil
			}

			// new nodes that dynamically created
			if target_node.child_nodes == nil {
				target_node.child_nodes = &Nodes{
					path_index:  make(map[byte]*Node),
					parent_node: target_node,
				}
			}

			return trie_db.target_nodes(target_node.child_nodes, left_path[len(target_node.path):], val)

		}

	} else if (len(left_path) < len(target_node.path)) && bytes.Equal(left_path, target_node.path[0:len(left_path)]) {
		// target_node.path start with left_path

		// nothing to do with nil (del action)
		if val == nil {
			return nil
		}

		new_node := Node{
			path:         left_path[:],
			parent_nodes: (*target_node).parent_nodes,
			child_nodes:  &Nodes{},
			val:          val,
		}

		new_node.child_nodes = &Nodes{
			path_index:  make(map[byte]*Node),
			parent_node: &new_node,
		}

		(*target_node).path = target_node.path[len(left_path):]
		(*target_node).parent_nodes.path_index[new_node.path[0]] = &new_node
		(*target_node).parent_nodes = new_node.child_nodes
		trie_db.recursive_mark_dirty(target_node)

		return nil

	} else {

		// nothing to do with nil (del action)
		if val == nil {
			return nil
		}

		////////// find the common bytes prefix
		min_len := len(left_path)
		node_path_len := len(target_node.path)
		if node_path_len < min_len {
			min_len = node_path_len
		}

		common_prefix_bytes := []byte{}
		for i := 0; i < min_len; i++ {
			if left_path[i] == target_node.path[i] {
				common_prefix_bytes = append(common_prefix_bytes, left_path[i])
			} else {
				break
			}
		}
		common_prefix_bytes_len := len(common_prefix_bytes)
		///////////

		new_node := Node{
			path:         common_prefix_bytes[:],
			parent_nodes: (*target_node).parent_nodes,
			child_nodes:  &Nodes{},
		}

		new_node.child_nodes = &Nodes{
			path_index:  make(map[byte]*Node),
			parent_node: &new_node,
		}

		new_node.child_nodes.path_index[left_path[common_prefix_bytes_len]] = &Node{
			path:         left_path[common_prefix_bytes_len:],
			parent_nodes: new_node.child_nodes,
			val:          val,
		}

		new_node.child_nodes.path_index[target_node.path[common_prefix_bytes_len]] = target_node

		target_node.path = target_node.path[common_prefix_bytes_len:]
		target_node.parent_nodes.path_index[new_node.path[0]] = &new_node
		target_node.parent_nodes = new_node.child_nodes
		trie_db.recursive_mark_dirty(target_node)

		return nil
	}

}
