# ---- build stage ----
    FROM golang:1.24.3-bullseye AS builder

    WORKDIR /app
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /srv/main ./cmd/server
    
    # ---- run stage ----
    FROM gcr.io/distroless/static-debian11
    WORKDIR /srv
    COPY --from=builder /srv/main .
    EXPOSE 8080
    ENV PORT=8080
    CMD ["./main"]
    