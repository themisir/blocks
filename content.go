package main

import (
	"html"
)

type Content struct {
	Markdown string
	Html     string
}

func CompileContent(markdown string) (content Content, err error) {
	content.Markdown = markdown
	content.Html = html.EscapeString(markdown)
	err = nil
	return
}
