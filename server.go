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
		fx.Invoke(types.RegisterAll),
		fx.Invoke(handlers.Register),

		serverfx.Module(serverfx.Options{
			Service:         "exodus",
			ManifestEnv:     "EXODUS_MANIFEST",
			DefaultManifest: "manifest.toml",
			ListenAddrEnv:   "SERVER_LISTEN_ADDRESS",
			DefaultListen:   ":5001", // pick your default; env overrides
			TLSCertEnv:      "SSL_SERVER_CERTIFICATE",
			TLSKeyEnv:       "SSL_SERVER_KEY",
		}),
	).Run()
}
