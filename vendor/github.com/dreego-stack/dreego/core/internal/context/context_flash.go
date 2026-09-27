package context

const flashPrefix = "flash_"

func (c *SSRContext) Flash(key, message string) {
	c.SetSessionVal(flashPrefix+key, message)
}

func (c *SSRContext) FlashGet(key string) string {
	value := c.SessionVal(flashPrefix + key)
	if value != "" {
		c.DelSessionVal(flashPrefix + key)
	}
	return value
}

func (c *SSRContext) FlashPeek(key string) string {
	return c.SessionVal(flashPrefix + key)
}
