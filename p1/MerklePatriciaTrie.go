package p1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"reflect"
	"strings"
)

//var KeyVal map[string]string

type Flag_value struct {
	encoded_prefix []uint8
	value          string
}

type Node struct {
	node_type    int // 0: Null, 1: Branch, 2: Ext or Leaf
	branch_value [17]string
	flag_value   Flag_value
}

type MerklePatriciaTrie struct {
	db     map[string]Node
	root   string
	KeyVal map[string]string
}

func (mpt *MerklePatriciaTrie) GetRoot() string {
	return mpt.root
}

func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {
	var strToAscii []uint8
	var decoded []uint8
	strToAscii = str_to_ascii(key)
	decoded = compact_decode_wt_prefix(strToAscii)
	root := mpt.db[mpt.root]
	value := mpt.get_helper(root, decoded)
	return value, errors.New("path_not_found")
}

func (mpt MerklePatriciaTrie) get_helper(current_node Node, new_path []uint8) string {
	if current_node.node_type == 0 {
		return "failure"
	} else if current_node.node_type == 1 {
		if len(new_path) == 0 {
			return current_node.branch_value[16]
		} else {
			return mpt.get_helper(mpt.db[current_node.branch_value[new_path[0]]], new_path[1:])
		}
	} else {
		var i int
		old_path := compact_decode(current_node.flag_value.encoded_prefix)
		for i = 0; i < len(new_path) && i < len(old_path); i += 1 {
			if new_path[i] != old_path[i] {
				break
			}
		}
		if i == len(new_path) && i == len(old_path) {
			return current_node.flag_value.value
		} else if i < len(old_path) && i < len(new_path) {
			return "failure"
		} else {
			return mpt.get_helper(mpt.db[current_node.flag_value.value], new_path[i:])
		}
	}
}

func (mpt *MerklePatriciaTrie) Insert(key string, new_value string) {
	if len(mpt.KeyVal) == 0 {
		mpt.KeyVal = make(map[string]string)
	}
	mpt.KeyVal[key] = new_value
	var strToAscii []uint8
	var decoded []uint8
	strToAscii = str_to_ascii(key)
	decoded = compact_decode_wt_prefix(strToAscii)
	root_node := mpt.insert_helper(decoded, new_value, mpt.db[mpt.root])
	mpt.root = root_node.hash_node()
	if mpt.db == nil {
		mpt.db = make(map[string]Node)
	}
	mpt.db[mpt.root] = root_node
	//fmt.Println("root: ", mpt.root)
	//for key, value := range mpt.db {
	//	fmt.Println("key: ", key, "value: ", value)
	//}
}

func (mpt *MerklePatriciaTrie) is_any_match(decoded_path_1 []uint8, decoded_path_2 []uint8) (bool, int) {
	length := len(decoded_path_1)
	var index int = 0
	if len(decoded_path_1) > len(decoded_path_2) {
		length = len(decoded_path_2)
	}
	for i := 0; i < length; i++ {
		//index++
		if decoded_path_1[i] != decoded_path_2[i] {
			break
		}
		index++
	}
	if index > 0 {
		return true, index
	}
	return false, index
}

