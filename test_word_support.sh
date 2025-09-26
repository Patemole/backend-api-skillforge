#!/bin/bash

echo "=== TEST DU SUPPORT COMPLET DES FICHIERS WORD ==="

# Vérifier que wvText est installé
echo "Vérification de wvText..."
if command -v wvText &> /dev/null; then
    echo "✅ wvText est installé: $(which wvText)"
else
    echo "❌ wvText n'est pas installé"
    exit 1
fi

# Vérifier que Docconv fonctionne
echo -e "\nTest de Docconv avec un fichier texte..."
echo "Test de contenu" > test_temp.txt
go run -c 'package main; import ("fmt"; "code.sajari.com/docconv"); func main() { fmt.Println("Docconv disponible") }' 2>/dev/null
if [ $? -eq 0 ]; then
    echo "✅ Docconv est disponible"
else
    echo "❌ Problème avec Docconv"
fi

# Nettoyer
rm -f test_temp.txt

echo -e "\n=== PRÊT POUR LES TESTS ==="
echo "Tu peux maintenant envoyer des vrais fichiers Word (.docx et .doc) pour tester l'extraction !"
echo ""
echo "Pour tester avec l'API :"
echo "curl -X POST -F 'file=@ton_fichier.docx' http://localhost:8080/extract"
echo "curl -X POST -F 'file=@ton_fichier.doc' http://localhost:8080/extract"
