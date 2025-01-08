package bp3

import "golang.org/x/exp/constraints"

// NodeDescriptor defines methods for reading and writing nodes.
// The generic parameters K and V are for the key and value types, respectively, where K must be ordered.
type NodeDescriptor[K constraints.Ordered, V any] interface {
	Read() *Node[K, V]  // Read returns a pointer to the node.
	Write() *Node[K, V] // Write returns a pointer to the node for modification.
}

// NodeLoader defines a method for loading node descriptors.
// The generic parameters K and V are for the key and value types, respectively, where K must be ordered.
type NodeLoader[K constraints.Ordered, V any] interface {
	Load(d NodeDescriptor[K, V]) error // Load loads the specified node descriptor.
}

// NodeBuilder defines methods for creating, updating, and deleting nodes.
// The generic parameters K and V are for the key and value types, respectively, where K must be ordered.
type NodeBuilder[K constraints.Ordered, V any] interface {
	Create(node *Node[K, V]) NodeDescriptor[K, V] // Create creates a new node descriptor for the given node.
	Update(d NodeDescriptor[K, V])                // Update updates the specified node descriptor.
	Delete(d NodeDescriptor[K, V])                // Delete deletes the specified node descriptor.
	Flush() error                                 // Flush flushes any pending changes.
}
