package routes

import (
	"github.com/danjac/photoshare/api/models"
	"github.com/danjac/photoshare/api/session"
	"github.com/danjac/photoshare/api/validation"
)

func logout(c *Context) *Result {

	if err := c.Logout(); err != nil {
		return c.Error(err)
	}

	return c.OK(session.NewSessionInfo(c.User))

}

func authenticate(c *Context) *Result {

	user, err := c.GetCurrentUser()
	if err != nil {
		return c.Error(err)
	}

	return c.OK(session.NewSessionInfo(user))
}

func login(c *Context) *Result {

	auth := &session.Authenticator{}
	if err := c.ParseJSON(auth); err != nil {
		return c.Error(err)
	}
	user, err := auth.Identify()
	if err != nil {
		if err == session.MissingLoginFields {
			return c.BadRequest("Missing email or password")
		}
		return c.Error(err)
	}
	if !user.IsAuthenticated {
		return c.BadRequest("Invalid email or password")
	}

	if err := c.Login(user); err != nil {
		return c.Error(err)
	}
	return c.OK(session.NewSessionInfo(user))
}

func signup(c *Context) *Result {

	user := &models.User{}

	if err := c.ParseJSON(user); err != nil {
		return c.Error(err)
	}

	// ensure nobody tries to make themselves an admin
	user.IsAdmin = false

	validator := &validation.UserValidator{user}

	if result, err := validator.Validate(); err != nil || !result.OK {
		if err != nil {
			return c.Error(err)
		}
		return c.BadRequest(result)
	}

	if err := userMgr.Insert(user); err != nil {
		return c.Error(err)
	}

	if err := c.Login(user); err != nil {
		return c.Error(err)
	}

	user.IsAuthenticated = true

	return c.OK(session.NewSessionInfo(user))

}
