#!/bin/bash

echo "🧪 Test de l'endpoint /boond/orgchart"
echo "====================================="

SERVER_URL="http://localhost:8081"
TEST_FILE="test_orgchart.json"

echo "📡 Envoi de la requête POST vers $SERVER_URL/boond/orgchart"
echo "📄 Utilisation du fichier: $TEST_FILE"
echo ""

curl -X POST \
  -H "Content-Type: application/json" \
  -d @$TEST_FILE \
  "$SERVER_URL/boond/orgchart" \
  -w "\n\n📊 Status HTTP: %{http_code}\n⏱️  Temps de réponse: %{time_total}s\n" \
  -v

echo ""
echo "✅ Test terminé"


