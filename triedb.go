package triedb

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/xlander-io/cache"
	"github.com/xlander-io/hash"
	"github.com/xlander-io/kv"
)

// Trie implementation
// 1.The val_hash of the root node is always nil.
// 2.Except the root node, a node's parent_node is never nil.
// 3.Root node never has folder_child
// 3.The cache is always synchronized with the disk KV database.
// 4.'attached_hash' stores all key-value hashes related to the trie and will be checked for key hashes to be removed during a commit.
// 5.Get, Update, Del and Commit operations use the same lock to prevent data inconsistency.

// 65535= 2^16 -1 , len can be put inside a uint16 , never change this
// setting a long path will decrease the speed of kvdb
const PATH_LEN_LIMIT = 65535        //65536 - 1
const PATH_FOLDER_DEPTH_LIMIT = 255 //2^8-1

var NODE_HASH_PREFIX []byte = []byte("node_hash_prefix")
var NODE_HASH_VAL_PREFIX []byte = []byte("node_hash_val_prefix")
var NODE_HASH_INDEX_PREFIX []byte = []byte("node_hash_index_prefix")
var NODES_HASH_PREFIX []byte = []byte("nodes_hash_prefix")

// // max bytes limit of full path
// func GetPathLenLimit() int {
// 	return PATH_LEN_LIMIT
// }

func Path(full_path ...[]byte) [][]byte {
	return full_path
}

type trie_cache_item struct {
	val []byte
}

func (item *trie_cache_item) CacheBytes() int {
	return len(item.val)
}

type TrieDBConfig struct {
	Root_hash *hash.Hash

	Read_only bool // Put Del not allowed

	Hash_prefix []byte //config from outside

	node_hash_prefix       []byte // Hash_prefix + NODE_HASH_PREFIX
	node_hash_val_prefix   []byte // Hash_prefix + NODE_HASH_VAL_PREFIX
	node_hash_index_prefix []byte // Hash_prefix + NODE_HASH_INDEX_PREFIX
	nodes_hash_prefix      []byte // Hash_prefix + NODES_HASH_PREFIX

	Update_val_len_limit int // max bytes len
	Commit_thread_limit  int // max concurrent threads during commit

}

type TrieDB struct {
	config *TrieDBConfig
	//
	kvdb  kv.KVDB
	cache *cache.Cache
	lock  sync.Mutex
	//
	root_node *Node
	//
	attached_hash map[string]*hash.Hash //hash => struct{}{} , all related hash in the trie
	//
	commit_thread_available chan struct{} //always >= 0, if 0 => new thread won't be created during commit
}

func NewTrieDB(kvdb_ kv.KVDB, cache_ *cache.Cache, user_config *TrieDBConfig) (*TrieDB, error) {

	if kvdb_ == nil {
		return nil, errors.New("err, kvdb is nil")
	}

	if cache_ == nil {
		return nil, errors.New("err, cache is nil")
	}

	//default config
	config := &TrieDBConfig{
		Root_hash:            nil,
		Update_val_len_limit: 4096 * 1024 * 1024, //4GB
		Commit_thread_limit:  10,
	}

	if user_config != nil {

		//
		hash_prefix := []byte{}
		if user_config.Hash_prefix != nil {
			hash_prefix = append(hash_prefix, user_config.Hash_prefix...)
		}
		user_config.Hash_prefix = hash_prefix
		user_config.node_hash_prefix = append(user_config.Hash_prefix, NODE_HASH_PREFIX...)
		user_config.node_hash_val_prefix = append(user_config.Hash_prefix, NODE_HASH_VAL_PREFIX...)
		user_config.node_hash_index_prefix = append(user_config.Hash_prefix, NODE_HASH_INDEX_PREFIX...)
		user_config.nodes_hash_prefix = append(user_config.Hash_prefix, NODES_HASH_PREFIX...)

		//
		if user_config.Read_only {
			config.Read_only = true
		}

		//
		if user_config.Root_hash != nil {
			config.Root_hash = user_config.Root_hash.Clone()
		}
		//
		if user_config.Update_val_len_limit < 0 {
			return nil, errors.New("config Update_val_len_limit err")
		} else if user_config.Update_val_len_limit == 0 {
			//use default val
		} else {
			config.Update_val_len_limit = user_config.Update_val_len_limit
		}
		//
		if user_config.Commit_thread_limit < 0 {
			return nil, errors.New("config Commit_thread_limit err")
		} else if user_config.Commit_thread_limit == 0 {
			//use default val
		} else {
			config.Commit_thread_limit = user_config.Commit_thread_limit
		}
	}

	//
	if hash.IsNilHash(config.Root_hash) {

		trie_db := TrieDB{
			config:        config,
			cache:         cache_,
			kvdb:          kvdb_,
			attached_hash: make(map[string]*hash.Hash),
			root_node: &Node{
				node_hash:                         hash.NIL_HASH,
				prefix_child_nodes_hash_recovered: true,
				folder_child_nodes_hash_recovered: true,
				val_hash_recovered:                true,
			},
			commit_thread_available: make(chan struct{}, config.Commit_thread_limit),
		}

		return &trie_db, nil

	} else {

		trie_db := TrieDB{
			config:        config,
			cache:         cache_,
			kvdb:          kvdb_,
			attached_hash: make(map[string]*hash.Hash),
			root_node: &Node{
				node_hash:                         config.Root_hash,
				val_hash_recovered:                false,
				prefix_child_nodes_hash_recovered: false,
				folder_child_nodes_hash_recovered: false,
			},
			commit_thread_available: make(chan struct{}, config.Commit_thread_limit),
		}

		root_node_err := trie_db.recover_root_node(trie_db.root_node)
		if root_node_err != nil {
			return nil, errors.New("recover_node err in NewTrieDB, err: " + root_node_err.Error())
		}

		return &trie_db, nil
	}

}

