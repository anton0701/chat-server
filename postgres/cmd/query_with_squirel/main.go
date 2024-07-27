package main

import (
	"context"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/brianvoe/gofakeit"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	dbDSN = "host=localhost port=54322 dbname=chat user=chat-user password=chat-password"
)

func main() {
	ctx := context.Background()

	pool, err := pgxpool.Connect(ctx, dbDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	builderInsert := sq.Insert("chats").
		PlaceholderFormat(sq.Dollar).
		Columns("name").
		Values(gofakeit.Name()).
		Suffix("RETURNING id")

	query, args, err := builderInsert.ToSql()
	if err != nil {
		log.Fatalf("failed to build insert query, error: %s", err)
	}

	var chatId int
	err = pool.QueryRow(ctx, query, args...).Scan(&chatId)
	if err != nil {
		log.Fatalf("failed to insert using builder, error: %s", err)
	}
}
