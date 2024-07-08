package triedb

import "errors"

type Iterator struct {
	triedb      *TrieDB
	root_node   *Node // also the min node
	cursor_node *Node // current pointed node
}

func NewIterator(parent_n *Node, cursor_n *Node) (*Iterator, error) {
	if parent_n == nil {
		return nil, errors.New("parent node nil")
	}

	if cursor_n == nil {
		return &Iterator{
			triedb:      parent_n.triedb,
			root_node:   parent_n,
			cursor_node: parent_n,
		}, nil
	} else {
		//check
		cursor_p_n, err := cursor_n.ParentNode()
		if err != nil {
			return nil, err
		}

		if parent_n != cursor_p_n {
			return nil, errors.New("cursor parent != parent node")
		} else {
			return &Iterator{
				triedb:      parent_n.triedb,
				root_node:   parent_n,
				cursor_node: cursor_p_n,
			}, nil
		}
	}
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

	recover_val_err := next_node.recover_node_val()
	if recover_val_err != nil {
		return nil, recover_val_err
	}
	return next_node, nil
}

func (iter *Iterator) next_recursive(current_node *Node) (*Node, error) {

	//
	if current_node.parent_nodes == iter.root_node.parent_nodes {
		return nil, nil
	}

	//check recover
	if current_node.child_nodes == nil && !current_node.child_nodes_hash_recovered && current_node.child_nodes_hash != nil {
		recover_child_err := current_node.recover_child_nodes()
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
		right_n, err := current_node.right_node()
		if err != nil {
			return nil, nil
		}

		//
		if right_n != nil {
			//
			right_n.recover_node_val()
			if right_n.val != nil {
				return right_n, nil
			}
			//
			return iter.next_recursive(right_n)
		}

		//return upper right
		upper_right_n, err := current_node.upper_right_node()
		if err != nil {
			return nil, nil
		}

		//
		if upper_right_n == nil || upper_right_n.parent_nodes == iter.root_node.parent_nodes {

			return nil, nil

		} else {
			//
			r_n_v_err := upper_right_n.recover_node_val()
			if r_n_v_err != nil {
				return nil, r_n_v_err
			}
			if upper_right_n.val != nil {
				return upper_right_n, nil
			}
			//
			return iter.next_recursive(upper_right_n)
		}

	} else {

		//// condition below current_node.child_nodes != nil

		child_iter := current_node.child_nodes.path_btree.Before(uint8(0))

		for child_iter.Next() {

			c_n := child_iter.Value.(*Node)
			recover_node_err := c_n.recover_node()
			if recover_node_err != nil {
				return nil, errors.New("next_recursive recover_node_err, " + recover_node_err.Error())
			}

			//try recover
			recover_node_v_err := c_n.recover_node_val()
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

		return nil, nil
	}
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

func (iter *Iterator) SkipNext() (bool, error) {
	r_n, err := iter.cursor_node.right_node()
	if err != nil {
		return false, err
	}

	if r_n != nil {
		r_n_v_err := r_n.recover_node_val()
		if r_n_v_err != nil {
			return false, r_n_v_err
		}
		if r_n.val != nil {
			iter.cursor_node = r_n
			return true, nil
		} else {
			next_n, err := iter.next_recursive(r_n)
			if err != nil {
				return false, err
			} else {
				iter.cursor_node = next_n
				return true, nil
			}
		}
	}

	up_r_n, err := iter.cursor_node.upper_right_node()
	if err != nil {
		return false, err
	}

	if up_r_n == nil {
		return false, nil
	} else {
		r_n_v_err := up_r_n.recover_node_val()
		if r_n_v_err != nil {
			return false, r_n_v_err
		}
		if up_r_n.val != nil {
			iter.cursor_node = up_r_n
			return true, nil
		} else {
			next_n, err := iter.next_recursive(up_r_n)
			if err != nil {
				return false, err
			} else {
				iter.cursor_node = next_n
				return true, nil
			}
		}
	}

}

// cursor won't move
func (iter *Iterator) GetNext() (*Node, error) {
	return iter.get_next()
}
