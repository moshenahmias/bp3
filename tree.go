package bp3

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"golang.org/x/exp/constraints"
)

type keyValue[K constraints.Ordered, V any] struct {
	key   K
	value V
}

type node[K constraints.Ordered, V any] struct {
	mins     []K
	children []*node[K, V]
	values   []keyValue[K, V]
}

func (n *node[K, V]) count() int {
	return len(n.children) + len(n.values)
}

func (n *node[K, V]) leaf() bool {
	return len(n.values) > 0
}

func insert[K constraints.Ordered, V any](root *node[K, V], minimum K, item keyValue[K, V], order int) (*node[K, V], K, *node[K, V], K) {

	if root == nil {
		kv := keyValue[K, V]{key: item.key, value: item.value}
		return &node[K, V]{values: []keyValue[K, V]{kv}}, item.key, nil, *new(K)
	}

	if root.leaf() {
		kv := keyValue[K, V]{key: item.key, value: item.value}

		i, found := slices.BinarySearchFunc(root.values, kv, func(a, b keyValue[K, V]) int {
			return cmp.Compare(a.key, b.key)
		})

		if found {
			root.values[i] = kv
		} else {
			root.values = slices.Insert(root.values, i, kv)
		}

		count := len(root.values)

		if count <= order {
			return root, root.values[0].key, nil, *new(K)
		}

		brother := &node[K, V]{values: slices.Clone(root.values[count/2:])}

		root.values = root.values[:count/2]

		return root, root.values[0].key, brother, brother.values[0].key
	}

	parentIdx, found := slices.BinarySearch(root.mins, item.key)

	if !found {
		parentIdx = parentIdx - 1
	}

	parentMin := minimum

	if parentIdx > -1 {
		parentMin = root.mins[parentIdx]
	}

	parent := root.children[parentIdx+1]

	_, parentMin, split, splitMin := insert(parent, parentMin, item, order)

	if split == nil {
		if parentIdx > -1 {

			if parentIdx == 0 {
				root.mins[0] = parentMin
			}

			return root, minimum, nil, *new(K)
		}

		return root, min(minimum, parentMin), nil, *new(K)
	}

	root.children = slices.Insert(root.children, parentIdx+2, split)
	root.mins = slices.Insert(root.mins, parentIdx+1, splitMin)

	if parentIdx < 0 {
		root.mins[0] = splitMin
	}

	count := len(root.children)

	if count <= order {
		return root, min(minimum, parentMin), nil, *new(K)
	}

	brother := &node[K, V]{children: slices.Clone(root.children[count/2:])}

	root.children = root.children[:count/2]

	c := len(root.mins)

	var brotherMin K

	if c%2 == 0 {
		brotherMin = root.mins[(c/2)-1]
		brother.mins = slices.Clone(root.mins[c/2:])
		root.mins = root.mins[:(c/2)-1]
	} else {
		brotherMin = root.mins[c/2]
		brother.mins = slices.Clone(root.mins[(c/2)+1:])
		root.mins = root.mins[:c/2]
	}

	return root, min(minimum, parentMin), brother, brotherMin
}

func find[K constraints.Ordered, V any](root *node[K, V], key K) (V, bool) {
	if root == nil {
		return *new(V), false
	}

	if root.leaf() {
		i, found := slices.BinarySearchFunc(root.values, key, func(a keyValue[K, V], b K) int {
			return cmp.Compare(a.key, b)
		})

		if !found {
			return *new(V), false
		}

		return root.values[i].value, true
	}

	i, found := slices.BinarySearch(root.mins, key)

	if found {
		return find(root.children[i+1], key)
	}

	return find(root.children[i], key)
}

