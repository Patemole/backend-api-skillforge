package nuextract

import "fmt"

// GetExtractionPrompt retourne le prompt pour l'extraction et la structuration des données CV
func GetExtractionPrompt(nuextractJSON string) string {
	return `Tu es un expert RH spécialisé dans l'analyse de CV. Je souhaite que tu analyses le dictionnaire JSON de l'extraction de CV que je te fournis en input et que tu extraies TOUTES les informations pertinentes sous la forme d'un dictionnaire structuré, pouvant être enregistré en JSON, selon le modèle suivant :

NE CHANGE SURTOUT PAS LES CLÉS DE CE DICTIONNAIRE, CAR IL DOIT ÊTRE UTILISÉ AUTREMENT PAR LA SUITE.

{
  "prenom": "",
  "age": "Si l'âge n'est pas explicitement mentionné dans le CV, laisse ce champ vide (\"\"). Ne l'estime pas.",
  "poste": "TITRE DU POSTE RECHERCHÉ - PAS le poste actuel mais le titre du poste visé",
  "diplome": "Formation principale (nom de l'école d'ingénieur, de commerce ou du M2)",
  "expérience": "Calcule l'expérience totale en années : trouve la date de début de l'expérience la plus ancienne et soustrais de l'année actuelle (2024). Si aucune date n'est disponible, laisse vide.",
  "mobilité": "Position géographique recherchée si précisée.",
  "disponibilité": "",
  "permis_B": "",
  "hobbies": ["Liste des centres d'intérêts"],
  "formations": [
    {
      "date_debut": "OBLIGATOIRE - Année de début (ex: 2020, 2018-2019)",
      "date_fin": "OBLIGATOIRE - Année de fin (ex: 2022, 2020-2021)",
      "diplome": "OBLIGATOIRE - Type de diplôme précis (ex: Master Ingénierie Mécanique, Diplôme d'Ingénieur, Bachelor Informatique, BTS Commerce, Diplôme de Médecine, MBA, etc.)",
      "ecole_cursus": "OBLIGATOIRE - Nom complet de l'école/université (ex: École Centrale Paris, Université Pierre et Marie Curie, HEC Paris, etc.)"
    }
  ],
  "expériences": [
    {
      "entreprise": "OBLIGATOIRE - Nom de l'entreprise",
      "durée": "OBLIGATOIRE - Durée précise (ex: 2 ans, 6 mois, 2020-2022)",
      "poste": "OBLIGATOIRE - Titre du poste occupé (ex: Ingénieur Conception, Développeur Senior, Chef de Projet, etc.)",
      "contexte": "Résume l'expérience succinctement pour présenter le projet réalisé en une phrase.",
      "projet": "Ici, étoffe autant que possible les objectifs / projets de cette expérience et reformule pour rendre cela le plus long possible, sous forme de titre, sans faire apparaître le nom du candidat.",
      "logiciels": ["OBLIGATOIRE - Extrais TOUS les logiciels/outils mentionnés dans cette expérience (ex: SolidWorks, Python, React, AWS, Docker, etc.) - même s'ils ne sont pas explicitement listés, déduis-les du contexte"],
      "réalisations": [
        "Liste les missions réalisées, reformulées pour apporter un maximum de détails. Ajoute autant d'éléments que possible en les reformulant pour qu'ils soient le plus long possible."
      ],
      "AI_suggest": ["Si tu peux déduire des éléments pertinents non présents dans le CV. Les suggestions doivent être spécifiques et adaptées à chaque expérience, pertinentes pour les recruteurs, leur nombre doit varier selon les expériences, sans redondance entre elles. N'en mets pas systématiquement : cela doit paraître naturel."]
    }
  ],
  "logiciels": [
    {
      "logiciel": "",
      "level": "Estime le niveau entre : Débutant, Intermédiaire, Avancé, Expert.",
      "temps_utilisation": "Estime le temps d'utilisation en mois."
    }
  ]
}

INSTRUCTIONS CRITIQUES :

1. **EXTRACTIONS OBLIGATOIRES** :
   - Extrais TOUTES les expériences professionnelles (stages, CDI, CDD, alternances, etc.) - NE PAS EN OUBLIER UNE SEULE
   - Pour chaque formation : date_debut, date_fin, diplome ET ecole_cursus sont OBLIGATOIRES
   - Pour chaque expérience : entreprise, durée, poste ET logiciels sont OBLIGATOIRES
   - **CRITIQUE** : Relis le CV plusieurs fois pour être sûr d'avoir extrait TOUTES les expériences mentionnées

2. **CHAMP "poste"** :
   - C'est le TITRE DU POSTE RECHERCHÉ, pas le poste actuel
   - Exemples : "Ingénieur Conception Mécanique", "Solution Architecte", "Data Engineer", "Développeur Full Stack", "Chef de Projet", "Consultant"

3. **DATES ET CALCUL D'EXPÉRIENCE** :
   - Extrais TOUJOURS les dates de début et fin des expériences
   - Format : "2020-2022", "6 mois", "2 ans", etc.
   - **CALCUL EXPÉRIENCE TOTALE** : Pour le champ "expérience", trouve la date de début de l'expérience la plus ancienne et calcule : 2024 - année_de_début = années d'expérience
   - Exemple : si la première expérience commence en 2018 → "6 ans d'expérience"

4. **FORMATIONS** :
   - Remplis TOUS les champs : date_debut, date_fin, diplome, ecole_cursus
   - Sois précis sur le type de diplôme : Master, Bachelor, BTS, Diplôme d'Ingénieur, MBA, etc.

5. **LOGICIELS DANS LES EXPÉRIENCES** :
   - Extrais TOUS les logiciels/outils mentionnés dans chaque expérience
   - Déduis-les du contexte si nécessaire (ex: si "développement web" → ajoute HTML, CSS, JavaScript)

6. **COMPLETUDE** :
   - Ne laisse AUCUNE expérience de côté - même les stages courts, les missions ponctuelles, les projets
   - Ne laisse AUCUNE formation de côté
   - Analyse TOUT le contenu du CV
   - **VÉRIFICATION** : Compte le nombre d'expériences mentionnées dans le CV et assure-toi d'en avoir extrait le même nombre

NB : Ne fais pas apparaître le type de contrat (exemple : Stage, Alternance, CDI, CDD...) dans les expériences.

Ajoute autant d'informations que possible en analysant le CV et en déduisant des éléments qui ne sont pas forcément présents, comme le ferait un expert RH.

L'output doit respecter EXACTEMENT le modèle ci-dessus. Si une information n'est pas présente et que tu ne peux pas l'estimer, laisse le champ vide (chaîne vide "").

Voici le JSON d'extraction à analyser :

` + nuextractJSON + `

Réponds UNIQUEMENT avec le JSON structuré, sans texte avant ou après.
`
}

// GetEmailPrompt retourne le prompt pour générer un email de présentation de candidat
func GetEmailPrompt(candidateData, need string) string {
	return fmt.Sprintf(`
Tu dois écrire un email professionnel pour présenter un candidat à des entreprises. 

**DONNÉES DU CANDIDAT :**
%s

**BESOIN DE L'ENTREPRISE (si fourni) :**
%s

**INSTRUCTIONS :**
1. Écris un email professionnel et engageant qui met en avant les points forts du candidat
2. Si un besoin est fourni, identifie et souligne les points communs entre le profil du candidat et les exigences du poste
3. Structure l'email de manière claire avec :
   - Un objet percutant
   - Une introduction personnalisée
   - Les compétences clés du candidat
   - Les expériences pertinentes
   - Les points de match avec le besoin (si applicable)
   - Une conclusion qui incite à l'action

4. Sois précis et utilise des exemples concrets des réalisations du candidat
5. Adapte le ton selon le niveau d'expérience et le secteur
6. L'email doit être convaincant et professionnel

**FORMAT DE SORTIE :**
Commence directement par "Objet: [objet de l'email]" suivi du contenu de l'email.
`, candidateData, need)
}
