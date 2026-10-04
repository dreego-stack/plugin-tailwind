package server

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	mw "github.com/dreego-stack/dreego/internal/middleware"
	sess "github.com/dreego-stack/dreego/internal/session"
)

type ProfileCookie struct {
	Name     string
	Path     string
	SameSite http.SameSite
}

type Profile struct {
	Session Store
	CSRF    *bool
	Cookie  *ProfileCookie
}

type profileBinding struct {
	pattern string
	name    string
}

func (a *App) Profile(name string, p Profile) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("dreego: profile name must not be empty")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.mutable(); err != nil {
		return err
	}
	if _, exists := a.profiles[name]; exists {
		return fmt.Errorf("dreego: profile %q is already registered", name)
	}
	applyProfileCookie(p.Session, p.Cookie)
	a.profiles[name] = p
	return nil
}

func (a *App) ApplyProfile(pattern, name string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.mutable(); err != nil {
		return err
	}
	if !strings.HasPrefix(pattern, "/") {
		return fmt.Errorf("dreego: ApplyProfile pattern %q must start with '/'; bind a route path such as \"/hooks\"", pattern)
	}
	a.profileBindings = append(a.profileBindings, profileBinding{pattern: pattern, name: name})
	return nil
}

func (a *App) validateProfileBindings() error {
	for _, binding := range a.profileBindings {
		if _, exists := a.profiles[binding.name]; !exists {
			return fmt.Errorf("dreego: unknown profile %q bound to route %q", binding.name, binding.pattern)
		}
	}
	return nil
}

func applyProfileCookie(store Store, cookie *ProfileCookie) {
	if store == nil || cookie == nil {
		return
	}
	cs, ok := store.(*sess.CookieStore)
	if !ok {
		return
	}
	cs.SetCookiePolicy(sess.CookiePolicy{
		Name:     cookie.Name,
		Path:     cookie.Path,
		SameSite: cookie.SameSite,
	})
}

func validateStore(store Store) error {
	v, ok := store.(storeValidator)
	if !ok {
		return nil
	}
	return v.Validate()
}

func (a *App) hasProfiles() bool {
	return len(a.profileBindings) > 0
}

func (a *App) hasSessionCoverage() bool {
	if a.sessionStore != nil {
		return true
	}
	for _, binding := range a.profileBindings {
		if a.profiles[binding.name].Session != nil {
			return true
		}
	}
	return false
}

type profileRoute struct {
	prefix  string
	handler http.Handler
}

type profileDispatcher struct {
	routes   []profileRoute
	fallback http.Handler
}

func (d *profileDispatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range d.routes {
		if pathUnderPrefix(route.prefix, r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	d.fallback.ServeHTTP(w, r)
}

func (a *App) buildSessionStack(base http.Handler) http.Handler {
	if len(a.profileBindings) == 0 {
		return wrapWithSession(base, a.sessionStore, a.csrfEnabled, a.errorHandlers[http.StatusForbidden])
	}
	routes := make([]profileRoute, 0, len(a.profileBindings))
	for _, binding := range a.profileBindings {
		p := a.profiles[binding.name]
		store := p.Session
		if store == nil {
			store = a.sessionStore
		}
		csrf := a.csrfEnabled
		if p.CSRF != nil {
			csrf = *p.CSRF
		}
		routes = append(routes, profileRoute{
			prefix:  patternPrefix(binding.pattern),
			handler: wrapWithSession(base, store, csrf, a.errorHandlers[http.StatusForbidden]),
		})
	}
	slices.SortStableFunc(routes, func(x, y profileRoute) int { return len(y.prefix) - len(x.prefix) })
	return &profileDispatcher{routes: routes, fallback: base}
}

func wrapWithSession(next http.Handler, store Store, csrf bool, onForbidden http.HandlerFunc) http.Handler {
	if store == nil {
		return next
	}
	h := next
	if csrf {
		h = mw.CSRFWithForbidden(store, onForbidden)(h)
	}
	h = sessionMiddlewareFor(store)(h)
	return h
}

func pathUnderPrefix(prefix, path string) bool {
	if prefix == "" || prefix == "/" {
		return true
	}
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func patternPrefix(pattern string) string {
	p := strings.TrimSpace(pattern)
	p = strings.TrimSuffix(p, "/{$}")
	p = strings.TrimSuffix(p, "/*")
	if i := strings.IndexByte(p, '{'); i >= 0 {
		p = p[:i]
	}
	p = strings.TrimSuffix(p, "/")
	if p == "" {
		return "/"
	}
	return p
}

func sessionMiddlewareFor(store Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, sess.WithStore(r, store))
		})
	}
}
