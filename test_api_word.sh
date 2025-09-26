#!/bin/bash

echo "=== TEST API AVEC FICHIER WORD/TXT ==="

# Démarrer le serveur en arrière-plan
echo "Démarrage du serveur..."
go run ./cmd/server/main.go &
SERVER_PID=$!

# Attendre que le serveur démarre
sleep 3

echo "Test avec fichier .txt:"
curl -X POST \
  -F "file=@test_cv_word.txt" \
  http://localhost:8080/extract \
  -H "Content-Type: multipart/form-data" \
  -v

echo -e "\n\n=== FIN DU TEST ==="

# Arrêter le serveur
kill $SERVER_PID 2>/dev/null
