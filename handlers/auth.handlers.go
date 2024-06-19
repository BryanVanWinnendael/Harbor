package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/BryanVanWinnendael/Harbor/services"
	"github.com/BryanVanWinnendael/Harbor/views/auth_views"

	"github.com/a-h/templ"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

const (
	auth_sessions_key string = "authenticate-sessions"
	auth_key          string = "authenticated"
	user_id_key       string = "user_id"
	username_key      string = "username"
	tzone_key         string = "time_zone"
)

/********** Handlers for Auth Views **********/

type UserServices interface {
	CheckUsername(username string) (services.User, error)
	ChangePassword(username, password string) error
}

func NewAuthHandler(us UserServices) *AuthHandler {

	return &AuthHandler{
		UserServices: us,
	}
}

type AuthHandler struct {
	UserServices UserServices
}

func (ah *AuthHandler) loginHandler(c echo.Context) error {
	loginView := auth_views.Login(false)
	isError = false

	if c.Request().Method == "POST" {
		// obtaining the time zone from the POST request of the login form
		tzone := ""
		if len(c.Request().Header["X-Timezone"]) != 0 {
			tzone = c.Request().Header["X-Timezone"][0]
		}

		user, err := ah.UserServices.CheckUsername(c.FormValue("username"))
		if err != nil {
			if strings.Contains(err.Error(), "no rows in result set") {
				setFlashmessages(c, "error", "There is no user with that username")

				return c.Redirect(http.StatusSeeOther, "/login")
			}

			return echo.NewHTTPError(
				echo.ErrInternalServerError.Code,
				fmt.Sprintf(
					"something went wrong: %s",
					err,
				))
		}

		err = bcrypt.CompareHashAndPassword(
			[]byte(user.Password),
			[]byte(c.FormValue("password")),
		)

		if err != nil {
			setFlashmessages(c, "error", "Incorrect password")

			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// Get Session and setting Cookies
		sess, _ := session.Get(auth_sessions_key, c)
		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   2629743, // in seconds
			HttpOnly: true,
		}

		// Set user as authenticated, their username,
		// their ID and the client's time zone
		sess.Values = map[interface{}]interface{}{
			auth_key:     true,
			user_id_key:  user.ID,
			username_key: user.Username,
			tzone_key:    tzone,
		}
		sess.Save(c.Request(), c.Response())

		if !user.ChangedPassword {
			return c.Redirect(http.StatusSeeOther, "/password")
		}

		return c.Redirect(http.StatusSeeOther, "/")
	}

	// check if the user is already authenticated
	sess, _ := session.Get(auth_sessions_key, c)
	if auth, ok := sess.Values[auth_key].(bool); ok && auth {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	return renderView(c, auth_views.LoginIndex(
		"| Login",
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		loginView,
	))
}

func (ah *AuthHandler) passwordHandler(c echo.Context) error {
	passwordView := auth_views.Password(false)
	isError = false

	sess, _ := session.Get(auth_sessions_key, c)

	username := sess.Values[username_key].(string)

	user, err := ah.UserServices.CheckUsername(username)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, "/password")
	}

	if user.ChangedPassword {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	if c.Request().Method == "POST" {
		tzone := ""
		if len(c.Request().Header["X-Timezone"]) != 0 {
			tzone = c.Request().Header["X-Timezone"][0]
		}

		newPassword := c.FormValue("password")

		err := ah.UserServices.ChangePassword(username, newPassword)
		if err != nil {
			return c.Redirect(http.StatusSeeOther, "/password")
		}

		sess.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   2629743, // in seconds
			HttpOnly: true,
		}

		sess.Values = map[interface{}]interface{}{
			auth_key:     true,
			user_id_key:  user.ID,
			username_key: user.Username,
			tzone_key:    tzone,
		}
		sess.Save(c.Request(), c.Response())

		return c.Redirect(http.StatusSeeOther, "/")
	}

	return renderView(c, auth_views.LoginIndex(
		"| Change Password",
		fromProtected,
		isError,
		getFlashmessages(c, "error"),
		getFlashmessages(c, "success"),
		passwordView,
	))
}

func (ah *AuthHandler) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, _ := session.Get(auth_sessions_key, c)
		if auth, ok := sess.Values[auth_key].(bool); !ok || !auth {
			fromProtected = false
			return echo.NewHTTPError(echo.ErrUnauthorized.Code, "Please provide valid credentials")
		}

		if userId, ok := sess.Values[user_id_key].(int); ok && userId != 0 {
			c.Set(user_id_key, userId) // set the user_id in the context
		}

		if username, ok := sess.Values[username_key].(string); ok && len(username) != 0 {
			c.Set(username_key, username) // set the username in the context
		}

		if tzone, ok := sess.Values[tzone_key].(string); ok && len(tzone) != 0 {
			c.Set(tzone_key, tzone) // set the client's time zone in the context
		}

		fromProtected = true

		return next(c)
	}
}

func renderView(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}
