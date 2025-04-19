package core

import (
	"context"
	"log/slog"
	"strings"
	"sync"
)

type Index struct {
	log     *slog.Logger
	db      DB
	mu      *sync.RWMutex
	entries map[string][]int
}

func NewIndex(log *slog.Logger, db DB) *Index {
	return &Index{
		log:     log,
		db:      db,
		mu:      &sync.RWMutex{},
		entries: make(map[string][]int),
	}
}

func (i *Index) Search(word string) []int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.entries[word]
}

func (i *Index) UpdateIndex(ctx context.Context) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.entries = make(map[string][]int)
	ids, err := i.db.IDs(ctx)
	if err != nil {
		i.log.Error("failed to get ids", "error", err)
		return
	}

	for _, id := range ids {
		comics, err := i.db.Get(ctx, id)
		if err != nil {
			i.log.Error("failed to get comics", "error", err)
			continue
		}
		words := strings.Split(comics.Description, " ")
		i.add(words, id)
	}
}

func (i *Index) add(phrase []string, comicsID int) {
	for _, word := range phrase {
		i.entries[word] = append(i.entries[word], comicsID)
	}
}
