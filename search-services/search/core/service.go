package core

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
)

type Service struct {
	log   *slog.Logger
	words Words
	index *Index
	db    DB
}

func NewService(log *slog.Logger, db DB, words Words, idx *Index) *Service {
	return &Service{
		log:   log,
		words: words,
		db:    db,
		index: idx,
	}
}

func (s *Service) Search(ctx context.Context, query string, limit int) ([]Comics, int, error) {
	if limit <= 0 {
		return nil, 0, ErrBadArguments
	}

	normQuery, err := s.words.Norm(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to normalize words: %w", err)
	}

	comicIDToHits := make(map[int]int)
	ids, err := s.db.IDs(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get ids: %w", err)
	}

	for _, id := range ids {
		comics, err := s.db.Get(ctx, id)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to get comic: %w", err)
		}
		for _, word := range normQuery {
			if strings.Contains(comics.Description, word) {
				comicIDToHits[id]++
			}
		}
	}

	type comicScore struct {
		ID    int
		Score int
	}
	var scoredComics []comicScore
	for id, score := range comicIDToHits {
		scoredComics = append(scoredComics, comicScore{id, score})
	}
	sort.Slice(scoredComics, func(i, j int) bool {
		return scoredComics[i].Score > scoredComics[j].Score
	})

	s.log.Debug("scoredComics", "scoredComics", scoredComics)

	resultComics := make([]Comics, 0, limit)
	for i := 0; i < len(scoredComics) && i < limit; i++ {
		comics, err := s.db.Get(ctx, scoredComics[i].ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to fetch comic %d: %v", scoredComics[i].ID, err)
		}
		resultComics = append(resultComics, comics)
	}

	s.log.Debug("resultComics", "resultComics", resultComics)

	return resultComics, len(resultComics), nil
}

func (s *Service) ISearch(ctx context.Context, query string, limit int) ([]Comics, int, error) {
	if limit <= 0 {
		return nil, 0, ErrBadArguments
	}

	uniqueWords, err := s.words.Norm(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to normalize words: %w", err)
	}

	comicIDToHits := make(map[int]int)
	for _, word := range uniqueWords {
		comicIDs := s.index.Search(word)
		for _, id := range comicIDs {
			comicIDToHits[id]++
		}
	}

	s.log.Debug("comicIDToHits", "comicIDToHits", comicIDToHits)

	type comicScore struct {
		ID    int
		Score int
	}
	var scoredComics []comicScore
	for id, score := range comicIDToHits {
		scoredComics = append(scoredComics, comicScore{id, score})
	}
	sort.Slice(scoredComics, func(i, j int) bool {
		return scoredComics[i].Score > scoredComics[j].Score
	})

	s.log.Debug("scoredComics", "scoredComics", scoredComics)

	resultComics := make([]Comics, 0, limit)
	for i := 0; i < len(scoredComics) && i < limit; i++ {
		comics, err := s.db.Get(ctx, scoredComics[i].ID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to fetch comic %d: %v", scoredComics[i].ID, err)
		}
		resultComics = append(resultComics, comics)
	}

	s.log.Debug("resultComics", "resultComics", resultComics)

	return resultComics, len(resultComics), nil
}
