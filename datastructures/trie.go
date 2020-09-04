package datastructures

// Node is a node of tree.
type Trie struct {
	IsWord   bool
	Children map[rune]*Trie
}

// NewTrie is create a root node.
func (n *Trie) Init() {
	*n = Trie{
		Children: make(map[rune]*Trie),
		IsWord:   false,
	}
}

// Insert is insert a word to tree.
func (n Trie) Insert(word string) {
	//fmt.Println(n.Children)

	runes := []rune(word)
	curNode := &n
	for _, r := range runes {
		if nextNode, ok := curNode.Children[r]; ok {
			curNode = nextNode
		} else {
			curNode.Children[r] = &Trie{
				Children: make(map[rune]*Trie),
				IsWord:   false,
			}
			curNode = curNode.Children[r]
		}
	}
	curNode.IsWord = true
}

// Search is search a word from a tree.
func (n Trie) Search(str string) bool {
	//fmt.Println(n.Children)

	runes := []rune(str)
	curNode := &n

	for _, r := range runes {
		if nextNode, ok := curNode.Children[r]; ok {
			curNode = nextNode
		} else {
			return false
		}
	}

	return curNode.IsWord
}

// Delete is delete a word from a tree.
func (n Trie) Delete(word string) {
	runes := []rune(word)
	curNode := &n

	for _, r := range runes {
		if nextNode, ok := curNode.Children[r]; ok {
			curNode = nextNode
		} else {
			return
		}
	}

	curNode.IsWord = false
}
