package triedb

import "errors"

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
	if check_node_hit(cursor_n) {
		return &Iterator{
			trie_db:      trie_db,
			parent_node:  n,
			current_node: cursor_n,
		}, nil
	}

	next_hit_node, err := recursive_hit_next(trie_db, n, cursor_n)
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
func check_node_hit(n *Node) bool {
	if n.has_val() || n.has_folder_child() {
		return true
	} else {
		return false
	}
}

func recursive_hit_next(trie_db *TrieDB, iter_parent_node *Node, n *Node) (*Node, error) {

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
		if check_node_hit(first_c_n) {
			return first_c_n, nil
		} else {
			return recursive_hit_next(trie_db, iter_parent_node, first_c_n)
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
			return recursive_hit_next(trie_db, iter_parent_node, btree_iter.Value.(*Node))
		}

	}

	upper_right_n := recursive_upper_right(iter_parent_node, n)
	if upper_right_n == nil {
		return nil, nil
	}

	if upper_right_n.has_val() {
		return upper_right_n, nil
	} else {
		return recursive_hit_next(trie_db, iter_parent_node, upper_right_n)
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
	next_hit_node, err := recursive_hit_next(iter.trie_db, iter.parent_node, iter.current_node)
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

func (iter *Iterator) HasVal() bool {
	return iter.current_node.has_val()
}

func (iter *Iterator) IsFolder() bool {
	return iter.current_node.has_folder_child()
}
