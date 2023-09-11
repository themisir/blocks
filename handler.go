package main

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
	"net/http"
	"net/url"
	"strconv"
)

type Handler struct {
	db      *sql.DB
	baseUrl string
}

func NewHandler(dbPath string, baseUrl string) (*Handler, error) {
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

	return &Handler{db, baseUrl}, nil
}

func (h *Handler) HandleGetPosts(c echo.Context) error {
	posts, err := QueryPosts(h.db)
	if err != nil {
		return fmt.Errorf("failed to query posts: %e", err)
	}

	user := GetUser(c)

	return c.Render(http.StatusOK, "index.html", echo.Map{
		"Posts":         posts,
		"EnablePosting": true,
		"User":          user,
	})
}

func (h *Handler) HandleGetSinglePost(c echo.Context) error {
	postId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return echo.ErrNotFound
	}

	post, err := QuerySinglePost(h.db, postId)
	if err != nil {
		return fmt.Errorf("failed to query single post (id: %v): %e", postId, err)
	}

	user := GetUser(c)

	return c.Render(http.StatusOK, "index.html", echo.Map{
		"Posts":         []Post{post},
		"User":          user,
		"EnablePosting": false,
	})
}

func (h *Handler) HandleCreatePost(c echo.Context) error {
	bodyContent := c.FormValue("content")

	var username string
	if user := GetUser(c); user != nil {
		username = user.Username()
	}

	if len(bodyContent) < 8 {
		return echo.ErrBadRequest
	}

	content, err := CompileContent(bodyContent)
	if err != nil {
		return fmt.Errorf("failed to compile content: %e", err)
	}

	_, err = InsertPost(h.db, Post{
		Content: content,
		Author:  username,
	})
	if err != nil {
		return fmt.Errorf("failed to create post: %e", err)
	}

	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *Handler) PageUrl(page string) string {
	result, _ := url.JoinPath(h.baseUrl, page)
	return result
}
