package bp3

import (
	"cmp"
	"fmt"
	"slices"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/exp/constraints"
)

type testNodeDescriptor[K constraints.Ordered, V any] struct {
	id      string
	node    *Node[K, V]
	builder NodeBuilder[K, V]
	loader  NodeLoader[K, V]
}

func (d *testNodeDescriptor[K, V]) Read() *Node[K, V] {
	if d.node == nil {
		d.loader.Load(d)
	}

	return d.node
}

func (d *testNodeDescriptor[K, V]) Write() *Node[K, V] {
	node := d.Read()
	d.builder.Update(d)
	return node
}

type record[K constraints.Ordered, V any] struct {
	mins     []K
	children []string
	values   []KeyValue[K, V]
	next     string
	prev     string
}

type testNodeBuilder[K constraints.Ordered, V any] struct {
	update map[string]*Node[K, V]
	disk   map[string]*record[K, V]
	delete []string
}

func (b *testNodeBuilder[K, V]) Load(d NodeDescriptor[K, V]) {
	td := d.(*testNodeDescriptor[K, V])
	saved := b.disk[td.id]

	var children []NodeDescriptor[K, V]

	if len(saved.children) > 0 {
		children = make([]NodeDescriptor[K, V], 0, len(saved.children))

		for _, id := range saved.children {
			children = append(children, &testNodeDescriptor[K, V]{id: id, builder: td.builder, loader: td.loader})
		}
	}

	var next NodeDescriptor[K, V]

	if len(saved.next) > 0 {
		next = &testNodeDescriptor[K, V]{id: saved.next, builder: td.builder, loader: td.loader}
	}

	var prev NodeDescriptor[K, V]

	if len(saved.prev) > 0 {
		prev = &testNodeDescriptor[K, V]{id: saved.prev, builder: td.builder, loader: td.loader}
	}

	td.node = &Node[K, V]{
		Mins:     saved.mins,
		Values:   saved.values,
		Children: children,
		Next:     next,
		Prev:     prev,
	}
}

func (b *testNodeBuilder[K, V]) Create(node *Node[K, V]) NodeDescriptor[K, V] {
	d := &testNodeDescriptor[K, V]{
		id:      uuid.NewString(),
		node:    node,
		builder: b,
		loader:  b,
	}

	b.Update(d)

	return d
}

func (b *testNodeBuilder[K, V]) Update(d NodeDescriptor[K, V]) {
	td := d.(*testNodeDescriptor[K, V])
	b.update[td.id] = td.node
}

func (b *testNodeBuilder[K, V]) Flush() {
	if len(b.delete) > 0 {
		for _, id := range b.delete {
			delete(b.disk, id)
			delete(b.update, id)
		}
	}

	for id, node := range b.update {
		var children []string

		if len(node.Children) > 0 {
			children = make([]string, 0, len(node.Children))

			for _, cd := range node.Children {
				children = append(children, cd.(*testNodeDescriptor[K, V]).id)
			}
		}

		var next string

		if node.Next != nil {
			next = node.Next.(*testNodeDescriptor[K, V]).id
		}

		var prev string

		if node.Prev != nil {
			prev = node.Prev.(*testNodeDescriptor[K, V]).id
		}

		b.disk[id] = &record[K, V]{
			mins:     slices.Clone(node.Mins),
			values:   slices.Clone(node.Values),
			children: children,
			next:     next,
			prev:     prev,
		}
	}

	b.delete = nil
	clear(b.update)
}

func (b *testNodeBuilder[K, V]) Delete(d NodeDescriptor[K, V]) {
	td := d.(*testNodeDescriptor[K, V])
	b.delete = append(b.delete, td.id)
}

func TestTreeInsertSync(t *testing.T) {
	test := func(order int, n int) {
		builder := &testNodeBuilder[int, string]{
			disk:   make(map[string]*record[int, string]),
			update: make(map[string]*Node[int, string]),
		}

		tree := &Instance[int, string]{Order: order, Builder: builder}

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		builder.Flush()

		s0 := slice(tree.Root)

		loaded := &Instance[int, string]{
			Root: &testNodeDescriptor[int, string]{id: tree.Root.(*testNodeDescriptor[int, string]).id, builder: builder, loader: builder},
		}

		s1 := slice(loaded.Root)

		if slices.CompareFunc(s0, s1, func(a, b KeyValue[int, string]) int {
			return cmp.Compare(a.Key, b.Key)
		}) != 0 {
			t.Fatalf("%v, %v", s0, s1)
		}
	}

	test(3, 100)
	test(10, 1000)
	test(15, 10000)
}

func TestTreeInsertDeleteSync(t *testing.T) {
	test := func(order int, n int) {
		builder := &testNodeBuilder[int, string]{
			disk:   make(map[string]*record[int, string]),
			update: make(map[string]*Node[int, string]),
		}

		tree := &Instance[int, string]{Order: order, Builder: builder}

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		tree.Delete(5)
		tree.Delete(10)
		tree.Delete(20)

		builder.Flush()

		s0 := slice(tree.Root)

		loaded := &Instance[int, string]{
			Root: &testNodeDescriptor[int, string]{id: tree.Root.(*testNodeDescriptor[int, string]).id, builder: builder, loader: builder},
		}

		s1 := slice(loaded.Root)

		if slices.CompareFunc(s0, s1, func(a, b KeyValue[int, string]) int {
			return cmp.Compare(a.Key, b.Key)
		}) != 0 {
			t.Fatalf("%v, %v", s0, s1)
		}
	}

	test(3, 100)
	test(10, 1000)
	test(15, 10000)
}

func TestTreeInsertFlushInsertSync(t *testing.T) {
	test := func(order int, n int) {
		builder := &testNodeBuilder[int, string]{
			disk:   make(map[string]*record[int, string]),
			update: make(map[string]*Node[int, string]),
		}

		tree := &Instance[int, string]{Order: order, Builder: builder}

		for i := 0; i < n/2; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		builder.Flush()

		for i := n / 2; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		builder.Flush()

		s0 := slice(tree.Root)

		loaded := &Instance[int, string]{
			Root: &testNodeDescriptor[int, string]{id: tree.Root.(*testNodeDescriptor[int, string]).id, builder: builder, loader: builder},
		}

		s1 := slice(loaded.Root)

		if slices.CompareFunc(s0, s1, func(a, b KeyValue[int, string]) int {
			return cmp.Compare(a.Key, b.Key)
		}) != 0 {
			t.Fatalf("%v, %v", s0, s1)
		}
	}

	test(3, 100)
	test(10, 1000)
	test(15, 10000)
}
