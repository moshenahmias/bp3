package bp3

import (
	"fmt"
	"testing"

	"golang.org/x/exp/constraints"
)

func slice[K constraints.Ordered, V any](root *node[K, V]) []keyValue[K, V] {
	if root == nil {
		return nil
	}

	if root.leaf() {
		return root.values
	}

	var s []keyValue[K, V]

	for _, child := range root.children {
		s = append(s, slice(child)...)
	}

	return s
}

func TestInsert(t *testing.T) {
	test := func(order int, n int) {
		tree := New[int, string](order)

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		if size := tree.Size(); size != n {
			t.Fatalf("size %d != %d", size, n)
		}

		s := slice(tree.root)

		if len(s) != n {
			t.Fatalf("slice size %d != %d", len(s), n)
		}

		for i := 0; i < n; i++ {
			if v, found := tree.Find(i); !found || fmt.Sprint(i) != v {

			}

			if s[i].key != i || s[i].value != fmt.Sprint(i) {
				t.Fatalf("%d not in position", i)
			}
		}

	}

	test(3, 100)
	test(10, 1000)
	test(15, 10000)
}

func TestDelete(t *testing.T) {
	test := func(order int, n int, del map[int]bool) {
		tree := New[int, string](order)

		for i := 0; i < n; i++ {
			if delete, found := del[i]; !found || delete {
				tree.Insert(i, fmt.Sprint(i))
			}
		}

		for i, delete := range del {
			if delete {
				if _, deleted := tree.Delete(i); !deleted {
					t.Fatalf("failed to delete %d", i)

				}
			}
		}

		if size := tree.Size(); size != n-len(del) {
			t.Fatalf("size %d != %d", size, n-len(del))
		}

		for i := 0; i < n; i++ {
			_, deleted := del[i]
			v, found := tree.Find(i)

			if !deleted {
				if !found {
					t.Fatalf("%d not deleted, but not found", i)
				} else if fmt.Sprint(i) != v {
					t.Fatalf("%d not deleted, but got wrong value %s", i, v)
				}
			} else if found {
				t.Fatalf("%d deleted, but found", i)
			}
		}
	}

	test(3, 100, map[int]bool{0: true, 1: true, 3: true})
	test(10, 1000, map[int]bool{4: true, 605: true, 900: true})
	test(15, 10000, map[int]bool{10: true, 20: true, 999: true})
	test(3, 100, map[int]bool{0: true, 1: false, 3: true})
	test(10, 1000, map[int]bool{4: true, 605: false, 900: true})
	test(15, 10000, map[int]bool{10: false, 20: true, 999: true})
	test(3, 1, map[int]bool{0: true})
}
