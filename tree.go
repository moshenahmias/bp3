package bp3

import (
	"cmp"
	"fmt"
	"iter"
	"math"
	"slices"

	"golang.org/x/exp/constraints"
)

// Instance represents a b+tree instance structure that contains a root node descriptor,
// a minimum value, the order of the structure, its size, and a node builder. The generic
// parameters K and V are for the key and value types, respectively.
type Instance[K constraints.Ordered, V any] struct {
	Root    NodeDescriptor[K, V] // Root is the descriptor for the root node.
	Min     K                    // Min is the minimum key value in the instance.
	Order   int                  // Order is the order of the structure.
	Size    int                  // Size is the number of elements in the instance.
	Builder NodeBuilder[K, V]    // Builder is used to create new nodes within the instance.
}

// New creates a new instance of b+tree structure with the specified order.
// The function ensures that the order is at least 3, and if it is not, it will panic
// with an appropriate error message.
func New[K constraints.Ordered, V any](order int) *Instance[K, V] {
	if order < 3 {
		panic(fmt.Sprintf("bp3: invalid tree order (%d)", order))
	}

	return &Instance[K, V]{Order: order, Builder: &defaultBuilder[K, V]{}}
}

func (t *Instance[K, V]) Insert(key K, value V) {
	kv := KeyValue[K, V]{Key: key, Value: value}
	child, minimum, brother, brotherMin := t.insert(t.Root, t.Min, kv)
	t.Min = minimum

	if brother == nil {
		t.Root = child
	} else {
		children := []NodeDescriptor[K, V]{child, brother}
		mins := []K{brotherMin}
		t.Root = t.Builder.Create(&Node[K, V]{Children: children, Mins: mins})
	}

	t.Size++
}

func (t *Instance[K, V]) Find(key K) (V, bool) {
	if node, i, found := find(t.Root, key); found {
		return node.Read().Values[i].Value, true
	}

	return *new(V), false
}

// RangeValue represents a range with a value and a flag indicating whether the range is closed.
// The generic parameter K is for the key type which must be ordered.
type RangeValue[K constraints.Ordered] struct {
	Value  K
	Closed bool
}

// Range returns a sequence of key-value pairs within the specified range.
// The range is defined by the 'from' and 'to' RangeValue parameters.
func (t *Instance[K, V]) Range(from, to RangeValue[K]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if to.Value < from.Value {
			return
		}

		node, i, _ := find(t.Root, from.Value)

		for node != nil && node.Read() != nil {
			for i < len(node.Read().Values) {
				key := node.Read().Values[i].Key

				if key > to.Value {
					return
				}

				if !to.Closed && key == to.Value {
					return
				}

				if !(key < from.Value || (!from.Closed && key == from.Value)) {
					if !yield(key, node.Read().Values[i].Value) {
						return
					}
				}

				i++
			}

			i = 0
			node = node.Read().Next
		}
	}
}

// RangeClosed returns a sequence of key-value pairs within the specified closed range [from, to].
func (t *Instance[K, V]) RangeClosed(from, to K) iter.Seq2[K, V] {
	return t.Range(RangeValue[K]{from, true}, RangeValue[K]{to, true})
}

// RangeOpened returns a sequence of key-value pairs within the specified open range (from, to).
func (t *Instance[K, V]) RangeOpened(from, to K) iter.Seq2[K, V] {
	return t.Range(RangeValue[K]{from, false}, RangeValue[K]{to, false})
}

// RangeLowHalfOpened returns a sequence of key-value pairs within the specified range (from, to].
func (t *Instance[K, V]) RangeLowHalfOpened(from, to K) iter.Seq2[K, V] {
	return t.Range(RangeValue[K]{from, false}, RangeValue[K]{to, true})
}

// RangeHighHalfOpened returns a sequence of key-value pairs within the specified range [from, to).
func (t *Instance[K, V]) RangeHighHalfOpened(from, to K) iter.Seq2[K, V] {
	return t.Range(RangeValue[K]{from, true}, RangeValue[K]{to, false})
}

// From returns a sequence of key-value pairs starting from the specified range value.
func (t *Instance[K, V]) From(from RangeValue[K]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		node, i, _ := find(t.Root, from.Value)

		for node != nil && node.Read() != nil {
			for i < len(node.Read().Values) {
				key := node.Read().Values[i].Key

				if !(key < from.Value || (!from.Closed && key == from.Value)) {
					if !yield(key, node.Read().Values[i].Value) {
						return
					}
				}

				i++
			}

			i = 0
			node = node.Read().Next
		}
	}
}

