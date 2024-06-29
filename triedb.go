package triedb

import (
	"bytes"
	"errors"
	"sync"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/kv"
	"github.com/xlander-io/triedb/util"
)

/*
 *		Trie implementation
 *		1. For the root node, its val_hash is always nil
 *		2. for a nodes, its parent_node is always not nil
 *		3. cache is always sync with disk kv db
 *		4. attached_hash stores all the kv k hash related in the trie, attached_hash will be checked
 *		   what key hash will be removed in commit
 *		5. Get,Put,Commit use the same lock to prevent data inconsistence
 *
 *
 *
 */

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

	kvdb  kv.KVDB
	cache *cache.Cache
	lock  sync.Mutex

	attached_hash map[string]struct{} //hash => struct{}{} , all related hash in the trie
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
			root_hash: nil,
			root_node: &Node{node_hash: nil},
		}, nil

	} else {

		root_hash_copy := util.NewHashFromBytes((*root_hash)[:])
		trie_db := TrieDB{
			cache:     cache_,
			kvdb:      kvdb_,
			root_hash: root_hash_copy,
			root_node: &Node{node_hash: root_hash_copy},
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
		})

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

// recursive_del will be called when del val happen
func (trie_db *TrieDB) recursive_del(node *Node) {

	//won't happen
	if node == nil || node.parent_nodes == nil {
		return
	}

	//del val of this node
	node.val = nil
	node.val_hash = nil

	//
	if node.child_nodes == nil {
		delete(node.parent_nodes.path_index, node.path[0])
		//what is left in parent_nodes
		if len(node.parent_nodes.path_index) == 0 {
			//
			node.parent_nodes.parent_node.child_nodes = nil
			trie_db.recursive_del(node.parent_nodes.parent_node)
			//
		} else if len(node.parent_nodes.path_index) == 1 {
			//  do simplification if possible

			var left_single_node *Node = nil
			for _, c_node_ := range node.parent_nodes.path_index {
				left_single_node = c_node_
			}

			// first !=nil checks the root node
			if left_single_node.parent_nodes.parent_node.parent_nodes != nil && left_single_node.parent_nodes.parent_node.val == nil {
				delete(left_single_node.parent_nodes.path_index, left_single_node.path[0])
				left_single_node.parent_nodes.parent_node.path = append(left_single_node.parent_nodes.parent_node.path, left_single_node.path...)
				left_single_node.parent_nodes.parent_node.child_nodes = left_single_node.child_nodes
				left_single_node.parent_nodes.parent_node.child_nodes_hash = left_single_node.child_nodes_hash
				left_single_node.parent_nodes.parent_node.val = left_single_node.val
				left_single_node.parent_nodes.parent_node.val_hash = nil
			}

		} else {
			//more then 1 node in parent nodes nothing to do
		}

	} else {
		//condition child_nodes !=nil && node.val==nil

		//check if child_nodes has only one node => replace it to this node
		if len(node.child_nodes.path_index) == 1 {
			//
			var single_child_node *Node = nil
			for _, c_node_ := range node.child_nodes.path_index {
				single_child_node = c_node_
			}
			//replace
			node.parent_nodes.path_index[node.path[0]] = single_child_node
			single_child_node.path = append(node.path, single_child_node.path...)
			single_child_node.parent_nodes = node.parent_nodes

			//better gc
			node.parent_nodes = nil
			node.child_nodes = nil
		}
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
		trie_db.root_node.child_nodes_hash = nil
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
	}

	return nil
}

// left_path has at least one byte same compared with target_node
func (trie_db *TrieDB) target_node(target_node *Node, left_path []byte, val []byte) error {

	//target exactly
	if bytes.Equal(target_node.path, left_path) {
		if val == nil {
			//del this node
			trie_db.recursive_del(target_node)
		} else {
			//update this node
			target_node.val = val
			target_node.val_hash = nil
		}
		return nil

	} else if (len(left_path) > len(target_node.path)) && bytes.Equal(target_node.path, left_path[0:len(target_node.path)]) {
		// left_path start with target_node.path
		if target_node.child_nodes != nil {
			///
			return trie_db.target_nodes(target_node.child_nodes, left_path[len(target_node.path):], val)
		} else {
			///
			recover_err := trie_db.recover_child_nodes(target_node)
			if recover_err != nil {
				return errors.New("target_node recover_child_nodes err:" + recover_err.Error())
			}
			// nothing to do with nil (del action)
			if target_node.child_nodes == nil {
				if val == nil {
					return nil
				}

				// new nodes that dynamically created
				target_node.child_nodes = &Nodes{
					path_index:  make(map[byte]*Node),
					parent_node: target_node,
				}
				target_node.child_nodes_hash = nil
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

		return nil

	} else {

		// target_node.path, left_path, they have common prefix path

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

		return nil
	}

}
