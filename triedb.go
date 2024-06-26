package triedb

type Nodes struct {
	Path_index map[byte]*Node // byte can only ranges from '0' to 'f' total 16 different values
	Hash       *Hash          //hash(self) , nil for new node or dirty node
}

type Node struct {
	Full_path  []byte //nil for root node
	Path       []byte //nil for root node
	Nodes      *Nodes
	Nodes_hash *Hash //nil for new node or dirty node
	Val        []byte
	Hash       *Hash //hash(self) , nil for new node or dirty node
}

func (n *Node) serialize() (result []byte) {
	//
}

// func (n *Node) deserialize() {

// }

type TrieDB struct {
	Root_hash *Hash
}

func NewTrieDB(root_hash *Hash) {

}
