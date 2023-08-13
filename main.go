package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
	"github.com/themisir/blocks/assets"
	"github.com/themisir/blocks/renderer"
)

func should(e *echo.Echo, err error) {
	if err != nil {
		e.Logger.Fatal(err)
	}
}

func env(name string, fallback string) string {
	if value, ok := os.LookupEnv(name); ok {
		return value
	}
	return fallback
}

func main() {
	e := echo.New()

	addr := env("ADDRESS", ":1323")
	db := env("DB", "data.sqlite3")

	handler, err := NewHandler(db)
	should(e, err)

	e.Renderer, err = renderer.Template(e, assets.FS, "views")
	should(e, err)

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Filesystem: http.FS(assets.FS),
		Root:       "static",
	}))

	e.GET("/", handler.HandleGetPosts)
	e.POST("/posts", handler.HandleCreatePost)
	e.GET("/posts/:id", handler.HandleGetSinglePost)

	should(e, e.Start(addr))
}
