// Package urlrule holds the redirect and rewrite rule validation shared by the
// runtime server and the code generator. It depends only on the standard
// library so the dependency-free core may use it.
package urlrule

import (
	"fmt"
	"net/http"
	"strings"
)

// ValidStatus reports whether status is an allowed redirect status code.
func ValidStatus(status int) bool {
	switch status {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	}
	return false
}

// Redirect validates a redirect rule (pattern, target, status, and self-loop).
func Redirect(from, to string, status int) error {
	if err := Pattern(from); err != nil {
		return err
	}
	if err := Target(to); err != nil {
		return err
	}
	if !ValidStatus(status) {
		return fmt.Errorf("dreego: invalid redirect status %d (allowed: 301,302,303,307,308)", status)
	}
	if err := Pair(from, to); err != nil {
		return err
	}
	if Loops(from, to) {
		return fmt.Errorf("dreego: redirect %q -> %q loops back to itself", from, to)
	}
	return nil
}

// Rewrite validates a rewrite rule (pattern, target, and self-loop).
func Rewrite(from, to string) error {
	if err := Pattern(from); err != nil {
		return err
	}
	if err := Target(to); err != nil {
		return err
	}
	if err := Pair(from, to); err != nil {
		return err
	}
	if Loops(from, to) {
		return fmt.Errorf("dreego: rewrite %q -> %q loops back to itself", from, to)
	}
	return nil
}

// Pair validates a from/to combination independent of the rule kind.
func Pair(from, to string) error {
	if strings.HasSuffix(from, "/*") && (to == "/" || to == "/*") {
		return fmt.Errorf("dreego: redirect/rewrite target %q with wildcard pattern %q emits //-prefixed targets", to, from)
	}
	if !strings.HasSuffix(from, "/*") && strings.HasSuffix(to, "/*") {
		return fmt.Errorf("dreego: redirect/rewrite target %q has wildcard but pattern %q is exact", to, from)
	}
	return nil
}

// Pattern validates a redirect/rewrite source pattern.
func Pattern(p string) error {
	if p == "" {
		return fmt.Errorf("dreego: redirect/rewrite pattern is empty")
	}
	if !strings.HasPrefix(p, "/") {
		return fmt.Errorf("dreego: redirect/rewrite pattern %q must start with /", p)
	}
	if strings.Contains(p, "//") {
		return fmt.Errorf("dreego: redirect/rewrite pattern %q contains //", p)
	}
	if strings.HasSuffix(p, "/") && p != "/" {
		return fmt.Errorf("dreego: redirect/rewrite pattern %q must not end with /", p)
	}
	wildcard := strings.HasSuffix(p, "/*")
	base := p
	if wildcard {
		base = strings.TrimSuffix(p, "/*")
	}
	if wildcard && base == "" {
		return fmt.Errorf("dreego: redirect/rewrite pattern %q has empty prefix", p)
	}
	if !wildcard && strings.Contains(p, "*") {
		return fmt.Errorf("dreego: redirect/rewrite pattern %q invalid wildcard (use /*)", p)
	}
	if wildcard && strings.Count(p, "/*") > 1 {
		return fmt.Errorf("dreego: redirect/rewrite pattern %q has multiple wildcards", p)
	}
	return nil
}

// Target validates a redirect/rewrite target.
func Target(p string) error {
	if p == "" {
		return fmt.Errorf("dreego: redirect/rewrite target is empty")
	}
	if !strings.HasPrefix(p, "/") {
		return fmt.Errorf("dreego: redirect/rewrite target %q must start with /", p)
	}
	if strings.Contains(p, "//") {
		return fmt.Errorf("dreego: redirect/rewrite target %q contains //", p)
	}
	if strings.HasSuffix(p, "/") && p != "/" {
		return fmt.Errorf("dreego: redirect/rewrite target %q must not end with /", p)
	}
	if strings.Contains(p, "*") && !strings.HasSuffix(p, "/*") {
		return fmt.Errorf("dreego: redirect/rewrite target %q invalid wildcard (use /*)", p)
	}
	if strings.Count(p, "/*") > 1 {
		return fmt.Errorf("dreego: redirect/rewrite target %q has multiple wildcards", p)
	}
	return nil
}

// Loops reports whether a rule points back under its own prefix.
func Loops(from, to string) bool {
	if from == to {
		return true
	}
	if !strings.HasSuffix(from, "/*") {
		return false
	}
	fromPrefix := strings.TrimSuffix(from, "/*")
	if !strings.HasSuffix(to, "/*") {
		return to == fromPrefix || strings.HasPrefix(to, fromPrefix+"/")
	}
	toPrefix := strings.TrimSuffix(to, "/*")
	return toPrefix == fromPrefix || strings.HasPrefix(toPrefix, fromPrefix+"/")
}

// Cycle reports whether the redirect and rewrite edges form a cycle.
func Cycle(redirects, rewrites [][2]string) error {
	const (
		gray  = 1
		black = 2
	)
	edges := map[string][]string{}
	for _, r := range rewrites {
		edges[r[0]] = append(edges[r[0]], r[1])
	}
	for _, r := range redirects {
		edges[r[0]] = append(edges[r[0]], r[1])
	}
	color := map[string]int{}
	var visit func(path string) bool
	visit = func(path string) bool {
		switch color[path] {
		case gray:
			return true
		case black:
			return false
		}
		color[path] = gray
		for _, next := range edges[path] {
			if visit(next) {
				return true
			}
		}
		color[path] = black
		return false
	}
	for from := range edges {
		if visit(from) {
			return fmt.Errorf("dreego: redirect/rewrite cycle detected involving %q", from)
		}
	}
	return nil
}
