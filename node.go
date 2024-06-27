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
	for _, node := range n.path_index {
		result = append(result, (*node.node_hash)[:]...)
		path_len_bytes := make([]byte, 16)
		binary.LittleEndian.PutUint16(path_len_bytes, uint16(len(node.path)))
		result = append(result, path_len_bytes...)
		result = append(result, node.path...)
	}
	n.nodes_bytes = result
}

// calculate nodes_hash
func (n *Nodes) cal_nodes_hash(prefix []byte) {
	result := []byte{}
	result = append(result, prefix...)
	result = append(result, n.full_path...)
	result = append(result, n.nodes_bytes...)
	n.nodes_hash = util.NewHashFromBytes(result)
}