// get first from cache then from kvdb
func (trie_db *TrieDB) get_from_cache_kvdb(key []byte) ([]byte, error) {

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
	}

	if node_val == nil {
		return nil, nil
	}

	//set to cache
	trie_db.cache.Set(key_str, &trie_cache_item{
		val: node_val,
	})

	return node_val, nil
}

// calculate val_hash
func (trie_db *TrieDB) cal_node_val_hash(n *Node) {
	if n.val == nil {
		n.val_hash = nil
	} else {
		result := []byte{}
		result = append(result, trie_db.config.node_hash_val_prefix...)
		result = append(result, n.node_path_flat()...)
		result = append(result, n.val...)
		n.val_hash = hash.CalHash(result)
	}
}

// calculate node_hash
func (trie_db *TrieDB) cal_node_hash(n *Node) {
	result := []byte{}
	result = append(result, trie_db.config.node_hash_prefix...)
	result = append(result, n.node_path_flat()...)
	result = append(result, n.node_bytes...)
	n.node_hash = hash.CalHash(result)
}

// calculate node index hash
func (trie_db *TrieDB) cal_index_hash(n *Node) {
	result := []byte{}
	result = append(result, trie_db.config.node_hash_index_prefix...)
	result = append(result, n.node_path_flat()...)
	n.index_hash = hash.CalHash(result)
}

// calculate nodes_hash
func (trie_db *TrieDB) cal_nodes_hash(n *nodes) {
	result := []byte{}
	result = append(result, trie_db.config.nodes_hash_prefix...)
	result = append(result, n.parent_node.node_path_flat()...)
	result = append(result, n.nodes_bytes...)
	if n.is_folder_child_nodes {
		n.parent_node.folder_child_nodes_hash = hash.CalHash(result)
	} else {
		n.parent_node.prefix_child_nodes_hash = hash.CalHash(result)
	}

}

func (trie_db *TrieDB) attach_hash(hash *hash.Hash) {
	if hash != nil {
		trie_db.attached_hash[string(hash.Bytes())] = hash
	}
}

func (trie_db *TrieDB) recover_node_val(node *Node) error {

	if node == nil {
		return errors.New("recover_node_val err, node nil")
	}

	//already read in the past or new created
	if node.val_hash_recovered || node.val_hash == nil {
		return nil
	}

	node_val_bytes, node_val_err := trie_db.get_from_cache_kvdb(node.val_hash.Bytes())
	if node_val_err != nil {
		return errors.New("recover_node_val get_from_cache_kvdb  err, " + node_val_err.Error())
	}

	//
	node.val = node_val_bytes
	node.val_hash_recovered = true
	//
	return nil
}

func (trie_db *TrieDB) recover_root_node(node *Node) error {

	if node == nil {
		return errors.New("recover_node err, node nil")
	}

	node_bytes, node_err := trie_db.get_from_cache_kvdb(node.node_hash.Bytes())
	if node_err != nil {
		return errors.New("recover_node get_from_cache_kvdb  err, node_hash: " + fmt.Sprintf("%x", node.node_hash.Bytes()))
	}

	if node_bytes == nil {
		return errors.New("recover_node get_from_cache_kvdb  err, node_hash not found")
	}
	//
	node.node_bytes = node_bytes
	node.deserialize()

	trie_db.attach_hash(node.index_hash)
	trie_db.attach_hash(node.val_hash)
	trie_db.attach_hash(node.folder_child_nodes_hash)
	trie_db.attach_hash(node.prefix_child_nodes_hash)
	trie_db.attach_hash(node.node_hash)

	//
	return nil
}

