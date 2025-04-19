package core

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDB реализует интерфейс DB для тестов
type MockDB struct {
	addFunc   func(ctx context.Context, comics Comics) error
	idsFunc   func(ctx context.Context) ([]int, error)
	statsFunc func(ctx context.Context) (DBStats, error)
	dropFunc  func(ctx context.Context) error
}

func (m MockDB) Add(ctx context.Context, comics Comics) error {
	if m.addFunc != nil {
		return m.addFunc(ctx, comics)
	}
	return nil
}

func (m MockDB) IDs(ctx context.Context) ([]int, error) {
	if m.idsFunc != nil {
		return m.idsFunc(ctx)
	}
	return []int{}, nil
}

func (m MockDB) Stats(ctx context.Context) (DBStats, error) {
	if m.statsFunc != nil {
		return m.statsFunc(ctx)
	}
	return DBStats{}, nil
}

func (m MockDB) Drop(ctx context.Context) error {
	if m.dropFunc != nil {
		return m.dropFunc(ctx)
	}
	return nil
}

// MockXKCD реализует интерфейс XKCD для тестов
type MockXKCD struct {
	getFunc    func(ctx context.Context, id int) (XKCDInfo, error)
	lastIDFunc func(ctx context.Context) (int, error)
}

func (m MockXKCD) Get(ctx context.Context, id int) (XKCDInfo, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return XKCDInfo{}, nil
}

func (m MockXKCD) LastID(ctx context.Context) (int, error) {
	if m.lastIDFunc != nil {
		return m.lastIDFunc(ctx)
	}
	return 0, nil
}

// MockWords реализует интерфейс Words для тестов
type MockWords struct {
	normFunc func(ctx context.Context, phrase string) ([]string, error)
}

func (m MockWords) Norm(ctx context.Context, phrase string) ([]string, error) {
	if m.normFunc != nil {
		return m.normFunc(ctx, phrase)
	}
	return []string{}, nil
}

func TestNewService(t *testing.T) {
	tests := []struct {
		name        string
		concurrency int
		wantError   bool
	}{
		{
			name:        "valid concurrency",
			concurrency: 1,
			wantError:   false,
		},
		{
			name:        "zero concurrency",
			concurrency: 0,
			wantError:   true,
		},
		{
			name:        "negative concurrency",
			concurrency: -1,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(os.Stdout, nil))
			db := MockDB{}
			xkcd := MockXKCD{}
			words := MockWords{}

			service, err := NewService(log, db, xkcd, words, tt.concurrency)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, service)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, service)
			assert.Equal(t, tt.concurrency, service.concurrency)
		})
	}
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name              string
		lastID            int
		lastIDError       error
		downloadedIDs     []int
		downloadedError   error
		comicsToAdd       int
		wantError         bool
		errorType         error
		processComicError bool
		normError         bool
		addError          bool
	}{
		{
			name:          "no comics to update",
			lastID:        5,
			downloadedIDs: []int{1, 2, 3, 4, 5},
			comicsToAdd:   0,
			wantError:     false,
		},
		{
			name:          "some comics to update",
			lastID:        5,
			downloadedIDs: []int{1, 3, 5},
			comicsToAdd:   2,
			wantError:     false,
		},
		{
			name:          "all comics to update",
			lastID:        5,
			downloadedIDs: []int{},
			comicsToAdd:   5,
			wantError:     false,
		},
		{
			name:          "lastID error",
			lastID:        0,
			lastIDError:   errors.New("xkcd error"),
			downloadedIDs: []int{},
			comicsToAdd:   0,
			wantError:     true,
			errorType:     ErrComicsCount,
		},
		{
			name:            "downloadedIDs error",
			lastID:          5,
			downloadedError: errors.New("db error"),
			downloadedIDs:   []int{},
			comicsToAdd:     0,
			wantError:       true,
			errorType:       ErrGetDownloadedComics,
		},
		{
			name:              "xkcd get error",
			lastID:            5,
			downloadedIDs:     []int{},
			comicsToAdd:       0,
			wantError:         false,
			processComicError: true,
		},
		{
			name:              "words norm error",
			lastID:            5,
			downloadedIDs:     []int{},
			comicsToAdd:       0,
			wantError:         false,
			processComicError: true,
		},
		{
			name:              "db add error",
			lastID:            5,
			downloadedIDs:     []int{},
			comicsToAdd:       0,
			wantError:         false,
			processComicError: true,
		},
		{
			name:              "norm error in processComic",
			lastID:            5,
			downloadedIDs:     []int{},
			comicsToAdd:       0,
			wantError:         false,
			processComicError: false,
			normError:         true,
		},
		{
			name:              "add error in processComic",
			lastID:            5,
			downloadedIDs:     []int{},
			comicsToAdd:       0,
			wantError:         false,
			processComicError: false,
			addError:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(os.Stdout, nil))
			addedComics := make(map[int]struct{})
			var mu sync.Mutex

			db := MockDB{
				idsFunc: func(ctx context.Context) ([]int, error) {
					return tt.downloadedIDs, tt.downloadedError
				},
				addFunc: func(ctx context.Context, comics Comics) error {
					if tt.processComicError || tt.addError {
						return errors.New("db add error")
					}
					mu.Lock()
					defer mu.Unlock()
					addedComics[comics.ID] = struct{}{}
					return nil
				},
			}

			xkcd := MockXKCD{
				lastIDFunc: func(ctx context.Context) (int, error) {
					return tt.lastID, tt.lastIDError
				},
				getFunc: func(ctx context.Context, id int) (XKCDInfo, error) {
					if tt.processComicError {
						return XKCDInfo{}, errors.New("xkcd get error")
					}
					return XKCDInfo{
						ID:          id,
						Title:       "Test Title",
						Description: "Test Description",
						URL:         "http://test.com",
					}, nil
				},
			}

			words := MockWords{
				normFunc: func(ctx context.Context, phrase string) ([]string, error) {
					if tt.processComicError || tt.normError {
						return nil, errors.New("words norm error")
					}
					return []string{"test", "words"}, nil
				},
			}

			service, err := NewService(log, db, xkcd, words, 2)
			require.NoError(t, err)

			err = service.Update(context.Background())
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
				return
			}

			assert.NoError(t, err)
			mu.Lock()
			assert.Equal(t, tt.comicsToAdd, len(addedComics))
			mu.Unlock()
		})
	}
}