// FromClosed returns a sequence of key-value pairs starting from the specified key, including the key itself.
func (t *Instance[K, V]) FromClosed(from K) iter.Seq2[K, V] {
	return t.From(RangeValue[K]{from, true})
}

// FromOpened returns a sequence of key-value pairs starting from the specified key, excluding the key itself.
func (t *Instance[K, V]) FromOpened(from K) iter.Seq2[K, V] {
	return t.From(RangeValue[K]{from, false})
}

// To returns a sequence of key-value pairs up to the specified range value.
func (t *Instance[K, V]) To(to RangeValue[K]) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		node := minimum(t.Root)
		i := 0

		for node != nil && node.Read() != nil {
			for i < len(node.Read().Values) {
				key := node.Read().Values[i].Key

				if key > to.Value {
					return
				}

				if !to.Closed && key == to.Value {
					return
				}

				if !yield(key, node.Read().Values[i].Value) {
					return
				}

				i++
			}

			i = 0
			node = node.Read().Next
		}
	}
}

// ToClosed returns a sequence of key-value pairs up to and including the specified key.
func (t *Instance[K, V]) ToClosed(to K) iter.Seq2[K, V] {
	return t.To(RangeValue[K]{to, true})
}

// ToOpened returns a sequence of key-value pairs up to but excluding the specified key.
func (t *Instance[K, V]) ToOpened(to K) iter.Seq2[K, V] {
	return t.To(RangeValue[K]{to, false})
}

// Delete removes the key-value pair associated with the specified key from the instance.
// It returns the value associated with the deleted key and a boolean indicating whether the key was found and deleted.
func (t *Instance[K, V]) Delete(key K) (V, bool) {
	v, deleted, newMin := t.delete(t.Root, key, t.Min)

	if deleted {
		t.Size--
		t.Min = newMin

		if len(t.Root.Read().Children) == 1 {
			t.Root = t.Root.Read().Children[0]
		} else if t.Root.Read().Count() == 0 {
			t.Root = nil
		}
	}

	return v, deleted
}

// Count returns the number of elements in the instance.
func (t *Instance[K, V]) Count() int {
	return t.Size
}

// Empty returns true if the instance is empty; otherwise, it returns false.
func (t *Instance[K, V]) Empty() bool {
	return t.Count() == 0
}

// Minimum returns the minimum value in the instance.
// It panics if the instance is empty.
func (t *Instance[K, V]) Minimum() V {
	if node := minimum(t.Root); node != nil && node.Read() != nil {
		return node.Read().Values[0].Value
	}

	panic("bp3: empty tree")
}

// Maximum returns the maximum value in the instance.
// It panics if the instance is empty.
func (t *Instance[K, V]) Maximum() V {
	if node := maximum(t.Root); node != nil && node.Read() != nil {
		return node.Read().Values[len(node.Read().Values)-1].Value
	}

	panic("bp3: empty tree")
}

