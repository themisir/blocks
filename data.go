package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"
)

type Post struct {
	Id        int64
	Content   Content
	Author    string
	CreatedAt time.Time
}

//go:embed schema.sql
var schemaSql string

func MigrateDb(db *sql.DB) error {
	_, err := db.Exec(schemaSql)
	return err
}

func QuerySinglePost(db *sql.DB, id int) (post Post, err error) {
	row := db.QueryRow(`SELECT id, content_markdown, content_html, author, created_at FROM posts WHERE deleted_at IS NULL AND id = $1`, id)
	err = row.Scan(&post.Id, &post.Content.Markdown, &post.Content.Html, &post.Author, &post.CreatedAt)
	return
}

func QueryPosts(db *sql.DB) ([]Post, error) {
	rows, err := db.Query(`SELECT id, content_markdown, content_html, author, created_at FROM posts WHERE deleted_at IS NULL ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.Id, &post.Content.Markdown, &post.Content.Html, &post.Author, &post.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to read a row: %e", err)
		}

		result = append(result, post)
	}

	return result, nil
}

func InsertPost(db *sql.DB, post Post) (int64, error) {
	res, err := db.Exec(`INSERT INTO posts (content_markdown, content_html, author) VALUES ($1, $2, $3)`, post.Content.Markdown, post.Content.Html, post.Author)
	if err != nil {
		return 0, fmt.Errorf("failed to insert a row: %e", err)
	}

	return res.LastInsertId()
}
