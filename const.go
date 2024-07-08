package triedb

// 65536= 2^16 , len can be put inside a uint16 , never change this
// setting a long path will decrease the speed of kvdb
const PATH_LEN_LIMIT = 65536
const PATH_B_TREE_DEGREE = 5

var HASH_NODE_PREFIX []byte = []byte("hash_node_prefix")
var HASH_NODE_VAL_PREFIX []byte = []byte("hash_node_val_prefix")
var HASH_NODES_PREFIX []byte = []byte("hash_nodes_prefix")

// max bytes limit of full path
func GetPathLenLimit() int {
	return PATH_LEN_LIMIT
}
