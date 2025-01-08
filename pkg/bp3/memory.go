package bp3

import (
	"golang.org/x/exp/constraints"
)

type memoryNodeDescriptor[K constraints.Ordered, V any] struct {
	node *Node[K, V]
}

func (d *memoryNodeDescriptor[K, V]) Read() *Node[K, V] {
	return d.node
}

func (d *memoryNodeDescriptor[K, V]) Write() *Node[K, V] {
	return d.node
}

type memoryBuilder[K constraints.Ordered, V any] struct{}

func (*memoryBuilder[K, V]) Create(node *Node[K, V]) NodeDescriptor[K, V] {
	return &memoryNodeDescriptor[K, V]{node}
}

func (*memoryBuilder[K, V]) Update(d NodeDescriptor[K, V]) {}

func (*memoryBuilder[K, V]) Delete(d NodeDescriptor[K, V]) {}

func (*memoryBuilder[K, V]) Flush() error {
	return nil
}

// New creates a new instance of b+tree structure with the specified order.
func New[K constraints.Ordered, V any](order int) *Instance[K, V] {
	return &Instance[K, V]{Order: max(order, MinOrder), Builder: &memoryBuilder[K, V]{}}
}
