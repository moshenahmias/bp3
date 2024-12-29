package bp3

import "golang.org/x/exp/constraints"

// KeyValue represents a key-value pair
// The generic parameters K and V are for the key and value types, respectively, where K must be ordered.
type KeyValue[K constraints.Ordered, V any] struct {
	Key   K
	Value V
}

// Node represents a node in the tree. It contains minimum keys, children node descriptors,
// key-value pairs, and pointers to the next and previous nodes. The generic parameters K and V are
// for the key and value types, respectively, where K must be ordered.
type Node[K constraints.Ordered, V any] struct {
	Mins     []K                    // Mins are the minimum keys for each child node, starting from the 2nd child.
	Children []NodeDescriptor[K, V] // Children are the descriptors for the child nodes.
	Values   []KeyValue[K, V]       // Values are the key-value pairs stored in the node.
	Next     NodeDescriptor[K, V]   // Next is the descriptor for the next node.
	Prev     NodeDescriptor[K, V]   // Prev is the descriptor for the previous node.
}

// Count returns the total number of children in the node.
func (n *Node[K, V]) Count() int {
	return len(n.Children) + len(n.Values)
}

// Leaf returns true if the node is a leaf node (i.e., contains values); otherwise, it returns false.
func (n *Node[K, V]) Leaf() bool {
	return len(n.Values) > 0
}

// NodeDescriptor is an interface that defines methods for reading and writing nodes.
// The generic parameters K and V are for the key and value types, respectively, where K must be ordered.
type NodeDescriptor[K constraints.Ordered, V any] interface {
	Read() *Node[K, V]  // Read returns a pointer to the node.
	Write() *Node[K, V] // Write returns a pointer to the node for modification.
}

// NodeLoader is an interface that defines a method for loading node descriptors.
// The generic parameters K and V are for the key and value types, respectively, where K must be ordered.
type NodeLoader[K constraints.Ordered, V any] interface {
	Load(d NodeDescriptor[K, V]) // Load loads the specified node descriptor.
}

// NodeBuilder is an interface that defines methods for creating, updating, deleting, and flushing nodes.
// The generic parameters K and V are for the key and value types, respectively, where K must be ordered.
type NodeBuilder[K constraints.Ordered, V any] interface {
	Create(node *Node[K, V]) NodeDescriptor[K, V] // Create creates a new node descriptor for the given node.
	Update(d NodeDescriptor[K, V])                // Update updates the specified node descriptor.
	Delete(d NodeDescriptor[K, V])                // Delete deletes the specified node descriptor.
	Flush()                                       // Flush flushes any pending changes.
}
