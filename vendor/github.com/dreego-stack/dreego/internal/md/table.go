package md

import (
	"regexp"
	"strings"
)

var tableSepRe = regexp.MustCompile(`^\s*\|?\s*:?-+:?\s*(\|\s*:?-+:?\s*)*\|?\s*$`)

func IsTableSeparator(s string) bool {
	return tableSepRe.MatchString(s)
}

func buildTable(header, sep string, body []string, r *Renderer) string {
	headers := splitCells(header)
	sepCells := splitCells(sep)
	n := len(headers)
	var b strings.Builder
	b.WriteString("<table><thead><tr>")
	for i, h := range headers {
		b.WriteString(cellTag("th", h, AlignAt(sepCells, i), r))
	}
	b.WriteString("</tr></thead><tbody>")
	for _, row := range body {
		cells := splitCells(row)
		b.WriteString("<tr>")
		for i := range n {
			c := ""
			if i < len(cells) {
				c = cells[i]
			}
			b.WriteString(cellTag("td", c, AlignAt(sepCells, i), r))
		}
		b.WriteString("</tr>")
	}
	b.WriteString("</tbody></table>")
	return b.String()
}

func splitCells(s string) []string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "|")
	s = strings.TrimSuffix(s, "|")
	parts := strings.Split(s, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

func CellAlign(s string) string {
	s = strings.TrimSpace(s)
	left := strings.HasPrefix(s, ":")
	right := strings.HasSuffix(s, ":")
	switch {
	case left && right:
		return "center"
	case right:
		return "right"
	case left:
		return "left"
	}
	return ""
}

func AlignAt(sepCells []string, i int) string {
	if i < len(sepCells) {
		return CellAlign(sepCells[i])
	}
	return ""
}

func cellTag(tag, content, align string, r *Renderer) string {
	var b strings.Builder
	b.WriteString("<")
	b.WriteString(tag)
	if align != "" {
		b.WriteString(` style="text-align: `)
		b.WriteString(align)
		b.WriteString(`"`)
	}
	b.WriteString(">")
	b.WriteString(r.RenderInline(content))
	b.WriteString("</")
	b.WriteString(tag)
	b.WriteString(">")
	return b.String()
}
