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

// New creates a new instance of b+tree structure with the specified options.
func New[K constraints.Ordered, V any](options ...Option) *Instance[K, V] {
	opts := buildOptions(options...)
	order := max(opts.order, MinOrder)
	return &Instance[K, V]{Order: order, Builder: &memoryBuilder[K, V]{}}
}

// Clear removes all key-value pairs from the B+ Tree, resetting its state.
func Clear[K constraints.Ordered, V any](tree *Instance[K, V]) {
	if _, ok := tree.Builder.(*memoryBuilder[K, V]); !ok {
		panic("bp3: invalid tree instance")
	}

	tree.Root = nil
	tree.Min = *new(K)
	tree.Size = 0
}