func (t *Instance[K, V]) insert(root NodeDescriptor[K, V], minimum K, item KeyValue[K, V]) (NodeDescriptor[K, V], K, NodeDescriptor[K, V], K) {

	if root == nil || root.Read() == nil {
		kv := KeyValue[K, V]{Key: item.Key, Value: item.Value}
		return t.Builder.Create(&Node[K, V]{Values: []KeyValue[K, V]{kv}}), item.Key, nil, *new(K)
	}

	if root.Read().Leaf() {
		kv := KeyValue[K, V]{Key: item.Key, Value: item.Value}

		i, found := slices.BinarySearchFunc(root.Read().Values, kv, func(a, b KeyValue[K, V]) int {
			return cmp.Compare(a.Key, b.Key)
		})

		if found {
			root.Write().Values[i] = kv
		} else {
			root.Write().Values = slices.Insert(root.Read().Values, i, kv)
		}

		count := len(root.Read().Values)

		if count <= t.Order {
			return root, root.Read().Values[0].Key, nil, *new(K)
		}

		brother := t.Builder.Create(&Node[K, V]{Values: slices.Clone(root.Read().Values[count/2:])})

		root.Write().Values = root.Read().Values[:count/2]

		if root.Read().Next != nil {
			brother.Read().Next = root.Read().Next
			root.Read().Next.Write().Prev = brother
		}

		brother.Read().Prev = root
		root.Write().Next = brother

		return root, root.Read().Values[0].Key, brother, brother.Read().Values[0].Key
	}

	parentIdx, found := slices.BinarySearch(root.Read().Mins, item.Key)

	if !found {
		parentIdx = parentIdx - 1
	}

	parentMin := minimum

	if parentIdx > -1 {
		parentMin = root.Read().Mins[parentIdx]
	}

	parent := root.Read().Children[parentIdx+1]

	_, parentMin, split, splitMin := t.insert(parent, parentMin, item)

	if split == nil {
		if parentIdx > -1 {

			if parentIdx == 0 {
				root.Write().Mins[0] = parentMin
			}

			return root, minimum, nil, *new(K)
		}

		return root, min(minimum, parentMin), nil, *new(K)
	}

	root.Write().Children = slices.Insert(root.Read().Children, parentIdx+2, split)
	root.Write().Mins = slices.Insert(root.Read().Mins, parentIdx+1, splitMin)

	if parentIdx < 0 {
		root.Write().Mins[0] = splitMin
	}

	count := len(root.Read().Children)

	if count <= t.Order {
		return root, min(minimum, parentMin), nil, *new(K)
	}

	brother := t.Builder.Create(&Node[K, V]{Children: slices.Clone(root.Read().Children[count/2:])})

	root.Write().Children = root.Read().Children[:count/2]

	c := len(root.Read().Mins)

	var brotherMin K

	if c%2 == 0 {
		brotherMin = root.Read().Mins[(c/2)-1]
		brother.Read().Mins = slices.Clone(root.Read().Mins[c/2:])
		root.Write().Mins = root.Read().Mins[:(c/2)-1]
	} else {
		brotherMin = root.Read().Mins[c/2]
		brother.Read().Mins = slices.Clone(root.Read().Mins[(c/2)+1:])
		root.Write().Mins = root.Read().Mins[:c/2]
	}

	return root, min(minimum, parentMin), brother, brotherMin
}