func (trie_db *TrieDB) recover_child_nodes(node *Node, folder_child bool, prefix_child bool) error {

	if node == nil {
		return errors.New("recover_child_nodes err, node nil")
	}

	if folder_child {

		if node.folder_child_nodes == nil && !node.folder_child_nodes_hash_recovered && node.folder_child_nodes_hash != nil {

			nodes_bytes, err := trie_db.get_from_cache_kvdb(node.folder_child_nodes_hash.Bytes())
			if err != nil {
				return errors.New("recover_folder_child_nodes get_from_cache_kvdb err : " + err.Error())
			}

			if nodes_bytes == nil {
				return errors.New("recover_folder_child_nodes err : folder_child_nodes_hash not found")
			}

			//
			child_nodes_ := nodes{
				is_folder_child_nodes: true,
				nodes_bytes:           nodes_bytes,
				parent_node:           node,
			}
			//
			child_nodes_.deserialize()

			path_b_iter := child_nodes_.btree.Before(uint8(0))
			for path_b_iter.Next() {
				c_n := path_b_iter.Value.(*Node)
				//
				trie_db.attach_hash(c_n.val_hash)
				trie_db.attach_hash(c_n.folder_child_nodes_hash)
				trie_db.attach_hash(c_n.prefix_child_nodes_hash)
				trie_db.attach_hash(c_n.index_hash)
				trie_db.attach_hash(c_n.node_hash)
				//
				c_n.parent_nodes = &child_nodes_
			}
			//
			node.folder_child_nodes = &child_nodes_
			node.folder_child_nodes_hash_recovered = true
		}
	}

	if prefix_child {

		if node.prefix_child_nodes == nil && !node.prefix_child_nodes_hash_recovered && node.prefix_child_nodes_hash != nil {

			nodes_bytes, err := trie_db.get_from_cache_kvdb(node.prefix_child_nodes_hash.Bytes())
			if err != nil {
				return errors.New("recover_prefix_child_nodes get_from_cache_kvdb err : " + err.Error())
			}

			if nodes_bytes == nil {
				return errors.New("recover_prefix_child_nodes err : folder_prefix_nodes_hash not found")
			}

			//
			child_nodes_ := nodes{
				is_folder_child_nodes: false,
				nodes_bytes:           nodes_bytes,
				parent_node:           node,
			}
			//
			child_nodes_.deserialize()

			path_b_iter := child_nodes_.btree.Before(uint8(0))
			for path_b_iter.Next() {
				c_n := path_b_iter.Value.(*Node)
				//
				trie_db.attach_hash(c_n.val_hash)
				trie_db.attach_hash(c_n.folder_child_nodes_hash)
				trie_db.attach_hash(c_n.prefix_child_nodes_hash)
				trie_db.attach_hash(c_n.index_hash)
				trie_db.attach_hash(c_n.node_hash)
				//
				c_n.parent_nodes = &child_nodes_
			}
			//
			node.prefix_child_nodes = &child_nodes_
			node.prefix_child_nodes_hash_recovered = true
		}
	}

	//
	return nil
}

func (trie_db *TrieDB) put_target_nodes(target_nodes *nodes, full_path [][]byte, path_level int, left_prefix []byte, val []byte, gen_hash_index bool) (*Node, error) {

	next_target_node_i := target_nodes.btree.Get(uint8(left_prefix[0]))
	if next_target_node_i != nil {
		return trie_db.put_target_node(next_target_node_i.(*Node), full_path, path_level, left_prefix, val, gen_hash_index)
	} else {

		is_final_path := ((len(full_path) - 1) == path_level)

		if is_final_path {

			new_node := &Node{
				prefix:                            left_prefix,
				parent_nodes:                      target_nodes,
				val:                               val,
				val_dirty:                         true,
				folder_child_nodes_hash_recovered: true,
				prefix_child_nodes_hash_recovered: true,
				val_hash_recovered:                true,
				dirty:                             true,
			}

			if gen_hash_index {
				trie_db.cal_index_hash(new_node)
			} else {
				new_node.index_hash = nil
			}

			target_nodes.btree.Set(uint8(left_prefix[0]), new_node)
			//
			target_nodes.mark_dirty()
			//
			return new_node, nil

		} else {

			new_node := &Node{
				prefix:                            left_prefix,
				parent_nodes:                      target_nodes,
				val:                               nil,
				folder_child_nodes_hash_recovered: true,
				prefix_child_nodes_hash_recovered: true,
				val_hash_recovered:                true,
				dirty:                             true,
			}

			new_node.folder_child_nodes = &nodes{
				is_folder_child_nodes: true,
				btree:                 new_nodes_btree(),
				parent_node:           new_node,
				dirty:                 true,
			}

			target_nodes.btree.Set(uint8(left_prefix[0]), new_node)
			//
			target_nodes.mark_dirty()
			//
			return trie_db.put_target_nodes(new_node.folder_child_nodes, full_path, path_level+1, full_path[path_level+1], val, gen_hash_index)

		}

	}

}