func delete[K constraints.Ordered, V any](root *node[K, V], key K, minimum K, order int) (V, bool, K) {
	if root == nil {
		return *new(V), false, minimum
	}

	if root.leaf() {
		i, found := slices.BinarySearchFunc(root.values, key, func(a keyValue[K, V], b K) int {
			return cmp.Compare(a.key, b)
		})

		if !found {
			return *new(V), false, minimum
		}

		v := root.values[i].value
		root.values = slices.Delete(root.values, i, i+1)

		if len(root.values) == 0 {
			return v, true, *new(K)
		}

		return v, true, root.values[0].key
	}

	parentIdx, found := slices.BinarySearch(root.mins, key)

	if !found {
		parentIdx = parentIdx - 1
	}

	parent := root.children[parentIdx+1]
	parentMin := minimum

	if parentIdx > -1 {
		parentMin = root.mins[parentIdx]
	}

	v, deleted, parentMin := delete(parent, key, parentMin, order)

	if !deleted {
		return v, deleted, minimum
	}

	newMin := minimum

	if parentIdx < 0 {
		newMin = parentMin
	}

	childMin := int(math.Ceil(float64(order) / 2))
	parentCount := parent.count()

	if childMin <= parentCount {
		return v, deleted, newMin
	}

	var rightUncle, leftUncle *node[K, V]
	var uncleCount int
	var uncleIdx int

	if parentIdx < 0 {
		uncleIdx = parentIdx + 1
		rightUncle = root.children[uncleIdx+1]
		uncleCount = rightUncle.count()
	} else {
		uncleIdx = parentIdx - 1
		leftUncle = root.children[uncleIdx+1]
		uncleCount = leftUncle.count()
	}

	if parent.leaf() {
		if uncleCount > childMin {
			// transfer child from uncle to parent

			if leftUncle != nil {
				// from left to right
				kv := leftUncle.values[uncleCount-1]
				parent.values = slices.Insert(parent.values, 0, kv)
				leftUncle.values = leftUncle.values[:uncleCount-1]
				parentMin = kv.key
				root.mins[parentIdx] = parentMin
			} else {
				// from right to left
				parent.values = append(parent.values, rightUncle.values[0])
				rightUncle.values = rightUncle.values[1:]
				root.mins[uncleIdx] = rightUncle.values[0].key
			}

			return v, deleted, newMin
		}

		// transfer all children from parent to uncle
		// delete parent from root

		if leftUncle != nil {
			// from right parent to left uncle
			leftUncle.values = append(leftUncle.values, parent.values...)
		} else {
			// from left parent to right uncle
			rightUncle.values = append(parent.values, rightUncle.values...)
			root.mins[uncleIdx] = rightUncle.values[0].key
		}

		parent.values = nil
		root.children = slices.Delete(root.children, parentIdx+1, parentIdx+2)

		if parentIdx < 0 {
			root.mins = root.mins[1:]
		} else {
			root.mins = slices.Delete(root.mins, parentIdx, parentIdx+1)
		}

		return v, deleted, newMin
	}

	// not a leaf

	if uncleCount > childMin {
		// transfer child from uncle to parent

		if leftUncle != nil {
			// from left to right
			kv := leftUncle.children[uncleCount-1]
			parent.children = slices.Insert(parent.children, 0, kv)
			leftUncle.children = leftUncle.children[:uncleCount-1]
			parent.mins = slices.Insert(parent.mins, 0, parentMin)
			parentMin = leftUncle.mins[uncleCount-2]
			leftUncle.mins = leftUncle.mins[:uncleCount-2]
			root.mins[parentIdx] = parentMin
		} else {
			// from right to left
			parent.children = append(parent.children, rightUncle.children[0])
			rightUncle.children = rightUncle.children[1:]
			parent.mins = append(parent.mins, root.mins[uncleIdx])
			root.mins[uncleIdx] = rightUncle.mins[0]
			rightUncle.mins = rightUncle.mins[1:]
		}

		return v, deleted, newMin
	}

	// transfer all children from parent to uncle
	// delete parent from root

	if leftUncle != nil {
		// from right parent to left uncle
		leftUncle.children = append(leftUncle.children, parent.children...)
		leftUncle.mins = append(append(leftUncle.mins, parentMin), parent.mins...)
	} else {
		// from left parent to right uncle
		rightUncle.children = append(parent.children, rightUncle.children...)
		rightUncle.mins = append(append(parent.mins, root.mins[uncleIdx]), rightUncle.mins...)
		root.mins[uncleIdx] = parentMin
	}

	parent.children = nil
	root.children = slices.Delete(root.children, parentIdx+1, parentIdx+2)

	if parentIdx < 0 {
		root.mins = root.mins[1:]
	} else {
		root.mins = slices.Delete(root.mins, parentIdx, parentIdx+1)
	}

	return v, deleted, newMin
}

type tree[K constraints.Ordered, V any] struct {
	root  *node[K, V]
	min   K
	order int
	size  int
}

func New[K constraints.Ordered, V any](order int) tree[K, V] {
	if order < 3 {
		panic(fmt.Sprintf("bp3: invalid tree order (%d)", order))
	}

	return tree[K, V]{order: order}
}

func (t *tree[K, V]) Insert(key K, value V) {
	kv := keyValue[K, V]{key: key, value: value}
	child, minimum, brother, brotherMin := insert(t.root, t.min, kv, t.order)
	t.min = minimum

	if brother == nil {
		t.root = child
	} else {
		children := []*node[K, V]{child, brother}
		mins := []K{brotherMin}
		t.root = &node[K, V]{children: children, mins: mins}
	}

	t.size++
}

func (t *tree[K, V]) Find(key K) (V, bool) {
	return find(t.root, key)
}

func (t *tree[K, V]) Delete(key K) (V, bool) {
	v, deleted, newMin := delete(t.root, key, t.min, t.order)

	if deleted {
		t.size--
		t.min = newMin

		if len(t.root.children) == 1 {
			t.root = t.root.children[0]
		} else if t.root.count() == 0 {
			t.root = nil
		}
	}

	return v, deleted
}

func (t *tree[K, V]) Size() int {
	return t.size
}

func (t *tree[K, V]) Empty() bool {
	return t.Size() == 0
}
