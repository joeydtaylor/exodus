// bundlefx/bundlefx.go
package bundlefx

import (
	"github.com/joeydtaylor/exodus/middleware/auth"
	"github.com/joeydtaylor/exodus/middleware/logger"
	"github.com/joeydtaylor/exodus/middleware/metrics"
	"go.uber.org/fx"
)

// Module provided to fx
var Module = fx.Options(
	auth.Module,
	logger.Module,
	metrics.Module,
)
