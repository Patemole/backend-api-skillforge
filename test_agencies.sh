#!/bin/bash

# Test de l'endpoint /agencies
echo "🧪 Test de l'endpoint /boond/agencies"
echo "====================================="

# URL du serveur (ajustez selon votre configuration)
SERVER_URL="http://localhost:8081"

# Fichier de test
TEST_FILE="test_agencies.json"

echo "📡 Envoi de la requête POST vers $SERVER_URL/boond/agencies"
echo "📄 Utilisation du fichier: $TEST_FILE"
echo ""

# Exécution de la requête
curl -X POST \
  -H "Content-Type: application/json" \
  -d @$TEST_FILE \
  "$SERVER_URL/boond/agencies" \
  -w "\n\n📊 Status HTTP: %{http_code}\n⏱️  Temps de réponse: %{time_total}s\n" \
  -v

echo ""
echo "✅ Test terminé"
