package triedb

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/xlander-io/btree"
	"github.com/xlander-io/hash"
)

func newNodesPathBTree() *btree.BTree {
	return btree.New(PATH_B_TREE_DEGREE, func(a, b interface{}) bool {
		return uint8(a.(uint8)) < uint8(b.(uint8))
	})
}

type Node struct {
	triedb *TrieDB

	path []byte //nil for root node

	full_path_cache []byte

	parent_nodes *nodes //nil for root node

	child_nodes                *nodes
	child_nodes_hash           *hash.Hash //nil for new node or dirty node
	child_nodes_hash_recovered bool

	val                []byte     //always nil for root node
	val_hash           *hash.Hash //always nil for root node
	val_hash_recovered bool

	node_bytes          []byte     //serialize(self) , nil for new node or dirty node
	node_hash           *hash.Hash //hash(self.node_bytes) , nil for new node or dirty node
	node_hash_recovered bool

	dirty bool //default false
}

func (n *Node) Val() []byte {
	return n.val
}

func (n *Node) Hash() *hash.Hash {
	return n.node_hash
}

func (n *Node) Path() []byte {
	return n.path
}

// serialize to node_bytes
func (n *Node) serialize() {
	var result []byte = []byte{}
	//child_nodes_hash
	if n.child_nodes_hash == nil {
		result = append(result, uint8(0))
	} else {
		result = append(result, uint8(32))
		result = n.child_nodes_hash.PrePend(result)
	}

	//val_hash
	if n.val_hash == nil {
		result = append(result, uint8(0))
	} else {
		result = append(result, uint8(32))
		result = n.val_hash.PrePend(result)
	}

	n.node_bytes = result
}

func (n *Node) deserialize() {

	deserialize_offset := 0
	//child_nodes_hash
	child_nodes_hash_len := uint8(n.node_bytes[deserialize_offset])
	deserialize_offset++
	if child_nodes_hash_len == 0 {
		n.child_nodes_hash = nil
	} else {
		// child_nodes_hash_len == 32
		n.child_nodes_hash = hash.NewHashFromBytes(n.node_bytes[deserialize_offset : deserialize_offset+32])
		deserialize_offset += 32
	}

	//val_hash
	val_hash_len := uint8(n.node_bytes[deserialize_offset])
	deserialize_offset++
	if val_hash_len == 0 {
		n.val_hash = nil
	} else {
		//val_hash_len == 32
		n.val_hash = hash.NewHashFromBytes(n.node_bytes[deserialize_offset : deserialize_offset+32])
		deserialize_offset += 32
	}

}

// return nil for root empty node
func (n *Node) FullPath() []byte {

	//
	if n.full_path_cache != nil {
		return n.full_path_cache
	}
	//
	if n.parent_nodes == nil {
		//may happen in root node
		return []byte{}
	}

	return append(n.parent_nodes.parent_node.FullPath(), n.path...)
}

// calculate val_hash
func (n *Node) cal_node_val_hash() {
	result := []byte{}
	result = append(result, HASH_NODE_VAL_PREFIX...)
	result = append(result, n.FullPath()...)
	result = append(result, n.val...)
	n.val_hash = hash.CalHash(result)
}

// calculate node_hash
func (n *Node) cal_node_hash() {
	result := []byte{}
	result = append(result, HASH_NODE_PREFIX...)
	result = append(result, n.FullPath()...)
	result = append(result, n.node_bytes...)
	n.node_hash = hash.CalHash(result)
}

// later will recalculate related value
func (node *Node) mark_dirty() {
	node.dirty = true

	if node.parent_nodes != nil && !node.parent_nodes.dirty {
		node.parent_nodes.mark_dirty()
	}
}

type nodes struct {
	path_btree  *btree.BTree //byte can only ranges from '0' to '255' total 16 different values
	parent_node *Node
	nodes_bytes []byte //serialize(self) , nil for new node or dirty node
	dirty       bool   //default false
}

// serialize to nodes_bytes
func (n *nodes) serialize() {

	var result []byte = []byte{}

	result = append(result, uint8(n.path_btree.Len()))

	iter := n.path_btree.Before(uint8(0))
	for iter.Next() {
		node := iter.Value.(*Node)
		result = node.node_hash.PrePend(result)
		path_len_bytes := make([]byte, 16)
		binary.LittleEndian.PutUint16(path_len_bytes, uint16(len(node.path)))
		result = append(result, path_len_bytes...)
		result = append(result, node.path...)
	}

	n.nodes_bytes = result
}

func (n *nodes) deserialize() {

	deserialize_offset := 0
	path_index_len := int(uint8(n.nodes_bytes[0]))
	deserialize_offset++
	if path_index_len != 0 {
		n.path_btree = newNodesPathBTree()
		for i := 0; i < path_index_len; i++ {
			node_ := Node{
				triedb: n.parent_node.triedb,
			}
			node_.node_hash = hash.NewHashFromBytes(n.nodes_bytes[deserialize_offset : deserialize_offset+32])
			deserialize_offset += 32
			path_len := int(binary.LittleEndian.Uint16(n.nodes_bytes[deserialize_offset : deserialize_offset+16]))
			deserialize_offset += 16
			node_.path = n.nodes_bytes[deserialize_offset : deserialize_offset+path_len]
			deserialize_offset += path_len
			n.path_btree.Set(uint8(node_.path[0]), &node_)
		}

	}
}

