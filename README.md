# 🚀 Exodus

**Exodus** is a high‑performance, extensible data transport framework built on [Electrician](https://github.com/joeydtaylor/electrician). It streams structured events between services with gRPC, S3, and custom adapters, supporting real‑time ingestion, transformation, and delivery at scale.

## 📦 Features

* **⚡ High‑throughput event streaming** over gRPC with AES‑GCM payload crypto
* **🛡 OAuth2, TLS** for authN/Z and transport security
* **🔌 Pluggable pipelines** with manifest‑defined transformers
* **📂 S3 Parquet sinks** with compression, batching, and rolling windows
* **🧩 Minimal edit surface** — you only touch 4 files
* **🛠 LocalStack‑friendly** for rapid dev

## ✍️ You Only Edit These

1. `./.env`
2. `./app/types/register.go`
3. `./app/handlers/inproc.go`
4. `./manifest.toml`

---

## 🧭 Quickstart (Recommended Flow)

### 1️⃣ Configure Environment (`.env`)

Use this working template:

```bash
# Exodus
EXODUS_MANIFEST=/app/manifest.toml
SERVER_LISTEN_ADDRESS=0.0.0.0:5001

# TLS for gRPC (if you keep ENV fallback)
ELECTRICIAN_RX_TLS_ENABLE=true
ELECTRICIAN_RX_TLS_SERVER_CRT=/app/etc/keys/tls/server.crt
ELECTRICIAN_RX_TLS_SERVER_KEY=/app/etc/keys/tls/server.key
ELECTRICIAN_RX_TLS_CA=/app/etc/keys/tls/ca.crt
ELECTRICIAN_RX_TLS_SERVER_NAME=localhost

# Payload crypto
ELECTRICIAN_COMPRESS=snappy
ELECTRICIAN_ENCRYPT=aesgcm
ELECTRICIAN_AES256_KEY_HEX=ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec

# S3 (container-to-container)
S3_REGION=us-east-1
S3_ENDPOINT=http://host.docker.internal:4566
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_ASSUME_ROLE_ARN=arn:aws:iam::000000000000:role/exodus-dev-role
S3_BUCKET=steeze-dev
S3_PREFIX_TEMPLATE=org={org}/feedback/demo/{yyyy}/{MM}/{dd}/{HH}/{mm}/
ORG_ID=4d948fa0-084e-490b-aad5-cfd01eeab79a

# SSE / KMS
S3_SSE_MODE=aws:kms
S3_KMS_KEY_ARN=arn:aws:kms:us-east-1:000000000000:alias/electrician-dev

# Writer knobs (fast dev flush)
PARQUET_COMPRESSION=zstd
ROLL_WINDOW_MS=2000         # roll every 2 seconds
ROLL_MAX_RECORDS=5          # or after 5 records
BATCH_MAX_RECORDS=5         # send batch to S3 after 5 records
BATCH_MAX_BYTES_MB=1        # or after ~1 MiB
BATCH_MAX_AGE_MS=2000       # or after 2 seconds

# Damocles session/assertion
SESSION_STATE_API=https://damocles:3000/api/auth/session
SESSION_COOKIE_NAME=s
ADMIN_ROLE_NAME=admin
DEVELOPER_ROLE_NAME=developer

ASSERTION_KEY_URL=https://damocles:3000/api/auth/oauth/public-key.pem
ASSERTION_COOKIE_NAME=assert
ASSERTION_ISSUER=auth-service
ASSERTION_AUDIENCE=session-assertion
ASSERTION_LEEWAY_SECONDS=60
```

### 2️⃣ Define Types & Transformers (`app/types/register.go`)

```go
// app/types/register.go
package types

import (
	"strings"

	"github.com/joeydtaylor/exodus/exodus"
	"github.com/joeydtaylor/exodus/exodus/codec"
	"github.com/joeydtaylor/exodus/exodus/transform"
	"github.com/joeydtaylor/exodus/pkg/electrician"
)

type Feedback struct {
	CustomerID string   `parquet:"name=customerId, type=BYTE_ARRAY, convertedtype=UTF8" json:"customerId"`
	Content    string   `parquet:"name=content, type=BYTE_ARRAY, convertedtype=UTF8" json:"content"`
	Category   string   `parquet:"name=category, type=BYTE_ARRAY, convertedtype=UTF8" json:"category,omitempty"`
	IsNegative bool     `parquet:"name=isNegative, type=BOOLEAN" json:"isNegative"`
	Tags       []string `parquet:"name=tags, type=LIST, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8" json:"tags,omitempty"`
}

func RegisterAll() {
	exodus.MustRegisterType[Feedback]("feedback.v1", codec.JSONStrict)
	electrician.EnableBuilderType[Feedback]("feedback.v1")

	// Manifest-visible transformers for feedback.v1
	transform.Register[Feedback]("feedback.v1", "sentiment", func(f Feedback) (Feedback, error) {
		low := strings.ToLower(f.Content)
		if strings.Contains(low, "love") || strings.Contains(low, "great") || strings.Contains(low, "happy") {
			f.Tags = append(f.Tags, "Positive Sentiment")
		} else {
			f.Tags = append(f.Tags, "Needs Attention")
		}
		return f, nil
	})
	transform.Register[Feedback]("feedback.v1", "tagger", func(f Feedback) (Feedback, error) {
		if f.IsNegative {
			f.Tags = append(f.Tags, "neg")
		}
		return f, nil
	})
	transform.Register[Feedback]("feedback.v1", "audit-only", func(f Feedback) (Feedback, error) {
		// no-op
		return f, nil
	})
}
```

### 3️⃣ Define In‑Process Handlers (`app/handlers/inproc.go`)

```go
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

	// POST /echo (echo request body; defaults to {} when empty)
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
```

### 4️⃣ Define Routes & Sinks (`manifest.toml`)

```toml
# =========================
# Receivers → wires
# =========================

[[receiver]]
address = "0.0.0.0:50053"
buffer_size = 1024
aes256_key_hex = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"

  [receiver.tls]
  enable = true
  server_cert = "/app/etc/keys/tls/server.crt"
  server_key  = "/app/etc/keys/tls/server.key"
  ca          = "/app/etc/keys/tls/ca.crt"
  server_name = "localhost"

  [receiver.oauth]
  mode = "merge"
  issuer_base = "https://damocles:3000"
  jwks_url = "https://damocles:3000/api/auth/.well-known/jwks.json"
  required_aud = ["your-api"]        # ← set to what Damocles actually mints
  required_scopes = ["write:data"]   # ← set to what you mint
  jwks_cache_seconds = 300
  introspect_url = "https://damocles:3000/api/auth/oauth/introspect"
  auth_type = "basic"
  client_id = "steeze-local-cli"
  client_secret = "local-secret"
  cache_seconds = 300

  [[receiver.pipeline]]
  datatype     = "feedback.v1"
  transformers = ["sentiment", "tagger"]
  output       = "wireA"

[[receiver]]
address = "0.0.0.0:50054"
buffer_size = 1024
aes256_key_hex = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"

  [receiver.tls]
  enable = true
  server_cert = "/app/etc/keys/tls/server.crt"
  server_key  = "/app/etc/keys/tls/server.key"
  ca          = "/app/etc/keys/tls/ca.crt"
  server_name = "localhost"

  [receiver.oauth]
  mode = "merge"
  issuer_base = "https://damocles:3000"
  jwks_url = "https://damocles:3000/api/auth/.well-known/jwks.json"
  required_aud = ["your-api"]
  required_scopes = ["write:data"]
  jwks_cache_seconds = 300
  introspect_url = "https://damocles:3000/api/auth/oauth/introspect"
  auth_type = "basic"
  client_id = "steeze-local-cli"
  client_secret = "local-secret"
  cache_seconds = 300

  [[receiver.pipeline]]
  datatype     = "feedback.v1"
  transformers = ["tagger"]
  output       = "wireB"

# =========================
# Sink: S3 parquet (fan-in)
# =========================

[[sink]]
type   = "s3"
name   = "debug-s3"
inputs = ["wireA", "wireB"]

  [sink.s3]
  bucket           = "steeze-dev"
  prefix_template  = "debug/{yyyy}/{MM}/{dd}/{HH}/{mm}/"  # drop {org} for now
  format           = "parquet"
  minute_partitions = false

    [sink.s3.batch]
    max_records = 1          # flush every record
    max_bytes   = 1
    max_age_ms  = 1000

    [sink.s3.parquet]
    compression      = "zstd"
    roll_window_ms   = 1000
    roll_max_records = 1

    [sink.s3.sse]
    type          = "aes256"   # avoid KMS during debug

    [sink.s3.aws]
    region           = "us-east-1"
    role_arn         = ""       # disable AssumeRole; use static creds
    session_name     = ""
    duration_minutes = 0
    endpoint_url     = "http://localstack:4566"
    use_path_style   = true
    static_access_key_id     = "test"
    static_secret_access_key = "test"
    insecure_skip_tls_verify = true
```

### 5️⃣ Run It

```bash
docker compose up --build
```

Then POST records matching `feedback.v1` to your receiver; Exodus will transform, batch, and write Parquet to S3 per your manifest.

---

## 📤 Output Characteristics

* Parquet with configurable compression (e.g., ZSTD)
* Record/time‑based rolling windows
* Fan‑in from multiple wires into one sink

## 🧩 How to Extend

* Add new payload structs and register with `MustRegisterType`
* Create transformers via `transform.Register`
* Reference datatypes/transformers in `manifest.toml`
* Add in‑proc handlers and route to them via your ingress

## 📜 License

Licensed under the terms of the [LICENSE](LICENSE) file.
