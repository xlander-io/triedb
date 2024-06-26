package triedb

import "encoding/binary"

const PATH_LEN_LIMIT = 64 * 1024         // in bytes uint16 limit
const VAL_LEN_LIMIT = 4096 * 1024 * 1024 // in bytes uint32 limit

type Node struct {
	Full_path        []byte //nil for root node
	Path             []byte //nil for root node
	Parent_nodes     *Nodes //nil for root node
	Child_nodes      *Nodes
	Child_nodes_hash *Hash //nil for new node or dirty node
	Val              []byte
	Hash             *Hash //hash(self) , nil for new node or dirty node
}

func (n *Node) serialize() (result []byte) {

	//child_nodes_hash
	if n.Child_nodes_hash == nil {
		result = append(result, uint8(0))
	} else {
		result = append(result, uint8(32))
		result = append(result, (*n.Child_nodes_hash)[:]...)
	}
	//full_path
	if n.Full_path == nil {
		result = append(result, ([]byte{0, 0})...)
	} else {
		f_p_len_bytes := make([]byte, 16)
		binary.LittleEndian.PutUint16(f_p_len_bytes, uint16(len(n.Full_path)))
		result = append(result, f_p_len_bytes...)
	}
	//val
	if len(n.Val) == 0 {
		result = append(result, ([]byte{0, 0, 0, 0})...)
	} else {
		val_len_bytes := make([]byte, 32)
		binary.LittleEndian.PutUint32(val_len_bytes, uint32(len(n.Val)))
		result = append(result, val_len_bytes...)
	}

	return
}

// func (n *Node) deserialize() {

// }

type Nodes struct {
	Path_index map[byte]*Node // byte can only ranges from '0' to 'f' total 16 different values
	Hash       *Hash          //hash(self) , nil for new node or dirty node
}

func (n *Nodes) serialize() (result []byte) {
	return nil
}

type TrieDB struct {
	Root_hash *Hash
}

func NewTrieDB(root_hash *Hash) {

}
