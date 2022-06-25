package trie

import (
	"fmt"
	"sort"
)

const (
	maxTopQueries = 5
)

type Trie struct {
	root *Node
}

func New() *Trie {
	return &Trie{
		root: &Node{},
	}
}

func (t *Trie) Add(word string) {
	rs := []rune(word)
	t.root.add(rs)
}

func (t *Trie) AutoComplete(word string) []*QueryItem {
	rs := []rune(word)
	n, _ := t.root.find(rs)
	return n.topQueries
}

type Node struct {
	word      string
	frequency int
	children  map[rune]*Node

	topQueries []*QueryItem // order by frequency desc
}

func (n *Node) find(rs []rune) (*Node, []rune) {
	if len(rs) == 0 {
		return n, nil
	}

	if n.children == nil {
		return n, rs
	}

	child, ok := n.children[rs[0]]
	if !ok {
		return n, rs
	}

	return child.find(rs[1:])
}

func (n *Node) add(rs []rune) *QueryItem {
	if len(rs) == 0 {
		n.frequency++
		return n.asQureryItem()
	}

	if n.children == nil {
		n.children = make(map[rune]*Node)
	}

	child, ok := n.children[rs[0]]
	if !ok {
		child = &Node{
			word: n.word + string(rs[0]),
		}
		n.children[rs[0]] = child
	}

	qi := child.add(rs[1:])
	return n.updateTopQueryItems(qi)
}

func (n *Node) remove(rs []rune) {
}

func (n *Node) updateTopQueryItems(item *QueryItem) *QueryItem {
	if n.word == "" { // root node
		return nil
	}

	if item == nil {
		return nil
	}

	if len(n.topQueries) == 0 {
		n.topQueries = append(n.topQueries, item)
		return item
	}

	for i, qi := range n.topQueries {
		if item.word == qi.word && item.frequency > qi.frequency {
			n.topQueries[i] = item
			n.sortTopQueryItems()
			return item
		}
	}

	n.topQueries = append(n.topQueries, item)
	n.sortTopQueryItems()

	if len(n.topQueries) <= maxTopQueries {
		return item
	} else {
		n.topQueries = n.topQueries[:maxTopQueries]
		lastItem := n.topQueries[len(n.topQueries)-1]
		if lastItem.Equals(item) {
			return nil
		} else {
			return lastItem
		}

	}
}

func (n *Node) sortTopQueryItems() {
	sort.Slice(
		n.topQueries,
		func(i, j int) bool {
			return n.topQueries[i].frequency > n.topQueries[j].frequency
		},
	)
}

func (n *Node) asQureryItem() *QueryItem {
	return &QueryItem{
		word:      n.word,
		frequency: n.frequency,
	}
}

type QueryItem struct {
	word      string
	frequency int
}

func (qi *QueryItem) Equals(other *QueryItem) bool {
	return qi.word == other.word && qi.frequency == other.frequency
}

func (qi *QueryItem) String() string {
	return fmt.Sprintf("%s: %d", qi.word, qi.frequency)
}
