#!/bin/bash

# Test des endpoints Boond Candidat
API_BASE_URL="http://localhost:8081"

echo "🧪 Test des endpoints Boond Candidat"
echo "=================================="

# Test 1: Suppression d'un candidat
echo ""
echo "1️⃣ Test suppression candidat"
echo "POST $API_BASE_URL/boond/candidat/delete"
curl -X POST "$API_BASE_URL/boond/candidat/delete" \
  -H "Content-Type: application/json" \
  -d @test_boond_candidate_delete.json \
  -w "\nStatus: %{http_code}\n" \
  -s

echo ""
echo "2️⃣ Test modification candidat"
echo "POST $API_BASE_URL/boond/candidat/modify"
curl -X POST "$API_BASE_URL/boond/candidat/modify" \
  -H "Content-Type: application/json" \
  -d @test_boond_candidate_modify.json \
  -w "\nStatus: %{http_code}\n" \
  -s

echo ""
echo "✅ Tests terminés"