func (trie_db *TrieDB) put_target_node(target_node *Node, full_path [][]byte, path_level int, left_prefix []byte, val []byte, gen_hash_index bool) (*Node, error) {

	is_final_path := ((len(full_path) - 1) == path_level)

	if target_node.parent_nodes == nil {
		//root node
		recover_child_err := trie_db.recover_child_nodes(target_node, true, false)
		if recover_child_err != nil {
			return nil, errors.New("put_target_node recover_child_nodes(*,true,false) err, " + recover_child_err.Error())
		}

		if target_node.folder_child_nodes == nil {
			target_node.folder_child_nodes = &nodes{
				is_folder_child_nodes: true,
				btree:                 new_nodes_btree(),
				parent_node:           target_node,
				nodes_bytes:           nil,
				dirty:                 true,
			}
		}

		return trie_db.put_target_nodes(target_node.folder_child_nodes, full_path, path_level, left_prefix, val, gen_hash_index)

	} else if bytes.Equal(target_node.prefix, left_prefix) {
		//target exactly
		if !is_final_path {

			recover_child_err := trie_db.recover_child_nodes(target_node, true, false)
			if recover_child_err != nil {
				return nil, errors.New("update_target_node recover_child_nodes(*,true,false) err, " + recover_child_err.Error())
			}

			if target_node.folder_child_nodes == nil {
				target_node.folder_child_nodes = &nodes{
					is_folder_child_nodes: true,
					btree:                 new_nodes_btree(),
					parent_node:           target_node,
					nodes_bytes:           nil,
					dirty:                 true,
				}
			}

			return trie_db.put_target_nodes(target_node.folder_child_nodes, full_path, path_level+1, full_path[path_level+1], val, gen_hash_index)

		} else {

			//update hash_index first
			if gen_hash_index {
				trie_db.cal_index_hash(target_node)
			} else {
				target_node.index_hash = nil
			}
			//
			target_node.val = val
			target_node.val_hash = nil
			target_node.val_hash_recovered = true
			target_node.val_dirty = true
			//
			target_node.mark_dirty()
			//
			return target_node, nil
		}

	} else if (len(left_prefix) > len(target_node.prefix)) && bytes.Equal(target_node.prefix, left_prefix[0:len(target_node.prefix)]) {
		// left_prefix start with target_node.prefix
		recover_err := trie_db.recover_child_nodes(target_node, false, true)
		if recover_err != nil {
			return nil, errors.New("update_target_node recover_child_nodes(*,false,true) err:" + recover_err.Error())
		}

		if target_node.prefix_child_nodes != nil {
			return trie_db.put_target_nodes(target_node.prefix_child_nodes, full_path, path_level, left_prefix[len(target_node.prefix):], val, gen_hash_index)
		} else {
			// new nodes that dynamically created
			target_node.prefix_child_nodes = &nodes{
				is_folder_child_nodes: false,
				btree:                 new_nodes_btree(),
				parent_node:           target_node,
				dirty:                 true,
			}
			//
			target_node.mark_dirty()
			//
			return trie_db.put_target_nodes(target_node.prefix_child_nodes, full_path, path_level, left_prefix[len(target_node.prefix):], val, gen_hash_index)
		}

	} else if (len(left_prefix) < len(target_node.prefix)) && bytes.Equal(left_prefix, target_node.prefix[0:len(left_prefix)]) {
		// target_node.prefix start with left_prefix
		new_node := &Node{
			prefix:                            left_prefix[:],
			parent_nodes:                      target_node.parent_nodes,
			val:                               nil,
			val_hash:                          nil,
			val_hash_recovered:                true,
			folder_child_nodes_hash_recovered: true,
			prefix_child_nodes_hash_recovered: true,
			dirty:                             true,
		}

		//
		new_node.prefix_child_nodes = &nodes{
			is_folder_child_nodes: false,
			btree:                 new_nodes_btree(),
			parent_node:           new_node,
			dirty:                 true,
		}

		target_node.prefix = target_node.prefix[len(left_prefix):]
		target_node.parent_nodes.btree.Set(uint8(left_prefix[0]), new_node)
		new_node.parent_nodes = target_node.parent_nodes
		target_node.parent_nodes = new_node.prefix_child_nodes
		new_node.prefix_child_nodes.btree.Set(target_node.prefix[0], target_node)

		if is_final_path {

			//
			if gen_hash_index {
				trie_db.cal_index_hash(new_node)
			} else {
				new_node.index_hash = nil
			}

			//
			new_node.val = val
			new_node.val_hash = nil
			new_node.val_hash_recovered = true
			new_node.val_dirty = true

			//mark dirty
			new_node.mark_dirty()
			//
			return new_node, nil

		} else {
			//
			new_node.folder_child_nodes = &nodes{
				is_folder_child_nodes: true,
				btree:                 new_nodes_btree(),
				parent_node:           new_node,
				dirty:                 true,
			}
			//mark dirty
			new_node.mark_dirty()
			//
			return trie_db.put_target_nodes(new_node.folder_child_nodes, full_path, path_level+1, full_path[path_level+1], val, gen_hash_index)
		}

	} else {
		// target_node.path, left_prefix, they have common prefix

		////////// find the common bytes prefix
		min_len := len(left_prefix)
		target_node_path_len := len(target_node.prefix)
		if target_node_path_len < min_len {
			min_len = target_node_path_len
		}

		common_prefix_bytes_len := 0
		for i := 0; i < min_len; i++ {
			if left_prefix[i] == target_node.prefix[i] {
				common_prefix_bytes_len++
			} else {
				break
			}
		}

		common_prefix_bytes := left_prefix[0:common_prefix_bytes_len]

		//
		new_parent_node := &Node{
			prefix:                            common_prefix_bytes[:],
			parent_nodes:                      (*target_node).parent_nodes,
			prefix_child_nodes_hash_recovered: true,
			folder_child_nodes_hash_recovered: true,
			val_hash_recovered:                true,
			dirty:                             true,
		}

		new_parent_node.prefix_child_nodes = &nodes{
			is_folder_child_nodes: false,
			btree:                 new_nodes_btree(),
			parent_node:           new_parent_node,
			dirty:                 true,
		}

		//
		new_parent_node.prefix_child_nodes.btree.Set(uint8(target_node.prefix[common_prefix_bytes_len]), target_node)
		target_node.prefix = target_node.prefix[common_prefix_bytes_len:]
		target_node.parent_nodes.btree.Set(uint8(new_parent_node.prefix[0]), new_parent_node)
		new_parent_node.parent_nodes = target_node.parent_nodes
		target_node.parent_nodes = new_parent_node.prefix_child_nodes

		//
		new_parent_node.parent_nodes.mark_dirty()

		return trie_db.put_target_nodes(new_parent_node.prefix_child_nodes, full_path, path_level, left_prefix[common_prefix_bytes_len:], val, gen_hash_index)

	}

}

