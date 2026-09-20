package core

import (
	stdctx "context"
	"net/http"

	"github.com/dreego-stack/dreego/core/internal/context"
	"github.com/dreego-stack/dreego/core/internal/render"
	"github.com/dreego-stack/dreego/core/internal/server"
	"github.com/dreego-stack/dreego/internal/session"
	"github.com/dreego-stack/dreego/internal/validate"
)

var ErrRedirect = context.ErrRedirect
var ErrAppBuilt = server.ErrAppBuilt
var ErrRouteConflict = server.ErrRouteConflict
var ErrRenderRouteNotFound = server.ErrRenderRouteNotFound
var ErrDynamicRenderRoute = server.ErrDynamicRenderRoute
var ErrStaticAssetNotFound = server.ErrStaticAssetNotFound
var ErrSessionTooLarge = session.ErrSessionTooLarge
var ErrCookiePathOverride = session.ErrCookiePathOverride

type Context = context.Context

type SSRContext = context.SSRContext
type RenderContext = context.RenderContext

type Component = render.Renderable

type ComponentFunc func(ctx RenderContext) (Result, error)

func (f ComponentFunc) Render(ctx RenderContext) (Result, error) {
	return f(ctx)
}

type App = server.App
type StaticAsset = server.StaticAsset

func New() *App {
	return server.New()
}

func NewSSR(w http.ResponseWriter, r *http.Request) *SSRContext {
	return context.NewSSR(w, r)
}

type Store = session.Store
type Options = session.Options
type CookieStore = session.CookieStore
type CookiePolicy = session.CookiePolicy

func NewCookieStore(secret []byte) *CookieStore {
	return session.NewCookieStore(secret)
}

func StoreFromCtx(ctx stdctx.Context) Store {
	return session.StoreFromCtx(ctx)
}

func ValidateForm(form any) map[string]string {
	return validate.ValidateForm(form)
}

type Setter = validate.Setter

func SaveOld(c *SSRContext, form any) {
	validate.SaveOld(c, form)
}

func SaveErrors(c *SSRContext, errs map[string]string) {
	validate.SaveErrors(c, errs)
}
