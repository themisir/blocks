package main

import (
	"time"

	"github.com/labstack/echo/v4"
)

type TzProfile struct {
	loc *time.Location
}

func TzProfileFor(c echo.Context) TzProfile {
	var tzName string
	if user := GetUser(c); user != nil {
		tzName = user.GetClaim("tz")
	}
	if tzName == "" {
		return TzProfile{}
	}

	loc, err := time.LoadLocation(tzName)
	if err != nil {
		c.Logger().Warnf("Unable to load timezone: %s", err)
		return TzProfile{}
	}

	return TzProfile{loc}
}

func (p TzProfile) AdjustPosts(posts []Post) TzProfile {
	if p.loc != nil {
		for i := range posts {
			p.AdjustPost(&posts[i])
		}
	}
	return p
}

func (p TzProfile) AdjustPost(post *Post) TzProfile {
	if p.loc != nil {
		post.CreatedAt = post.CreatedAt.In(p.loc)
	}
	return p
}
