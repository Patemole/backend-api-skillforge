# Schéma de Réponse - Endpoint /extract

## Vue d'ensemble
L'endpoint `/extract` renvoie maintenant le champ `DC_language` dans toutes les réponses, contenant exactement la valeur reçue du frontend.

## Structure de Réponse

### Cas 1 : Sans JWT Boond (extraction simple)
```json
{
  "prenom": "Jean",
  "nom": "Dupont",
  "email": "jean.dupont@email.com",
  "phone": "+33 1 23 45 67 89",
  "summary": "Ingénieur avec 5 ans d'expérience...",
  "age": "",
  "poste": "Ingénieur Conception Mécanique",
  "diplome": "École Centrale Paris",
  "expérience": "5 ans d'expérience",
  "mobilité": "Paris",
  "disponibilité": "",
  "permis_B": "",
  "hobbies": ["Sport", "Musique"],
  "languages": [
    {
      "language": "Français",
      "level": "Natif"
    },
    {
      "language": "Anglais", 
      "level": "B2"
    }
  ],
  "secteurs_activites": ["Automobile", "Aéronautique"],
  "domaines_expertise": ["Conception Mécanique", "CAO"],
  "formations": [
    {
      "date_debut": "2018",
      "date_fin": "2020",
      "diplome": "Master Ingénierie Mécanique",
      "ecole_cursus": "École Centrale Paris"
    }
  ],
  "expériences": [
    {
      "date_debut": "01/20",
      "date_fin": "12/22",
      "entreprise": "Renault",
      "detail_entreprise": "Constructeur automobile français...",
      "durée": "2 ans",
      "poste": "Ingénieur Conception",
      "contexte": "Développement de nouveaux véhicules",
      "projet": "Conception et développement de systèmes de freinage...",
      "logiciels": ["SolidWorks", "CATIA", "ANSYS"],
      "réalisations": [
        "Conception de pièces mécaniques",
        "Tests de résistance",
        "Optimisation des performances"
      ],
      "AI_suggest": ["Expertise en simulation numérique"]
    }
  ],
  "logiciels": [
    {
      "logiciel": "SolidWorks",
      "level": "Expert",
      "temps_utilisation": "24"
    }
  ],
  "DC_language": "fr"
}
```

### Cas 2 : Avec JWT Boond (extraction + création candidat)
```json
{
  "prenom": "Jean",
  "nom": "Dupont",
  "email": "jean.dupont@email.com",
  "phone": "+33 1 23 45 67 89",
  "summary": "Ingénieur avec 5 ans d'expérience...",
  "age": "",
  "poste": "Ingénieur Conception Mécanique",
  "diplome": "École Centrale Paris",
  "expérience": "5 ans d'expérience",
  "mobilité": "Paris",
  "disponibilité": "",
  "permis_B": "",
  "hobbies": ["Sport", "Musique"],
  "languages": [
    {
      "language": "Français",
      "level": "Natif"
    },
    {
      "language": "Anglais", 
      "level": "B2"
    }
  ],
  "secteurs_activites": ["Automobile", "Aéronautique"],
  "domaines_expertise": ["Conception Mécanique", "CAO"],
  "formations": [
    {
      "date_debut": "2018",
      "date_fin": "2020",
      "diplome": "Master Ingénierie Mécanique",
      "ecole_cursus": "École Centrale Paris"
    }
  ],
  "expériences": [
    {
      "date_debut": "01/20",
      "date_fin": "12/22",
      "entreprise": "Renault",
      "detail_entreprise": "Constructeur automobile français...",
      "durée": "2 ans",
      "poste": "Ingénieur Conception",
      "contexte": "Développement de nouveaux véhicules",
      "projet": "Conception et développement de systèmes de freinage...",
      "logiciels": ["SolidWorks", "CATIA", "ANSYS"],
      "réalisations": [
        "Conception de pièces mécaniques",
        "Tests de résistance",
        "Optimisation des performances"
      ],
      "AI_suggest": ["Expertise en simulation numérique"]
    }
  ],
  "logiciels": [
    {
      "logiciel": "SolidWorks",
      "level": "Expert",
      "temps_utilisation": "24"
    }
  ],
  "boond_candidate_id": "12345",
  "DC_language": "en"
}
```

## Champ `DC_language`

### Description
- **Nom du champ** : `DC_language`
- **Type** : `string`
- **Valeurs possibles** : `"fr"` ou `"en"`
- **Valeur par défaut** : `"fr"` (si non fournie dans la requête)
- **Comportement** : Retourne exactement la valeur reçue du frontend

### Exemples de valeurs
- `"fr"` : Français (valeur par défaut)
- `"en"` : Anglais

### Utilisation
Le champ `DC_language` est maintenant inclus dans toutes les réponses de l'endpoint `/extract`, permettant au frontend de savoir quelle langue a été utilisée pour l'extraction et la génération du DC (Document de Candidature).

## Notes importantes
1. Le champ `DC_language` est ajouté à la fin de la réponse JSON
2. Il contient exactement la valeur reçue du paramètre `language` de la requête
3. Si aucun paramètre `language` n'est fourni, la valeur par défaut `"fr"` est utilisée
4. Ce champ est présent dans tous les cas (avec ou sans JWT Boond)
