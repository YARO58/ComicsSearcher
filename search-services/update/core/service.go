package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

type Service struct {
	log         *slog.Logger
	db          DB
	xkcd        XKCD
	words       Words
	concurrency int
	mutex       sync.Mutex
}

func NewService(
	log *slog.Logger, db DB, xkcd XKCD, words Words, concurrency int,
) (*Service, error) {
	if concurrency < 1 {
		return nil, fmt.Errorf("wrong concurrency specified: %d", concurrency)
	}
	return &Service{
		log:         log,
		db:          db,
		xkcd:        xkcd,
		words:       words,
		concurrency: concurrency,
		mutex:       sync.Mutex{},
	}, nil
}

func (s *Service) processComic(ctx context.Context, i int, semaphore chan struct{}, wg *sync.WaitGroup) {
	semaphore <- struct{}{}
	defer func() {
		<-semaphore
		wg.Done()
	}()

	s.log.Debug("downloading comic", "id", i)
	comicsInfo, err := s.xkcd.Get(ctx, i)
	if err != nil {
		s.log.Error("failed to get comic", "error", err)
		return
	}

	description := comicsInfo.Title + " " + comicsInfo.Description

	normedDescription, err := s.words.Norm(ctx, description)
	if err != nil {
		s.log.Error("failed to normalize words", "error", err)
		return
	}

	err = s.db.Add(ctx, Comics{
		ID:    i,
		URL:   comicsInfo.URL,
		Words: normedDescription,
	})
	if err != nil {
		s.log.Error("failed to add comic", "error", err)
		return
	}

	s.log.Debug("added comic", "id", i)
}

func (s *Service) Update(ctx context.Context) (err error) {
	if !s.mutex.TryLock() {
		s.log.Debug("update already in progress")
		return ErrAlreadyExists
	}
	defer s.mutex.Unlock()

	s.log.Debug("update started")

	comicsTotal, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("failed to get comics total", "error", err)
		return ErrComicsCount
	}

	downloadedComics, err := s.db.IDs(ctx)
	if err != nil {
		s.log.Error("failed to get downloaded comics", "error", err)
		return ErrGetDownloadedComics
	}

	downloadedComicsMap := make(map[int]struct{})
	for _, id := range downloadedComics {
		downloadedComicsMap[id] = struct{}{}
	}

	s.log.Debug("downloaded comics", "count :", len(downloadedComicsMap))
	s.log.Debug("downloading comics", "count :", comicsTotal-len(downloadedComicsMap))

	semaphore := make(chan struct{}, s.concurrency)
	defer close(semaphore)
	wg := sync.WaitGroup{}
	for i := 1; i <= comicsTotal; i++ {
		if _, ok := downloadedComicsMap[i]; ok {
			continue
		}
		wg.Add(1)
		go s.processComic(ctx, i, semaphore, &wg)
	}
	wg.Wait()
	return nil
}

func (s *Service) Stats(ctx context.Context) (ServiceStats, error) {
	comicsTotal, err := s.xkcd.LastID(ctx)
	if err != nil {
		s.log.Error("failed to get comics total", "error", err)
		return ServiceStats{}, ErrComicsCount
	}

	dbStats, err := s.db.Stats(ctx)
	if err != nil {
		s.log.Error("failed to get db stats", "error", err)
		return ServiceStats{}, ErrGetDBStats
	}

	return ServiceStats{
		DBStats:     dbStats,
		ComicsTotal: comicsTotal,
	}, nil
}

func (s *Service) Status(ctx context.Context) ServiceStatus {
	if !s.mutex.TryLock() {
		s.log.Debug("status already in progress")
		return StatusRunning
	}
	s.mutex.Unlock()
	s.log.Debug("status idle")
	return StatusIdle
}

func (s *Service) Drop(ctx context.Context) error {
	s.log.Debug("dropping db")
	return s.db.Drop(ctx)
}
