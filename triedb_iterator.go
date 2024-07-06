package triedb

import "errors"

type Iterator struct {
	triedb      *TrieDB
	root_node   *Node // also the min node
	cursor_node *Node // current pointed node
}

func (iter *Iterator) Clone() *Iterator {
	return &Iterator{
		triedb:      iter.triedb,
		root_node:   iter.root_node,
		cursor_node: iter.cursor_node,
	}
}

func (iter *Iterator) Reset() {
	iter.cursor_node = iter.root_node
}

func (iter *Iterator) Value() []byte {
	return iter.cursor_node.val
}

func (iter *Iterator) FullPath() []byte {
	return iter.cursor_node.get_full_path()
}

func (iter *Iterator) Path() []byte {
	return iter.cursor_node.path
}

/*
 *  if hasnext then cursor will point to the next node and return true
 *	otherwise stay in the current cursor position and return false
 */
func (iter *Iterator) Next() (bool, error) {
	//
	next_n, err := iter.get_next()
	if err != nil {
		return false, err
	}
	//
	if next_n == nil {
		return false, nil
	} else {
		iter.cursor_node = next_n
		return true, nil
	}
}

// not suggested to call, don't need to check HasNext() before calling Next()
// cursor won't move after calling this function
func (iter *Iterator) HasNext() (bool, error) {
	next_n, err := iter.get_next()
	if err != nil {
		return false, err
	}
	if next_n == nil {
		return false, nil
	} else {
		return true, nil
	}
}

func (iter *Iterator) get_next() (*Node, error) {
	if iter.cursor_node == nil {
		return nil, nil
	}

	next_node, err := iter.next_recursive(iter.cursor_node)
	if err != nil {
		return nil, err
	}

	if next_node == nil {
		return nil, nil
	}

	recover_val_err := iter.triedb.recover_node_val(next_node)
	if recover_val_err != nil {
		return nil, recover_val_err
	}
	return next_node, nil
}

func (iter *Iterator) next_recursive(current_node *Node) (*Node, error) {

	//check recover
	if current_node.child_nodes == nil && !current_node.child_nodes_hash_recovered && current_node.child_nodes_hash != nil {
		recover_child_err := iter.triedb.recover_child_nodes(current_node)
		if recover_child_err != nil {
			return nil, errors.New("next_recursive recover_child_err, " + recover_child_err.Error())
		}
	}

	//root node & no child
	if current_node.child_nodes == nil && iter.root_node == current_node {
		return nil, nil
	}

	// not root node
	if current_node.child_nodes == nil {
		//check my right neighbor node
		right_n, err := iter.triedb.right(current_node)
		if err != nil {
			return nil, nil
		}

		//
		if right_n != nil {
			//
			iter.triedb.recover_node_val(right_n)
			if right_n.val != nil {
				return right_n, nil
			}
			//
			return iter.next_recursive(right_n)
		}
		//
		if current_node.parent_nodes.parent_node == iter.root_node {
			return nil, nil
		}
		//return upper right
		upper_right_n, err := iter.triedb.upper_right(current_node)
		if err != nil {
			return nil, nil
		}

		if upper_right_n != nil {
			//
			iter.triedb.recover_node_val(upper_right_n)
			if upper_right_n.val != nil {
				return upper_right_n, nil
			}
			//
			return iter.next_recursive(upper_right_n)
		}

		return nil, nil

	} else {

		//// condition below current_node.child_nodes != nil

		child_iter := current_node.child_nodes.path_btree.Before(uint8(0))

		for child_iter.Next() {

			c_n := child_iter.Value.(*Node)
			recover_node_err := iter.triedb.recover_node(c_n)
			if recover_node_err != nil {
				return nil, errors.New("next_recursive recover_node_err, " + recover_node_err.Error())
			}

			//try recover
			recover_node_v_err := iter.triedb.recover_node_val(c_n)
			if recover_node_v_err != nil {
				return nil, errors.New("next_recursive recover_node_val_err, " + recover_node_v_err.Error())
			}

			if c_n.val != nil {
				return c_n, nil
			}

			target_n, err := iter.next_recursive(child_iter.Value.(*Node))
			if err != nil {
				return nil, err
			}

			if target_n != nil {
				return target_n, nil
			}

		}

		//impossible to happen here
		//panic("program err remove after debug")
		return nil, nil
	}
}