type PutError struct {
	Err   error
	Fatal bool //trie_db won't be used any more as status chaos may exist
}

func (trie_db *TrieDB) Put(full_path [][]byte, val []byte, gen_hash_index bool) *PutError {
	_, err := trie_db.put_(full_path, val, gen_hash_index)
	return err
}

// update the target and return the related node
func (trie_db *TrieDB) put_(full_path [][]byte, val []byte, gen_hash_index bool) (*Node, *PutError) {

	if trie_db.config.Read_only {
		return nil, &PutError{
			Err:   errors.New("put is not allowed for read only trie"),
			Fatal: false,
		}
	}

	//val check
	if len(val) == 0 {
		return nil, &PutError{
			Err:   errors.New("update val empty"),
			Fatal: false,
		}
	}
	if len(val) > trie_db.config.Update_val_len_limit {
		return nil, &PutError{
			Err:   errors.New("trie val size over limit"),
			Fatal: false,
		}
	}

	//path limit check
	if len(full_path) <= 0 {
		return nil, &PutError{
			Err:   errors.New("full_path empty"),
			Fatal: false,
		}
	}
	if len(full_path) > PATH_FOLDER_DEPTH_LIMIT {
		return nil, &PutError{
			Err:   errors.New("full_path folder depth over limit"),
			Fatal: false,
		}
	}

	full_prefix := []byte{}
	for _, path := range full_path {
		if len(path) == 0 {
			return nil, &PutError{
				Err:   errors.New("empty path error"),
				Fatal: false,
			}
		}
		full_prefix = append(full_prefix, path...)
		if len(full_prefix) > PATH_LEN_LIMIT {
			return nil, &PutError{
				Err:   errors.New("full_path over limit error"),
				Fatal: false,
			}
		}
	}

	trie_db.lock.Lock()
	defer trie_db.lock.Unlock()

	put_n, put_err := trie_db.put_target_node(trie_db.root_node, full_path, 0, full_path[0], val, gen_hash_index)
	if put_err != nil {
		return nil, &PutError{
			Err:   put_err,
			Fatal: true,
		}
	} else {
		return put_n, nil
	}
}

