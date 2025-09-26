#!/bin/bash

echo "=== TEST API AVEC SUPPORT WORD ==="

# Démarrer le serveur en arrière-plan
echo "Démarrage du serveur..."
go run ./cmd/server/main.go > server.log 2>&1 &
SERVER_PID=$!

# Attendre que le serveur démarre
echo "Attente du démarrage du serveur..."
sleep 5

# Vérifier que le serveur est en cours d'exécution
if ! kill -0 $SERVER_PID 2>/dev/null; then
    echo "❌ Le serveur n'a pas démarré correctement"
    cat server.log
    exit 1
fi

echo "✅ Serveur démarré (PID: $SERVER_PID)"

# Test avec fichier texte
echo -e "\n📄 Test avec fichier .txt:"
curl -X POST \
  -F "file=@test_cv_word.txt" \
  http://localhost:8080/extract \
  -H "Content-Type: multipart/form-data" \
  -s \
  -w "\nStatus: %{http_code}\nTemps: %{time_total}s\n" \
  -o response_txt.json

echo "Réponse:"
cat response_txt.json

echo -e "\n=== FIN DU TEST ==="

# Arrêter le serveur
echo "Arrêt du serveur..."
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null

# Nettoyer
rm -f server.log response_txt.json
