package main

import (
	"strings"
)

type Content struct {
	Markdown string
	Html     string
}

func CompileContent(markdown string) (content Content, err error) {
	content.Markdown = markdown
	content.Html = compile(markdown)
	err = nil
	return
}

var (
	htmlEscapeMap = map[byte]string{
		'&':  "&amp;",
		'\'': "&#39;", // "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
		'<':  "&lt;",
		'>':  "&gt;",
		'"':  "&#34;",
	}
)

func compile(s string) string {
	var sb strings.Builder

	newLines := 0

	for _, c := range []byte(s) {
		if c == '\r' {
			continue
		}

		if c == '\n' {
			newLines++
			if newLines <= 2 {
				sb.WriteString("<br/>")
			}
			continue
		} else {
			newLines = 0
		}

		if rep, ok := htmlEscapeMap[c]; ok {
			sb.WriteString(rep)
		} else {
			sb.WriteByte(c)
		}
	}

	return sb.String()
}
