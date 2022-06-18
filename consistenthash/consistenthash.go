package consistenthash

import (
	"sort"
	"strconv"
	"sync"
)

type Hash func(data []byte) uint32

// CH means Consistent Hash
type CH struct {
	hash      Hash
	nodesMap  map[string]*Node
	vnodes    []*VirtualNode
	vnodesMap map[uint32]*VirtualNode

	mu sync.RWMutex
}

func New(hash Hash, nodes ...*Node) *CH {
	nodesMap := make(map[string]*Node, len(nodes))
	for _, n := range nodes {
		nodesMap[n.name] = n
	}

	var vnodes []*VirtualNode
	for _, n := range nodes {
		vnodes = append(vnodes, n.VNodes(hash)...)
	}
	sort.Slice(
		vnodes,
		func(i, j int) bool { return vnodes[i].hash < vnodes[j].hash },
	)

	vnodesMap := make(map[uint32]*VirtualNode, len(vnodes))
	for _, vnode := range vnodes {
		vnodesMap[vnode.hash] = vnode
	}

	return &CH{
		hash:      hash,
		nodesMap:  nodesMap,
		vnodes:    vnodes,
		vnodesMap: vnodesMap,
	}
}

type RedistributeTask struct {
	KeyRange [2]uint32
	Source   *Node
	Target   *Node
}

func (ch *CH) AddNode(node *Node) []*RedistributeTask {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if _, ok := ch.nodesMap[node.name]; ok {
		return nil
	}

	ch.nodesMap[node.name] = node
	vnodes := node.VNodes(ch.hash)
	for _, vnode := range vnodes {
		ch.vnodesMap[vnode.hash] = vnode
	}
	ch.vnodes = append(ch.vnodes, vnodes...)
	sort.Slice(
		ch.vnodes,
		func(i, j int) bool { return ch.vnodes[i].hash < ch.vnodes[j].hash },
	)

	var tasks []*RedistributeTask

	for _, vnode := range vnodes {
		prevVnodeIndex, nextVnodeIndex := ch.prevAndNextVnodeIndexes(vnode.hash)
		task := &RedistributeTask{
			KeyRange: [2]uint32{ch.vnodes[prevVnodeIndex].hash, vnode.hash},
			Source:   ch.vnodes[nextVnodeIndex].node,
			Target:   node,
		}
		tasks = append(tasks, task)
	}

	return tasks
}

func (ch *CH) RemoveNode(name string) []*RedistributeTask {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	node, ok := ch.nodesMap[name]
	if !ok {
		return nil
	}

	vnodes := node.VNodes(ch.hash)

	var tasks []*RedistributeTask

	for _, vnode := range vnodes {
		prevVnodeIndex, nextVnodeIndex := ch.prevAndNextVnodeIndexes(vnode.hash)
		task := &RedistributeTask{
			KeyRange: [2]uint32{ch.vnodes[prevVnodeIndex].hash, vnode.hash},
			Source:   node,
			Target:   ch.vnodes[nextVnodeIndex].node,
		}
		tasks = append(tasks, task)
	}

	delete(ch.nodesMap, name)
	for _, vnode := range vnodes {
		delete(ch.vnodesMap, vnode.hash)
	}

	newVnodes := make([]*VirtualNode, 0, len(ch.vnodes)-len(vnodes))
	for _, vnode := range ch.vnodesMap {
		newVnodes = append(newVnodes, vnode)
	}
	ch.vnodes = newVnodes

	return tasks
}

// NodeOfKey returns the node that should owen the key
func (ch *CH) NodeOfKey(key []byte) *Node {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	index := sort.Search(
		len(ch.vnodes),
		func(i int) bool { return ch.vnodes[i].hash >= ch.hash(key) },
	)

	if index == len(ch.nodesMap) {
		index = 0
	}
	vnode := ch.vnodes[index]

	return vnode.node
}

func (ch *CH) prevAndNextVnodeIndexes(hash uint32) (int, int) {
	index := sort.Search(
		len(ch.vnodes),
		func(i int) bool { return ch.vnodes[i].hash >= hash },
	)

	prevVnodeIndex := index - 1
	if prevVnodeIndex < 0 {
		prevVnodeIndex = len(ch.vnodes) - 1
	}

	nextVnodeIndex := index + 1
	if nextVnodeIndex >= len(ch.vnodes) {
		nextVnodeIndex = 0
	}

	return prevVnodeIndex, nextVnodeIndex
}

type Node struct {
	name     string
	vnodeNum int
}

func NewNode(name string, vnodeNum int) *Node {
	return &Node{
		name:     name,
		vnodeNum: vnodeNum,
	}
}

func (n *Node) Name() string {
	return n.name
}

func (n *Node) VNodes(hash Hash) []*VirtualNode {
	var vnodes []*VirtualNode

	for i := 0; i < n.vnodeNum; i++ {
		vNodeName := n.name + "-" + strconv.Itoa(i)

		vnodes = append(vnodes, &VirtualNode{
			node: n,
			hash: hash([]byte(vNodeName)),
		})
	}

	return vnodes
}

type VirtualNode struct {
	node *Node
	// (, hash]
	hash uint32
}
