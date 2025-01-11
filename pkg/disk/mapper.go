package disk

import (
	"encoding/gob"
	"errors"
	"hash/fnv"
	"io"

	"github.com/golanguzb70/lrucache"
	"github.com/google/uuid"
)

type ReadWriteSeekSyncTruncater interface {
	ReadWriteSeekSyncer
	Truncate(size int64) error
}

type mapper struct {
	pages   []ReadWriteSeekSyncTruncater
	updates map[int]map[uuid.UUID]int64
	cache   lrucache.LRUCache[int, map[uuid.UUID]int64]
}

func newMapper(capacity int, pages []ReadWriteSeekSyncTruncater) mapper {
	return mapper{
		pages:   pages,
		updates: make(map[int]map[uuid.UUID]int64),
		cache:   lrucache.New[int, map[uuid.UUID]int64](capacity, 0),
	}
}

func (m *mapper) hash(id uuid.UUID) int {
	hasher := fnv.New64a()
	hasher.Write(id[:])
	hashValue := hasher.Sum64()

	return int(hashValue % uint64(len(m.pages)))
}

func (m *mapper) load(h int) (map[uuid.UUID]int64, error) {
	f := m.pages[h]

	var p map[uuid.UUID]int64

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	if err := gob.NewDecoder(f).Decode(&p); err != nil {
		if err == io.EOF {
			if offset, err := f.Seek(0, io.SeekEnd); err == nil {
				if offset == 0 {
					p = make(map[uuid.UUID]int64)
					m.updates[h] = p
				} else {
					return nil, errors.New("bp3store: corrupted page file")
				}
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	m.cache.Put(h, p)

	return p, nil
}

func (m *mapper) flush() error {
	for h, p := range m.updates {

		f := m.pages[h]

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return err
		}

		if err := gob.NewEncoder(f).Encode(p); err != nil {
			return err
		}

		if size, err := f.Seek(0, io.SeekEnd); err == nil {
			if err := f.Truncate(size); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	clear(m.updates)
	return nil
}

func (m *mapper) get(k uuid.UUID) (int64, error) {
	h := m.hash(k)

	p, found := m.cache.Get(h)

	if !found {

		p, found = m.updates[h]

		if !found {
			var err error

			if p, err = m.load(h); err != nil {
				return 0, err
			}
		}
	}

	return p[k], nil
}

func (m *mapper) set(k uuid.UUID, v int64) error {
	h := m.hash(k)

	p, found := m.cache.Get(h)

	if !found {

		p, found = m.updates[h]

		if !found {
			var err error

			if p, err = m.load(h); err != nil {
				return err
			}
		}
	}

	m.updates[h] = p
	p[k] = v

	return nil
}
