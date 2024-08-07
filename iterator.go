package triedb

import (
	"bytes"
	"errors"
)

type Iterator struct {
	trie_db      *TrieDB
	parent_node  *Node
	current_node *Node
	next_node    *Node
}

func (trie_db *TrieDB) NewIterator(folder_full_path [][]byte) (*Iterator, error) {
	if !trie_db.config.Read_only {
		return nil, errors.New("iterator is only allowed for read only trie")
	}

	n, err := trie_db.get_(folder_full_path)
	if err != nil {
		return nil, errors.New("get folder err, " + err.Error())
	}

	if n == nil {
		return nil, errors.New("folder not exist")
	}

	recover_err := trie_db.recover_child_nodes(n, true, false)
	if recover_err != nil {
		return nil, recover_err
	}

	if n.folder_child_nodes == nil {
		return nil, errors.New("folder has no child")
	}

	_, first_c_n_i := n.folder_child_nodes.btree.Min()
	if first_c_n_i == nil {
		return nil, errors.New("folder 0 child ")
	}

	cursor_n := first_c_n_i.(*Node)
	if iter_check_node_hit(cursor_n) {
		return &Iterator{
			trie_db:      trie_db,
			parent_node:  n,
			current_node: cursor_n,
		}, nil
	}

	next_hit_node, err := iter_recursive_hit_next(trie_db, n, cursor_n)
	if err != nil {
		return nil, err
	}
	return &Iterator{
		trie_db:      trie_db,
		parent_node:  n,
		current_node: next_hit_node,
	}, nil

}

// n has val or n has folder child
func iter_check_node_hit(n *Node) bool {
	if n.has_val() || n.has_folder_child() {
		return true
	} else {
		return false
	}
}

func iter_recursive_hit_next(trie_db *TrieDB, iter_parent_node *Node, n *Node) (*Node, error) {

	//load prefix child
	recover_err := trie_db.recover_child_nodes(n, false, true)
	if recover_err != nil {
		return nil, recover_err
	}

	if n.prefix_child_nodes != nil {
		_, first_c_n_i := n.prefix_child_nodes.btree.Min()
		if first_c_n_i == nil {
			return nil, errors.New("folder 0 child ")
		}

		first_c_n := first_c_n_i.(*Node)
		if iter_check_node_hit(first_c_n) {
			return first_c_n, nil
		} else {
			return iter_recursive_hit_next(trie_db, iter_parent_node, first_c_n)
		}
	}

	//load right neighbour
	btree_iter := n.parent_nodes.btree.Before(uint8(n.prefix[0]))
	btree_iter.Next()
	if btree_iter.Next() {
		//has right  neighbour
		right_n := btree_iter.Value.(*Node)

		if right_n.has_val() {
			return right_n, nil
		} else {
			return iter_recursive_hit_next(trie_db, iter_parent_node, btree_iter.Value.(*Node))
		}

	}

	upper_right_n := recursive_upper_right(iter_parent_node, n)
	if upper_right_n == nil {
		return nil, nil
	}

	if iter_check_node_hit(upper_right_n) {
		return upper_right_n, nil
	} else {
		return iter_recursive_hit_next(trie_db, iter_parent_node, upper_right_n)
	}

}

func recursive_upper_right(iter_parent_node *Node, n *Node) *Node {
	if n.parent_nodes.parent_node == iter_parent_node {
		return nil
	}

	btree_iter := n.parent_nodes.parent_node.parent_nodes.btree.Before(uint8(n.parent_nodes.parent_node.prefix[0]))
	btree_iter.Next()
	if btree_iter.Next() {
		return btree_iter.Value.(*Node)
	}

	return recursive_upper_right(iter_parent_node, n.parent_nodes.parent_node)
}

func (iter *Iterator) get_next_node() (*Node, error) {
	next_hit_node, err := iter_recursive_hit_next(iter.trie_db, iter.parent_node, iter.current_node)
	if err != nil {
		return nil, err
	}
	return next_hit_node, nil
}

func (iter *Iterator) HasNext() (bool, error) {
	if iter.next_node != nil {
		return true, nil
	}

	next_n, err := iter.get_next_node()
	if err != nil {
		return false, err
	}

	iter.next_node = next_n
	return true, nil
}

func (iter *Iterator) Next() (bool, error) {

	if iter.next_node != nil {
		iter.current_node = iter.next_node
		return true, nil
	}

	//
	next_n, err := iter.get_next_node()
	if err != nil {
		return false, err
	}

	if next_n != nil {
		iter.current_node = next_n
		iter.next_node = nil
		return true, nil
	}

	return false, nil

}

func (iter *Iterator) recursive_get_child_node(target_node *Node, left_prefix []byte) (*Node, error) {

	if target_node == iter.parent_node {

		if !target_node.has_folder_child() {
			return nil, nil
		}

		recover_err := iter.trie_db.recover_child_nodes(target_node, true, false)
		if recover_err != nil {
			return nil, recover_err
		}

		target_c_n_i := target_node.folder_child_nodes.btree.Get(uint8(left_prefix[0]))
		if target_c_n_i == nil {
			return nil, nil
		}

		return iter.recursive_get_child_node(target_c_n_i.(*Node), left_prefix)

	}

	if bytes.Equal(target_node.prefix, left_prefix) {

		if target_node.has_folder_child() || target_node.has_val() {
			return target_node, nil
		} else {
			return nil, nil
		}

	} else if (len(left_prefix) > len(target_node.prefix)) && bytes.Equal(target_node.prefix, left_prefix[0:len(target_node.prefix)]) {
		// left_prefix start with target_node.prefix

		if !target_node.has_prefix_child() {
			//not found
			return nil, nil
		}

		recover_err := iter.trie_db.recover_child_nodes(target_node, false, true)
		if recover_err != nil {
			return nil, recover_err
		}

		target_c_n_i := target_node.prefix_child_nodes.btree.Get(uint8(left_prefix[len(target_node.prefix)]))
		if target_c_n_i == nil {
			return nil, nil
		}

		return iter.recursive_get_child_node(target_c_n_i.(*Node), left_prefix[len(target_node.prefix):])

	} else {
		// target_node.prefix start with left_prefix
		// target_node.path, left_prefix, they have common prefix

		//not found
		return nil, nil
	}

}

func (iter *Iterator) SetCursorWithFullPath(full_path [][]byte) (bool, error) {

	if len(full_path) != len(iter.parent_node.node_path())+1 {
		return false, errors.New("full_path not a child of this iterator")
	}

	return iter.SetCursor(full_path[len(full_path)-1])
}

func (iter *Iterator) SetCursor(child_prefix []byte) (bool, error) {

	if len(child_prefix) == 0 {
		return false, errors.New("child path empty")
	}

	n, err := iter.recursive_get_child_node(iter.parent_node, child_prefix)
	if err != nil {
		return false, err
	}

	if n != nil {
		iter.current_node = n
		iter.next_node = nil
		return true, nil
	} else {
		return false, nil
	}
}

func (iter *Iterator) Val() ([]byte, error) {
	recover_err := iter.trie_db.recover_node_val(iter.current_node)
	if recover_err != nil {
		return nil, recover_err
	} else {
		return iter.current_node.val, nil
	}
}

func (iter *Iterator) FullPath() [][]byte {
	return iter.current_node.node_path()
}

func (iter *Iterator) FullPathFlatStr() string {
	return iter.current_node.node_path_flat_str()
}

func (iter *Iterator) HasVal() bool {
	return iter.current_node.has_val()
}

func (iter *Iterator) IsFolder() bool {
	return iter.current_node.has_folder_child()
}
