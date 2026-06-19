// Package utm extracts campaign parameters from page paths.
package utm

import (
	"net/url"
	"strings"
)

// Params holds standard UTM query fields.
type Params struct {
	Source   string
	Medium   string
	Campaign string
	Term     string
	Content  string
}

// FromPath splits path and query, extracts utm_* params, and returns a path
// without utm_* query keys (other query params are preserved).
func FromPath(rawPath string) (path string, p Params) {
	path = rawPath
	if path == "" {
		path = "/"
	}
	i := strings.IndexByte(path, '?')
	if i < 0 {
		return path, p
	}
	base, qstr := path[:i], path[i+1:]
	if qstr == "" {
		return base, p
	}
	vals, err := url.ParseQuery(qstr)
	if err != nil {
		return path, p
	}
	p.Source = vals.Get("utm_source")
	p.Medium = vals.Get("utm_medium")
	p.Campaign = vals.Get("utm_campaign")
	p.Term = vals.Get("utm_term")
	p.Content = vals.Get("utm_content")
	for _, k := range []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content"} {
		vals.Del(k)
	}
	if len(vals) == 0 {
		return base, p
	}
	return base + "?" + vals.Encode(), p
}