func (mpt *MerklePatriciaTrie) insert_helper(path []uint8, new_value string, current_node Node) Node {
	previous_node := current_node
	fmt.Println("new_ path: ", path, "value: ", new_value)
	if current_node.node_type == 0 || mpt.root == "" {
		//create leaf node
		fmt.Println("creating leaf condition 1")
		newnode := Node{2, [17]string{}, Flag_value{compact_encode(append(path, 16)), new_value}}
		current_node = newnode
		//return n
	} else if current_node.node_type == 1 {
		fmt.Println("Current is Branch")
		if len(path) == 0 {
			current_node.branch_value[16] = new_value
		} else {
			fmt.Println("ELSE IN BRANCH!")
			n := mpt.insert_helper(path[1:], new_value, mpt.db[current_node.branch_value[path[0]]])
			current_node.branch_value[path[0]] = n.hash_node()
			return current_node
		}
	} else if current_node.node_type == 2 {
		existing_node_path := compact_decode(current_node.flag_value.encoded_prefix)
		fmt.Println("existing_ path:", existing_node_path)
		length := len(existing_node_path)
		var index int = 0
		if len(existing_node_path) > len(path) {
			length = len(path)
		}
		for i := 0; i < length; i++ {
			//index++
			if existing_node_path[i] != path[i] {
				break
			}
			index++
		}
		fmt.Println("index", index, "path length: ", len(path))
		if index == len(existing_node_path) && index == len(path) { //full match in both paths (hex values)
			fmt.Println("check 0")
			if compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3 { //if current node is leaf
				current_node.flag_value.value = new_value
			} else { //if nodes (current node and node_to_be_inserted) have same path and current is an extension node, then just store the new value in next branch node, and  override previous
				n := mpt.insert_helper(path[index:], new_value, mpt.db[current_node.flag_value.value])
				current_node.flag_value.value = n.hash_node()
			}
		} else if index == len(existing_node_path) && index < len(path) {
			fmt.Println("check 1")
			new_node := Node{2, [17]string{}, Flag_value{compact_encode(append(path[index+1:], 16)), new_value}} // make leaf of remaining path that would not be stored in extension node
			branch_node := Node{1, [17]string{""}, Flag_value{}}
			if compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3 { //if current node is leaf
				fmt.Println("is leaf")
				if len(existing_node_path) == 0 {
					fmt.Println("path length is zero")
					branch_node.branch_value[16] = current_node.flag_value.value
					branch_node.branch_value[path[index]] = new_node.hash_node() //HINT.TEST still testing
					mpt.db[branch_node.hash_node()] = branch_node
					mpt.db[new_node.hash_node()] = new_node
					delete(mpt.db, current_node.hash_node())
					return branch_node
				} else {
					fmt.Println("path length more than zero")
					current_node.flag_value.encoded_prefix = compact_encode(existing_node_path[:index])
					branch_node.branch_value[16] = current_node.flag_value.value
					branch_node.branch_value[path[index]] = new_node.hash_node() //HINT.TEST: copied from below and addded here
					current_node.flag_value.value = branch_node.hash_node()
					//fmt.Println("current node: ", current_node)
					//fmt.Println("current node hash: ", current_node.hash_node())
					//fmt.Println("branch node hash", branch_node.hash_node())
					mpt.db[current_node.hash_node()] = current_node
					mpt.db[new_node.hash_node()] = new_node       //HINT.TEST: copied from below and addded here
					mpt.db[branch_node.hash_node()] = branch_node //HINT.TEST: copied from below and addded here
				}
			} else {
				fmt.Println("is extension")
				new_node = mpt.insert_helper(path[index:], new_value, mpt.db[current_node.flag_value.value])
				fmt.Println("NEW NODE:", new_node)
				current_node.flag_value.value = new_node.hash_node()
				fmt.Println("CurrentNode: ", current_node)
				fmt.Println("Current Node hash:", current_node.hash_node())

				//HINT: Leaving this block for future optimization. Consider how to replicate same solutions for extension with partial match done in next IF statement below
				//ext_node2 := Node{2, [17]string{}, Flag_value{compact_encode(path[index + 1:]), new_value}}
				//branch_node := Node{1, [17]string{""}, Flag_value{}}
				//branch_node.branch_value[16] = current_node.flag_value.value
				//branch_node.branch_value[path[index]] = ext_node2.hash_node()
				//ext_node1 := Node{2, [17]string{}, Flag_value{compact_encode(path[:index]), branch_node.hash_node()}}
				//delete(mpt.db, current_node.hash_node())
				//mpt.db[ext_node2.hash_node()] = ext_node2
				//mpt.db[branch_node.hash_node()] = branch_node
				//mpt.db[ext_node1.hash_node()] = ext_node1
				//current_node = ext_node1
				//return current_node
			}
			fmt.Println("Index: ", index)
			branch_node.branch_value[path[index]] = new_node.hash_node()
			mpt.db[new_node.hash_node()] = new_node
			mpt.db[branch_node.hash_node()] = branch_node
		} else if index < len(existing_node_path) && index == len(path) {
			fmt.Println("check 2")
			current_node_value := current_node.flag_value.value
			fmt.Println(current_node_value)
			branch_node := Node{1, [17]string{""}, Flag_value{}}
			new_flag := Flag_value{encoded_prefix: compact_encode(append(existing_node_path[index+1:], 16)), value: current_node.flag_value.value}
			new_node := Node{node_type: 2, branch_value: [17]string{}, flag_value: new_flag}
			branch_node.branch_value[existing_node_path[index]] = new_node.hash_node()
			//fmt.Println(new_node)
			if compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 1 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 0 {
				//if extension node
				fmt.Println("is extension")
				ext_node2 := Node{2, [17]string{}, Flag_value{compact_encode(existing_node_path[index+1:]), current_node_value}}
				branch_node := Node{1, [17]string{""}, Flag_value{}}
				branch_node.branch_value[16] = new_value
				branch_node.branch_value[existing_node_path[index]] = ext_node2.hash_node()
				ext_node1 := Node{2, [17]string{}, Flag_value{compact_encode(existing_node_path[:index]), branch_node.hash_node()}}
				delete(mpt.db, current_node.hash_node())
				mpt.db[ext_node2.hash_node()] = ext_node2
				mpt.db[branch_node.hash_node()] = branch_node
				mpt.db[ext_node1.hash_node()] = ext_node1
				current_node = ext_node1
				return current_node
			} else {
				fmt.Println("is leaf", len(path))
				if len(path) == 0 {
					branch_node.branch_value[16] = new_value //put current value in 16th index
					mpt.db[new_node.hash_node()] = new_node
					mpt.db[branch_node.hash_node()] = branch_node
					delete(mpt.db, current_node.hash_node())
				} else {
					current_node.flag_value.encoded_prefix = compact_encode(path[:index]) //convert to extension node
					branch_node.branch_value[16] = new_value
					current_node.flag_value.value = branch_node.hash_node()
					mpt.db[current_node.hash_node()] = current_node
				}
				mpt.db[new_node.hash_node()] = new_node //update db
				mpt.db[branch_node.hash_node()] = branch_node
				return branch_node
			}
		} else if index == 0 {
			fmt.Println("check 3")
			var new_node Node
			branch_node := Node{1, [17]string{""}, Flag_value{}}
			if compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3 { //if current node is leaf
				fmt.Println("is leaf")
				new_node = Node{2, [17]string{}, Flag_value{compact_encode(append(existing_node_path[index+1:], 16)), current_node.flag_value.value}}
				mpt.db[new_node.hash_node()] = new_node
				branch_node.branch_value[existing_node_path[index]] = new_node.hash_node()
			} else { //extension
				fmt.Println("is extension")
				if len(existing_node_path) == 1 {
					fmt.Println("length is 1")
					node_connected_to_extension := mpt.db[current_node.flag_value.value]
					branch_node.branch_value[existing_node_path[index]] = node_connected_to_extension.hash_node() //HINT.TEST: connect branch connecting to current extension to new branch, before removing extension
					fmt.Println("current node: ", current_node)
					fmt.Println("current node hash: ", current_node.hash_node())
					delete(mpt.db, current_node.hash_node())
				} else {
					fmt.Println("len is more than 1")
					delete(mpt.db, current_node.hash_node())                                                                            //HINT.TEST: remove current node from db before updating
					current_node.flag_value.encoded_prefix = compact_encode(compact_decode(current_node.flag_value.encoded_prefix)[1:]) //HINT.TEST: (p,aaa,aap) Remove first index from extension (current node) as we use that in path to connect it to parent branch
					mpt.db[current_node.hash_node()] = current_node                                                                     //HINT.TEST: Add updated current node to db
					branch_node.branch_value[existing_node_path[index]] = current_node.hash_node()                                      //HINT.TEST: (p,aaa,aap) Assign Branch node value to current node value
				}
			}
			leaf_node := Node{2, [17]string{}, Flag_value{compact_encode(append(path[index+1:], 16)), new_value}}
			branch_node.branch_value[path[index]] = leaf_node.hash_node()
			fmt.Println("branch node: ", branch_node)
			fmt.Println("branch node hash: ", branch_node.hash_node())
			mpt.db[leaf_node.hash_node()] = leaf_node
			mpt.db[branch_node.hash_node()] = branch_node
			return branch_node
		} else {
			fmt.Println("check 4")
			fmt.Println("Any other case enters here")
			var new_node Node
			var isextension bool
			if compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3 { //if current node is leaf
				fmt.Println("is leaf")
				new_node = Node{2, [17]string{}, Flag_value{compact_encode(append(existing_node_path[index+1:], 16)), current_node.flag_value.value}}
			} else { //extension
				fmt.Println("is extension")
				isextension = true
				new_node = Node{2, [17]string{}, Flag_value{compact_encode(existing_node_path[index+1:]), current_node.flag_value.value}}
			}
			prefix := compact_encode(path[:index])
			current_node.flag_value.encoded_prefix = prefix
			branch_node := Node{1, [17]string{""}, Flag_value{}}
			mpt.db[new_node.hash_node()] = new_node
			fmt.Println("node prefix: ", compact_decode(new_node.flag_value.encoded_prefix))
			if isextension && len(compact_decode(new_node.flag_value.encoded_prefix)) == 0 { //HINT.TEST: added for aa,ap,b insert. Ensures we only add new node to branch if it doesn't have empty prefix
				fmt.Println("ELSE... node is extension with empty prefix")
				branch_node.branch_value[existing_node_path[index]] = current_node.flag_value.value
			} else {
				branch_node.branch_value[existing_node_path[index]] = new_node.hash_node()
			}
			leaf_node := Node{2, [17]string{}, Flag_value{compact_encode(append(path[index+1:], 16)), new_value}}
			branch_node.branch_value[path[index]] = leaf_node.hash_node()
			mpt.db[leaf_node.hash_node()] = leaf_node
			current_node.flag_value.value = branch_node.hash_node()
			mpt.db[current_node.hash_node()] = current_node
			mpt.db[branch_node.hash_node()] = branch_node
			fmt.Println("Current node: ", current_node)
			fmt.Println("Current node hash: ", current_node.hash_node())
			fmt.Println("New node: ", new_node)
			fmt.Println("New node hash: ", new_node.hash_node())
			//current_node = new_node
			return current_node //SEE IF THIS DOESN"T AFFECT OTHER CASES (works for p, aa, ap)
		}
	}
	if previous_node.hash_node() != current_node.hash_node() {
		fmt.Println("remove node :", previous_node)
		fmt.Println("hash of remove node :", previous_node.hash_node())
		delete(mpt.db, previous_node.hash_node())
		mpt.db[current_node.hash_node()] = current_node
	}
	return current_node
}

