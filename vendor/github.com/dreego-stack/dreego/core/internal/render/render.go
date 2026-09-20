// Package render defines the internal, target-neutral render contract.
// SSR and desktop hosts consume Result directly. It must not import net/http.
package render

import (
	"github.com/dreego-stack/dreego/core/internal/context"
)

type Result struct {
	HTML []byte
}

type Renderable interface {
	Render(context.RenderContext) (Result, error)
}

func Component(fn func(c context.RenderContext) (Result, error)) (Result, error) {
	ctx := context.NewRender(nil)
	result, err := fn(ctx)
	if err != nil {
		return Result{}, err
	}
	return result, nil
}
