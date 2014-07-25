package api

import (
	"github.com/coopernurse/gorp"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
	"net/http"
)

// HTTPError is a wrapper for all HTTP-based errors
type HTTPError struct {
	Status      int
	Description string
}

func (h HTTPError) Error() string {
	if h.Description == "" {
		return http.StatusText(h.Status)
	}
	return h.Description
}

func httpError(status int, description string) HTTPError {
	return HTTPError{status, description}
}

type appHandler func(c web.C, w http.ResponseWriter, r *http.Request) error

func (h appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h(web.C{}, w, r)
	handleError(w, r, err)
}
func (h appHandler) ServeHTTPC(c web.C, w http.ResponseWriter, r *http.Request) {
	err := h(c, w, r)
	handleError(w, r, err)
}

// AppContext tracks all global handler stuff
type AppContext struct {
	config     *AppConfig
	ds         *DataStores
	mailer     *Mailer
	fs         FileStorage
	sessionMgr SessionManager
	cache      Cache
}

// NewAppContext creates new AppContext instance
func NewAppContext(config *AppConfig, dbMap *gorp.DbMap) (*AppContext, error) {

	photoDS := NewPhotoDataStore(dbMap)
	userDS := NewUserDataStore(dbMap)

	ds := &DataStores{
		photos: photoDS,
		users:  userDS,
	}

	fs := NewFileStorage(config)
	mailer := NewMailer(config)
	cache := NewCache(config)

	sessionMgr, err := NewSessionManager(config)
	if err != nil {
		return nil, err
	}

	a := &AppContext{
		config:     config,
		ds:         ds,
		fs:         fs,
		sessionMgr: sessionMgr,
		mailer:     mailer,
		cache:      cache,
	}
	return a, nil
}

// GetRouter generates a new set of routes
func GetRouter(config *AppConfig, dbMap *gorp.DbMap) (*web.Mux, error) {

	a, err := NewAppContext(config, dbMap)
	if err != nil {
		return nil, err
	}
	r := web.New()

	r.Use(middleware.EnvInit)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.AutomaticOptions)

	r.Get("/api/photos/", appHandler(a.getPhotos))
	r.Post("/api/photos/", appHandler(a.upload))
	r.Get("/api/photos/search", appHandler(a.searchPhotos))
	r.Get("/api/photos/owner/:ownerID", appHandler(a.photosByOwnerID))

	r.Get("/api/photos/:id", appHandler(a.photoDetail))
	r.Delete("/api/photos/:id", appHandler(a.deletePhoto))
	r.Patch("/api/photos/:id/title", appHandler(a.editPhotoTitle))
	r.Patch("/api/photos/:id/tags", appHandler(a.editPhotoTags))
	r.Patch("/api/photos/:id/upvote", appHandler(a.voteUp))
	r.Patch("/api/photos/:id/downvote", appHandler(a.voteDown))

	r.Get("/api/tags/", appHandler(a.getTags))

	r.Get("/api/auth/", appHandler(a.getSessionInfo))
	r.Post("/api/auth/", appHandler(a.login))
	r.Delete("/api/auth/", appHandler(a.logout))
	r.Post("/api/auth/signup", appHandler(a.signup))
	r.Put("/api/auth/recoverpass", appHandler(a.recoverPassword))
	r.Put("/api/auth/changepass", appHandler(a.changePassword))

	r.Get("/feeds/", appHandler(a.latestFeed))
	r.Get("/feeds/popular/", appHandler(a.popularFeed))
	r.Get("/feeds/owner/:ownerID", appHandler(a.ownerFeed))

	r.Handle("/api/messages/*", messageHandler)
	r.Handle("/*", http.FileServer(http.Dir(config.PublicDir)))
	return r, nil
}
