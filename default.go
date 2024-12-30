package bp3

import "golang.org/x/exp/constraints"

type defaultNodeDescriptor[K constraints.Ordered, V any] struct {
	node *Node[K, V]
}

func (d *defaultNodeDescriptor[K, V]) Read() *Node[K, V] {
	return d.node
}

func (d *defaultNodeDescriptor[K, V]) Write() *Node[K, V] {
	return d.node
}

type defaultBuilder[K constraints.Ordered, V any] struct{}

func (*defaultBuilder[K, V]) Create(node *Node[K, V]) NodeDescriptor[K, V] {
	return &defaultNodeDescriptor[K, V]{node}
}

func (*defaultBuilder[K, V]) Update(d NodeDescriptor[K, V]) {}

func (*defaultBuilder[K, V]) Delete(d NodeDescriptor[K, V]) {}

func (*defaultBuilder[K, V]) Flush() {}
