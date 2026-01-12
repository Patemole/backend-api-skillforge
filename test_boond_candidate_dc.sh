#!/bin/bash

# Test de l'endpoint /boond/candidat/dc
API_BASE_URL="http://localhost:8081"

echo "🧪 Test endpoint /boond/candidat/dc"
echo "=================================="

# Créer un fichier de test
echo "Contenu du dossier de compétences" > test_dc.txt

echo ""
echo "📄 Test upload dossier de compétences"
echo "POST $API_BASE_URL/boond/candidat/dc"
curl -X POST "$API_BASE_URL/boond/candidat/dc" \
  -F "file=@test_dc.txt" \
  -F "boondCandidateId=5471" \
  -F "boondJwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyVG9rZ..." \
  -F "filename=test_dc.txt" \
  -w "\nStatus: %{http_code}\n" \
  -s

# Nettoyer
rm -f test_dc.txt

echo ""
echo "✅ Test terminé"
