# 📧 Guide de l'Endpoint Email de Présentation V2

## 🎯 **Vue d'ensemble**

Le nouvel endpoint `/api/email/generate-presentation-v2` génère des emails de présentation professionnels avec :
- **Reformatage du français** pour améliorer la fluidité
- **Sélection intelligente** d'expériences pertinentes selon le besoin
- **Sélection de logiciels** pertinents selon le besoin
- **Respect du template** fourni (si présent)

## 📡 **Endpoint**

```
POST http://localhost:8081/api/email/generate-presentation-v2
Content-Type: application/json
```

## 📥 **Structure de la requête**

```json
{
  "candidateData": {
    "prenom": "string (obligatoire)",
    "titre_poste": "string (obligatoire)", 
    "nombre_experience": "number (obligatoire)",
    "disponibilite": "string (obligatoire)",
    "mobilite": "string (obligatoire)",
    "diplome": "string (obligatoire)",
    "langues": "string (optionnel)",
    "certifications": "string (optionnel)",
    "hobbies": "string (optionnel)",
    "experience": "string (toutes les expériences formatées)",
    "experience_count": "number",
    "logiciel": "string (logiciel principal)",
    "logiciels": "string (liste des logiciels)", 
    "logiciels_count": "number",
    "competences_techniques": "string",
    "competences_fonctionnelles": "string",
    "projets": "string",
    "projets_count": "number"
  },
  "need": "string (optionnel - description du besoin)",
  "templateId": "string (optionnel)",
  "template": {
    "id": "string",
    "name": "string",
    "subject": "string (avec variables remplacées)",
    "content": "string (avec variables remplacées)",
    "isDefault": "boolean"
  }
}
```

## 📤 **Structure de la réponse**

```json
{
  "emailContent": "string (contenu de l'email généré)",
  "success": "boolean",
  "error": "string (optionnel - message d'erreur)"
}
```

## 🚀 **Fonctionnalités**

### **1. Reformatage du français**
- Améliore la fluidité en ajoutant les liaisons manquantes
- Corrige les tournures de phrases pour un français naturel
- Ne change pas le sens, améliore seulement la fluidité

### **2. Sélection d'expériences (si besoin présent)**
- Sélectionne 2-3 expériences les plus pertinentes
- Les expériences correspondent au besoin mentionné
- Présentation concise et pertinente

### **3. Sélection de logiciels (si besoin présent)**
- Sélectionne 3-5 logiciels les plus pertinents
- Les logiciels correspondent au besoin mentionné
- N'invente pas de logiciels, utilise seulement ceux fournis

### **4. Respect du template**
- Suit exactement le template fourni si présent
- Si pas de template, utilise une structure professionnelle standard
- Ne rajoute pas d'informations non fournies

## 📝 **Exemple de requête**

```json
{
  "candidateData": {
    "prenom": "Jean",
    "titre_poste": "Ingénieur Mécanique",
    "nombre_experience": 3,
    "disponibilite": "immédiatement",
    "mobilite": "région parisienne",
    "diplome": "Master en Génie Mécanique",
    "langues": "Français (natif), Anglais (courant)",
    "experience": "Formation 3DExperience chez Alten (3 mois) - Conception de systèmes complexes pour le secteur naval\nAssistant Chef de Projet chez Sidas (5 mois) - Développement de solutions de moulage innovantes",
    "experience_count": 3,
    "logiciels": "SolidWorks (Expert), Créo (Intermédiaire), 3DExperience (Débutant), AutoCAD (Intermédiaire)",
    "logiciels_count": 4,
    "competences_techniques": "Conception 3D, Simulation FEA, Optimisation de processus"
  },
  "need": "Recherche d'un ingénieur mécanique spécialisé en conception 3D et simulation pour des projets dans le secteur automobile. Expérience en SolidWorks et Créo requise.",
  "template": {
    "id": "template-123",
    "name": "Template Ingénieur Mécanique",
    "subject": "{{prenom}}, {{titre_poste}} – {{nombre_experience}} ans d'expériences",
    "content": "Bonjour,\n\nJe vous présente {{prenom}}, {{titre_poste}} avec {{nombre_experience}} ans d'expérience disponible {{disponibilite}} et mobile en {{mobilite}}.\n\nDernières expériences pertinentes :\n- {{experiences}}\n\nLogiciels maîtrisés : {{logiciels}}\n\nCe profil correspond-il à vos besoins ?\n\nCordialement",
    "isDefault": false
  }
}
```

## 📧 **Exemple de réponse**

```json
{
  "emailContent": "Objet: Jean, Ingénieur Mécanique – 3 ans d'expériences\n\nBonjour,\n\nJe vous présente Jean, Ingénieur Mécanique avec 3 ans d'expérience, disponible immédiatement et mobile en région parisienne.\n\nDernières expériences pertinentes :\n- Formation 3DExperience chez Alten (3 mois) - Conception de systèmes complexes pour le secteur naval\n- Assistant Chef de Projet chez Sidas (5 mois) - Développement de solutions de moulage innovantes\n\nLogiciels maîtrisés : SolidWorks (Expert), Créo (Intermédiaire), 3DExperience (Débutant)\n\nCe profil correspond-il à vos besoins ? Souhaitez-vous le rencontrer pour échanger plus techniquement avec lui ?\n\nCordialement",
  "success": true
}
```

## 🧪 **Tests**

### **Test automatisé**
```bash
./test_presentation_email_v2.sh
```

### **Test manuel**
```bash
curl -X POST http://localhost:8081/api/email/generate-presentation-v2 \
  -H "Content-Type: application/json" \
  -d @test_presentation_email_v2.json
```

## ⚠️ **Validation**

### **Champs obligatoires**
- `candidateData.prenom`
- `candidateData.titre_poste`
- `candidateData.nombre_experience` (≥ 0)
- `candidateData.disponibilite`
- `candidateData.mobilite`
- `candidateData.diplome`

### **Codes d'erreur**
- `400` : Données de requête invalides
- `500` : Erreur interne du serveur

## 🔧 **Différences avec V1**

| Fonctionnalité | V1 | V2 |
|---|---|---|
| Structure des données | Ancienne structure | Nouvelle structure optimisée |
| Reformatage français | ❌ | ✅ |
| Sélection d'expériences | ❌ | ✅ |
| Sélection de logiciels | ❌ | ✅ |
| Support template | ❌ | ✅ |
| Validation renforcée | ❌ | ✅ |

## 🎉 **Avantages**

- ✅ **Français fluide** : Amélioration automatique de la fluidité
- ✅ **Sélection intelligente** : Expériences et logiciels pertinents
- ✅ **Respect du template** : Suit exactement le template fourni
- ✅ **Validation robuste** : Contrôles stricts des données
- ✅ **Flexibilité** : Fonctionne avec ou sans template/need
