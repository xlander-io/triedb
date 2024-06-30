package triedb

import (
	"encoding/binary"

	"github.com/xlander-io/triedb/util"
)

var HASH_NODE_PREFIX []byte = []byte("hash_node_prefix")
var HASH_NODE_VAL_PREFIX []byte = []byte("hash_node_val_prefix")
var HASH_NODES_PREFIX []byte = []byte("hash_nodes_prefix")

type Node struct {
	path []byte //nil for root node

	parent_nodes *Nodes //nil for root node

	child_nodes      *Nodes
	child_nodes_hash *util.Hash //nil for new node or dirty node

	val      []byte     //always nil for root node
	val_hash *util.Hash //always nil for root node

	node_bytes []byte     //serialize(self) , nil for new node or dirty node
	node_hash  *util.Hash //hash(self.node_bytes) , nil for new node or dirty node

	dirty bool //default false

}

// serialize to node_bytes
func (n *Node) serialize() {
	var result []byte = []byte{}
	//child_nodes_hash
	if n.child_nodes_hash == nil {
		result = append(result, uint8(0))
	} else {
		result = append(result, uint8(32))
		result = append(result, (*n.child_nodes_hash)[:]...)
	}

	//val_hash
	if n.val_hash == nil {
		result = append(result, uint8(0))
	} else {
		result = append(result, uint8(32))
		result = append(result, (*n.val_hash)[:]...)
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
	} else if child_nodes_hash_len == 32 {
		n.child_nodes_hash = util.NewHashFromBytes(n.node_bytes[deserialize_offset : deserialize_offset+32])
		deserialize_offset += 32
	} else {
		//remove this else in final production
		panic("node deserialize child_nodes_hash err")
	}

	//val_hash
	val_hash_len := uint8(n.node_bytes[deserialize_offset])
	deserialize_offset++
	if val_hash_len == 0 {
		n.val_hash = nil
	} else if val_hash_len == 32 {
		n.val_hash = util.NewHashFromBytes(n.node_bytes[deserialize_offset : deserialize_offset+32])
		deserialize_offset += 32
	} else {
		//remove this else in final production
		panic("node deserialize val_hash err")
	}

}

// return nil for root empty node
func (n *Node) get_full_path() []byte {
	if n.path == nil {
		//may happen in root node
		return nil
	}
	//
	full_path := n.path
	parent_nodes := n.parent_nodes

	//when parent_nodes != nil always -> parent_nodes.parent_node !=nil
	for parent_nodes != nil && parent_nodes.parent_node.path != nil {
		full_path = append(parent_nodes.parent_node.path, full_path...)
	}

	return full_path
}

// calculate val_hash
func (n *Node) cal_node_val_hash() {
	result := []byte{}
	result = append(result, HASH_NODE_VAL_PREFIX...)
	result = append(result, n.get_full_path()...)
	result = append(result, n.val...)
	n.val_hash = util.NewHashFromBytes(result)
}

// calculate node_hash
func (n *Node) cal_node_hash() {
	result := []byte{}
	result = append(result, HASH_NODE_PREFIX...)
	result = append(result, n.get_full_path()...)
	result = append(result, n.node_bytes...)
	n.node_hash = util.NewHashFromBytes(result)
}

// later will recalculate related value
func (node *Node) mark_dirty() {
	node.dirty = true
	if node.parent_nodes != nil {
		node.parent_nodes.mark_dirty()
	}
	//node.node_bytes = nil
	//node.node_hash = nil
}

type Nodes struct {
	path_index  map[byte]*Node //byte can only ranges from '0' to 'f' total 16 different values
	parent_node *Node
	nodes_bytes []byte     //serialize(self) , nil for new node or dirty node
	nodes_hash  *util.Hash //hash(self) , nil for new node or dirty node
	dirty       bool       //default false
}

// serialize to nodes_bytes
func (n *Nodes) serialize() {
	var result []byte = []byte{}
	result = append(result, uint8(len(n.path_index)))
	for _, node := range n.path_index {
		result = append(result, (*node.node_hash)[:]...)
		path_len_bytes := make([]byte, 16)
		binary.LittleEndian.PutUint16(path_len_bytes, uint16(len(node.path)))
		result = append(result, path_len_bytes...)
		result = append(result, node.path...)
	}
	n.nodes_bytes = result
}

func (n *Nodes) deserialize() {
	deserialize_offset := 0
	path_index_len := int(uint8(n.nodes_bytes[0]))
	deserialize_offset++
	if path_index_len == 0 {
		n.path_index = make(map[byte]*Node)
	} else {
		for i := 0; i < path_index_len; i++ {
			node_ := Node{}
			node_.node_hash = util.NewHashFromBytes(n.nodes_bytes[deserialize_offset : deserialize_offset+32])
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
func (n *Nodes) cal_nodes_hash() {
	result := []byte{}
	result = append(result, HASH_NODES_PREFIX...)
	result = append(result, n.parent_node.get_full_path()...)
	result = append(result, n.nodes_bytes...)
	n.nodes_hash = util.NewHashFromBytes(result)
}

// later will recalculate related value
func (n *Nodes) mark_dirty() {

	n.dirty = true
	if n.parent_node != nil {
		n.parent_node.mark_dirty()
	}

	//n.nodes_bytes = nil
	//n.nodes_hash = nil
}
