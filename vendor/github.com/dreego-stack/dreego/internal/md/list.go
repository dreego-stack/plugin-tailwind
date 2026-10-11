package md

import "strings"

func parseList(lines []string, start int, r *Renderer) ([]string, int) {
	items, next := parseListAt(lines, start, indentOf(lines[start]), r)
	return items, next
}

func parseListAt(lines []string, start, indent int, r *Renderer) ([]string, int) {
	var items []string
	i := start
	ordered := isOL(strings.TrimSpace(lines[start]))
	for i < len(lines) {
		line := lines[i]
		ind := indentOf(line)
		if ind < indent {
			break
		}
		content := strings.TrimSpace(line)
		if isOL(content) != ordered {
			break
		}
		var text string
		if m := ULItem.FindStringSubmatch(content); m != nil {
			text = m[1]
		} else if m := OLItem.FindStringSubmatch(content); m != nil {
			text = m[2]
		} else {
			break
		}
		child := nestedList(lines, i, ind, r)
		items = append(items, "<li>"+r.RenderInline(text)+child+"</li>")
		i++
		for i < len(lines) && indentOf(lines[i]) > ind {
			i++
		}
	}
	return items, i - 1
}

func nestedList(lines []string, parent, parentIndent int, r *Renderer) string {
	i := parent + 1
	if i >= len(lines) || indentOf(lines[i]) <= parentIndent {
		return ""
	}
	childIndent := indentOf(lines[i])
	items, next := parseListAt(lines, i, childIndent, r)
	if len(items) == 0 {
		return ""
	}
	_ = next
	if ULItem.MatchString(strings.TrimSpace(lines[i])) {
		return "<ul>" + strings.Join(items, "") + "</ul>"
	}
	return "<ol>" + strings.Join(items, "") + "</ol>"
}

func indentOf(s string) int {
	n := 0
	for _, r := range s {
		if r == ' ' {
			n++
		} else if r == '\t' {
			n += 2
		} else {
			break
		}
	}
	return n
}

func isOL(raw string) bool { return OLItem.MatchString(raw) }
