package bp3store

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"slices"

	"github.com/google/uuid"
	"github.com/moshenahmias/bp3/pkg/bp3"
	"golang.org/x/exp/constraints"
)

type ReadWriteSeekSyncer interface {
	io.ReadWriteSeeker
	Sync() error
}

type nodeDescriptor[K constraints.Ordered, V any] struct {
	id      string
	offset  int64
	size    int64 // on store
	node    *bp3.Node[K, V]
	builder bp3.NodeBuilder[K, V]
	loader  bp3.NodeLoader[K, V]
}

func (d *nodeDescriptor[K, V]) Read() *bp3.Node[K, V] {
	if d.node == nil {
		if err := d.loader.Load(d); err != nil {
			panic(fmt.Sprintf("bp3store: %v", err))
		}
	}

	return d.node
}

func (d *nodeDescriptor[K, V]) Write() *bp3.Node[K, V] {
	node := d.Read()
	d.builder.Update(d)
	return node
}

type nodeRecord[K constraints.Ordered, V any] struct {
	Id       string
	Mins     []K
	Children []string
	Values   []bp3.KeyValue[K, V]
	Next     string
	Prev     string
}

type nodeBuilder[K constraints.Ordered, V any] struct {
	update map[string]*nodeDescriptor[K, V]
	delete map[string]*nodeDescriptor[K, V]
	store  ReadWriteSeekSyncer
	index  mapper
}

func (b *nodeBuilder[K, V]) Load(d bp3.NodeDescriptor[K, V]) error {
	desc := d.(*nodeDescriptor[K, V])
	var record nodeRecord[K, V]

	if desc.offset == 0 {
		if offset, err := b.index.get(desc.id); err == nil {
			desc.offset = offset
		} else {
			return err
		}
	}

	if _, err := b.store.Seek(desc.offset, io.SeekStart); err != nil {
		return err
	}

	decoder := gob.NewDecoder(b.store)

	if err := decoder.Decode(&record); err != nil {
		return err
	}

	currentOffset, err := b.store.Seek(0, io.SeekCurrent)

	if err != nil {
		return err
	}

	desc.size = currentOffset - desc.offset

	var children []bp3.NodeDescriptor[K, V]

	if len(record.Children) > 0 {
		children = make([]bp3.NodeDescriptor[K, V], 0, len(record.Children))

		for _, id := range record.Children {
			children = append(children, &nodeDescriptor[K, V]{id: id, builder: desc.builder, loader: desc.loader})
		}
	}

	var next bp3.NodeDescriptor[K, V]

	if len(record.Next) > 0 {
		next = &nodeDescriptor[K, V]{id: record.Next, builder: desc.builder, loader: desc.loader}
	}

	var prev bp3.NodeDescriptor[K, V]

	if len(record.Prev) > 0 {
		prev = &nodeDescriptor[K, V]{id: record.Prev, builder: desc.builder, loader: desc.loader}
	}

	desc.node = &bp3.Node[K, V]{
		Mins:     record.Mins,
		Values:   record.Values,
		Children: children,
		Next:     next,
		Prev:     prev,
	}

	return nil
}

func (b *nodeBuilder[K, V]) Create(node *bp3.Node[K, V]) bp3.NodeDescriptor[K, V] {
	d := &nodeDescriptor[K, V]{
		id:      uuid.NewString(),
		node:    node,
		builder: b,
		loader:  b,
	}

	b.Update(d)

	return d
}

func (b *nodeBuilder[K, V]) Update(d bp3.NodeDescriptor[K, V]) {
	desc := d.(*nodeDescriptor[K, V])
	b.update[desc.id] = desc
}

