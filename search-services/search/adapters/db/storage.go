package db

import (
	"context"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"yadro.com/course/search/core"
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

func (db *DB) Get(ctx context.Context, id int) (core.Comics, error) {
	var comic core.Comics
	err := db.conn.GetContext(ctx, &comic, "SELECT comic_id as id, url, description FROM comics WHERE comic_id = $1", id)
	if err != nil {
		return core.Comics{}, core.ErrNotFound
	}

	db.log.Debug("comics", "id", id)

	return comic, nil
}

func (db *DB) IDs(ctx context.Context) ([]int, error) {
	var ids []int
	err := db.conn.SelectContext(ctx, &ids, "SELECT comic_id FROM comics")
	if err != nil {
		return nil, core.ErrNotFound
	}

	db.log.Debug("comics ids", "ids", ids)

	return ids, nil
}
