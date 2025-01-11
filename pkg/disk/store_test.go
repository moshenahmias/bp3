package disk_test

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"testing"

	"github.com/moshenahmias/bp3/pkg/bp3"
	"github.com/moshenahmias/bp3/pkg/disk"
	"github.com/spf13/afero"
)

func TestTreeInsertSync(t *testing.T) {
	test := func(order int, n int, p int) {
		fs := afero.NewMemMapFs()

		file, err := fs.Create("testo")

		if err != nil {
			t.Fatal(err)
		}

		defer file.Close()

		var pages []disk.ReadWriteSeekSyncTruncater

		for i := 0; i < p; i++ {
			if pf, err := fs.Create(fmt.Sprintf("page_%d", i)); err == nil {
				pages = append(pages, pf)
				defer pf.Close()
			} else {
				t.Fatal(err)
			}
		}

		tree, err := disk.Initialize[int, string](
			file,
			pages[0],
			disk.WithOrder(order),
			disk.WithIndexPages(pages[1:]),
		)

		if err != nil {
			t.Fatal(err)
		}

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		if err := disk.Flush(tree); err != nil {
			t.Fatal(err)
		}

		s0 := bp3.Slice(tree.Root)

		if _, err := file.Seek(0, io.SeekStart); err != nil {
			t.Fatal(err)
		}

		loaded, err := disk.Load[int, string](
			file,
			pages[0],
			disk.WithOrder(order),
			disk.WithIndexPages(pages[1:]),
		)

		if err != nil {
			t.Fatal(err)
		}

		s1 := bp3.Slice(loaded.Root)

		if slices.CompareFunc(s0, s1, func(a, b bp3.KeyValue[int, string]) int {
			return cmp.Compare(a.Key, b.Key)
		}) != 0 {
			t.Fatalf("%v, %v", s0, s1)
		}
	}

	test(3, 100, 1)
	test(3, 100, 5)
	test(10, 1000, 10)
	test(15, 10000, 100)
}

func TestTreeInserDeleteSync(t *testing.T) {
	test := func(order int, n int, p int) {
		fs := afero.NewMemMapFs()

		file, err := fs.Create("testo")

		if err != nil {
			t.Fatal(err)
		}

		defer file.Close()

		var pages []disk.ReadWriteSeekSyncTruncater

		for i := 0; i < p; i++ {
			if pf, err := fs.Create(fmt.Sprintf("page_%d", i)); err == nil {
				pages = append(pages, pf)
				defer pf.Close()
			} else {
				t.Fatal(err)
			}
		}

		tree, err := disk.Initialize[int, string](
			file,
			pages[0],
			disk.WithOrder(order),
			disk.WithIndexPages(pages[1:]),
		)

		if err != nil {
			t.Fatal(err)
		}

		for i := 0; i < n; i++ {
			tree.Insert(i, fmt.Sprint(i))
		}

		tree.Delete(5)
		tree.Delete(10)
		tree.Delete(20)

		if err := disk.Flush(tree); err != nil {
			t.Fatal(err)
		}

		s0 := bp3.Slice(tree.Root)

		if _, err := file.Seek(0, io.SeekStart); err != nil {
			t.Fatal(err)
		}

		loaded, err := disk.Load[int, string](
			file,
			pages[0],
			disk.WithOrder(order),
			disk.WithIndexPages(pages[1:]),
		)

		if err != nil {
			t.Fatal(err)
		}

		s1 := bp3.Slice(loaded.Root)

		if slices.CompareFunc(s0, s1, func(a, b bp3.KeyValue[int, string]) int {
			return cmp.Compare(a.Key, b.Key)
		}) != 0 {
			t.Fatalf("%v, %v", s0, s1)
		}
	}

	test(3, 100, 1)
	test(3, 100, 5)
	test(10, 1000, 10)
	test(15, 10000, 100)
}