func (b *nodeBuilder[K, V]) Flush() error {
	for id := range b.delete {
		delete(b.update, id)
	}

	clear(b.delete)

	for _, dd := range b.update {
		var children []string

		if len(dd.node.Children) > 0 {
			children = make([]string, 0, len(dd.node.Children))

			for _, d := range dd.node.Children {
				childDesc := d.(*nodeDescriptor[K, V])
				children = append(children, childDesc.id)
			}
		}

		var next string

		if dd.node.Next != nil {
			nextDesc := dd.node.Next.(*nodeDescriptor[K, V])
			next = nextDesc.id
		}

		var prev string

		if dd.node.Prev != nil {
			prevDesc := dd.node.Prev.(*nodeDescriptor[K, V])
			prev = prevDesc.id
		}

		record := nodeRecord[K, V]{
			Id:       dd.id,
			Mins:     slices.Clone(dd.node.Mins),
			Values:   slices.Clone(dd.node.Values),
			Children: children,
			Next:     next,
			Prev:     prev,
		}

		var buffer bytes.Buffer

		encoder := gob.NewEncoder(&buffer)

		if err := encoder.Encode(record); err != nil {
			return err
		}

		encodedSize := int64(buffer.Len())

		if encodedSize > 0 {
			offset := dd.offset
			whench := io.SeekStart

			if dd.size < encodedSize || dd.offset == 0 {
				offset = 0
				whench = io.SeekEnd
			}

			if i, err := b.store.Seek(offset, whench); err != nil {
				return err
			} else {
				offset = i
			}

			if _, err := buffer.WriteTo(b.store); err != nil {
				return err
			}

			dd.offset = offset

			if err := b.index.set(dd.id, offset); err != nil {
				return err
			}
		}

		dd.size = encodedSize
	}

	if err := b.store.Sync(); err != nil {
		return err
	}

	clear(b.update)

	return nil
}

func (b *nodeBuilder[K, V]) Delete(d bp3.NodeDescriptor[K, V]) {
	dd := d.(*nodeDescriptor[K, V])
	b.delete[dd.id] = dd
}

type treeRecord[K constraints.Ordered, V any] struct {
	Root  string
	Min   K
	Order int
	Size  int
}

// Initialize sets up a new B+ Tree instance with the given order, where to store the tree and index page/s.
func Initialize[K constraints.Ordered, V any](order int, store ReadWriteSeekSyncer, page ReadWriteSeekSyncTruncater, rest ...ReadWriteSeekSyncTruncater) (*bp3.Instance[K, V], error) {
	order = max(order, bp3.MinOrder)

	record := treeRecord[K, V]{
		Order: order,
	}

	if err := gob.NewEncoder(store).Encode(record); err != nil {
		return nil, err
	}

	return &bp3.Instance[K, V]{Order: order, Builder: &nodeBuilder[K, V]{
		store:  store,
		update: make(map[string]*nodeDescriptor[K, V]),
		delete: make(map[string]*nodeDescriptor[K, V]),
		index:  newMapper(append([]ReadWriteSeekSyncTruncater{page}, rest...))},
	}, nil
}

// Load retrieves a B+ Tree instance from the given tree store and page/s.
func Load[K constraints.Ordered, V any](store ReadWriteSeekSyncer, page ReadWriteSeekSyncTruncater, rest ...ReadWriteSeekSyncTruncater) (*bp3.Instance[K, V], error) {
	var record treeRecord[K, V]

	if _, err := store.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	if err := gob.NewDecoder(store).Decode(&record); err != nil {
		return nil, err
	}

	builder := &nodeBuilder[K, V]{
		store:  store,
		update: make(map[string]*nodeDescriptor[K, V]),
		delete: make(map[string]*nodeDescriptor[K, V]),
		index:  newMapper(append([]ReadWriteSeekSyncTruncater{page}, rest...)),
	}

	var root *nodeDescriptor[K, V]

	if len(record.Root) > 0 {
		root = &nodeDescriptor[K, V]{id: record.Root, builder: builder, loader: builder}
	}

	return &bp3.Instance[K, V]{
		Root:    root,
		Order:   record.Order,
		Size:    record.Size,
		Min:     record.Min,
		Builder: builder,
	}, nil
}

// Flush writes the current state of the B+ Tree.
func Flush[K constraints.Ordered, V any](tree *bp3.Instance[K, V]) error {
	var root string

	if tree.Root != nil {
		root = tree.Root.(*nodeDescriptor[K, V]).id
	}

	builder := tree.Builder.(*nodeBuilder[K, V])

	if _, err := builder.store.Seek(0, io.SeekStart); err != nil {
		return err
	}

	record := treeRecord[K, V]{
		Order: tree.Order,
		Min:   tree.Min,
		Size:  tree.Size,
		Root:  root,
	}

	if err := gob.NewEncoder(builder.store).Encode(record); err != nil {
		return err
	}

	if err := builder.Flush(); err != nil {
		return err
	}

	return builder.index.flush()
}
