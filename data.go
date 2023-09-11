package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Post struct {
	Id         int64
	Content    Content
	Author     string
	ParentId   *int64
	ReplyCount int
	CreatedAt  time.Time
}

func QuerySinglePost(db *sql.DB, id int) (post Post, err error) {
	row := db.QueryRow(`SELECT id, content_markdown, content_html, author, parent_id, children_count, created_at FROM posts WHERE deleted_at IS NULL AND id = $1`, id)
	err = row.Scan(&post.Id, &post.Content.Markdown, &post.Content.Html, &post.Author, &post.ParentId, &post.ReplyCount, &post.CreatedAt)
	return
}

func QueryPosts(db *sql.DB) ([]Post, error) {
	rows, err := db.Query(`SELECT id, content_markdown, content_html, author, parent_id, children_count, created_at FROM posts WHERE deleted_at IS NULL ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	return readPosts(rows)
}

func (p *Post) GetChildren(db *sql.DB) ([]Post, error) {
	rows, err := db.Query(`SELECT id, content_markdown, content_html, author, parent_id, children_count, created_at FROM posts WHERE deleted_at IS NULL AND parent_id = $1 ORDER BY id`, p.Id)
	if err != nil {
		return nil, err
	}
	return readPosts(rows)
}

func readPosts(rows *sql.Rows) ([]Post, error) {
	defer rows.Close()

	var result []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.Id, &post.Content.Markdown, &post.Content.Html, &post.Author, &post.ParentId, &post.ReplyCount, &post.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to read a row: %e", err)
		}

		result = append(result, post)
	}

	return result, nil
}

func InsertPost(db *sql.DB, content Content, username string, parentPostId *int64) (int64, error) {
	res, err := db.Exec(`INSERT INTO posts (content_markdown, content_html, author, parent_id) VALUES ($1, $2, $3, $4)`, content.Markdown, content.Html, username, parentPostId)
	if err != nil {
		return 0, fmt.Errorf("failed to insert a row: %e", err)
	}
	if parentPostId != nil {
		updateChildCount(db, *parentPostId)
	}

	return res.LastInsertId()
}

func DeletePost(db *sql.DB, postId int64, username string) error {
	var parentId *int64
	err := db.QueryRow(`UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE deleted_at IS NULL AND id = $1 AND author = $2 RETURNING parent_id`, postId, username).Scan(&parentId)
	if err != nil {
		return err
	}
	if parentId != nil {
		updateChildCount(db, *parentId)
	}
	return nil
}

func updateChildCount(db *sql.DB, postId int64) {
	_, err := db.Exec(`UPDATE posts SET children_count = (SELECT count(*) FROM posts c WHERE c.deleted_at IS NULL AND c.parent_id = posts.id) WHERE id = $1`, postId)
	if err != nil {
		log.Printf("Failed to update child count: %s", err)
	}
}
