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

/*
 *		Trie implementation
 *		1. For the root node, its val_hash is always nil
 *		2. for a nodes, its parent_node is always not nil
 *		3. cache is always sync with disk kv db
 *		4. attached_hash stores all the kv k hash related in the trie, attached_hash will be checked
 *		   what key hash will be removed in commit
 *		5. Get,Put,Commit use the same lock to prevent data inconsistence
 */

// ///////////////////////////
type trie_cache_item struct {
	val []byte
}

func (item *trie_cache_item) CacheBytes() int {
	return len(item.val)
}

/////////////////////////////

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

type TrieDBConfig struct {
	Root_hash            *hash.Hash
	Update_val_len_limit int // max bytes len
	Commit_thread_limit  int // max concurrent threads during commit
}

func NewTrieDB(kvdb_ kv.KVDB, cache_ *cache.Cache, user_config *TrieDBConfig) (*TrieDB, error) {

	if kvdb_ == nil {
		return nil, errors.New("NewTrieDB kvdb is nil")
	}

	if cache_ == nil {
		return nil, errors.New("NewTrieDB cache is nil")
	}

	//default config
	config := &TrieDBConfig{
		Root_hash:            nil,
		Update_val_len_limit: 4096 * 1024 * 1024, //4GB
		Commit_thread_limit:  10,
	}

	if user_config != nil {
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

	if hash.IsNilHash(config.Root_hash) {

		trie_db := TrieDB{
			config:        config,
			cache:         cache_,
			kvdb:          kvdb_,
			attached_hash: make(map[string]*hash.Hash),
			root_node: &Node{
				node_hash:                  hash.NIL_HASH,
				child_nodes_hash_recovered: true,
				val_hash_recovered:         true,
				node_hash_recovered:        true,
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
				node_hash:                  config.Root_hash,
				val_hash_recovered:         false,
				child_nodes_hash_recovered: false,
				node_hash_recovered:        false,
			},
			commit_thread_available: make(chan struct{}, config.Commit_thread_limit),
		}

		root_node_err := trie_db.recover_node(trie_db.root_node)
		if root_node_err != nil {
			return nil, errors.New("recover_node err in NewTrieDB, err: " + root_node_err.Error())
		}

		return &trie_db, nil
	}
}

func (trie_db *TrieDB) attachHash(hash *hash.Hash) {
	if hash != nil {
		trie_db.attached_hash[string(hash.Bytes())] = hash
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

func (trie_db *TrieDB) recover_node(node *Node) error {

	if node == nil {
		return errors.New("recover_node err, node nil")
	}

	defer func() {
		// delete to prevent double recover
		node.node_hash_recovered = true
	}()

	//already read in the past or new created
	if node.node_hash_recovered || node.node_hash == nil {
		return nil
	}

	node_bytes, node_err := trie_db.getFromCacheKVDB(node.node_hash.Bytes())
	if node_err != nil {
		return errors.New("recover_node getFromCacheKVDB  err, node_hash: " + fmt.Sprintf("%x", node.node_hash.Bytes()))
	}

	if node_bytes == nil {
		return errors.New("recover_node getFromCacheKVDB  err, node_hash not found")
	}
	//
	node.node_bytes = node_bytes
	node.deserialize()

	trie_db.attachHash(node.node_hash)
	trie_db.attachHash(node.val_hash)         //don't froget val_hash as it may be affected during direct del action
	trie_db.attachHash(node.child_nodes_hash) //better put this

	//
	return nil
}

func (trie_db *TrieDB) recover_node_val(node *Node) error {

	if node == nil {
		return errors.New("recover_node_val err, node nil")
	}

	defer func() {
		//delete to prevent double recover
		node.val_hash_recovered = true
	}()

	//already read in the past or new created
	if node.val_hash_recovered || node.val_hash == nil {
		return nil
	}

	node_val_bytes, node_val_err := trie_db.getFromCacheKVDB(node.val_hash.Bytes())
	if node_val_err != nil {
		return errors.New("recover_node_val getFromCacheKVDB  err, " + node_val_err.Error())
	}

	//
	node.val = node_val_bytes

	//
	return nil
}

func (trie_db *TrieDB) recover_child_nodes(node *Node) error {

	if node == nil {
		return errors.New("recover_child_nodes err, node nil")
	}

	defer func() {
		//delete to prevent double recover
		node.child_nodes_hash_recovered = true
	}()

	if node.child_nodes == nil && !node.child_nodes_hash_recovered && node.child_nodes_hash != nil {

		nodes_bytes, err := trie_db.getFromCacheKVDB(node.child_nodes_hash.Bytes())
		if err != nil {
			return errors.New("recover_child_nodes err : " + err.Error())
		}

		if nodes_bytes == nil {
			return errors.New("recover_child_nodes err : child_nodes_hash not found")
		}

		//
		child_nodes_ := nodes{
			nodes_bytes: nodes_bytes,
			parent_node: node,
		}
		//
		child_nodes_.deserialize()

		for _, c_n := range child_nodes_.path_index {
			trie_db.attachHash(c_n.node_hash)
			c_n.parent_nodes = &child_nodes_
		}

		//
		node.child_nodes = &child_nodes_

	}

	//
	return nil
}

// recursive_del will be called when del val happen
func (trie_db *TrieDB) update_recursive_del(node *Node) error {

	//won't happen
	if node == nil || node.parent_nodes == nil {
		return nil
	}

	//
	r_err := trie_db.recover_child_nodes(node)
	if r_err != nil {
		return r_err
	}
	//
	if node.child_nodes == nil {
		//
		delete(node.parent_nodes.path_index, node.path[0])

		//what is left in parent_nodes
		if len(node.parent_nodes.path_index) == 0 {
			//
			node.parent_nodes.parent_node.child_nodes = nil
			node.parent_nodes.parent_node.child_nodes_hash = nil
			//don't forget to mark_dirty before next potential recursive
			node.parent_nodes.parent_node.mark_dirty()
			//
			if node.parent_nodes.parent_node.val == nil {
				return trie_db.update_recursive_del(node.parent_nodes.parent_node)
			}

		} else if len(node.parent_nodes.path_index) == 1 &&
			node.parent_nodes.parent_node.parent_nodes != nil && // !=nil checks the root node
			node.parent_nodes.parent_node.val == nil {

			// do simplification if possible
			var left_single_node *Node
			for _, c_n := range node.parent_nodes.path_index {
				left_single_node = c_n
				break
			}

			//
			left_single_node.path = append(left_single_node.parent_nodes.parent_node.path, left_single_node.path...)
			left_single_node.full_path_cache = nil //reset pull path cache

			node.parent_nodes.parent_node.parent_nodes.path_index[left_single_node.path[0]] = left_single_node
			left_single_node.parent_nodes = node.parent_nodes.parent_node.parent_nodes

			//because path changes, recover val is required
			recover_err := trie_db.recover_node_val(left_single_node)
			if recover_err != nil {
				return errors.New("recover_node_val err, key:" + string(left_single_node.FullPath()) + ", err:" + recover_err.Error())
			}
			//mark dirty
			left_single_node.mark_dirty()

		} else {
			//more then 1 node in parent nodes nothing to do
			node.mark_dirty()
		}

	} else {
		//condition child_nodes !=nil && node.val==nil
		//check if child_nodes has only one node => replace it to this node
		if len(node.child_nodes.path_index) == 1 {

			var single_child_node *Node
			for _, c_n := range node.child_nodes.path_index {
				single_child_node = c_n
				break
			}

			//replace
			node.parent_nodes.path_index[node.path[0]] = single_child_node
			single_child_node.path = append(node.path, single_child_node.path...)
			single_child_node.full_path_cache = nil
			single_child_node.parent_nodes = node.parent_nodes

			//because path changes, recover val is required
			recover_err := trie_db.recover_node_val(single_child_node)
			if recover_err != nil {
				return errors.New("recover_node_val err, key:" + string(single_child_node.FullPath()) + ", err:" + recover_err.Error())
			}

			//path changes mark dirty
			single_child_node.mark_dirty()
		} else {
			node.val = nil
			node.val_hash_recovered = true
			node.mark_dirty()
		}

	}

	return nil
}

// len(left_path) is >0
func (trie_db *TrieDB) update_target_nodes(target_nodes *nodes, left_path []byte, val []byte) (*Node, error) {

	/////// target the next node

	next_target_node := target_nodes.path_index[left_path[0]]
	if next_target_node != nil {
		return trie_db.update_target_node(next_target_node, left_path, val)
	}

	////// no common first byte
	if val == nil {
		//nothing todo
		return nil, nil
	} else {
		//simply add a new node
		new_node := &Node{
			path:                       left_path,
			parent_nodes:               target_nodes,
			val:                        val,
			dirty:                      true,
			child_nodes_hash_recovered: true,
			node_hash_recovered:        true,
			val_hash_recovered:         true,
		}

		//
		target_nodes.path_index[left_path[0]] = new_node
		//mark dirty
		new_node.mark_dirty()
		return new_node, nil
	}
}

// left_path has at least one byte same compared with target_node
func (trie_db *TrieDB) update_target_node(target_node *Node, left_path []byte, val []byte) (*Node, error) {

	r_err := trie_db.recover_node(target_node)
	if r_err != nil {
		return nil, errors.New("update_target_node recover_node err, " + r_err.Error())
	}

	//target exactly
	if bytes.Equal(target_node.path, left_path) {

		if val == nil {
			//del this node
			update_r_del_err := trie_db.update_recursive_del(target_node)
			if update_r_del_err != nil {
				return nil, update_r_del_err
			} else {
				return nil, nil
			}
		} else {
			//update this node
			target_node.val = val
			//to prevent recover again later inside calhash process
			target_node.val_hash_recovered = true
			//mark dirty
			target_node.mark_dirty()
			//
			return target_node, nil
		}

	} else if (len(left_path) > len(target_node.path)) && bytes.Equal(target_node.path, left_path[0:len(target_node.path)]) {
		// left_path start with target_node.path

		if !target_node.child_nodes_hash_recovered {
			recover_err := trie_db.recover_child_nodes(target_node)
			if recover_err != nil {
				return nil, errors.New("update_target_node recover_child_nodes err:" + recover_err.Error())
			}
		}

		if target_node.child_nodes != nil {
			///
			return trie_db.update_target_nodes(target_node.child_nodes, left_path[len(target_node.path):], val)
		} else {

			//nothing to do
			if val == nil {
				return nil, nil
			}

			// new nodes that dynamically created
			target_node.child_nodes = &nodes{
				path_index:  make(map[byte]*Node),
				parent_node: target_node,
				dirty:       true,
			}
			//first call child_nodes.mark_dirty() which is required
			target_node.child_nodes.mark_dirty()

			//
			return trie_db.update_target_nodes(target_node.child_nodes, left_path[len(target_node.path):], val)
		}

	} else if (len(left_path) < len(target_node.path)) && bytes.Equal(left_path, target_node.path[0:len(left_path)]) {
		// target_node.path start with left_path

		// nothing to do with nil (del action)
		if val == nil {
			return nil, nil
		}

		new_node := &Node{
			path:                       left_path[:],
			parent_nodes:               (*target_node).parent_nodes,
			child_nodes:                nil,
			val:                        val,
			dirty:                      true,
			child_nodes_hash_recovered: true,
			val_hash_recovered:         true,
			node_hash_recovered:        true,
		}

		//
		new_node.child_nodes = &nodes{
			path_index:  make(map[byte]*Node),
			parent_node: new_node,
			dirty:       true,
		}

		target_node.path = target_node.path[len(left_path):]
		target_node.parent_nodes.path_index[new_node.path[0]] = new_node
		new_node.parent_nodes = target_node.parent_nodes
		target_node.parent_nodes = new_node.child_nodes
		new_node.child_nodes.path_index[target_node.path[0]] = target_node
		//because path changes, recover val is required
		recover_err := trie_db.recover_node_val(target_node)
		if recover_err != nil {
			return nil, errors.New("recover_node_val err, key:" + string(target_node.FullPath()) + ", err:" + recover_err.Error())
		}

		//mark dirty
		target_node.dirty = true
		new_node.parent_nodes.mark_dirty()

		//
		return new_node, nil

	} else {

		// target_node.path, left_path, they have common prefix path
		// nothing to do with nil (del action)
		if val == nil {
			return nil, nil
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

		new_parent_node := Node{
			path:                       common_prefix_bytes[:],
			parent_nodes:               (*target_node).parent_nodes,
			child_nodes:                nil,
			child_nodes_hash_recovered: true,
			val_hash_recovered:         true,
			node_hash_recovered:        true,
			dirty:                      true,
		}

		new_parent_node.child_nodes = &nodes{
			path_index:  make(map[byte]*Node),
			parent_node: &new_parent_node,
			dirty:       true,
		}

		new_node := Node{
			path:                       left_path[common_prefix_bytes_len:],
			parent_nodes:               new_parent_node.child_nodes,
			val:                        val,
			dirty:                      true,
			child_nodes_hash_recovered: true,
			val_hash_recovered:         true,
			node_hash_recovered:        true,
		}

		new_parent_node.child_nodes.path_index[left_path[common_prefix_bytes_len]] = &new_node
		new_parent_node.child_nodes.path_index[target_node.path[common_prefix_bytes_len]] = target_node

		target_node.path = target_node.path[common_prefix_bytes_len:]
		target_node.parent_nodes.path_index[new_parent_node.path[0]] = &new_parent_node
		new_parent_node.parent_nodes = target_node.parent_nodes
		target_node.parent_nodes = new_parent_node.child_nodes

		//because path changes, recover val is required
		recover_node_val_err := trie_db.recover_node_val(target_node)
		if recover_node_val_err != nil {
			return nil, recover_node_val_err
		}
		//mark dirty
		target_node.dirty = true
		new_parent_node.parent_nodes.mark_dirty()

		return &new_node, nil
	}

}

// full_path len !=0 and <= PATH_LEN_LIMIT is required
// val == nil stands for del
// return error may be caused by kvdb io as get reading may happen inside update
func (trie_db *TrieDB) update_(full_path []byte, val []byte) (*Node, error) {

	if len(full_path) == 0 || len(full_path) > PATH_LEN_LIMIT {
		return nil, errors.New("full_path len err")
	}

	trie_db.lock.Lock()
	defer trie_db.lock.Unlock()

	return trie_db.update_target_node(trie_db.root_node, full_path, val)
}

// update the target and return the updated related Iterator
func (trie_db *TrieDB) Update(full_path []byte, val []byte) error {

	if len(val) == 0 {
		return errors.New("update val empty")
	}

	if len(val) > trie_db.config.Update_val_len_limit {
		return errors.New("trie val size over limit")
	}

	_, update_err := trie_db.update_(full_path, val)
	if update_err != nil {
		return update_err
	}

	return nil
}

func (trie_db *TrieDB) Delete(full_path []byte) error {
	_, del_err := trie_db.update_(full_path, nil)
	return del_err
}

//////////////////////////////////GET/////////////////////////////////////////

func (trie_db *TrieDB) get_recursive(target_node *Node, left_path []byte) (*Node, error) {

	//
	left_path_len := len(left_path)
	//
	if target_node == nil || left_path_len == 0 {
		return nil, errors.New("get_recursive err, target_node == nil || len(left_path) == 0")
	}
	//
	if len(target_node.path) > left_path_len {
		return nil, nil
	}
	//
	r_err := trie_db.recover_node(target_node)
	if r_err != nil {
		return nil, errors.New("recover_node err," + r_err.Error())
	}
	//
	if bytes.Equal(target_node.path, left_path) {
		r_err := trie_db.recover_node_val(target_node)
		if r_err != nil {
			return nil, errors.New("recover_node_val err," + r_err.Error())
		}
		return target_node, nil
	}

	//may be root node
	if target_node.parent_nodes == nil || bytes.Equal(target_node.path, left_path[0:len(target_node.path)]) {

		//check child nodes
		r_err := trie_db.recover_child_nodes(target_node)
		if r_err != nil {
			return nil, errors.New("recover_child_nodes err," + r_err.Error())
		}
		//not find
		if target_node.child_nodes == nil {
			return nil, nil
		}

		next_left_path := left_path[len(target_node.path):] //0 len for root node path
		next_target_node := target_node.child_nodes.path_index[next_left_path[0]]
		if next_target_node == nil {
			return nil, nil
		} else {
			return trie_db.get_recursive(next_target_node, next_left_path)
		}
	} else {
		return nil, nil
	}

}

// 1.get from internal nodes which is the lastest val(dirty or not dirty val)
// 2.get from cache
// 3.get from kvdb
func (trie_db *TrieDB) Get(full_path []byte) ([]byte, error) {

	//should never exsit related value
	if len(full_path) == 0 || len(full_path) > PATH_LEN_LIMIT {
		return nil, nil
	}

	trie_db.lock.Lock()
	defer trie_db.lock.Unlock()
	//
	get_node, get_err := trie_db.get_recursive(trie_db.root_node, full_path)
	//
	if get_err != nil {
		return nil, get_err
	}

	if get_node == nil {
		return nil, nil
	} else {
		return get_node.val, nil
	}

}

////////////////////////commit ///////////////////////////////

// k_v_map to collected all the k_v , string(key) => []byte(value)
func (trie_db *TrieDB) cal_hash_recursive(node *Node, k_v_map *sync.Map) (*hash.Hash, error) {

	//root node and empty trie check
	if node.parent_nodes == nil && node.child_nodes_hash == nil && (node.child_nodes == nil || len(node.child_nodes.path_index) == 0) {
		node.node_hash = hash.NIL_HASH
		return hash.NIL_HASH, nil
	}

	//
	if node.child_nodes != nil && len(node.child_nodes.path_index) != 0 {

		child_result_chan := make(chan error, len(node.child_nodes.path_index))

		for _, c_n := range node.child_nodes.path_index {
			go func(cn *Node) {
				trie_db.commit_thread_available <- struct{}{}
				_, cn_h_err := trie_db.cal_hash_recursive(cn, k_v_map)
				child_result_chan <- cn_h_err
				<-trie_db.commit_thread_available
			}(c_n)
		}

		//
		<-trie_db.commit_thread_available //give out a thread-slot

		//make sure all sub-thread done
		for range node.child_nodes.path_index {
			cn_err := <-child_result_chan
			if cn_err != nil {
				return nil, cn_err
			}
		}
		//
		trie_db.commit_thread_available <- struct{}{} //get back the thread-slot
	}

	//
	if node.dirty {
		//
		if node.child_nodes != nil && node.child_nodes.dirty {
			node.child_nodes.serialize()
			node.child_nodes.cal_nodes_hash()
		}
		//if dirty maybe cause by path change ,so val_hash has to be recalculated
		r_err := trie_db.recover_node_val(node)
		if r_err != nil {
			return nil, errors.New("cal_hash_recursive recover_node_val err, " + r_err.Error())
		}

		//cal val hash
		if node.val != nil {
			node.cal_node_val_hash()
		} else {
			node.val_hash = nil
		}

		//cal node hash
		node.serialize()
		node.cal_node_hash()
	}

	//////////////// store all related //////////////////////////
	if node.val_hash != nil {
		k_v_map.Store(string(node.val_hash.Bytes()), node.val)
	}

	if node.child_nodes_hash != nil {
		if node.child_nodes != nil {
			k_v_map.Store(string(node.child_nodes_hash.Bytes()), node.child_nodes.nodes_bytes)
		} else {
			//may caused by no loading because of lazy loading feature
			k_v_map.Store(string(node.child_nodes_hash.Bytes()), []byte{})
		}
	}

	k_v_map.Store(string(node.node_hash.Bytes()), node.node_bytes)

	return node.node_hash, nil
}

// return root_hash, update/insert hash map, del hash map ,error
func (trie_db *TrieDB) CalHash() (*hash.Hash, map[string][]byte, map[string]*hash.Hash, error) {
	trie_db.lock.Lock()
	defer trie_db.lock.Unlock()

	//
	all_trie_k_v := sync.Map{}
	//
	trie_db.commit_thread_available <- struct{}{} //main thread-slot
	//
	_, cal_hash_err := trie_db.cal_hash_recursive(trie_db.root_node, &all_trie_k_v)
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

		value_bytes := value.([]byte)
		//because of lazy load len may be 0, e.g : lazy load of val_hash
		if len(value_bytes) != 0 {
			if _, exist := trie_db.attached_hash[key.(string)]; !exist {
				update_k_v[key.(string)] = value_bytes
			}
		}
		return true
	})

	return trie_db.root_node.node_hash, update_k_v, del_k_v, nil
}
