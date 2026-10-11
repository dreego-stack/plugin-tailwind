package context

import "html"

func (c *SSRContext) CSRFInput() string {
	return `<input type="hidden" name="csrf_token" value="` + html.EscapeString(c.CSRFToken()) + `">`
}
