// app/handlers/inproc.go
package handlers

import (
	"context"
	"net/http"

	"github.com/joeydtaylor/exodus/exodus"
)

// Register in-process HTTP handlers referenced by manifest "inproc" routes.
func Register() {
	// GET /healthz
	exodus.Register("health.ok", func(ctx context.Context, _ []byte) ([]byte, int, error) {
		return []byte(`{"status":"ok"}`), http.StatusOK, nil
	})

	// POST /echo  (echo request body; defaults to {} when empty)
	exodus.Register("echo.body", func(ctx context.Context, in []byte) ([]byte, int, error) {
		if len(in) == 0 {
			in = []byte(`{}`)
		}
		return in, http.StatusOK, nil
	})

	// GET /admin/ping (guarded by role in manifest)
	exodus.Register("admin.ping", func(ctx context.Context, _ []byte) ([]byte, int, error) {
		return []byte(`{"pong":true}`), http.StatusOK, nil
	})
}
