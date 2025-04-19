package db

import (
	"context"
	"log/slog"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/update/core"
)

type DB struct {
	log  *slog.Logger
	conn *sqlx.DB
}

func New(log *slog.Logger, address string) (*DB, error) {
	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		conn: db,
	}, nil
}

func (db *DB) Add(ctx context.Context, comic core.Comics) error {
	query := `
		INSERT INTO comics (comic_id, url, description)
		VALUES ($1, $2, $3)
		ON CONFLICT (comic_id) DO UPDATE
		SET url = EXCLUDED.url,
			description = EXCLUDED.description
	`

	_, err := db.conn.ExecContext(ctx, query, comic.ID, comic.URL, strings.Join(comic.Words, " "))
	if err != nil {
		db.log.Error("failed to insert comic",
			"error", err,
			"comic_id", comic.ID,
			"words_count", len(comic.Words),
			"url", comic.URL)
		return err
	}

	db.log.Debug("comic added",
		"comic_id", comic.ID,
		"words_count", len(comic.Words),
		"url", comic.URL)

	return nil
}

func (db *DB) Stats(ctx context.Context) (core.DBStats, error) {
	var stats core.DBStats

	err := db.conn.GetContext(ctx, &stats.ComicsFetched, "SELECT COUNT(*) FROM comics")
	if err != nil {
		return core.DBStats{}, err
	}

	db.log.Debug("comics fetched", "count", stats.ComicsFetched)

	var descriptions []string
	err = db.conn.SelectContext(ctx, &descriptions, "SELECT description FROM comics WHERE description IS NOT NULL")
	if err != nil {
		return core.DBStats{}, err
	}

	wordsMap := make(map[string]struct{})
	for _, desc := range descriptions {
		words := strings.Fields(desc)
		stats.WordsTotal += len(words)
		for _, word := range words {
			wordsMap[word] = struct{}{}
		}
	}
	stats.WordsUnique = len(wordsMap)

	db.log.Debug("words unique", "count", stats.WordsUnique)
	db.log.Debug("words total", "count", stats.WordsTotal)

	return stats, nil
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	err := db.conn.SelectContext(ctx, &ids, "SELECT comic_id FROM comics")
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (db *DB) Drop(ctx context.Context) error {
	if _, err := db.conn.ExecContext(ctx, "TRUNCATE TABLE comics"); err != nil {
		db.log.Error("failed to truncate table", "error", err)
		return core.ErrTruncateTable
	}

	db.log.Debug("table truncated")

	return nil
}
