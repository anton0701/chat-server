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

	log.Printf("inserted chat with id: %d", chatId)

	//builderSelect := sq.Select("id", "name", "email", "role", "created_at", "updated_at").
	//	From("chat").
	//	PlaceholderFormat(sq.Dollar).
	//	OrderBy("id").
	//	Limit(10)
	//
	//query, args, err = builderSelect.ToSql()
	//if err != nil {
	//	log.Fatalf("failed to build select query, error: %s", err)
	//}
	//
	//rows, err := pool.Query(ctx, query, args...)
	//
	//log.Printf("Select result:\n\n")
	//for rows.Next() {
	//	var id int
	//	var name, email string
	//	var createdAt time.Time
	//	var updatedAt sql.NullTime
	//	var role desc.UserRole
	//
	//	err = rows.Scan(&id, &name, &email, &role, &createdAt, &updatedAt)
	//	if err != nil {
	//		log.Fatalf("failed to scan row, error: %s", err)
	//	}
	//
	//	log.Printf("id: %v, name: %v, email: %v, createdAt: %v, updatedAt: %v, role: %v", id, name, email, createdAt, updatedAt, role)
	//}
}