func TestService_Update_Concurrent(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service, err := NewService(log, MockDB{}, MockXKCD{}, MockWords{}, 1)
	require.NoError(t, err)

	// Захватываем мьютекс
	service.mutex.Lock()

	// Создаем канал для получения результата
	resultChan := make(chan error)

	// Пытаемся запустить Update в другой горутине
	go func() {
		resultChan <- service.Update(context.Background())
	}()

	// Проверяем, что получили ErrAlreadyExists
	assert.ErrorIs(t, <-resultChan, ErrAlreadyExists)

	// Освобождаем мьютекс
	service.mutex.Unlock()
}

func TestService_Stats(t *testing.T) {
	tests := []struct {
		name         string
		lastID       int
		lastIDError  error
		dbStats      DBStats
		dbStatsError error
		wantError    bool
		errorType    error
	}{
		{
			name:      "successful stats",
			lastID:    100,
			dbStats:   DBStats{WordsTotal: 100, WordsUnique: 50, ComicsFetched: 10},
			wantError: false,
		},
		{
			name:        "xkcd error",
			lastID:      0,
			lastIDError: errors.New("xkcd error"),
			dbStats:     DBStats{},
			wantError:   true,
			errorType:   ErrComicsCount,
		},
		{
			name:         "db stats error",
			lastID:       100,
			dbStats:      DBStats{},
			dbStatsError: errors.New("db error"),
			wantError:    true,
			errorType:    ErrGetDBStats,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(os.Stdout, nil))

			xkcd := MockXKCD{
				lastIDFunc: func(ctx context.Context) (int, error) {
					return tt.lastID, tt.lastIDError
				},
			}

			db := MockDB{
				statsFunc: func(ctx context.Context) (DBStats, error) {
					return tt.dbStats, tt.dbStatsError
				},
			}

			service, err := NewService(log, db, xkcd, MockWords{}, 1)
			require.NoError(t, err)

			stats, err := service.Stats(context.Background())
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.lastID, stats.ComicsTotal)
			assert.Equal(t, tt.dbStats, stats.DBStats)
		})
	}
}

func TestService_Status(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service, err := NewService(log, MockDB{}, MockXKCD{}, MockWords{}, 1)
	require.NoError(t, err)

	// Первый вызов должен вернуть StatusIdle
	assert.Equal(t, StatusIdle, service.Status(context.Background()))

	// Захватываем мьютекс
	service.mutex.Lock()

	// Второй вызов должен вернуть StatusRunning
	assert.Equal(t, StatusRunning, service.Status(context.Background()))

	// Освобождаем мьютекс
	service.mutex.Unlock()

	// Третий вызов должен снова вернуть StatusIdle
	assert.Equal(t, StatusIdle, service.Status(context.Background()))
}

func TestService_Drop(t *testing.T) {
	tests := []struct {
		name      string
		dropError error
		wantError bool
	}{
		{
			name:      "successful drop",
			dropError: nil,
			wantError: false,
		},
		{
			name:      "drop error",
			dropError: errors.New("drop error"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.New(slog.NewTextHandler(os.Stdout, nil))

			db := MockDB{
				dropFunc: func(ctx context.Context) error {
					return tt.dropError
				},
			}

			service, err := NewService(log, db, MockXKCD{}, MockWords{}, 1)
			require.NoError(t, err)

			err = service.Drop(context.Background())
			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
