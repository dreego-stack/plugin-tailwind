package server

import (
	"github.com/dreego-stack/dreego/internal/urlrule"
)

func validateRedirect(from, to string, status int) error {
	return urlrule.Redirect(from, to, status)
}

func validateRewrite(from, to string) error {
	return urlrule.Rewrite(from, to)
}

// validateRedirectCycles builds a transitively-closed graph over the redirect
// and rewrite rules and reports a cycle. Rewrites are applied before redirects,
// so the union of both rule sets forms a single rewrite/redirect chain.
func validateRedirectCycles(redirects []redirectRule, rewrites []rewriteRule) error {
	var rp, wp [][2]string
	for _, r := range redirects {
		rp = append(rp, [2]string{r.from, r.to})
	}
	for _, r := range rewrites {
		wp = append(wp, [2]string{r.from, r.to})
	}
	return urlrule.Cycle(rp, wp)
}
