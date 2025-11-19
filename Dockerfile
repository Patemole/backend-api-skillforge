# ---- build stage ----
    FROM golang:1.24.3-bullseye AS builder

    # Install Tesseract and dependencies for CGO build
    RUN apt-get update && apt-get install -y \
        tesseract-ocr \
        libtesseract-dev \
        libleptonica-dev \
        && rm -rf /var/lib/apt/lists/*

    WORKDIR /app
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    # Enable CGO for Tesseract OCR support
    RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /srv/main ./cmd/server
    
    # ---- run stage ----
    FROM debian:bullseye-slim
    
    # Installer Python 3, pip et les dépendances système pour WeasyPrint
    # + Tesseract OCR et poppler-utils pour l'extraction OCR des PDFs image-based
    RUN apt-get update && apt-get install -y \
        python3 \
        python3-pip \
        python3-dev \
        libpango-1.0-0 \
        libpangoft2-1.0-0 \
        libgdk-pixbuf2.0-0 \
        libffi7 \
        libcairo2 \
        libharfbuzz0b \
        libfreetype6 \
        libjpeg62-turbo \
        shared-mime-info \
        fontconfig \
        fonts-dejavu-core \
        fonts-dejavu \
        fonts-liberation \
        fonts-noto-core \
        tesseract-ocr \
        tesseract-ocr-fra \
        tesseract-ocr-eng \
        poppler-utils \
        && rm -rf /var/lib/apt/lists/*
    
    # Installer WeasyPrint via pip et s'assurer que le script est dans le PATH
    RUN pip3 install --no-cache-dir weasyprint==66.0 && \
        ln -sf /usr/local/bin/weasyprint /usr/bin/weasyprint 2>/dev/null || true
    
    # Configurer fontconfig pour garantir un DPI uniforme (96 DPI standard web)
    # Cela assure un rendu identique entre local et production
    RUN printf '<?xml version="1.0"?>\n<!DOCTYPE fontconfig SYSTEM "fonts.dtd">\n<fontconfig>\n  <match target="pattern">\n    <edit name="dpi" mode="assign"><double>96</double></edit>\n  </match>\n</fontconfig>\n' > /etc/fonts/local.conf && \
        fc-cache -fv
    
    WORKDIR /srv
    COPY --from=builder /srv/main .
    
    EXPOSE 8080
    ENV PORT=8080
    CMD ["./main"]
    