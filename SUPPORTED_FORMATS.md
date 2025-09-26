# 📄 Formats de CV Supportés

## ✅ Formats Completement Supportés

### **PDF** 
- **Extraction** : UniPDF (3 méthodes de fallback)
- **Qualité** : Excellente
- **Usage** : Standard universel

### **DOCX** 
- **Extraction** : Docconv
- **Qualité** : Excellente  
- **Usage** : Microsoft Word moderne

### **DOC**
- **Extraction** : Docconv + wvText
- **Qualité** : Excellente
- **Usage** : Microsoft Word ancien

### **PAGES**
- **Extraction** : LibreOffice → DOCX → Docconv
- **Qualité** : Excellente
- **Usage** : Apple Pages

### **RTF**
- **Extraction** : Docconv
- **Qualité** : Bonne
- **Usage** : Rich Text Format (Microsoft)

### **ODT**
- **Extraction** : Docconv
- **Qualité** : Bonne
- **Usage** : LibreOffice/OpenOffice

### **HTML**
- **Extraction** : Docconv (avec readability)
- **Qualité** : Bonne
- **Usage** : CV en page web

### **Markdown**
- **Extraction** : Traitement direct
- **Qualité** : Excellente
- **Usage** : Développeurs, GitHub

### **TXT**
- **Extraction** : Traitement direct
- **Qualité** : Excellente
- **Usage** : Texte brut

## 🔧 Détection Automatique

Le système détecte automatiquement le type de fichier basé sur :
- **Extension** : `.pdf`, `.docx`, `.doc`, `.pages`, `.rtf`, `.odt`, `.html`, `.md`, `.txt`
- **Magic Numbers** : Vérification du contenu binaire
- **Contenu** : Analyse des premiers bytes

## 📊 Statistiques

- **9 formats** supportés
- **Détection automatique** pour tous
- **Extraction robuste** avec fallbacks
- **Gestion d'erreurs** complète

## 🚀 Utilisation

```bash
# Tous ces formats sont automatiquement supportés
curl -X POST -F "file=@cv.pdf" http://localhost:8080/extract
curl -X POST -F "file=@cv.docx" http://localhost:8080/extract  
curl -X POST -F "file=@cv.pages" http://localhost:8080/extract
curl -X POST -F "file=@cv.rtf" http://localhost:8080/extract
curl -X POST -F "file=@cv.odt" http://localhost:8080/extract
curl -X POST -F "file=@cv.html" http://localhost:8080/extract
curl -X POST -F "file=@cv.md" http://localhost:8080/extract
curl -X POST -F "file=@cv.txt" http://localhost:8080/extract
```

## 💡 Recommandations

- **PDF** : Meilleur format universel
- **DOCX** : Excellent pour l'édition
- **PAGES** : Parfait pour les utilisateurs Apple
- **Markdown** : Idéal pour les développeurs
