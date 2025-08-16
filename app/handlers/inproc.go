// app/handlers/inproc.go
package handlers

import (
	"context"
	"net/http"

	"github.com/joeydtaylor/steeze-core/pkg/core"
)

// Register in-process HTTP handlers referenced by manifest "inproc" routes.
func Register() {
	// GET /healthz
	core.Register("health.ok", func(ctx context.Context, _ []byte) ([]byte, int, error) {
		return []byte(`{"status":"ok"}`), http.StatusOK, nil
	})

}