// recursively simplify
func (trie_db *TrieDB) recursive_simplify(node *Node) error {
	//
	if node == nil {
		return errors.New("simplify nil node error")
	}

	//root node nothing to simplify
	if node.parent_nodes == nil {
		return nil
	}

	if node.has_folder_child() {
		//has folder child nothing to simplify
		return nil

	} else if node.has_prefix_child() {

		recover_child_nodes_err := trie_db.recover_child_nodes(node, false, true)
		if recover_child_nodes_err != nil {
			return recover_child_nodes_err
		}

		//simplify
		if node.prefix_child_nodes.btree.Len() == 1 {
			_, p_c_single_node_i := node.prefix_child_nodes.btree.Min()
			p_c_single_node := p_c_single_node_i.(*Node)
			p_c_single_node.prefix = append(node.prefix, p_c_single_node.prefix...)
			node.parent_nodes.btree.Set(uint8(p_c_single_node.prefix[0]), p_c_single_node)
			p_c_single_node.parent_nodes = node.parent_nodes
			//
			p_c_single_node.parent_nodes.mark_dirty()
			//
			return nil

		} else {
			//more then one child
			return nil
		}

	} else {

		//no any prefix|folder child

		//del first
		node.parent_nodes.btree.Delete(uint8(node.prefix[0]))

		if node.parent_nodes.is_folder_child_nodes {

			//condition folder child

			//simplify
			if node.parent_nodes.btree.Len() == 0 {
				node.parent_nodes.parent_node.folder_child_nodes = nil
				node.parent_nodes.parent_node.mark_dirty()
				//simplify
				if !node.parent_nodes.parent_node.has_val() {
					return trie_db.recursive_simplify(node.parent_nodes.parent_node)
				} else {
					return nil
				}
			} else {
				node.parent_nodes.mark_dirty()
				return nil
			}

		} else {

			//condition prefix child

			//simplify
			if node.parent_nodes.btree.Len() == 0 {
				//
				node.parent_nodes.parent_node.prefix_child_nodes = nil
				node.parent_nodes.parent_node.mark_dirty()

				//simplify
				if !node.parent_nodes.parent_node.has_val() {
					return trie_db.recursive_simplify(node.parent_nodes.parent_node)
				} else {
					return nil
				}

			} else if node.parent_nodes.btree.Len() == 1 &&
				!node.parent_nodes.parent_node.has_folder_child() &&
				!node.parent_nodes.parent_node.has_val() &&
				node.parent_nodes.parent_node.parent_nodes != nil {

				//
				_, left_single_node_i := node.parent_nodes.btree.Min()
				left_single_node := left_single_node_i.(*Node)
				left_single_node.prefix = append(node.parent_nodes.parent_node.prefix, left_single_node.prefix...)

				node.parent_nodes.parent_node.parent_nodes.btree.Set(uint8(left_single_node.prefix[0]), left_single_node)
				left_single_node.parent_nodes = node.parent_nodes.parent_node.parent_nodes
				//
				left_single_node.parent_nodes.mark_dirty()
				return nil

			} else {
				node.parent_nodes.mark_dirty()
				return nil
			}

		}

	}

}

func (trie_db *TrieDB) del_target_node(target_node *Node, full_path [][]byte, path_level int, left_prefix []byte) (bool, error) {

	if target_node.parent_nodes == nil {
		//root node
		recover_err := trie_db.recover_child_nodes(target_node, true, false)
		if recover_err != nil {
			return false, recover_err
		}

		//not found
		if target_node.folder_child_nodes == nil {
			return false, nil
		}

		c_n_i := target_node.folder_child_nodes.btree.Get(uint8(left_prefix[0]))
		if c_n_i == nil {
			//not found
			return false, nil
		}

		return trie_db.del_target_node(c_n_i.(*Node), full_path, path_level, left_prefix)

	} else if bytes.Equal(target_node.prefix, left_prefix) {
		//target exactly
		//
		is_final_path := ((len(full_path) - 1) == path_level)
		//
		if is_final_path {

			if !target_node.has_val() {
				//not found
				return false, nil
			}

			//del
			target_node.val = nil
			target_node.val_hash = nil
			target_node.val_hash_recovered = true
			target_node.val_dirty = true
			target_node.index_hash = nil
			target_node.mark_dirty()

			//
			simplify_err := trie_db.recursive_simplify(target_node)
			if simplify_err != nil {
				return false, simplify_err
			} else {
				return true, nil
			}

		} else {
			recover_err := trie_db.recover_child_nodes(target_node, true, false)
			if recover_err != nil {
				return false, recover_err
			}

			if target_node.folder_child_nodes == nil {
				//not found
				return false, nil
			}

			c_n_i := target_node.folder_child_nodes.btree.Get(uint8(full_path[path_level+1][0]))
			if c_n_i == nil {
				//not found
				return false, nil
			}

			return trie_db.del_target_node(c_n_i.(*Node), full_path, path_level+1, full_path[path_level+1])
		}
	} else if (len(left_prefix) > len(target_node.prefix)) && bytes.Equal(target_node.prefix, left_prefix[0:len(target_node.prefix)]) {
		// left_prefix start with target_node.prefix

		recover_err := trie_db.recover_child_nodes(target_node, false, true)
		if recover_err != nil {
			return false, recover_err
		}

		if target_node.prefix_child_nodes == nil {
			//not found
			return false, nil
		}

		c_n_i := target_node.prefix_child_nodes.btree.Get(uint8(left_prefix[len(target_node.prefix)]))
		if c_n_i == nil {
			//not found
			return false, nil
		}

		return trie_db.del_target_node(c_n_i.(*Node), full_path, path_level, left_prefix[len(target_node.prefix):])

	} else {
		// conditon target_node.prefix start with left_prefix or
		// condition target_node.path, left_prefix, they have common prefix

		//not found
		return false, nil
	}

}

type DelError struct {
	Err   error
	Fatal bool //trie_db won't be used any more as status chaos may exist
}

