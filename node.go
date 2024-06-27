package triedb

import (
	"encoding/binary"

	"github.com/xlander-io/triedb/util"
)

type Node struct {
	full_path []byte //nil for root node
	path      []byte //nil for root node

	parent_nodes *Nodes //nil for root node

	child_nodes      *Nodes
	child_nodes_hash *util.Hash //nil for new node or dirty node

	val      []byte
	val_hash *util.Hash

	node_bytes []byte     //serialize(self) , nil for new node or dirty node
	node_hash  *util.Hash //hash(self.node_bytes) , nil for new node or dirty node
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

// calculate node_hash
func (n *Node) cal_node_hash(prefix []byte) {
	result := []byte{}
	result = append(result, prefix...)
	result = append(result, n.full_path...)
	result = append(result, n.node_bytes...)
	n.node_hash = util.NewHashFromBytes(result)
}

// calculate val_hash
func (n *Node) cal_val_hash(prefix []byte) {
	result := []byte{}
	result = append(result, prefix...)
	result = append(result, n.full_path...)
	result = append(result, n.val...)
	n.val_hash = util.NewHashFromBytes(result)
}

type Nodes struct {
	full_path  []byte
	path_index map[byte]*Node //byte can only ranges from '0' to 'f' total 16 different values

	parent_node *Node

	nodes_bytes []byte     //serialize(self) , nil for new node or dirty node
	nodes_hash  *util.Hash //hash(self) , nil for new node or dirty node
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
func (n *Nodes) cal_nodes_hash(prefix []byte) {
	result := []byte{}
	result = append(result, prefix...)
	result = append(result, n.full_path...)
	result = append(result, n.nodes_bytes...)
	n.nodes_hash = util.NewHashFromBytes(result)
}