func (t *Instance[K, V]) delete(root NodeDescriptor[K, V], key K, minimum K) (V, bool, K) {
	if root == nil {
		return *new(V), false, minimum
	}

	if root.Read().Leaf() {
		i, found := slices.BinarySearchFunc(root.Read().Values, key, func(a KeyValue[K, V], b K) int {
			return cmp.Compare(a.Key, b)
		})

		if !found {
			return *new(V), false, minimum
		}

		v := root.Read().Values[i].Value
		root.Write().Values = slices.Delete(root.Read().Values, i, i+1)

		if len(root.Read().Values) == 0 {
			return v, true, *new(K)
		}

		return v, true, root.Read().Values[0].Key
	}

	parentIdx, found := slices.BinarySearch(root.Read().Mins, key)

	if !found {
		parentIdx = parentIdx - 1
	}

	parent := root.Read().Children[parentIdx+1]
	parentMin := minimum

	if parentIdx > -1 {
		parentMin = root.Read().Mins[parentIdx]
	}

	v, deleted, parentMin := t.delete(parent, key, parentMin)

	if !deleted {
		return v, deleted, minimum
	}

	newMin := minimum

	if parentIdx < 0 {
		newMin = parentMin
	}

	childMin := int(math.Ceil(float64(t.Order) / 2))
	parentCount := parent.Read().Count()

	if childMin <= parentCount {
		return v, deleted, newMin
	}

	var rightUncle, leftUncle NodeDescriptor[K, V]
	var uncleCount int
	var uncleIdx int

	if parentIdx < 0 {
		uncleIdx = parentIdx + 1
		rightUncle = root.Read().Children[uncleIdx+1]
		uncleCount = rightUncle.Read().Count()
	} else {
		uncleIdx = parentIdx - 1
		leftUncle = root.Read().Children[uncleIdx+1]
		uncleCount = leftUncle.Read().Count()
	}

	if parent.Read().Leaf() {
		if uncleCount > childMin {
			// transfer child from uncle to parent

			if leftUncle != nil {
				// from left to right
				kv := leftUncle.Read().Values[uncleCount-1]
				parent.Write().Values = slices.Insert(parent.Read().Values, 0, kv)
				leftUncle.Write().Values = leftUncle.Read().Values[:uncleCount-1]
				parentMin = kv.Key
				root.Write().Mins[parentIdx] = parentMin
			} else {
				// from right to left
				parent.Write().Values = append(parent.Read().Values, rightUncle.Read().Values[0])
				rightUncle.Write().Values = rightUncle.Read().Values[1:]
				root.Write().Mins[uncleIdx] = rightUncle.Read().Values[0].Key
			}

			return v, deleted, newMin
		}

		// transfer all children from parent to uncle
		// delete parent from root

		if leftUncle != nil {
			// from right parent to left uncle
			leftUncle.Write().Values = append(leftUncle.Read().Values, parent.Read().Values...)
		} else {
			// from left parent to right uncle
			rightUncle.Write().Values = append(parent.Read().Values, rightUncle.Read().Values...)
			root.Write().Mins[uncleIdx] = rightUncle.Read().Values[0].Key
		}

		parent.Write().Values = nil

		if parent.Read().Prev != nil && parent.Read().Next != nil {
			// middle
			parent.Write().Prev, parent.Read().Next.Write().Prev = parent.Read().Next, parent.Read().Prev
		} else if parent.Read().Prev != nil {
			// last
			parent.Read().Prev.Write().Next = nil
		} else if parent.Read().Next != nil {
			// first
			parent.Read().Next.Write().Prev = nil
		}

		t.Builder.Delete(parent)
		root.Write().Children = slices.Delete(root.Read().Children, parentIdx+1, parentIdx+2)

		if parentIdx < 0 {
			root.Write().Mins = root.Read().Mins[1:]
		} else {
			root.Write().Mins = slices.Delete(root.Read().Mins, parentIdx, parentIdx+1)
		}

		return v, deleted, newMin
	}

	// not a leaf

	if uncleCount > childMin {
		// transfer child from uncle to parent

		if leftUncle != nil {
			// from left to right
			kv := leftUncle.Read().Children[uncleCount-1]
			parent.Write().Children = slices.Insert(parent.Read().Children, 0, kv)
			leftUncle.Write().Children = leftUncle.Read().Children[:uncleCount-1]
			parent.Write().Mins = slices.Insert(parent.Read().Mins, 0, parentMin)
			parentMin = leftUncle.Read().Mins[uncleCount-2]
			leftUncle.Write().Mins = leftUncle.Read().Mins[:uncleCount-2]
			root.Write().Mins[parentIdx] = parentMin
		} else {
			// from right to left
			parent.Write().Children = append(parent.Read().Children, rightUncle.Read().Children[0])
			rightUncle.Write().Children = rightUncle.Read().Children[1:]
			parent.Write().Mins = append(parent.Read().Mins, root.Read().Mins[uncleIdx])
			root.Write().Mins[uncleIdx] = rightUncle.Read().Mins[0]
			rightUncle.Write().Mins = rightUncle.Read().Mins[1:]
		}

		return v, deleted, newMin
	}

	// transfer all children from parent to uncle
	// delete parent from root

	if leftUncle != nil {
		// from right parent to left uncle
		leftUncle.Write().Children = append(leftUncle.Read().Children, parent.Read().Children...)
		leftUncle.Write().Mins = append(append(leftUncle.Read().Mins, parentMin), parent.Read().Mins...)
	} else {
		// from left parent to right uncle
		rightUncle.Write().Children = append(parent.Read().Children, rightUncle.Read().Children...)
		rightUncle.Write().Mins = append(append(parent.Read().Mins, root.Read().Mins[uncleIdx]), rightUncle.Read().Mins...)
		root.Write().Mins[uncleIdx] = parentMin
	}

	parent.Read().Children = nil
	t.Builder.Delete(parent)
	root.Write().Children = slices.Delete(root.Read().Children, parentIdx+1, parentIdx+2)

	if parentIdx < 0 {
		root.Write().Mins = root.Read().Mins[1:]
	} else {
		root.Write().Mins = slices.Delete(root.Read().Mins, parentIdx, parentIdx+1)
	}

	return v, deleted, newMin
}

func find[K constraints.Ordered, V any](root NodeDescriptor[K, V], key K) (NodeDescriptor[K, V], int, bool) {
	if root == nil {
		return nil, *new(int), false
	}

	if root.Read().Leaf() {
		i, found := slices.BinarySearchFunc(root.Read().Values, key, func(a KeyValue[K, V], b K) int {
			return cmp.Compare(a.Key, b)
		})

		return root, i, found
	}

	i, found := slices.BinarySearch(root.Read().Mins, key)

	if found {
		return find(root.Read().Children[i+1], key)
	}

	return find(root.Read().Children[i], key)
}

func minimum[K constraints.Ordered, V any](root NodeDescriptor[K, V]) NodeDescriptor[K, V] {
	if root == nil || root.Read().Leaf() {
		return root
	}

	return minimum(root.Read().Children[0])
}

func maximum[K constraints.Ordered, V any](root NodeDescriptor[K, V]) NodeDescriptor[K, V] {
	if root == nil || root.Read().Leaf() {
		return root
	}

	return maximum(root.Read().Children[len(root.Read().Children)-1])
}
