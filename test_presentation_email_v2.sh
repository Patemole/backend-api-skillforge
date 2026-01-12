#!/bin/bash

# Test script pour l'endpoint de génération d'emails de présentation V2
echo "🧪 Test de l'endpoint POST /api/email/generate-presentation-v2"
echo "============================================================="

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
echo "🚀 Test de génération d'email de présentation V2..."
echo "---------------------------------------------------"

response=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -d @test_presentation_email_v2.json \
  http://localhost:8081/api/email/generate-presentation-v2)

echo "📤 Réponse reçue:"
echo "$response" | jq '.' 2>/dev/null || echo "$response"

# Vérifier le statut de la réponse
http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -d @test_presentation_email_v2.json \
  http://localhost:8081/api/email/generate-presentation-v2)

echo ""
echo "📊 Code de statut HTTP: $http_code"

if [ "$http_code" -eq 200 ]; then
    echo "✅ Test réussi !"
    echo ""
    echo "🔍 Analyse de la réponse :"
    echo "-------------------------"
    
    # Extraire et afficher le contenu de l'email
    email_content=$(echo "$response" | jq -r '.emailContent' 2>/dev/null)
    success=$(echo "$response" | jq -r '.success' 2>/dev/null)
    
    echo "📧 Contenu de l'email généré :"
    echo "==============================="
    echo "$email_content"
    echo ""
    echo "✅ Succès : $success"
else
    echo "❌ Test échoué avec le code $http_code"
    echo ""
    echo "🔍 Détails de l'erreur :"
    echo "$response" | jq -r '.error' 2>/dev/null || echo "Erreur non formatée"
fi
