package triedb

import (
	"encoding/binary"

	"github.com/xlander-io/btree"
	"github.com/xlander-io/hash"
)

func newNodesPathBTree() *btree.BTree {
	return btree.New(PATH_B_TREE_DEGREE, func(a, b interface{}) bool {
		return uint8(a.(uint8)) < uint8(b.(uint8))
	})
}

type Node struct {
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
	path_index  map[byte]*Node //byte can only ranges from '0' to '255' total 16 different values
	parent_node *Node
	nodes_bytes []byte //serialize(self) , nil for new node or dirty node
	dirty       bool   //default false
}

// serialize to nodes_bytes
func (n *nodes) serialize() {

	var result []byte = []byte{}

	result = append(result, uint8(len(n.path_index)))

	for _, node := range n.path_index {
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
		n.path_index = make(map[byte]*Node)

		for i := 0; i < path_index_len; i++ {
			node_ := Node{}
			node_.node_hash = hash.NewHashFromBytes(n.nodes_bytes[deserialize_offset : deserialize_offset+32])
			deserialize_offset += 32
			path_len := int(binary.LittleEndian.Uint16(n.nodes_bytes[deserialize_offset : deserialize_offset+16]))
			deserialize_offset += 16
			node_.path = n.nodes_bytes[deserialize_offset : deserialize_offset+path_len]
			deserialize_offset += path_len
			n.path_index[node_.path[0]] = &node_
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
