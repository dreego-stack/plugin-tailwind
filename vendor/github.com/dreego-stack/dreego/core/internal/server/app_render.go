package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	rendercontext "github.com/dreego-stack/dreego/core/internal/context"
	rendercore "github.com/dreego-stack/dreego/core/internal/render"
)

var ErrRenderRouteNotFound = errors.New("dreego: render route not found")
var ErrDynamicRenderRoute = errors.New("dreego: dynamic render route is unsupported")

func (a *App) RegisterRender(routePath string, component rendercore.Renderable) error {
	if component == nil {
		return errors.New("dreego: render component is nil")
	}
	if routePath == "" || routePath[0] != '/' || path.Clean(routePath) != routePath {
		return fmt.Errorf("dreego: invalid render route %q", routePath)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if err := a.mutable(); err != nil {
		return err
	}
	if _, exists := a.renderPages[routePath]; exists {
		return fmt.Errorf("%w: GET %s", ErrRouteConflict, routePath)
	}
	a.renderPages[routePath] = component
	return nil
}

func (a *App) RenderPage(routePath string) (rendercore.Result, error) {
	a.mu.RLock()
	component, exists := a.renderPages[routePath]
	routes := append([]route(nil), a.routes...)
	a.mu.RUnlock()
	if !exists {
		pattern, err := a.matchingDynamicGETRoute(routes, routePath)
		if err != nil {
			return rendercore.Result{}, err
		}
		if pattern != "" {
			return rendercore.Result{}, fmt.Errorf("%w: %s matches %s; use SSR or define a literal desktop route", ErrDynamicRenderRoute, routePath, pattern)
		}
		return rendercore.Result{}, fmt.Errorf("%w: %s", ErrRenderRouteNotFound, routePath)
	}
	return component.Render(rendercontext.NewRender(nil))
}

func (a *App) matchingDynamicGETRoute(routes []route, routePath string) (string, error) {
	mux := http.NewServeMux()
	if err := a.handleMux(func() {
		for _, candidate := range routes {
			if candidate.method != http.MethodGet || !strings.Contains(candidate.pattern, "{") {
				continue
			}
			mux.HandleFunc(http.MethodGet+" "+candidate.pattern, candidate.handler)
		}
	}); err != nil {
		return "", err
	}
	request := &http.Request{Method: http.MethodGet, URL: &url.URL{Path: routePath}}
	_, pattern := mux.Handler(request)
	return strings.TrimPrefix(pattern, http.MethodGet+" "), nil
}