func (mpt *MerklePatriciaTrie) Delete(key string) (string, error) {
	var strToAscii []uint8
	var decoded []uint8
	strToAscii = str_to_ascii(key)
	decoded = compact_decode_wt_prefix(strToAscii)
	if mpt.root == "" || len(mpt.db) == 0 {
		return "", errors.New("path_not_found")
	}
	root := mpt.db[mpt.root]
	node := mpt.delete_helper(decoded, root, Node{}) //passing empty parent node as it isn't used here
	if node.node_type == 0 {
		return "", errors.New("path_not_found")
	}
	fmt.Println("successfully deleted")
	fmt.Println("Node returned from delete_helper: ", node)
	mpt.root = node.hash_node() //?? just added, hope it's fine
	//fmt.Println("CHECK root: ", mpt.root)
	//for key, value := range mpt.db {
	//	fmt.Println("key: ", key, "value: ", value)
	//}
	return "", errors.New("")
}

func find_index_branch(current_node Node) []uint8 {
	var index_of_branch []uint8
	fmt.Println(len(current_node.branch_value))
	for i := 0; i < len(current_node.branch_value); i++ {
		if current_node.branch_value[i] != "" {
			index_of_branch = append(index_of_branch, uint8(i))
		}
	}
	return index_of_branch
}

