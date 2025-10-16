#!/bin/bash

# Test script pour vérifier la fonctionnalité de sélection de langue
# Ce script teste l'endpoint /extract avec les paramètres language="fr" et language="en"

echo "🧪 Test de la fonctionnalité de sélection de langue"
echo "=================================================="

# URL du serveur (ajustez selon votre configuration)
SERVER_URL="http://localhost:8080"

# Fichier de test (utilisez un CV existant)
TEST_FILE="test_cv_poste.txt"

echo "📁 Fichier de test: $TEST_FILE"

# Test 1: Français (par défaut)
echo ""
echo "🇫🇷 Test avec language=fr (français)"
echo "------------------------------------"
curl -X POST \
  -F "file=@$TEST_FILE" \
  -F "language=fr" \
  "$SERVER_URL/extract" \
  -H "Content-Type: multipart/form-data" \
  -w "\n\nStatus: %{http_code}\nTime: %{time_total}s\n" \
  -s | head -20

echo ""
echo "🇬🇧 Test avec language=en (anglais)"
echo "------------------------------------"
curl -X POST \
  -F "file=@$TEST_FILE" \
  -F "language=en" \
  "$SERVER_URL/extract" \
  -H "Content-Type: multipart/form-data" \
  -w "\n\nStatus: %{http_code}\nTime: %{time_total}s\n" \
  -s | head -20

echo ""
echo "🔍 Test sans paramètre language (doit utiliser fr par défaut)"
echo "------------------------------------------------------------"
curl -X POST \
  -F "file=@$TEST_FILE" \
  "$SERVER_URL/extract" \
  -H "Content-Type: multipart/form-data" \
  -w "\n\nStatus: %{http_code}\nTime: %{time_total}s\n" \
  -s | head -20

echo ""
echo "✅ Tests terminés"
