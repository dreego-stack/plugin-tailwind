package core

import (
	"log/slog"
	"os"
	"sync"

	"github.com/dreego-stack/dreego/internal/md"
)

var markdownLogger = slog.New(slog.NewJSONHandler(os.Stderr, nil))

var trustedWarnOnce sync.Once

func MarkdownToHTML(src string) (string, error) {
	return md.ToHTML(src, md.ModeSafe)
}

func MarkdownToHTMLTrusted(src string) (string, error) {
	trustedWarnOnce.Do(func() {
		markdownLogger.Warn("MarkdownToHTMLTrusted: raw HTML passthrough enabled — only use with trusted content")
	})
	return md.ToHTML(src, md.ModeTrusted)
}
