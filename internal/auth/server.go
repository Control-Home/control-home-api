package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/schema"

	"github.com/gbaranski/houseflow/pkg/types"
)

// Database is database interface
type Database interface {
	AddUser(ctx context.Context, user types.User) (id string, err error)
	GetUserByEmail(ctx context.Context, email string) (*types.User, error)
}

// Auth hold root server state
type Auth struct {
	db     Database
	Router *chi.Mux
	opts   Options
}

var decoder = schema.NewDecoder()
var encoder = schema.NewEncoder()

// New creates server, it won't run till Auth.Router.Start
func New(db Database, opts Options) Auth {
	a := Auth{
		db:     db,
		Router: chi.NewRouter(),
		opts:   opts,
	}
	a.Router.Use(middleware.Logger)
	a.Router.Use(middleware.Recoverer)
	a.Router.Use(middleware.RealIP)
	a.Router.Use(middleware.Heartbeat("/ping"))
	a.Router.Use(middleware.Timeout(time.Second * 10))

	a.Router.Get("/auth", a.onAuthSite)

	a.Router.Post("/login", a.onLogin)
	a.Router.Post("/register", a.onRegister)
	a.Router.Post("/token", a.onToken)

	return a
}

func (a *Auth) onAuthSite(w http.ResponseWriter, r *http.Request) {
	var query LoginPageQuery

	if err := decoder.Decode(&query, r.URL.Query()); err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if query.ClientID != a.opts.ClientID {
		http.Error(w, "ClientID is invalid", http.StatusForbidden)
		return
	}
	if !a.validateRedirectURI(query.RedirectURI) {
		http.Error(w, "redirect_uri is invalid", http.StatusBadRequest)
		return
	}
	a.opts.LoginSiteTemplate.Execute(w, map[string]string{
		"redirect_uri": query.RedirectURI,
		"state":        query.State,
	})
}
