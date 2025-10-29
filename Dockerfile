# ---- build stage ----
    FROM golang:1.24.3-bullseye AS builder

    WORKDIR /app
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /srv/main ./cmd/server
    
    # ---- run stage ----
    FROM debian:bullseye-slim
    
    # Installer Python 3, pip et les dépendances système pour WeasyPrint
    RUN apt-get update && apt-get install -y \
        python3 \
        python3-pip \
        python3-dev \
        libpango-1.0-0 \
        libpangoft2-1.0-0 \
        libgdk-pixbuf2.0-0 \
        libffi8 \
        libcairo2 \
        libharfbuzz0b \
        libfreetype6 \
        libjpeg62-turbo \
        shared-mime-info \
        && rm -rf /var/lib/apt/lists/*
    
    # Installer WeasyPrint via pip et s'assurer que le script est dans le PATH
    RUN pip3 install --no-cache-dir weasyprint==66.0 && \
        ln -sf /usr/local/bin/weasyprint /usr/bin/weasyprint 2>/dev/null || true
    
    WORKDIR /srv
    COPY --from=builder /srv/main .
    
    EXPOSE 8080
    ENV PORT=8080
    CMD ["./main"]
    