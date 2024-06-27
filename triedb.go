package triedb

import (
	"encoding/binary"

	"github.com/xlander-io/triedb/util"
)

const PATH_LEN_LIMIT = 64 * 1024         // in bytes uint16 limit
const VAL_LEN_LIMIT = 4096 * 1024 * 1024 // in bytes uint32 limit

type Node struct {
	full_path        []byte //nil for root node
	path             []byte //nil for root node
	parent_nodes     *Nodes //nil for root node
	child_nodes      *Nodes
	child_nodes_hash *util.Hash //nil for new node or dirty node
	val              []byte
	node_bytes       []byte     //serialize(self) , nil for new node or dirty node
	node_hash        *util.Hash //hash(self.node_bytes) , nil for new node or dirty node
}

func (n *Node) serialize() {
	var result []byte = []byte{}
	//child_nodes_hash
	if n.child_nodes_hash == nil {
		result = append(result, uint8(0))
	} else {
		result = append(result, uint8(32))
		result = append(result, (*n.child_nodes_hash)[:]...)
	}

	//val
	if len(n.val) == 0 {
		result = append(result, ([]byte{0, 0, 0, 0})...)
	} else {
		val_len_bytes := make([]byte, 32)
		binary.LittleEndian.PutUint32(val_len_bytes, uint32(len(n.val)))
		result = append(result, val_len_bytes...)
	}

	n.node_bytes = result
}

func (n *Node) deserialize() {

}

func (n *Node) hash(prefix []byte) {
	result := []byte{}
	result = append(result, prefix...)
	result = append(result, n.full_path...)
	result = append(result, n.node_bytes...)
	n.node_hash = util.NewHashFromBytes(result)
}

type Nodes struct {
	full_path  []byte
	path_index map[byte]*Node //byte can only ranges from '0' to 'f' total 16 different values
	node_bytes []byte         //serialize(self) , nil for new node or dirty node
	nodes_hash *util.Hash     //hash(self) , nil for new node or dirty node
}

func (n *Nodes) serialize() {
	var result []byte = []byte{}
	for _, node := range n.path_index {
		result = append(result, (*node.node_hash)[:]...)
		path_len_bytes := make([]byte, 16)
		binary.LittleEndian.PutUint16(path_len_bytes, uint16(len(node.path)))
		result = append(result, path_len_bytes...)
		result = append(result, node.path...)
	}
	n.node_bytes = result
}

func (n *Nodes) hash(prefix []byte) {
	result := []byte{}
	result = append(result, prefix...)
	result = append(result, n.full_path...)
	result = append(result, n.node_bytes...)
	n.nodes_hash = util.NewHashFromBytes(result)
}

type TrieDB struct {
	//Root_hash *util.Hash
}

func NewTrieDB(root_hash *util.Hash) {

}