// return params:
//
//	 	bool:
//					true:  node found and node.val exist
//					false: node not found or node found but node.val not exist
//		error:	may exist make sure you check the fatal, if fatal is true the trie can't be used anymore as the status inside may be chaotic
func (trie_db *TrieDB) Del(full_path [][]byte) (bool, *DelError) {

	if trie_db.config.Read_only {
		return false, &DelError{
			Err:   errors.New("del is not allowed for read only trie"),
			Fatal: false,
		}
	}

	//path limit check
	if len(full_path) <= 0 {
		return false, &DelError{
			Err:   errors.New("full_path empty"),
			Fatal: false,
		}
	}

	//path empty check
	for _, path := range full_path {
		if len(path) == 0 {
			return false, &DelError{
				Err:   errors.New("empty path error"),
				Fatal: false,
			}
		}
	}

	trie_db.lock.Lock()
	defer trie_db.lock.Unlock()

	del_success, del_err := trie_db.del_target_node(trie_db.root_node, full_path, 0, full_path[0])
	if del_err != nil {
		return false, &DelError{
			Err:   del_err,
			Fatal: true,
		}
	} else {
		return del_success, nil
	}
}

////////

func (trie_db *TrieDB) get_target_nodes(target_nodes *nodes, full_path [][]byte, path_level int, left_prefix []byte) (*Node, error) {
	next_target_node_i := target_nodes.btree.Get(uint8(left_prefix[0]))
	if next_target_node_i != nil {
		return trie_db.get_target_node(next_target_node_i.(*Node), full_path, path_level, left_prefix)
	}
	//not found
	return nil, nil
}

func (trie_db *TrieDB) get_target_node(target_node *Node, full_path [][]byte, path_level int, left_prefix []byte) (*Node, error) {
	is_final_path := ((len(full_path) - 1) == path_level)

	if target_node.parent_nodes == nil {
		//root node
		recover_err := trie_db.recover_child_nodes(target_node, true, false)
		if recover_err != nil {
			return nil, recover_err
		}

		//not found
		if target_node.folder_child_nodes == nil {
			return nil, nil
		}

		return trie_db.get_target_nodes(target_node.folder_child_nodes, full_path, path_level, left_prefix)

	} else if bytes.Equal(target_node.prefix, left_prefix) {
		//target exactly

		if !is_final_path {

			recover_err := trie_db.recover_child_nodes(target_node, true, false)
			if recover_err != nil {
				return nil, recover_err
			}

			//not found
			if target_node.folder_child_nodes == nil {
				return nil, nil
			}

			return trie_db.get_target_nodes(target_node.folder_child_nodes, full_path, path_level+1, full_path[path_level+1])

		} else {

			if !target_node.has_val() && !target_node.has_folder_child() {
				return nil, nil
			}

			// recover_err := trie_db.recover_node_val(target_node)
			// if recover_err != nil {
			// 	return nil, recover_err
			// }

			return target_node, nil

		}
	} else if (len(left_prefix) > len(target_node.prefix)) && bytes.Equal(target_node.prefix, left_prefix[0:len(target_node.prefix)]) {
		// left_prefix start with target_node.prefix
		recover_err := trie_db.recover_child_nodes(target_node, false, true)
		if recover_err != nil {
			return nil, recover_err
		}

		//not found
		if target_node.prefix_child_nodes == nil {
			return nil, nil
		}

		return trie_db.get_target_nodes(target_node.prefix_child_nodes, full_path, path_level, left_prefix[len(target_node.prefix):])

	} else {
		// conditions:
		// target_node.prefix start with left_prefixelse
		// target_node.path, left_prefix, they have common prefix
		return nil, nil
	}
}

func (trie_db *TrieDB) get_(full_path [][]byte) (*Node, error) {

	if len(full_path) == 0 {
		return trie_db.root_node, nil
	}

	//path empty check
	for _, path := range full_path {
		if len(path) == 0 {
			return nil, errors.New("empty path error")
		}
	}

	trie_db.lock.Lock()
	defer trie_db.lock.Unlock()

	return trie_db.get_target_node(trie_db.root_node, full_path, 0, full_path[0])
}

func (trie_db *TrieDB) Get(full_path [][]byte) ([]byte, error) {
	target_node, err := trie_db.get_(full_path)
	if err != nil {
		return nil, err
	}

	recover_err := trie_db.recover_node_val(target_node)
	if recover_err != nil {
		return nil, recover_err
	}

	if target_node.val == nil {
		//is just a folder
		return nil, nil
	}

	return target_node.val, nil
}

func (trie_db *TrieDB) FolderExist(full_path [][]byte) (bool, error) {
	target_node, err := trie_db.get_(full_path)
	if err != nil {
		return false, err
	}

	if target_node.has_folder_child() {
		return true, nil
	} else {
		return false, nil
	}
}

