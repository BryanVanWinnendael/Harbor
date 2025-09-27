package handlers

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

const (
	sessionName     = "fmessages"
	sessionFlashKey = "flashmessages-key"
)

// cookie store
func getCookieStore() *sessions.CookieStore {
	return sessions.NewCookieStore([]byte(sessionFlashKey))
}

// SetFlash adds a flash message to the session
func setFlashmessages(c echo.Context, kind, message string) {
	session, _ := getCookieStore().Get(c.Request(), sessionName)
	session.AddFlash(message, kind)
	session.Save(c.Request(), c.Response())
}

// GetFlash retrieves flash messages for a kind
func getFlashmessages(c echo.Context, kind string) []string {
	session, _ := getCookieStore().Get(c.Request(), sessionName)
	flashes := session.Flashes(kind)
	if len(flashes) > 0 {
		session.Save(c.Request(), c.Response())
		result := []string{}
		for _, f := range flashes {
			result = append(result, f.(string))
		}
		return result
	}
	return nil
}
