package main

import (
	"github.com/joeydtaylor/exodus/app/handlers"
	"github.com/joeydtaylor/exodus/app/types"
	"github.com/joeydtaylor/steeze-core/pkg/serverfx"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

func main() {
	_ = godotenv.Load()
	fx.New(
		serverfx.Module(
			serverfx.WithService("exodus"),
			serverfx.WithManifestEnv("EXODUS_MANIFEST"),
		),
		// App-specific registrations:
		fx.Invoke(types.RegisterAll),
		fx.Invoke(handlers.Register),
	).Run()
}
