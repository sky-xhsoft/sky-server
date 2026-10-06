package middleware

import "testing"

func TestShouldAudit(t *testing.T) {
	for method, want := range map[string]bool{"GET": false, "HEAD": false, "OPTIONS": false, "POST": true, "PUT": true, "PATCH": true, "DELETE": true} {
		if got := shouldAudit(method); got != want {
			t.Errorf("shouldAudit(%s) = %v, want %v", method, got, want)
		}
	}
}