// calculate nodes_hash
func (n *nodes) cal_nodes_hash() {
	result := []byte{}
	result = append(result, HASH_NODES_PREFIX...)
	result = append(result, n.parent_node.FullPath()...)
	result = append(result, n.nodes_bytes...)
	n.parent_node.child_nodes_hash = hash.CalHash(result)
}

// later will recalculate related value
func (n *nodes) mark_dirty() {
	n.dirty = true

	if n.parent_node != nil && !n.parent_node.dirty {
		n.parent_node.mark_dirty()
	}
}

/////////

func (node *Node) recover_node() error {

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

	node_bytes, node_err := node.triedb.getFromCacheKVDB(node.node_hash.Bytes())
	if node_err != nil {
		return errors.New("recover_node getFromCacheKVDB  err, node_hash: " + fmt.Sprintf("%x", node.node_hash.Bytes()))
	}

	if node_bytes == nil {
		return errors.New("recover_node getFromCacheKVDB  err, node_hash not found")
	}
	//
	node.node_bytes = node_bytes
	node.deserialize()

	node.triedb.attachHash(node.node_hash)
	node.triedb.attachHash(node.val_hash)         //don't froget val_hash as it may be affected during direct del action
	node.triedb.attachHash(node.child_nodes_hash) //better put this

	//
	return nil
}

func (node *Node) recover_node_val() error {

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

	node_val_bytes, node_val_err := node.triedb.getFromCacheKVDB(node.val_hash.Bytes())
	if node_val_err != nil {
		return errors.New("recover_node_val getFromCacheKVDB  err, " + node_val_err.Error())
	}

	//
	node.val = node_val_bytes

	//
	return nil
}

func (node *Node) recover_child_nodes() error {

	if node == nil {
		return errors.New("recover_child_nodes err, node nil")
	}

	defer func() {
		//delete to prevent double recover
		node.child_nodes_hash_recovered = true
	}()

	if node.child_nodes == nil && !node.child_nodes_hash_recovered && node.child_nodes_hash != nil {

		nodes_bytes, err := node.triedb.getFromCacheKVDB(node.child_nodes_hash.Bytes())
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

		path_b_iter := child_nodes_.path_btree.Before(uint8(0))
		for path_b_iter.Next() {
			c_n := path_b_iter.Value.(*Node)
			node.triedb.attachHash(c_n.node_hash)
			c_n.parent_nodes = &child_nodes_
		}

		//
		node.child_nodes = &child_nodes_

	}

	//
	return nil
}

//////////////////// Neibour////////////////////////////////////

// return the right neibour node in the same nodes container
// val not related
func (node *Node) right_node() (*Node, error) {
	//
	if node == nil || node.parent_nodes == nil {
		return nil, nil
	}
	//
	if node.parent_nodes.path_btree == nil || node.parent_nodes.path_btree.Len() <= 1 {
		return nil, nil
	}
	//
	iter := node.parent_nodes.path_btree.Before(uint8(node.path[0]))
	if iter == nil {
		return nil, nil
	}
	//
	iter.Next()
	if iter.Next() {
		return iter.Value.(*Node), nil
	}
	//
	return nil, nil
}

// return the upper right neibour node
// val not related
func (node *Node) upper_right_node() (*Node, error) {
	//
	if node == nil || node.parent_nodes == nil ||
		node.parent_nodes.parent_node.parent_nodes == nil ||
		node.parent_nodes.parent_node.parent_nodes.path_btree.Len() <= 1 {
		return nil, nil
	}

	//
	iter := node.parent_nodes.parent_node.parent_nodes.path_btree.Before(uint8(node.parent_nodes.parent_node.path[0]))
	if iter == nil {
		return nil, nil
	}
	//
	iter.Next()
	if iter.Next() {
		return iter.Value.(*Node), nil
	}
	//
	return nil, nil
}

// return the parentNode whose val must not be nil
func (node *Node) ParentNode() (*Node, error) {

	var target_node *Node = node

	for {
		if target_node == nil || target_node.parent_nodes == nil || target_node.parent_nodes.parent_node == nil {
			return nil, nil
		}

		if target_node.parent_nodes.parent_node.val != nil {
			return target_node.parent_nodes.parent_node, nil
		}

		if target_node.parent_nodes.parent_node.val_hash == nil {
			target_node = target_node.parent_nodes.parent_node
		} else {
			recover_node_val_err := target_node.parent_nodes.parent_node.recover_node_val()
			if recover_node_val_err != nil {
				return nil, recover_node_val_err
			}
			return target_node.parent_nodes.parent_node, nil
		}
	}

}
