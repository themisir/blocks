package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
	_ "github.com/mattn/go-sqlite3"
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

func (self *Handler) PageUrl(page string) string {
	result, _ := url.JoinPath(self.baseUrl, page)
	return result
}

func (self *Handler) HandleGetRssFeed(c echo.Context) error {
	posts, err := QueryPosts(self.db)
	if err != nil {
		return fmt.Errorf("failed to query posts: %e", err)
	}

	feed := &feeds.Feed{
		Title: "Blocks",
		Link: &feeds.Link{
			Href: self.PageUrl("/"),
		},
		Created: time.Now(),
		Items:   make([]*feeds.Item, len(posts)),
	}

	for i, post := range posts {
		description := post.Content.Markdown
		if len(description) > 256 {
			description = description[:256] + "..."
		}

		feed.Items[i] = &feeds.Item{
			Title:       fmt.Sprintf("Post by %s", post.Author),
			Link:        &feeds.Link{Href: self.PageUrl(fmt.Sprintf("/posts/%v", post.Id))},
			Description: description,
			Author:      &feeds.Author{Name: post.Author},
			Created:     post.CreatedAt,
			Content:     post.Content.Html,
		}
	}

	c.Response().Header().Add(echo.HeaderContentType, "application/rss+xml; charset=utf-8")
	c.Response().WriteHeader(http.StatusOK)
	feed.WriteRss(c.Response())
	return nil
}
