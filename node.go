package triedb

import (
	"encoding/binary"
	"strconv"

	"github.com/xlander-io/btree"
	"github.com/xlander-io/hash"
)

const NODES_B_TREE_DEGREE = 5

func new_nodes_btree() *btree.BTree {
	return btree.New(NODES_B_TREE_DEGREE, func(a, b interface{}) bool {
		return uint8(a.(uint8)) < uint8(b.(uint8))
	})
}

type Node struct {
	index_hash *hash.Hash

	prefix []byte //nil for root node

	parent_nodes *nodes //nil for root node

	prefix_child_nodes                *nodes
	prefix_child_nodes_hash           *hash.Hash //nil for new node or dirty node
	prefix_child_nodes_hash_recovered bool

	folder_child_nodes                *nodes
	folder_child_nodes_hash           *hash.Hash //nil for new node or dirty node
	folder_child_nodes_hash_recovered bool

	val                []byte     //always nil for root node
	val_hash           *hash.Hash //always nil for root node
	val_hash_recovered bool
	val_dirty          bool

	node_bytes []byte     //serialize(self) , nil for new node or dirty node
	node_hash  *hash.Hash //hash(self.node_bytes) , nil for new node or dirty node

	node_path_cache [][]byte

	dirty bool //default false
}

func (n *Node) has_folder_child() bool {
	return n.folder_child_nodes != nil || (!n.folder_child_nodes_hash_recovered && n.folder_child_nodes_hash != nil)
}

func (n *Node) has_prefix_child() bool {
	return n.prefix_child_nodes != nil || (!n.prefix_child_nodes_hash_recovered && n.prefix_child_nodes_hash != nil)
}

func (n *Node) has_val() bool {
	return n.val != nil || (!n.val_hash_recovered && n.val_hash != nil)
}

func (n *Node) encode_node_type() uint8 {
	// bitwise node_type : 	|0|0|0|has_folder_child_nodes_hash|has_prefix_child_nodes_hash|has_val_hash|has_hash_index|
	var node_type uint8 = 0

	if n.index_hash != nil {
		node_type += 1
	}

	if n.val_hash != nil {
		node_type += 2
	}

	if n.prefix_child_nodes_hash != nil {
		node_type += 4
	}

	if n.folder_child_nodes_hash != nil {
		node_type += 8
	}

	return node_type
}

// serialize to node_bytes
func (n *Node) serialize() {

	var result []byte = []byte{}

	//node_type
	result = append(result, n.encode_node_type())

	//index_hash
	if n.index_hash != nil {
		result = n.index_hash.PrePend(result)
	}

	//val_hash
	if n.val_hash != nil {
		result = n.val_hash.PrePend(result)
	}

	//child_nodes_hash
	if n.prefix_child_nodes_hash != nil {
		result = n.prefix_child_nodes_hash.PrePend(result)
	}

	if n.folder_child_nodes_hash != nil {
		result = n.folder_child_nodes_hash.PrePend(result)
	}

	n.node_bytes = result
}

func (n *Node) deserialize() {

	// bitwise node_type : 	|0|0|0|has_folder_child_nodes_hash|has_prefix_child_nodes_hash|has_val_hash|has_hash_index|
	node_type := uint8(n.node_bytes[0])
	offset := 1

	if node_type&1 != 0 {
		n.index_hash = hash.NewHashFromBytes(n.node_bytes[offset : offset+32])
		offset += 32
	}

	if node_type&2 != 0 {
		n.val_hash = hash.NewHashFromBytes(n.node_bytes[offset : offset+32])
		offset += 32
	}

	if node_type&4 != 0 {
		n.prefix_child_nodes_hash = hash.NewHashFromBytes(n.node_bytes[offset : offset+32])
		offset += 32
	}

	if node_type&8 != 0 {
		n.folder_child_nodes_hash = hash.NewHashFromBytes(n.node_bytes[offset : offset+32])
	}

}

