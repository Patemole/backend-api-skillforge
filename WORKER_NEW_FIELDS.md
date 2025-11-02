# 📋 Nouveaux champs dans le payload pour le worker Word

## 🎯 Vue d'ensemble

Trois nouveaux champs ont été ajoutés au payload `competence_dossier` envoyé au worker pour la génération de fichiers Word. Le worker doit les prendre en compte dans son code.

---

## 📦 Structure du payload

Le worker reçoit un payload JSON dans la table `jobs` avec cette structure :

```json
{
  "competence_dossier": {
    // ... autres champs existants ...
    "competence_fonctionnelle": [...],  // ⭐ NOUVEAU
    "expériences": [
      {
        // ... autres champs existants ...
        "projets_name": [...],  // ⭐ NOUVEAU
        "result": "...",        // ⭐ NOUVEAU
      }
    ],
    "logiciels": [
      {
        "logiciel": "...",
        "level": "...",
        "temps_utilisation": "..."  // ⚠️ Déjà existant mais à vérifier
      }
    ]
  },
  "template_url": "...",
  "organization_name": "...",
  "dossier_id": "..."
}
```

---

## 🆕 Nouveaux champs détaillés

### 1. `competence_fonctionnelle` (Niveau racine de `competence_dossier`)

**Type** : `array[string]`

**Description** : Liste des compétences fonctionnelles (soft skills, compétences comportementales) du candidat.

**Exemple** :
```json
{
  "competence_dossier": {
    "competence_fonctionnelle": [
      "Gestion de projet",
      "Leadership",
      "Communication interculturelle",
      "Résolution de problèmes",
      "Travail en équipe"
    ]
  }
}
```

**Utilisation** : Ces compétences peuvent être affichées dans une section dédiée du CV Word, similaire à la section `technical_skills`.

---

### 2. `projets_name` (Dans chaque objet de `expériences`)

**Type** : `array[string]`

**Description** : Liste des noms de projets spécifiques mentionnés pour cette expérience professionnelle.

**Exemple** :
```json
{
  "expériences": [
    {
      "entreprise": "Acme Corp",
      "poste": "Développeur Senior",
      "projet": "Développement d'une plateforme de gestion...",
      "projets_name": [
        "Projet Alpha",
        "Projet Beta - Refonte système",
        "Bureau des Italiens"
      ],
      "result": "...",
      // ... autres champs
    }
  ]
}
```

**Cas particuliers** :
- Peut être un tableau vide `[]` si aucun nom de projet n'est mentionné
- Peut contenir un seul élément ou plusieurs
- Les noms de projets sont souvent extraits du contexte ou des réalisations

**Utilisation** : À afficher dans la section expérience, par exemple :
- Soit dans une sous-liste après le projet
- Soit intégré dans le texte du projet
- Format suggéré : "Projets : [Projet Alpha, Projet Beta]"

---

### 3. `result` (Dans chaque objet de `expériences`)

**Type** : `string`

**Description** : Résultat/réalisation principale de l'expérience professionnelle, formulé de manière concise.

**Exemple** :
```json
{
  "expériences": [
    {
      "entreprise": "Tech Solutions",
      "poste": "Chef de projet",
      "projet": "Mise en place d'un système de monitoring...",
      "result": "Augmentation de 30% des performances et réduction des temps de traitement de 40%",
      // ... autres champs
    }
  ]
}
```

**Cas particuliers** :
- Peut être une chaîne vide `""` si aucun résultat n'est mentionné
- C'est un résumé concis, différent du champ `réalisations` qui est un tableau détaillé
- Différent aussi du champ `projet` qui décrit le contexte et l'histoire du projet

**Utilisation** : À afficher dans la section expérience :
- Soit comme une phrase après le projet
- Soit dans une section "Résultats" séparée
- Format suggéré : "Résultat : [texte du result]"

---

### 4. `temps_utilisation` (Dans chaque objet de `logiciels`) - Vérification nécessaire

**Type** : `string`

**Description** : Durée d'utilisation du logiciel (en mois généralement).

**Exemple** :
```json
{
  "logiciels": [
    {
      "logiciel": "Docker",
      "level": "Avancé",
      "temps_utilisation": "24"  // 24 mois
    },
    {
      "logiciel": "Kubernetes",
      "level": "Intermédiaire",
      "temps_utilisation": "12"
    }
  ]
}
```

**Cas particuliers** :
- Peut être une chaîne vide `""` si la durée n'est pas disponible
- Format généralement numérique (nombre de mois)
- Ce champ existait déjà dans le modèle mais il faut vérifier qu'il est bien utilisé par le worker

**Utilisation** : À afficher avec les logiciels :
- Format suggéré : "Docker (Avancé - 24 mois)"
- Ou dans un tableau avec colonnes : Logiciel | Niveau | Durée

---

## 🔍 Structure complète d'une expérience

Voici un exemple complet d'un objet `expérience` avec tous les champs (anciens + nouveaux) :

