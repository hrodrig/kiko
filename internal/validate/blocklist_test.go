package validate

import "testing"

func TestReferrerSpam(t *testing.T) {
	if !ReferrerSpam("https://semalt.com/crawler") {
		t.Fatal("expected spam")
	}
	if ReferrerSpam("https://google.com/") {
		t.Fatal("google not spam")
	}
}

func TestIgnoreIP(t *testing.T) {
	if !IgnoreIP("10.0.0.1", []string{"10.0.0.0/8"}) {
		t.Fatal("expected ignore")
	}
}

func TestDatacenterBlocklist(t *testing.T) {
	bl, err := NewDatacenterBlocklist(nil)
	if err != nil {
		t.Fatal(err)
	}
	if !bl.Contains("35.200.1.1") {
		t.Fatal("expected GCP range match")
	}
	if bl.Contains("203.0.113.1") {
		t.Fatal("TEST-NET should not match")
	}
}