//**** HINT.SOPE: Changed the rebalance trie function get current branch indexes instead of count
//HINT.SOPE: Added previous_node parameter due to issues noticed in TestExt 111 case
func (mpt *MerklePatriciaTrie) rebalance_trie(current_node Node, previous_node Node) Node { //expects a branch node as input
	fmt.Println("GETS HERE!")
	index_of_branch := find_index_branch(current_node)
	if len(index_of_branch) == 1 {
		next_node := mpt.db[current_node.branch_value[index_of_branch[0]]]
		fmt.Println("Index of Branch:", index_of_branch[0])
		fmt.Println("next node =", current_node)
		if next_node.node_type == 2 {
			fmt.Println("next node is leaf or extension.....")
			if compact_decode_wt_prefix(next_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(next_node.flag_value.encoded_prefix)[0] == 3 {
				// if next node is leaf, convert current branch node to leaf node containing both branch & leaf values
				current_node.flag_value.value = next_node.flag_value.value
				fmt.Println("flg: ", next_node.flag_value.value)
				prefix := append(index_of_branch, compact_decode(next_node.flag_value.encoded_prefix)...)
				current_node.flag_value.encoded_prefix = compact_encode(append(prefix, 16))
				current_node.node_type = 2
				fmt.Println("CURR NODE: ", current_node)
				current_node.branch_value = [17]string{}
				return current_node
			} else if compact_decode_wt_prefix(next_node.flag_value.encoded_prefix)[0] == 1 || compact_decode_wt_prefix(next_node.flag_value.encoded_prefix)[0] == 0 {
				//if extension, merge current branch with only one element with child extension and to make a new extension
				fmt.Println("Inside extension merge...")
				next_node_extension := compact_decode(next_node.flag_value.encoded_prefix)
				new_extension_prefix := append(index_of_branch, next_node_extension...)
				current_node.flag_value.encoded_prefix = compact_encode(new_extension_prefix) //merge 2 extension into one
				current_node.flag_value.value = next_node.flag_value.value
				current_node.node_type = 2
				current_node.branch_value = [17]string{}
				//fmt.Println("new extension: ", current_node)
				//fmt.Println("new extension hash: ", current_node.hash_node())
				return current_node
			}
		} else if next_node.node_type == 1 {
			//if next node is branch, merge current branch and next node into one extension
			fmt.Println("next node is branch")
			index_of_new_branch := find_index_branch(next_node)
			if len(index_of_new_branch) == 1 {
				current_node.flag_value.encoded_prefix = compact_encode(append(index_of_branch, index_of_new_branch...))
				current_node.flag_value.value = current_node.branch_value[index_of_new_branch[0]]
				current_node.node_type = 2
				current_node.branch_value = [17]string{}
				return current_node
			} else { //If next node is branch and has more than one node, convert current node to extension and link value to child branch
				fmt.Println("SEEING THIS")
				if previous_node.node_type == 2 { //HINT.TEST: Fix for TestExt 111
					fmt.Println("previous node: ", previous_node)
					fmt.Println("got here")
					return current_node //JUST RETURN CURRENT NODE
				} else {
					new_node := Node{2, [17]string{}, Flag_value{compact_encode(index_of_branch), current_node.branch_value[index_of_branch[0]]}} //extension node
					current_node = new_node
				}
				//return current_node	//JUST RETURN CURRENT NODE
			}
		}
	}
	//fmt.Println("rebalanced node: ", current_node)
	//fmt.Println("rebalanced node hash: ", current_node.hash_node())
	return current_node
}

//HINT.SOPE: Added parent_node parameter due to issues noticed in TestExt 111 case
func (mpt *MerklePatriciaTrie) delete_helper(path []uint8, current_node Node, parent_node Node) Node {
	fmt.Println("********************************DELETING********************************")
	fmt.Println("delete path", path)
	var previous_node Node
	previous_node = current_node
	if current_node.node_type == 1 { //branch node
		if len(path) == 0 {
			fmt.Println("entered in branch where path len == 0, node: ", current_node)
			current_node.branch_value[16] = ""
			assigned_branch_indices := find_index_branch(current_node)
			if len(assigned_branch_indices) == 1 {
				node := mpt.rebalance_trie(current_node, parent_node)
				current_node = node
			}
		} else {
			//decided how to recurse below
			fmt.Println("entered in branch where path len > 0, node: ", current_node)
			var node Node
			fmt.Println("recurse further from branch, path: ", path[1:])
			node = mpt.delete_helper(path[1:], mpt.db[current_node.branch_value[path[0]]], previous_node)
			fmt.Println("entered in branch where path len > 0, result node: ", node)
			if node.node_type == 0 {
				current_node.branch_value[path[0]] = ""
				fmt.Println("Set curr node index to empty. curr node: ", current_node)
				if current_node.branch_value[16] != "" {
					fmt.Println("Entered IF in branch where path len > 0")
					if len(find_index_branch(current_node)) < 2 {
						var new_path []uint8
						new_node := Node{2, [17]string{}, Flag_value{compact_encode(append(new_path, 16)), current_node.branch_value[16]}} //convert to leaf node
						current_node = new_node
					}
				} else {
					fmt.Println("Branch 16th value is empty")
					new_node := mpt.rebalance_trie(current_node, parent_node)
					current_node = new_node
				}
			} else {
				fmt.Println("Entered ELSE in branch where path len > 0")
				current_node.branch_value[path[0]] = node.hash_node()
			}

		}
	} else if current_node.node_type == 2 {
		existing_node_path := compact_decode(current_node.flag_value.encoded_prefix)
		fmt.Println("existing_ path:", existing_node_path)
		length := len(existing_node_path)
		var index int = 0
		if len(existing_node_path) > len(path) {
			length = len(path)
		}
		for i := 0; i < length; i++ {
			if existing_node_path[i] != path[i] {
				break
			}
			index++
		}
		fmt.Println("index", index, "path length: ", len(path))
		if index == len(path) && index == len(existing_node_path) && (compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 3) { //if paths match fully and node isLeaf
			fmt.Println("full match deleting")
			delete(mpt.db, current_node.hash_node())
			current_node.flag_value.value = ""
			current_node.flag_value.encoded_prefix = nil
			current_node.node_type = 0
			fmt.Println("after deletion := ", current_node)
		} else if index == len(existing_node_path) && (compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 1 || compact_decode_wt_prefix(current_node.flag_value.encoded_prefix)[0] == 0) { //is extension
			fmt.Println("partial match deleting from path: ", path[index:])
			node := mpt.delete_helper(path[index:], mpt.db[current_node.flag_value.value], previous_node)
			fmt.Println("finished recursing from extension parent. returned value: ", node)
			fmt.Println("current node: ", current_node)
			fmt.Println("node: ", node)
			fmt.Println("node hash: ", node.hash_node())
			if current_node.flag_value.value != node.hash_node() {
				if node.node_type == 0 {
					fmt.Println("Entered zeroth condition")
					current_node.flag_value.value = ""
					current_node = Node{}
				} else if node.node_type == 1 {
					fmt.Println("Entered one condition")
					branch_indexes := find_index_branch(node)
					if len(branch_indexes) == 1 { //if branch has just one index, link extension to branch child and move branch index to extension path
						current_node.flag_value.value = node.branch_value[branch_indexes[0]]
						current_node.flag_value.encoded_prefix = compact_encode(append(compact_decode(current_node.flag_value.encoded_prefix), branch_indexes...))
					} else {
						current_node.flag_value.value = node.hash_node()
					}

				} else if node.node_type == 2 {
					if compact_decode_wt_prefix(node.flag_value.encoded_prefix)[0] == 2 || compact_decode_wt_prefix(node.flag_value.encoded_prefix)[0] == 3 { //is leaf
						fmt.Println("Entered two condition -- leaf")
						decoded_prefix1 := compact_decode(current_node.flag_value.encoded_prefix)
						decoded_prefix2 := compact_decode(node.flag_value.encoded_prefix)
						combined_pref := append(decoded_prefix1, decoded_prefix2...)
						new_node := Node{2, [17]string{}, Flag_value{compact_encode(append(combined_pref, 16)), node.flag_value.value}} //convert to leaf node
						current_node = new_node
					} else if compact_decode_wt_prefix(node.flag_value.encoded_prefix)[0] == 0 || compact_decode_wt_prefix(node.flag_value.encoded_prefix)[0] == 1 {
						fmt.Println("Entered two condition -- extension")
						decoded_prefix1 := compact_decode(current_node.flag_value.encoded_prefix)
						decoded_prefix2 := compact_decode(node.flag_value.encoded_prefix)
						current_node.flag_value.encoded_prefix = compact_encode(append(decoded_prefix1, decoded_prefix2...))
						current_node.flag_value.value = node.flag_value.value //HINT.TEST: Check if needs to be uncommented
					}
				}
			}
		}
	}
	if previous_node.hash_node() != current_node.hash_node() {
		delete(mpt.db, previous_node.hash_node())
		if current_node.node_type != 0 {
			mpt.db[current_node.hash_node()] = current_node
		}
	}
	return current_node
}

func compact_encode(hex_array []uint8) []uint8 {
	var term int
	if len(hex_array) == 0 { //HINT.TEST: Added to solve null path issue
		term = 0
	} else if hex_array[len(hex_array)-1] == 16 {
		term = 1
	} else {
		term = 0
	}
	if term == 1 {
		hex_array = hex_array[:len(hex_array)-1]
	}
	var oddlen int = len(hex_array) % 2
	var flags []uint8 = []uint8{uint8(2*term + oddlen)}
	if oddlen > 0 {
		hex_array = append(flags, hex_array...)
	} else {
		var zeroArr []uint8 = []uint8{0}
		flags = append(flags, zeroArr...)
		hex_array = append(flags, hex_array...)
	}
	var result []uint8
	fmt.Println(hex_array)
	for i := 0; i < len(hex_array); i += 2 {
		result = append(result, 16*hex_array[i]+hex_array[i+1])
	}
	fmt.Print(term, oddlen, flags)
	return result
}

// If Leaf, ignore 16 at the end
func compact_decode(encoded_arr []uint8) []uint8 {
	result := compact_decode_wt_prefix(encoded_arr)
	firstNibble := result[0]
	if firstNibble == 0 || firstNibble == 2 {
		result = result[2:]
	} else {
		result = result[1:]
	}
	return result
}

func compact_decode_wt_prefix(encoded_arr []uint8) []uint8 {
	var result []uint8
	for i := 0; i < len(encoded_arr); i += 1 {
		result = append(result, encoded_arr[i]/16)
		result = append(result, encoded_arr[i]%16)
	}
	return result
}

func str_to_ascii(input string) []uint8 {
	if len(input) == 0 {
		return nil
	}
	return []uint8(input)
}

func test_compact_encode() {
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))
}

