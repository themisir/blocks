package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	db      *sql.DB
	baseUrl string
}

func NewHandler(dbPath string, baseUrl string) (*Handler, error) {
	if dbPath == "" {
		dbPath = "data.sqlite3"
	}
	db, err := CreateMigratedDb(dbPath)
	if err != nil {
		return nil, err
	}

	return &Handler{db, baseUrl}, nil
}

func (h *Handler) HandleGetPosts(c echo.Context) error {
	posts, err := QueryPosts(h.db)
	if err != nil {
		return fmt.Errorf("failed to query posts: %w", err)
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
		return fmt.Errorf("failed to query single post (id: %v): %w", postId, err)
	}

	children, err := QueryChildrenPosts(h.db, post.Id)
	if err != nil {
		return fmt.Errorf("failed to query children: %w", err)
	}

	user := GetUser(c)

	return c.Render(http.StatusOK, "index.html", echo.Map{
		"Children":      children,
		"Post":          post,
		"User":          user,
		"EnablePosting": false,
	})
}

func (h *Handler) HandleCreatePost(c echo.Context) error {
	bodyContent := c.FormValue("content")

	postIdParam := c.Param("id")
	var parentPostId *int64

	if postIdParam != "" {
		postId, err := strconv.Atoi(postIdParam)
		if err != nil {
			return echo.ErrNotFound
		}
		postId64 := int64(postId)
		parentPostId = &postId64
	}

	var username string
	if user := GetUser(c); user != nil {
		username = user.Username()
	}

	if len(bodyContent) < 8 {
		return echo.ErrBadRequest
	}

	content, err := CompileContent(bodyContent)
	if err != nil {
		return fmt.Errorf("failed to compile content: %w", err)
	}

	createdPostId, err := InsertPost(h.db, content, username, parentPostId)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}

	if parentPostId == nil {
		return c.Redirect(http.StatusSeeOther, "/")
	} else {
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/posts/%v#reply-%v", *parentPostId, createdPostId))
	}
}

func (h *Handler) HandleDeletePost(c echo.Context) error {
	postIdParam := c.Param("id")
	postId, err := strconv.Atoi(postIdParam)
	if err != nil {
		return echo.ErrNotFound
	}

	user := GetUser(c)

	if user == nil {
		return echo.ErrUnauthorized
	}

	if err := DeletePost(h.db, int64(postId), user.Username(), user.IsAdmin()); err != nil {
		return err
	}

	return c.Redirect(http.StatusSeeOther, "/")
}
