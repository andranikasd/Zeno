# ---------- builder ----------
FROM golang:1.23-alpine@sha256:fb7ea5cd19bc4eea3eb0d1972919ec0f6229b138985ce4b35ce5846c6bc02973 AS builder

# enable cgo‐free static build
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

# cache deps
COPY go.mod go.sum ./
RUN go mod download

# copy & build
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o /app/zeno ./cmd/zeno

# ---------- runtime ----------
FROM scratch

# drop in the static binary
COPY --from=builder /app/zeno /zeno

ENTRYPOINT ["/zeno"]
