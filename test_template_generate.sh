#!/bin/bash

# Test script pour l'endpoint de génération de templates
echo "🧪 Test de l'endpoint POST /api/templates/generate"
echo "=================================================="

# Vérifier que le serveur est en cours d'exécution
echo "📡 Vérification du serveur..."
if ! curl -s http://localhost:8081/health > /dev/null; then
    echo "❌ Le serveur n'est pas en cours d'exécution sur le port 8081"
    echo "💡 Lancez d'abord le serveur avec: go run cmd/server/main.go"
    exit 1
fi

echo "✅ Serveur accessible"

# Test de l'endpoint
echo ""
echo "🚀 Test de génération de template..."
echo "-----------------------------------"

response=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -d @test_template_generate.json \
  http://localhost:8081/api/templates/generate)

echo "📤 Réponse reçue:"
echo "$response" | jq '.' 2>/dev/null || echo "$response"

# Vérifier le statut de la réponse
http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -d @test_template_generate.json \
  http://localhost:8081/api/templates/generate)

echo ""
echo "📊 Code de statut HTTP: $http_code"

if [ "$http_code" -eq 200 ]; then
    echo "✅ Test réussi !"
else
    echo "❌ Test échoué avec le code $http_code"
fi
