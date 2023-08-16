package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"
)

type Post struct {
	Id           int
	Content      Content
	IsAnonymous  bool
	User         UserRef
	Depth        int
	ParentPostId int
	CreatedAt    time.Time
	Children     []*Post
	ReplyCount   int
}

func (p Post) CanBeDeleted() bool {
	delta := time.Now().Sub(p.CreatedAt)
	return delta < 3*time.Minute
}

func (p Post) Author() string {
	if p.IsAnonymous {
		return "anonymous"
	} else {
		return p.User.Username
	}
}

type UserRef struct {
	Id       int
	Username string
}

func QuerySinglePost(db *sql.DB, id int) (*Post, error) {
	rows, err := db.Query(`SELECT p.id,
       p.content_markdown,
       p.content_html,
       p.is_anonymous,
       coalesce(p.user_id, 0) as user_id,
       coalesce(u.username, '') as username,
       p.depth,
       coalesce(p.parent_post_id, 0) as parent_post_id,
       p.created_at
FROM posts p
         LEFT JOIN main.users u on u.id = p.user_id
WHERE (p.id = $1 OR p.top_level_post_id = $1)
  AND p.deleted_at IS NULL
ORDER BY p.depth ASC, p.id DESC`, id)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	posts := make(map[int]*Post)

	for rows.Next() {
		post := new(Post)
		err = rows.Scan(&post.Id, &post.Content.Markdown, &post.Content.Html, &post.IsAnonymous, &post.User.Id, &post.User.Username, &post.Depth, &post.ParentPostId, &post.CreatedAt)
		if err != nil {
			return nil, err
		}

		posts[post.Id] = post

		if parent, ok := posts[post.ParentPostId]; ok {
			parent.Children = append(parent.Children, post)
		}
	}

	if top, ok := posts[id]; ok {
		return top, nil
	} else {
		return nil, fmt.Errorf("failed to select post, empty post map")
	}
}

func QueryPosts(db *sql.DB) ([]Post, error) {
	rows, err := db.Query(`SELECT p.id,
       p.content_markdown,
       p.content_html,
       p.is_anonymous,
       coalesce(p.user_id, 0) as user_id,
       coalesce(u.username, '') as username,
       p.depth,
       coalesce(p.parent_post_id, 0) as parent_post_id,
       p.created_at,
       (SELECT count(*) FROM posts c WHERE c.top_level_post_id = p.id AND c.deleted_at IS NULL) as reply_count
FROM posts p
         LEFT JOIN main.users u on u.id = p.user_id
WHERE p.depth = 0
  AND p.deleted_at IS NULL
ORDER BY p.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Post
	for rows.Next() {
		var post Post
		err := rows.Scan(&post.Id, &post.Content.Markdown, &post.Content.Html, &post.IsAnonymous, &post.User.Id, &post.User.Username, &post.Depth, &post.ParentPostId, &post.CreatedAt, &post.ReplyCount)
		if err != nil {
			return nil, fmt.Errorf("failed to read a row: %e", err)
		}

		result = append(result, post)
	}

	return result, nil
}

type CreatePost struct {
	Content      Content
	IsAnonymous  bool
	UserId       int
	ParentPostId int
}

func InsertPost(db *sql.DB, post CreatePost) (int, error) {
	var res sql.Result
	var err error

	if post.ParentPostId > 0 {
		var parentPostId, topLevelPostId, parentDepth int

		err := db.QueryRow(`SELECT p.id, coalesce(p.top_level_post_id, 0), p.depth FROM posts p WHERE id = $1`, post.ParentPostId).Scan(&parentPostId, &topLevelPostId, &parentDepth)
		if err != nil {
			return 0, fmt.Errorf("failed to find parent post: %v", err)
		}
		if topLevelPostId == 0 {
			topLevelPostId = parentPostId
		}

		res, err = db.Exec(`INSERT INTO posts (content_markdown, content_html, is_anonymous, user_id, depth, top_level_post_id, parent_post_id) VALUES ($1, $2, $3, $4, $5, $6, $7)`, post.Content.Markdown, post.Content.Html, post.IsAnonymous, post.UserId, parentDepth+1, topLevelPostId, parentPostId)
	} else {
		res, err = db.Exec(`INSERT INTO posts (content_markdown, content_html, is_anonymous, user_id, depth, top_level_post_id, parent_post_id) VALUES ($1, $2, $3, $4, 0, NULL, NULL)`, post.Content.Markdown, post.Content.Html, post.IsAnonymous, post.UserId)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to insert a row: %e", err)
	}
	return id32(res)
}

func DeletePost(db *sql.DB, id int64, userId int) error {
	_, err := db.Exec(`DELETE FROM posts WHERE id = $1 AND user_id = $2 AND (UNIXEPOCH('now') - UNIXEPOCH(created_at)) < 180`, id, userId)
	return err
}

func id32(res sql.Result) (int, error) {
	id, err := res.LastInsertId()
	return int(id), err
}
