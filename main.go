package main

import (
	"net/http"

	"github.com/tomocy/goron/session/cookie"
	"github.com/tomocy/goron/session/manager"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

var sessionManager manager.Manager

func init() {
	initTemplates()
	sessionManager, _ = manager.New("memory")
}

func main() {
	e := echo.New()

	t := &Template{}
	e.Renderer = t

	e.Use(middleware.Logger(), middleware.Recover())

	// e.GET("/greet/create", greetNew)
	// e.POST("/greet/create", greetCreate)

	e.GET("/count", countIndex)

	e.Start(":8080")
}

func countIndex(c echo.Context) error {
	sessionID, err := cookie.GetSessionID(c)
	if err != nil {
		session := sessionManager.CreateSession()
		sessionID = session.ID()

		cookie.SetSessionID(c, sessionID)
	}

	session, err := sessionManager.GetSession(sessionID)
	if err != nil {
		return c.Render(http.StatusInternalServerError, "error", err)
	}

	cnt, ok := session.Get("count").(int)
	if !ok {
		cnt = 1
	} else {
		cnt++
	}

	session.Set("count", cnt)

	dat := struct {
		Cnt interface{}
	}{
		Cnt: cnt,
	}

	return c.Render(http.StatusOK, "countIndex", dat)
}