func (node *Node) hash_node() string {
	var str string
	switch node.node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.branch_value {
			str += v
		}
	case 2:
		str = node.flag_value.value
	}

	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func (node *Node) String() string {
	str := "empty string"
	switch node.node_type {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.branch_value[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("value=%s]", node.branch_value[16])
	case 2:
		encoded_prefix := node.flag_value.encoded_prefix
		node_name := "Leaf"
		if is_ext_node(encoded_prefix) {
			node_name = "Ext"
		}
		ori_prefix := strings.Replace(fmt.Sprint(compact_decode(encoded_prefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, value=\"%s\">", node_name, ori_prefix, node.flag_value.value)
	}
	return str
}

func node_to_string(node Node) string {
	return node.String()
}

func (mpt *MerklePatriciaTrie) Initial() {
	mpt.db = make(map[string]Node)
}

func is_ext_node(encoded_arr []uint8) bool {
	return encoded_arr[0]/16 < 2
}

func TestCompact() {
	test_compact_encode()
}

func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.root)
	for hash := range mpt.db {
		content += fmt.Sprintf("%s: %s\n", hash, node_to_string(mpt.db[hash]))
	}
	return content
}

func (mpt *MerklePatriciaTrie) Order_nodes() string {
	raw_content := mpt.String()
	content := strings.Split(raw_content, "\n")
	root_hash := strings.Split(strings.Split(content[0], "HashStart")[1], "HashEnd")[0]
	queue := []string{root_hash}
	i := -1
	rs := ""
	cur_hash := ""
	for len(queue) != 0 {
		last_index := len(queue) - 1
		cur_hash, queue = queue[last_index], queue[:last_index]
		i += 1
		line := ""
		for _, each := range content {
			if strings.HasPrefix(each, "HashStart"+cur_hash+"HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart"+cur_hash+"HashEnd", fmt.Sprintf("Hash%v", i), -1)
			}
		}
		temp2 := strings.Split(line, "HashStart")
		flag := true
		for _, each := range temp2 {
			if flag {
				flag = false
				continue
			}
			queue = append(queue, strings.Split(each, "HashEnd")[0])
		}
	}
	return rs
}

//func (mpt *MerklePatriciaTrie) GetMptKeyValues() map[string]string {
//	return keyVal
//}
