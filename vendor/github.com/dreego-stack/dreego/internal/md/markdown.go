package md

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

type Mode int

const (
	ModeTrusted Mode = iota
	ModeSafe
)

var (
	AtxHeading     = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	ULItem         = regexp.MustCompile(`^[-*+]\s+(.*)$`)
	OLItem         = regexp.MustCompile(`^(\d+)\.\s+(.*)$`)
	listMarker     = regexp.MustCompile(`^([-*+]|\d+\.)\s*$`)
	codeRe         = regexp.MustCompile("`([^`]*)`")
	linkRe         = regexp.MustCompile(`\[([^\]]*)\]\(([^)]*)\)`)
	strongRe       = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	emRe           = regexp.MustCompile(`\*([^*]+)\*`)
	HTMLBlockStart = regexp.MustCompile(`^</?([a-zA-Z][a-zA-Z0-9-]*)(\s|>|/)`)
	fenceLanguage  = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_+.#-]*$`)
)

func SafeFenceLanguage(info string) string {
	if fenceLanguage.MatchString(info) {
		return info
	}
	return ""
}

func ToHTML(src string, mode Mode) (string, error) {
	nodes, err := ParseBlocks(src, mode)
	if err != nil {
		return "", err
	}
	return strings.Join(nodes, ""), nil
}

func ParseBlocks(src string, mode Mode) ([]string, error) {
	return parseBlocks(src, NewRenderer(mode))
}

func parseBlocks(src string, r *Renderer) ([]string, error) {
	lines := strings.Split(src, "\n")
	var nodes []string
	var para []string

	flushPara := func() {
		if len(para) > 0 {
			nodes = append(nodes, "<p>"+r.RenderInline(strings.Join(para, " "))+"</p>")
			para = nil
		}
	}

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			flushPara()
			lang := SafeFenceLanguage(strings.TrimSpace(strings.TrimPrefix(trimmed, "```")))
			var code []string
			i++
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
				code = append(code, html.EscapeString(lines[i]))
				i++
			}
			open := "<pre><code"
			if lang != "" {
				open += ` class="language-` + html.EscapeString(lang) + `"`
			}
			open += ">"
			nodes = append(nodes, open+strings.Join(code, "\n")+"</code></pre>")
			continue
		}

		if HTMLBlockStart.MatchString(trimmed) {
			flushPara()
			var raw []string
			for i < len(lines) && strings.TrimSpace(lines[i]) != "" {
				if r.mode == ModeSafe {
					raw = append(raw, html.EscapeString(lines[i]))
				} else {
					raw = append(raw, lines[i])
				}
				i++
			}
			i--
			nodes = append(nodes, strings.Join(raw, "\n"))
			continue
		}

		if m := AtxHeading.FindStringSubmatch(trimmed); m != nil {
			flushPara()
			level := len(m[1])
			nodes = append(nodes, fmt.Sprintf("<h%d>%s</h%d>", level, r.RenderInline(m[2]), level))
			continue
		}

		if IsHR(trimmed) {
			flushPara()
			nodes = append(nodes, "<hr>")
			continue
		}

		if strings.HasPrefix(trimmed, ">") {
			flushPara()
			var q []string
			for i < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i]), ">") {
				q = append(q, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[i]), ">")))
				i++
			}
			i--
			nodes = append(nodes, "<blockquote>"+r.RenderInline(strings.Join(q, " "))+"</blockquote>")
			continue
		}

		if listMarker.MatchString(trimmed) {
			return nil, fmt.Errorf("md: unparseable list marker on line %d", i+1)
		}

		if m := ULItem.FindStringSubmatch(trimmed); m != nil {
			flushPara()
			items, next := parseList(lines, i, r)
			i = next
			nodes = append(nodes, "<ul>"+strings.Join(items, "")+"</ul>")
			continue
		}

		if m := OLItem.FindStringSubmatch(trimmed); m != nil {
			flushPara()
			items, next := parseList(lines, i, r)
			i = next
			nodes = append(nodes, "<ol>"+strings.Join(items, "")+"</ol>")
			continue
		}

		if strings.Contains(trimmed, "|") && i+1 < len(lines) && IsTableSeparator(lines[i+1]) {
			flushPara()
			header := trimmed
			sep := strings.TrimSpace(lines[i+1])
			i += 2
			var body []string
			for i < len(lines) && strings.TrimSpace(lines[i]) != "" && strings.Contains(lines[i], "|") {
				body = append(body, strings.TrimSpace(lines[i]))
				i++
			}
			i--
			nodes = append(nodes, buildTable(header, sep, body, r))
			continue
		}

		if m := FootnoteDefRe.FindStringSubmatch(trimmed); m != nil {
			flushPara()
			r.AddDef(m[1], m[2])
			continue
		}

		if trimmed == "" {
			flushPara()
			continue
		}

		para = append(para, trimmed)
	}

	flushPara()
	if r.HasDefs() {
		nodes = append(nodes, r.FootnotesSection())
	}
	return nodes, nil
}

func IsHR(s string) bool {
	if !strings.HasPrefix(s, "-") {
		return false
	}
	for _, r := range s {
		if r != '-' {
			return false
		}
	}
	return len(s) >= 3
}

func isHTMLNameStart(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z'
}
