package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Parent struct {
	Id      int64
	Content Content
	Author  string
}

type Post struct {
	Id         int64
	Content    Content
	Author     string
	Parent     *Parent
	ReplyCount int
	CreatedAt  time.Time
}

func AuthorDisplay(post any) string {
	switch post := post.(type) {
	case *Post:
		return authorDisplay(post.Author)
	case Post:
		return authorDisplay(post.Author)
	case *Parent:
		return authorDisplay(post.Author)
	case Parent:
		return authorDisplay(post.Author)
	default:
		return "<invalid>"
	}
}

func authorDisplay(s string) string {
	if s == "" {
		return "anonymous"
	} else {
		return "@" + s
	}
}

func QuerySinglePost(db *sql.DB, id int) (post Post, err error) {
	row := db.QueryRow(`SELECT c.id,
       c.content_markdown,
       c.content_html,
       coalesce(c.author, ''),
       c.children_count,
       c.created_at,
       (c.parent_id IS NOT NULL AND p.deleted_at IS NULL) as has_parent,
       coalesce(p.id, 0),
       coalesce(p.content_markdown, ''),
       coalesce(p.content_html, ''),
       coalesce(p.author, '')
FROM posts c
         LEFT JOIN posts p ON p.id = c.parent_id AND p.deleted_at IS NULL
WHERE c.deleted_at IS NULL AND c.id = $1`, id)

	var hasParent bool
	var parent Parent

	err = row.Scan(&post.Id, &post.Content.Markdown, &post.Content.Html, &post.Author, &post.ReplyCount, &post.CreatedAt, &hasParent,
		&parent.Id, &parent.Content.Markdown, &parent.Content.Html, &parent.Author)

	if err == nil && hasParent {
		post.Parent = &parent
	}

	return
}

func QueryPosts(db *sql.DB) ([]Post, error) {
	rows, err := db.Query(`SELECT c.id,
       c.content_markdown,
       c.content_html,
       coalesce(c.author, ''),
       c.children_count,
       c.created_at,
       (c.parent_id IS NOT NULL AND p.deleted_at IS NULL) as has_parent,
       coalesce(p.id, 0),
       coalesce(p.content_markdown, ''),
       coalesce(p.content_html, ''),
       coalesce(p.author, '')
FROM posts c
         LEFT JOIN posts p ON p.id = c.parent_id AND p.deleted_at IS NULL
WHERE c.deleted_at IS NULL
ORDER BY c.id DESC`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Post
	for rows.Next() {
		var post Post
		var hasParent bool
		var parent Parent

		err := rows.Scan(&post.Id, &post.Content.Markdown, &post.Content.Html, &post.Author, &post.ReplyCount, &post.CreatedAt, &hasParent,
			&parent.Id, &parent.Content.Markdown, &parent.Content.Html, &parent.Author)
		if err != nil {
			return nil, fmt.Errorf("failed to read a row: %w", err)
		}

		if hasParent {
			post.Parent = &parent
		}

		result = append(result, post)
	}

	return result, nil
}

func QueryChildrenPosts(db *sql.DB, parentPostId int64) ([]Post, error) {
	rows, err := db.Query(`SELECT c.id,
       c.content_markdown,
       c.content_html,
       coalesce(c.author, ''),
       c.children_count,
       c.created_at
FROM posts c
         LEFT JOIN posts p ON p.id = c.parent_id AND p.deleted_at IS NULL
WHERE c.deleted_at IS NULL AND c.parent_id = $1
ORDER BY c.id`, parentPostId)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Post
	for rows.Next() {
		var post Post

		err := rows.Scan(&post.Id, &post.Content.Markdown, &post.Content.Html, &post.Author, &post.ReplyCount, &post.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to read a row: %w", err)
		}

		result = append(result, post)
	}

	return result, nil
}

func InsertPost(db *sql.DB, content Content, username string, parentPostId *int64) (int64, error) {
	res, err := db.Exec(`INSERT INTO posts (content_markdown, content_html, author, parent_id) VALUES ($1, $2, $3, $4)`, content.Markdown, content.Html, username, parentPostId)
	if err != nil {
		return 0, fmt.Errorf("failed to insert a row: %w", err)
	}
	if parentPostId != nil {
		updateChildCount(db, *parentPostId)
	}

	return res.LastInsertId()
}

func DeletePost(db *sql.DB, postId int64, username string, isAdmin bool) error {
	var parentId *int64
	err := db.QueryRow(`UPDATE posts SET deleted_at = CURRENT_TIMESTAMP WHERE deleted_at IS NULL AND id = $1 AND (author = $2 OR $3) RETURNING parent_id`, postId, username, isAdmin).Scan(&parentId)
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