// k_v_map to collected all the k_v , string(key) => []byte(value)
func (trie_db *TrieDB) commit_recursive(node *Node, k_v_map *sync.Map) (*hash.Hash, error) {

	//
	child_num := 0
	if node.prefix_child_nodes != nil {
		child_num += node.prefix_child_nodes.btree.Len()
	}
	if node.folder_child_nodes != nil {
		child_num += node.folder_child_nodes.btree.Len()
	}

	if child_num != 0 {

		child_result_chan := make(chan error, child_num)

		//
		if node.prefix_child_nodes != nil && node.prefix_child_nodes.btree.Len() > 0 {
			p_c_iter := node.prefix_child_nodes.btree.Before(uint8(0))
			for p_c_iter.Next() {
				go func(cn *Node) {
					trie_db.commit_thread_available <- struct{}{}
					_, cn_h_err := trie_db.commit_recursive(cn, k_v_map)
					child_result_chan <- cn_h_err
					<-trie_db.commit_thread_available
				}(p_c_iter.Value.(*Node))
			}
			<-trie_db.commit_thread_available //give out a thread-slot
		}

		//
		if node.folder_child_nodes != nil && node.folder_child_nodes.btree.Len() > 0 {
			f_c_iter := node.folder_child_nodes.btree.Before(uint8(0))
			for f_c_iter.Next() {
				go func(cn *Node) {
					trie_db.commit_thread_available <- struct{}{}
					_, cn_h_err := trie_db.commit_recursive(cn, k_v_map)
					child_result_chan <- cn_h_err
					<-trie_db.commit_thread_available
				}(f_c_iter.Value.(*Node))
			}
			<-trie_db.commit_thread_available //give out a thread-slot
		}

		//make sure all sub-thread done
		for range child_num {
			cn_err := <-child_result_chan
			if cn_err != nil {
				return nil, cn_err
			}
		}

		//
		if node.prefix_child_nodes != nil && node.prefix_child_nodes.btree.Len() > 0 {
			trie_db.commit_thread_available <- struct{}{} //get back the thread-slot
		}

		//
		if node.folder_child_nodes != nil && node.folder_child_nodes.btree.Len() > 0 {
			trie_db.commit_thread_available <- struct{}{} //get back the thread-slot
		}
	}

	//
	if node.dirty {
		//
		if node.prefix_child_nodes != nil && node.prefix_child_nodes.dirty {
			node.prefix_child_nodes.serialize()
			trie_db.cal_nodes_hash(node.prefix_child_nodes)
		}
		//
		if node.folder_child_nodes != nil && node.folder_child_nodes.dirty {
			node.folder_child_nodes.serialize()
			trie_db.cal_nodes_hash(node.folder_child_nodes)
		}

		//cal val hash
		if node.val_dirty {
			trie_db.cal_node_val_hash(node)
		}

		//cal node hash
		node.serialize()
		trie_db.cal_node_hash(node)
	}

	//////////////// store all related //////////////////////////
	if node.val_hash != nil {
		//node.val may be nil because of lazy loading
		k_v_map.Store(string(node.val_hash.Bytes()), node.val)
	}

	if node.index_hash != nil {
		k_v_map.Store(string(node.index_hash.Bytes()), node.node_path_flat())
	}

	if node.prefix_child_nodes_hash != nil {
		if node.prefix_child_nodes != nil {
			k_v_map.Store(string(node.prefix_child_nodes_hash.Bytes()), node.prefix_child_nodes.nodes_bytes)
		} else {
			//may caused by no loading because of lazy loading feature
			k_v_map.Store(string(node.prefix_child_nodes_hash.Bytes()), nil)
		}
	}

	if node.folder_child_nodes_hash != nil {
		if node.folder_child_nodes != nil {
			k_v_map.Store(string(node.folder_child_nodes_hash.Bytes()), node.folder_child_nodes.nodes_bytes)
		} else {
			//may caused by no loading because of lazy loading feature
			k_v_map.Store(string(node.folder_child_nodes_hash.Bytes()), nil)
		}
	}

	if node.node_hash == nil {
		fmt.Println(node)
	}
	k_v_map.Store(string(node.node_hash.Bytes()), node.node_bytes)

	return node.node_hash, nil
}

// return root_hash, updated hash map, del hash map ,error
func (trie_db *TrieDB) Commit() (*hash.Hash, map[string][]byte, map[string]*hash.Hash, error) {
	trie_db.lock.Lock()
	defer trie_db.lock.Unlock()

	//
	all_trie_k_v := sync.Map{}
	//
	trie_db.commit_thread_available <- struct{}{} //main thread-slot
	//
	_, cal_hash_err := trie_db.commit_recursive(trie_db.root_node, &all_trie_k_v)
	if cal_hash_err != nil {
		return nil, nil, nil, cal_hash_err
	}
	//
	<-trie_db.commit_thread_available //main thread-slot
	//

	/////////////////////////////////////////////
	update_k_v := make(map[string][]byte)
	del_k_v := make(map[string]*hash.Hash)

	// what to delete
	for key_str, key_hash := range trie_db.attached_hash {
		if _, found := all_trie_k_v.Load(key_str); !found {
			del_k_v[key_str] = key_hash
		}
	}

	//what to update
	all_trie_k_v.Range(func(key, value any) bool {
		if _, exist := trie_db.attached_hash[key.(string)]; !exist {
			update_k_v[key.(string)] = value.([]byte)
		}
		return true
	})

	return trie_db.root_node.node_hash, update_k_v, del_k_v, nil
}
