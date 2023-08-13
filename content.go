package main

import (
	"io"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type Content struct {
	Markdown string
	Html     string
}

func CompileContent(markdown string) (content Content, err error) {
	content.Markdown = markdown
	// todo: markdown compile
	content.Html = string(mdToHtml([]byte(markdown)))
	err = nil
	return
}

func mdToHtml(md []byte) []byte {
	extensions := parser.CommonExtensions | parser.NoEmptyLineBeforeBlock
	doc := parser.NewWithExtensions(extensions).Parse(md)
	renderer := html.NewRenderer(html.RendererOptions{
		Flags:          html.CommonFlags | html.HrefTargetBlank | html.NoopenerLinks,
		RenderNodeHook: htmlRenderHook,
	})

	return markdown.Render(doc, renderer)
}

func htmlRenderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	// set minimum heading level to 4
	if heading, ok := node.(*ast.Heading); ok {
		if heading.Level < 4 {
			heading.Level = 4
		}
	}

	// do not parse html
	if _, ok := node.(*ast.HTMLBlock); ok {
		return ast.GoToNext, true
	}
	if _, ok := node.(*ast.HTMLSpan); ok {
		return ast.GoToNext, true
	}

	return ast.GoToNext, false
}
