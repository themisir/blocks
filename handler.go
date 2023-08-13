package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
)

type Handler struct {
	db *sql.DB
}

func NewHandler(dbPath string) (*Handler, error) {
	if dbPath == "" {
		dbPath = "data.sqlite3"
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %e", err)
	}

	err = MigrateDb(db)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %e", err)
	}

	return &Handler{db}, nil
}

func (self *Handler) HandleGetPosts(c echo.Context) error {
	posts, err := QueryPosts(self.db)
	if err != nil {
		return fmt.Errorf("failed to query posts: %e", err)
	}

	return c.Render(http.StatusOK, "index.html", echo.Map{
		"Posts":         posts,
		"EnablePosting": true,
	})
}

func (self *Handler) HandleGetSinglePost(c echo.Context) error {
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrNotFound
	}

	post, err := QuerySinglePost(self.db, postId)
	if err != nil {
		return fmt.Errorf("failed to query single post (id: %v): %e", postId, err)
	}

	return c.Render(http.StatusOK, "index.html", echo.Map{
		"Posts":         []Post{post},
		"EnablePosting": false,
	})
}

func (self *Handler) HandleCreatePost(c echo.Context) error {
	bodyContent := c.FormValue("content")
	bodyAuthor := c.FormValue("author")

	if len(bodyAuthor) > 32 {
		bodyAuthor = bodyAuthor[:32]
	}
	if len(bodyContent) < 8 {
		return echo.ErrBadRequest
	}

	content, err := CompileContent(bodyContent)
	if err != nil {
		return fmt.Errorf("failed to compile content: %e", err)
	}

	_, err = InsertPost(self.db, Post{
		Content: content,
		Author:  bodyAuthor,
	})
	if err != nil {
		return fmt.Errorf("failed to create post: %e", err)
	}

	return c.Redirect(http.StatusSeeOther, "/")
}
