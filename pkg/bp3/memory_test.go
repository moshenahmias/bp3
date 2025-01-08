package bp3_test

import (
	"cmp"
	"fmt"
	"iter"
	"slices"
	"testing"

	"github.com/moshenahmias/bp3/pkg/bp3"
	"golang.org/x/exp/constraints"
)

func TestInsert(t *testing.T) {
	test := func(order int, n int) {
		tree := bp3.New[int, string](order)

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		if size := tree.Size; size != n {
			t.Fatalf("size %d != %d", size, n)
		}

		s := bp3.Slice(tree.Root)

		if len(s) != n {
			t.Fatalf("slice size %d != %d", len(s), n)
		}

		for i := 0; i < n; i++ {
			if v, found := tree.Find(i); !found || fmt.Sprint(i) != v {

			}

			if s[i].Key != i || s[i].Value != fmt.Sprint(i) {
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
		tree := bp3.New[int, string](order)

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

		if size := tree.Size; size != n-len(del) {
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

func TestRangeClosed(t *testing.T) {
	test := func(order int, n int, from, to int) {
		tree := bp3.New[int, string](order)

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		s := slices.Collect(SeqFirst(tree.RangeClosed(from, to)))
		master := slices.Collect(Range_(from, to+1))

		if slices.Compare(s, master) != 0 {
			t.Fatalf("%v, %v", s, master)
		}
	}

	test(3, 100, 3, 50)
	test(10, 1000, 100, 700)
	test(15, 10000, 500, 2005)
}

func TestRangeOpened(t *testing.T) {
	test := func(order int, n int, from, to int) {
		tree := bp3.New[int, string](order)

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		s := slices.Collect(SeqFirst(tree.RangeOpened(from, to)))
		master := slices.Collect(Range_(from+1, to))

		if slices.Compare(s, master) != 0 {
			t.Fatalf("%v, %v", s, master)
		}
	}

	test(3, 100, 3, 50)
	test(10, 1000, 100, 700)
	test(15, 10000, 500, 2005)
}

func TestRangeLowHalfOpened(t *testing.T) {
	test := func(order int, n int, from, to int) {
		tree := bp3.New[int, string](order)

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		s := slices.Collect(SeqFirst(tree.RangeLowHalfOpened(from, to)))
		master := slices.Collect(Range_(from+1, to+1))

		if slices.Compare(s, master) != 0 {
			t.Fatalf("%v, %v", s, master)
		}
	}

	test(3, 100, 3, 50)
	test(10, 1000, 100, 700)
	test(15, 10000, 500, 2005)
}

func TestRangeHighHalfOpened(t *testing.T) {
	test := func(order int, n int, from, to int) {
		tree := bp3.New[int, string](order)

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		s := slices.Collect(SeqFirst(tree.RangeHighHalfOpened(from, to)))
		master := slices.Collect(Range_(from, to))

		if slices.Compare(s, master) != 0 {
			t.Fatalf("%v, %v", s, master)
		}
	}

	test(3, 100, 3, 50)
	test(10, 1000, 100, 700)
	test(15, 10000, 500, 2005)
}

func TestRangeMissingLeft(t *testing.T) {
	tree := bp3.New[int, string](3)

	for i := 20; i < 50; i++ {
		tree.Insert(i, fmt.Sprint(i))
	}

	s := slices.Collect(SeqFirst(tree.RangeClosed(10, 30)))
	master := slices.Collect(Range_(20, 31))

	if slices.Compare(s, master) != 0 {
		t.Fatalf("%v, %v", s, master)
	}
}

func TestRangeMissingRight(t *testing.T) {
	tree := bp3.New[int, string](3)

	for i := 20; i < 50; i++ {
		tree.Insert(i, fmt.Sprint(i))
	}

	s := slices.Collect(SeqFirst(tree.RangeClosed(30, 60)))
	master := slices.Collect(Range_(30, 50))

	if slices.Compare(s, master) != 0 {
		t.Fatalf("%v, %v", s, master)
	}
}

func TestRangeMissingLeftRight(t *testing.T) {
	tree := bp3.New[int, string](3)

	for i := 20; i < 50; i++ {
		tree.Insert(i, fmt.Sprint(i))
	}

	s := slices.Collect(SeqFirst(tree.RangeClosed(10, 60)))
	master := slices.Collect(Range_(20, 50))

	if slices.Compare(s, master) != 0 {
		t.Fatalf("%v, %v", s, master)
	}
}

func TestRangeMissingMiddle(t *testing.T) {
	tree := bp3.New[int, string](3)

	for i := 20; i < 30; i++ {
		tree.Insert(i, fmt.Sprint(i))
	}

	for i := 31; i < 50; i++ {
		tree.Insert(i, fmt.Sprint(i))
	}

	s := slices.Collect(SeqFirst(tree.RangeClosed(30, 50)))
	master := slices.Collect(Range_(31, 50))

	if slices.Compare(s, master) != 0 {
		t.Fatalf("%v, %v", s, master)
	}
}

func TestDeleteRange(t *testing.T) {
	tree := bp3.New[int, string](3)

	for i := 0; i < 5; i++ {
		tree.Insert(i, fmt.Sprint(i))
	}

	if _, deleted := tree.Delete(3); !deleted {
		t.Fail()
	}

	s := slices.Collect(SeqFirst(tree.RangeClosed(0, 5)))
	master := []int{0, 1, 2, 4}

	if slices.Compare(s, master) != 0 {
		t.Fatalf("%v, %v", s, master)
	}
}

func TestFromClosed(t *testing.T) {
	test := func(order int, n int, from int) {
		tree := bp3.New[int, string](order)

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		s := slices.Collect(SeqFirst(tree.FromClosed(from)))
		master := slices.Collect(Range_(from, n))

		if slices.Compare(s, master) != 0 {
			t.Fatalf("%v, %v", s, master)
		}
	}

	test(3, 100, 3)
	test(10, 1000, 100)
	test(15, 10000, 500)
}

func TestFromOpened(t *testing.T) {
	test := func(order int, n int, from int) {
		tree := bp3.New[int, string](order)

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		s := slices.Collect(SeqFirst(tree.FromOpened(from)))
		master := slices.Collect(Range_(from+1, n))

		if slices.Compare(s, master) != 0 {
			t.Fatalf("%v, %v", s, master)
		}
	}

	test(3, 100, 3)
	test(10, 1000, 100)
	test(15, 10000, 500)
}

func TestToClosed(t *testing.T) {
	test := func(order int, n int, to int) {
		tree := bp3.New[int, string](order)

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		s := slices.Collect(SeqFirst(tree.ToClosed(to)))
		master := slices.Collect(Range_(0, to+1))

		if slices.Compare(s, master) != 0 {
			t.Fatalf("%v, %v", s, master)
		}
	}

	test(3, 100, 3)
	test(10, 1000, 100)
	test(15, 10000, 500)
}

func TestToOpened(t *testing.T) {
	test := func(order int, n int, to int) {
		tree := bp3.New[int, string](order)

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		s := slices.Collect(SeqFirst(tree.ToOpened(to)))
		master := slices.Collect(Range_(0, to))

		if slices.Compare(s, master) != 0 {
			t.Fatalf("%v, %v", s, master)
		}
	}

	test(3, 100, 3)
	test(10, 1000, 100)
	test(15, 10000, 500)
}

func TestMinimum(t *testing.T) {
	tree := bp3.New[int, string](3)

	for i := 0; i < 5; i++ {
		tree.Insert(i, fmt.Sprint(i))
	}

	if minimum := tree.Minimum(); minimum != "0" {
		t.Fatal(minimum)
	}
}

func TestMaximum(t *testing.T) {
	tree := bp3.New[int, string](3)

	for i := 0; i < 5; i++ {
		tree.Insert(i, fmt.Sprint(i))
	}

	if maximum := tree.Maximum(); maximum != "4" {
		t.Fatal(maximum)
	}
}

func slice[K constraints.Ordered, V any](root bp3.NodeDescriptor[K, V]) []bp3.KeyValue[K, V] {
	if root == nil || root.Read() == nil {
		return nil
	}

	if root.Read().Leaf() {
		return root.Read().Values
	}

	var s []bp3.KeyValue[K, V]

	for _, child := range root.Read().Children {
		s = append(s, slice(child)...)
	}

	return s
}

func TestSlice(t *testing.T) {
	test := func(order int, n int, delete []int) {

		tree := bp3.New[int, string](order)

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		master := slices.Collect(Range_(0, n))
		d := 0

		for _, k := range delete {
			tree.Delete(k)
			master = slices.Delete(master, k-d, k-d+1)
			d++
		}

		s := bp3.Slice(tree.Root)

		if slices.CompareFunc(s, master, func(kv bp3.KeyValue[int, string], x int) int {
			return cmp.Compare(kv.Key, x)
		}) != 0 {
			t.Fatalf("%v, %v", s, master)
		}

		v := slice(tree.Root)

		if slices.CompareFunc(v, master, func(kv bp3.KeyValue[int, string], x int) int {
			return cmp.Compare(kv.Key, x)
		}) != 0 {
			t.Fatalf("%v, %v", v, master)
		}
	}

	test(3, 8, []int{4})
	test(3, 600, []int{100, 560})
	test(10, 1000, []int{50, 100, 560})
	test(15, 10000, []int{1, 999})
}

// TODO: move to a different pacakge

func SeqFirst[K, V any](seq iter.Seq2[K, V]) iter.Seq[K] {
	return func(yield func(K) bool) {
		for k, _ := range seq {
			if !yield(k) {
				return
			}
		}
	}
}

func SeqSecond[K, V any](seq iter.Seq2[K, V]) iter.Seq[V] {
	return func(yield func(V) bool) {
		for _, v := range seq {
			if !yield(v) {
				return
			}
		}
	}
}

func Range_[T constraints.Integer](from, to T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for from < to {
			if !yield(from) {
				return
			}

			from++
		}
	}
}
