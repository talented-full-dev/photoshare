package routes

import (
	"encoding/json"
	"fmt"
	"github.com/danjac/photoshare/api/models"
	"github.com/danjac/photoshare/api/session"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Result struct {
	Context *Context
	Code    int
	Payload interface{}
	Error   error
}

func (r *Result) Render() error {
	if r.Error != nil {
		http.Error(r.Context.Response, "Sorry, an error has occurred", r.Code)
		return r.Error
	}

	r.Context.Response.WriteHeader(r.Code)

	if r.Payload != nil {
		r.Context.Response.Header().Set("Content-type", "application/json")
		return json.NewEncoder(r.Context.Response).Encode(r.Payload)
	}

	return nil
}

type Context struct {
	*http.Request
	Response http.ResponseWriter
	Params   map[string]string
	User     *models.User
}

func (c *Context) Result(code int, payload interface{}, err error) *Result {
	return &Result{c, code, payload, err}
}

// Renders feed in Atom format
func (c *Context) Atomize(feed *feeds.Feed) *Result {
	atom, err := feed.ToAtom()
	if err != nil {
		return c.Error(err)
	}

	c.Response.Header().Set("Content-Type", "application/atom+xml")
	c.Response.Write([]byte(atom))

	return c.Result(http.StatusOK, nil, nil)
}

func (c *Context) Param(name string) string {
	return c.Params[name]
}

func (c *Context) Scheme() string {
	if c.Request.TLS == nil {
		return "http"
	}
	return "https"
}

func (c *Context) BaseURL() string {
	return fmt.Sprintf("%s://%s", c.Scheme(), c.Request.Host)
}

func (c *Context) GetCurrentUser() (*models.User, error) {
	var err error
	if c.User != nil {
		return c.User, nil
	}
	c.User, err = session.GetCurrentUser(c.Request)
	return c.User, err
}

func (c *Context) Login(user *models.User) error {
	c.User = user
	_, err := session.Login(c.Response, user)
	return err
}

func (c *Context) Logout() error {
	if c.User != nil {
		c.User.IsAuthenticated = false
	}
	_, err := session.Logout(c.Response)
	return err
}

func (c *Context) OK(value interface{}) *Result {
	return c.Result(http.StatusOK, value, nil)
}

func (c *Context) Unauthorized(value interface{}) *Result {
	return c.Result(http.StatusUnauthorized, value, nil)
}

func (c *Context) Forbidden(value interface{}) *Result {
	return c.Result(http.StatusForbidden, value, nil)
}

func (c *Context) BadRequest(value interface{}) *Result {
	return c.Result(http.StatusBadRequest, value, nil)
}

func (c *Context) NotFound(value interface{}) *Result {
	return c.Result(http.StatusNotFound, value, nil)
}

func (c *Context) Error(err error) *Result {
	return c.Result(http.StatusInternalServerError, nil, err)
}

func (c *Context) ParseJSON(value interface{}) error {
	return json.NewDecoder(c.Request.Body).Decode(value)
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{r, w, mux.Vars(r), nil}
}

type AppHandlerFunc func(c *Context) *Result

func MakeAppHandler(fn AppHandlerFunc, loginRequired bool) http.HandlerFunc {

	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
		}
	}()
	return func(w http.ResponseWriter, r *http.Request) {
		c := NewContext(w, r)
		if loginRequired {
			if user, err := c.GetCurrentUser(); err != nil || user == nil {
				if err != nil {
					panic(err)
					return
				}
				c.Unauthorized("You must be logged in")
				return
			}
		}

		result := fn(c)

		if err := result.Render(); err != nil {
			log.Println("ERROR:", c.Request)
			panic(err)
		}
	}

}
