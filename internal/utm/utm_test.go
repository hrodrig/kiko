package utm

import "testing"

func TestFromPath(t *testing.T) {
	path, p := FromPath("/blog?utm_source=newsletter&utm_medium=email&utm_campaign=spring&foo=bar")
	if path != "/blog?foo=bar" {
		t.Errorf("path = %q", path)
	}
	if p.Source != "newsletter" || p.Medium != "email" || p.Campaign != "spring" {
		t.Errorf("utm = %+v", p)
	}
}

func TestFromPathNoQuery(t *testing.T) {
	path, p := FromPath("/")
	if path != "/" || p.Source != "" {
		t.Errorf("path=%q utm=%+v", path, p)
	}
}