func (n *Node) node_path() [][]byte {

	if n.parent_nodes == nil {
		return make([][]byte, 0)
	}

	if n.node_path_cache != nil {
		return n.node_path_cache
	}

	if n.parent_nodes.is_folder_child_nodes {
		n.node_path_cache = append(n.parent_nodes.parent_node.node_path(), n.prefix)
		return n.node_path_cache

	} else {

		//deep copy
		path_array := make([][]byte, len(n.parent_nodes.parent_node.node_path()))
		for n, path := range n.parent_nodes.parent_node.node_path() {
			path_array[n] = append([]byte{}, path...)
		}

		path_array[len(path_array)-1] = append(path_array[len(path_array)-1], n.prefix...)
		n.node_path_cache = path_array
		return n.node_path_cache
	}

}

func (n *Node) node_path_flat() []byte {
	node_path := n.node_path()
	node_path_flat := []byte{}

	for _, path := range node_path {
		path_len_bytes := make([]byte, 16)
		binary.LittleEndian.PutUint16(path_len_bytes, uint16(len(path)))
		node_path_flat = append(node_path_flat, path_len_bytes...)
		node_path_flat = append(node_path_flat, path...)
	}
	return node_path_flat
}

func (n *Node) node_path_flat_str() string {
	path_flat := n.node_path_flat()
	if len(path_flat) == 0 {
		return ""
	}

	result := ""
	offset := 0
	for {
		if len(path_flat) >= offset+16 {
			path_len := int(binary.LittleEndian.Uint16(path_flat[offset : offset+16]))
			result = result + strconv.Itoa(path_len) + string(path_flat[offset+16:offset+16+path_len])
			offset = offset + 16 + path_len
		} else {
			break
		}
	}

	return result
}

// later will recalculate related value
func (node *Node) mark_dirty() {
	//
	node.dirty = true
	//
	if node.parent_nodes != nil && !node.parent_nodes.dirty {
		node.parent_nodes.mark_dirty()
	}
}

type nodes struct {
	is_folder_child_nodes bool         //folder_child_nodes | prefix_child_nodes
	btree                 *btree.BTree //byte can only ranges from '0' to '255' total 16 different values
	parent_node           *Node
	nodes_bytes           []byte //serialize(self) , nil for new node or dirty node
	dirty                 bool   //default false
}

// serialize to nodes_bytes
func (n *nodes) serialize() {

	var result []byte = []byte{}

	result = append(result, uint8(n.btree.Len()))

	iter := n.btree.Before(uint8(0))
	for iter.Next() {
		node := iter.Value.(*Node)
		result = append(result, node.node_hash.Bytes()...)
		prefix_len_bytes := make([]byte, 16)
		binary.LittleEndian.PutUint16(prefix_len_bytes, uint16(len(node.prefix)))
		result = append(result, prefix_len_bytes...)
		result = append(result, node.prefix...)
		result = append(result, uint8(len(node.node_bytes)))
		result = append(result, node.node_bytes...)
	}

	n.nodes_bytes = result
}

func (n *nodes) deserialize() {

	deserialize_offset := 0
	prefix_index_len := int(uint8(n.nodes_bytes[0]))
	deserialize_offset++
	if prefix_index_len != 0 {

		n.btree = new_nodes_btree()

		for i := 0; i < prefix_index_len; i++ {
			node_ := Node{}
			//
			node_.node_hash = hash.NewHashFromBytes(n.nodes_bytes[deserialize_offset : deserialize_offset+32])
			deserialize_offset += 32
			//
			prefix_len := int(binary.LittleEndian.Uint16(n.nodes_bytes[deserialize_offset : deserialize_offset+16]))
			deserialize_offset += 16
			node_.prefix = n.nodes_bytes[deserialize_offset : deserialize_offset+prefix_len]
			deserialize_offset += prefix_len
			//
			node_bytes_len := int(uint8(n.nodes_bytes[deserialize_offset]))
			deserialize_offset++
			node_.node_bytes = n.nodes_bytes[deserialize_offset : deserialize_offset+node_bytes_len]
			node_.deserialize()
			deserialize_offset += node_bytes_len
			//
			n.btree.Set(uint8(node_.prefix[0]), &node_)
		}

	}
}

// later will recalculate related value
func (n *nodes) mark_dirty() {
	//
	n.dirty = true
	//
	if n.parent_node != nil && !n.parent_node.dirty {
		n.parent_node.mark_dirty()
	}
}
