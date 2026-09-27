package server

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
)

var ErrStaticAssetNotFound = errors.New("dreego: static asset not found")

type StaticAsset struct {
	MIME    string
	Content []byte
}

func (a *App) RegisterStatic(assetPath, mime string, content []byte) error {
	if !validStaticAssetPath(assetPath) {
		return fmt.Errorf("dreego: invalid static asset path %q", assetPath)
	}
	data := append([]byte(nil), content...)
	if err := a.Register(http.MethodGet, assetPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", mime)
		_, _ = w.Write(data)
	}); err != nil {
		return err
	}
	a.mu.Lock()
	a.staticAssets[assetPath] = StaticAsset{MIME: mime, Content: data}
	a.mu.Unlock()
	return nil
}

func (a *App) StaticAsset(assetPath string) (StaticAsset, error) {
	if !validStaticAssetPath(assetPath) {
		return StaticAsset{}, fmt.Errorf("%w: %s", ErrStaticAssetNotFound, assetPath)
	}
	a.mu.RLock()
	asset, exists := a.staticAssets[assetPath]
	a.mu.RUnlock()
	if !exists {
		return StaticAsset{}, fmt.Errorf("%w: %s", ErrStaticAssetNotFound, assetPath)
	}
	asset.Content = append([]byte(nil), asset.Content...)
	return asset, nil
}

func validStaticAssetPath(assetPath string) bool {
	return strings.HasPrefix(assetPath, "/") &&
		!strings.ContainsAny(assetPath, "\\\x00") &&
		!strings.Contains(assetPath, "%") &&
		path.Clean(assetPath) == assetPath
}
