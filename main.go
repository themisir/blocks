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

func main() {
	e := echo.New()

	addr, addrSet := os.LookupEnv("ADDRESS")
	if !addrSet {
		addr = ":1323"
	}

	handler, err := NewHandler()
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
