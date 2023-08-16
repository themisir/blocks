package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	db, err := CreateMigratedDb(dbPath)
	if err != nil {
		return nil, err
	}

	return &Handler{db, baseUrl}, nil
}

func (h *Handler) HandleGetPosts(c echo.Context) error {
	posts, err := QueryPosts(h.db)
	if err != nil {
		return fmt.Errorf("failed to query posts: %e", err)
	}

	return c.Render(http.StatusOK, "list.html", echo.Map{
		"Posts":         posts,
		"EnablePosting": true,
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

	return c.Render(http.StatusOK, "single.html", echo.Map{
		"Post":          post,
		"EnablePosting": false,
	})
}

type CreatePostRequest struct {
	ID               int    `param:"id"`
	Content          string `json:"content" form:"content"`
	IsAnonymous      bool   `json:"isAnonymous" form:"isAnonymous"`
	RedirectToPostId int    `json:"redirectToPostId" form:"redirectToPostId" query:"redirectToPostId"`
}

func (h *Handler) createPost(req *CreatePostRequest) (id int, err error) {
	// compile and validate content
	req.Content = strings.TrimSpace(req.Content)
	if len(req.Content) < 8 {
		return 0, echo.ErrBadRequest
	}
	content, err := CompileContent(req.Content)
	if err != nil {
		return 0, fmt.Errorf("failed to compile content: %e", err)
	}
	if len(content.Html) < 8 {
		return 0, echo.ErrBadRequest
	}

	userId := 0

	return InsertPost(h.db, CreatePost{
		Content:      content,
		IsAnonymous:  req.IsAnonymous,
		UserId:       userId,
		ParentPostId: req.ID,
	})
}

func (h *Handler) HandleDeletePost(c echo.Context) error {
	type DeletePostRequest struct {
		ID int64 `param:"id"`
	}

	req := new(DeletePostRequest)
	if err := c.Bind(req); err != nil {
		return err
	}

	userId := 0

	if err := DeletePost(h.db, req.ID, userId); err != nil {
		return c.String(http.StatusTeapot, "Imma be a teapot for ya?")
	}

	redirectTo := c.QueryParam("redirectTo")
	if redirectTo == "" {
		redirectTo = "/"
	}

	return c.Redirect(http.StatusSeeOther, redirectTo)
}

func (h *Handler) HandleCreatePost(c echo.Context) error {
	req := new(CreatePostRequest)
	if err := c.Bind(req); err != nil {
		return err
	}

	_, err := h.createPost(req)
	if err != nil {
		return err
	}

	if req.RedirectToPostId > 0 {
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/posts/%v", req.RedirectToPostId))
	}
	if req.ID > 0 {
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/posts/%v", req.ID))
	}
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *Handler) PageUrl(page string) string {
	result, _ := url.JoinPath(h.baseUrl, page)
	return result
}

func (h *Handler) HandleGetRssFeed(c echo.Context) error {
	posts, err := QueryPosts(h.db)
	if err != nil {
		return fmt.Errorf("failed to query posts: %e", err)
	}

	feed := &feeds.Feed{
		Title: "Blocks",
		Link: &feeds.Link{
			Href: h.PageUrl("/"),
		},
		Created: time.Now(),
		Items:   make([]*feeds.Item, len(posts)),
	}

	for i, post := range posts {
		description := post.Content.Markdown
		if len(description) > 256 {
			description = description[:256] + "..."
		}

		author := post.User.Username
		if post.IsAnonymous {
			author = "@anonymous"
		} else {
			author = fmt.Sprintf("@%s", author)
		}

		feed.Items[i] = &feeds.Item{
			Title:       fmt.Sprintf("Block by %s", author),
			Link:        &feeds.Link{Href: h.PageUrl(fmt.Sprintf("/posts/%v", post.Id))},
			Description: description,
			Author:      &feeds.Author{Name: author},
			Created:     post.CreatedAt,
			Content:     post.Content.Html,
		}
	}

	c.Response().Header().Add(echo.HeaderContentType, "application/rss+xml; charset=utf-8")
	c.Response().WriteHeader(http.StatusOK)
	return feed.WriteRss(c.Response())
}

func (h *Handler) HandleGetLogin(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", echo.Map{})
}

func (h *Handler) HandlePostLogin(c echo.Context) error {
	type LoginRequest struct {
		Username string `form:"username" json:"username"`
		Password string `form:"username" json:"password"`
	}

	req := new(LoginRequest)
	if err := c.Bind(req); err != nil {
		return err
	}

	user, err := TryLogin(h.db, req.Username, req.Password)
	if err != nil {
		return c.Render(http.StatusOK, "login.html", echo.Map{
			"Error": "Invalid username or password",
		})
	}

	sessionId, err := CreateUserSession(h.db, user.ID, c.Request().UserAgent(), c.Request().RemoteAddr)
	if err != nil {
		c.Logger().Errorf("unable to create session: %s", err)
		return echo.ErrInternalServerError
	}

	cookieTtl := 30 * 24 * time.Hour

	c.SetCookie(&http.Cookie{
		Name:     "_session",
		Value:    sessionId,
		Path:     "/",
		Expires:  time.Now().Add(cookieTtl),
		MaxAge:   int(cookieTtl.Seconds()),
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return c.Redirect(http.StatusSeeOther, "/")
}
