#!/bin/bash

# Test script pour l'endpoint de génération de templates amélioré
echo "🧪 Test de l'endpoint POST /api/templates/generate (Version Améliorée)"
echo "====================================================================="

# Vérifier que le serveur est en cours d'exécution
echo "📡 Vérification du serveur..."
if ! curl -s http://localhost:8081/health > /dev/null; then
    echo "❌ Le serveur n'est pas en cours d'exécution sur le port 8081"
    echo "💡 Lancez d'abord le serveur avec: go run cmd/server/main.go"
    exit 1
fi

echo "✅ Serveur accessible"

# Test de l'endpoint avec les nouvelles variables granulaires
echo ""
echo "🚀 Test de génération de template avec variables granulaires..."
echo "---------------------------------------------------------------"

response=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -d @test_template_generate_improved.json \
  http://localhost:8081/api/templates/generate)

echo "📤 Réponse reçue:"
echo "$response" | jq '.' 2>/dev/null || echo "$response"

# Vérifier le statut de la réponse
http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -d @test_template_generate_improved.json \
  http://localhost:8081/api/templates/generate)

echo ""
echo "📊 Code de statut HTTP: $http_code"

if [ "$http_code" -eq 200 ]; then
    echo "✅ Test réussi !"
    echo ""
    echo "🔍 Analyse de la réponse :"
    echo "-------------------------"
    
    # Extraire et afficher le template généré
    template_subject=$(echo "$response" | jq -r '.generated_template.subject' 2>/dev/null)
    template_content=$(echo "$response" | jq -r '.generated_template.content' 2>/dev/null)
    detected_vars=$(echo "$response" | jq -r '.detected_variables[]' 2>/dev/null | tr '\n' ' ')
    confidence=$(echo "$response" | jq -r '.confidence_score' 2>/dev/null)
    
    echo "📝 Sujet du template :"
    echo "$template_subject"
    echo ""
    echo "📄 Contenu du template :"
    echo "$template_content"
    echo ""
    echo "🏷️  Variables détectées : $detected_vars"
    echo "📊 Score de confiance : $confidence"
else
    echo "❌ Test échoué avec le code $http_code"
fi
