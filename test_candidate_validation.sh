#!/bin/bash

# Script de test pour l'endpoint /candidate-validation
# Assurez-vous que le serveur est démarré sur le port 8081

echo "🧪 Test de l'endpoint /candidate-validation"
echo "=========================================="

# URL de l'endpoint
URL="http://localhost:8081/candidate-validation"

# Fichier de test
TEST_FILE="test_candidate_validation.json"

echo "📤 Envoi de la requête de test..."
echo "URL: $URL"
echo "Fichier: $TEST_FILE"
echo ""

# Vérifier que le fichier de test existe
if [ ! -f "$TEST_FILE" ]; then
    echo "❌ Erreur: Le fichier $TEST_FILE n'existe pas"
    exit 1
fi

# Envoyer la requête POST
echo "🚀 Exécution de la requête..."
echo ""

response=$(curl -s -w "\n%{http_code}" -X POST \
  -H "Content-Type: application/json" \
  -d @"$TEST_FILE" \
  "$URL")

# Séparer la réponse et le code de statut
http_code=$(echo "$response" | tail -n1)
response_body=$(echo "$response" | head -n -1)

echo "📊 Résultat du test:"
echo "==================="
echo "Code HTTP: $http_code"
echo ""
echo "Réponse:"
echo "$response_body" | jq . 2>/dev/null || echo "$response_body"
echo ""

# Analyser le résultat
if [ "$http_code" -eq 200 ]; then
    echo "✅ Test réussi ! L'endpoint fonctionne correctement."
    echo "📧 L'email de notification devrait être envoyé à l'inviteur."
elif [ "$http_code" -eq 400 ]; then
    echo "⚠️  Erreur de validation (400) - Vérifiez les données d'entrée"
elif [ "$http_code" -eq 500 ]; then
    echo "❌ Erreur serveur (500) - Vérifiez les logs du serveur"
    echo "💡 Assurez-vous que RESEND_API_KEY est configuré dans .env"
else
    echo "❌ Erreur inattendue (HTTP $http_code)"
fi

echo ""
echo "🔍 Pour plus de détails, consultez les logs du serveur."