```json
{
  "date_debut": "01/20",
  "date_fin": "12/22",
  "entreprise": "Acme Corporation",
  "detail_entreprise": "Startup spécialisée en IA et Machine Learning, 50 employés",
  "durée": "2 ans",
  "poste": "Développeur Full-Stack",
  "contexte": "Équipe de 5 développeurs travaillant sur une plateforme SaaS",
  "projet": "Développement d'une plateforme de gestion de projets avec intégration IA pour l'automatisation des tâches. Le projet incluait la refonte complète de l'architecture backend et l'implémentation d'une nouvelle interface utilisateur.",
  "projets_name": [
    "Plateforme ProjetManager",
    "Refonte Backend 2022"
  ],
  "result": "Lancement réussi de la plateforme avec 500+ utilisateurs actifs dès le premier mois et amélioration de 40% du temps de traitement des tâches.",
  "logiciels": [
    "React",
    "Node.js",
    "PostgreSQL",
    "Docker"
  ],
  "réalisations": [
    "Développement de l'API REST complète",
    "Implémentation de l'authentification JWT",
    "Création de composants React réutilisables",
    "Configuration CI/CD avec GitHub Actions"
  ],
  "AI_suggest": []
}
```

---

## 📝 Modifications à faire dans le worker

### 1. Parser les nouveaux champs

Le worker doit vérifier que lors du parsing du JSON `competence_dossier`, il lit bien :

**Niveau racine** :
- `competence_fonctionnelle` → tableau de strings (peut être vide)

**Dans chaque expérience** :
- `projets_name` → tableau de strings (peut être vide)
- `result` → string (peut être vide)

**Dans chaque logiciel** :
- `temps_utilisation` → string (peut être vide, vérifier que c'est bien lu)

### 2. Intégration dans le template Word

Selon le template utilisé, le worker doit :

1. **Pour `competence_fonctionnelle`** :
   - Créer une section similaire à `technical_skills`
   - Ou les ajouter dans une section "Compétences" globale
   - Format de liste à puces ou virgules

2. **Pour `projets_name` (dans expériences)** :
   - Les afficher dans la section expérience
   - Format suggéré : "Projets : [liste des projets]"
   - Ou les intégrer dans le texte du projet

3. **Pour `result` (dans expériences)** :
   - Afficher après le projet ou dans une sous-section
   - Format suggéré : "Résultat : [texte]"
   - Mise en évidence possible (gras, italique)

4. **Pour `temps_utilisation` (dans logiciels)** :
   - Vérifier que ce champ est bien affiché avec les logiciels
   - Format : "[Logiciel] ([Niveau] - [Durée] mois)"
   - Ou dans un tableau avec colonne dédiée

### 3. Gestion des valeurs vides

Tous ces nouveaux champs peuvent être vides :
- `competence_fonctionnelle` → `[]` (tableau vide)
- `projets_name` → `[]` (tableau vide)
- `result` → `""` (chaîne vide)
- `temps_utilisation` → `""` (chaîne vide)

Le worker doit gérer ces cas et ne rien afficher ou afficher une version par défaut.

---

## 🧪 Exemple de payload complet

```json
{
  "competence_dossier": {
    "prenom": "Jean",
    "nom": "Dupont",
    "poste": "Développeur Senior",
    "technical_skills": [
      "Python",
      "TypeScript",
      "React"
    ],
    "competence_fonctionnelle": [
      "Gestion de projet",
      "Leadership",
      "Communication"
    ],
    "expériences": [
      {
        "entreprise": "Tech Corp",
        "poste": "Développeur",
        "projet": "Développement d'une application web...",
        "projets_name": [
          "Projet Alpha",
          "Projet Beta"
        ],
        "result": "Augmentation de 30% des performances",
        "logiciels": ["React", "Node.js"]
      }
    ],
    "logiciels": [
      {
        "logiciel": "Docker",
        "level": "Avancé",
        "temps_utilisation": "24"
      },
      {
        "logiciel": "Kubernetes",
        "level": "Intermédiaire",
        "temps_utilisation": "12"
      }
    ]
  },
  "template_url": "https://...",
  "organization_name": "Ma Société",
  "dossier_id": "123e4567-e89b-12d3-a456-426614174000"
}
```

---

## ✅ Checklist pour le développeur du worker

- [ ] Vérifier que le parsing JSON lit `competence_fonctionnelle` dans `competence_dossier`
- [ ] Vérifier que le parsing JSON lit `projets_name` dans chaque expérience
- [ ] Vérifier que le parsing JSON lit `result` dans chaque expérience
- [ ] Vérifier que le parsing JSON lit `temps_utilisation` dans chaque logiciel
- [ ] Gérer les cas où ces champs sont vides
- [ ] Intégrer `competence_fonctionnelle` dans le template Word
- [ ] Intégrer `projets_name` dans les sections expériences
- [ ] Intégrer `result` dans les sections expériences
- [ ] Vérifier l'affichage de `temps_utilisation` avec les logiciels
- [ ] Tester avec des payloads réels (certains champs vides, certains remplis)

---

## 📞 Contact

Si vous avez des questions sur la structure des données ou besoin de clarification, contactez l'équipe backend.

