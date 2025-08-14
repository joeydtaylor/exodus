# ---- build stage ----
FROM golang:1.24.6-alpine AS build
RUN apk add --no-cache ca-certificates git
WORKDIR /src

# cache deps
COPY go.mod go.sum ./
RUN go mod download

# build
COPY . .
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags "-s -w" -o /out/exodus server.go

# ---- runtime stage ----
FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl
WORKDIR /app

# binary
COPY --from=build /out/exodus /app/exodus
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# non-root
RUN adduser -D -H -u 10001 exodus && chown -R exodus:exodus /app
USER exodus

EXPOSE 5001 50053 50054
ENTRYPOINT ["/app/exodus"]
