package nuextract

import "fmt"

// GetExtractionPrompt retourne le prompt pour l'extraction et la structuration des données CV
func GetExtractionPrompt(nuextractJSON string) string {
	return `Tu es un expert RH spécialisé dans l'analyse de CV. Je souhaite que tu analyses le dictionnaire JSON de l'extraction de CV que je te fournis en input et que tu extraies TOUTES les informations pertinentes sous la forme d'un dictionnaire structuré, pouvant être enregistré en JSON, selon le modèle suivant :

NE CHANGE SURTOUT PAS LES CLÉS DE CE DICTIONNAIRE, CAR IL DOIT ÊTRE UTILISÉ AUTREMENT PAR LA SUITE.

{
  "prenom": "",
  "email": "Adresse email du candidat. Si elle n'est pas explicitement présente dans le CV, laisse ce champ vide (\"\"). Ne l'invente pas.",
  "phone": "Numéro de téléphone du candidat. Si il n'est pas explicitement présent dans le CV, laisse ce champ vide (\"\"). Ne l'invente pas.",
  "summary": "Résumé professionnel en 2-3 lignes maximum présentant le candidat, ses compétences clés et son expérience principale. Sois concis mais impactant pour donner une vision d'ensemble du profil.",
  "age": "Si l'âge n'est pas explicitement mentionné dans le CV, laisse ce champ vide (\"\"). Ne l'estime pas.",
  "poste": "TITRE DU POSTE RECHERCHÉ - Analyse les expériences passées et déduis le titre de poste le plus approprié en etant precis si il a un domaine d'activite precis. Si le candidat cherche un poste spécifique, utilise-le. Sinon, déduis du poste le plus récent ou le plus représentatif de son profil. Exemples : 'Ingénieur Conception Mécanique', 'Solution Architecte', 'Data Engineer', 'Chef de Projet', 'Développeur Full Stack'",
  "diplome": "Formation principale (nom de l'école d'ingénieur, de commerce ou du M2)",
  "expérience": "Calcule l'expérience totale en années : trouve la date de début de l'expérience la plus ancienne et soustrais de l'année actuelle (2025). Si aucune date n'est disponible, laisse vide.",
  "mobilité": "Position géographique recherchée si précisée.",
  "disponibilité": "",
  "permis_B": "",
  "hobbies": ["Liste des centres d'intérêts"],
  "languages": ["Liste des langues parlées (ex: Français, Anglais, Allemand, Espagnol, etc.)"],
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
      "date_debut": "OBLIGATOIRE - Date de début au format mois et année (ex: Février 2020, Janvier 2018, Septembre 2019)",
      "date_fin": "OBLIGATOIRE - Date de fin au format mois et année (ex: Décembre 2022, Août 2020, En cours). Si l'expérience est en cours, utilise 'En cours'",
      "entreprise": "OBLIGATOIRE - Nom de l'entreprise",
      "durée": "OBLIGATOIRE - Durée calculée automatiquement (ex: 2 ans, 6 mois, 1 an 3 mois). Si inférieur à 1 an, affiche en mois. Si supérieur ou égal à 1 an, affiche en années.",
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
   - Pour chaque expérience : date_debut, date_fin, entreprise, durée, poste ET logiciels sont OBLIGATOIRES
   - **CRITIQUE** : Relis le CV plusieurs fois pour être sûr d'avoir extrait TOUTES les expériences mentionnées

2. **NOUVEAUX CHAMPS** :
   - **"phone"** : Extrais le numéro de téléphone s'il est présent dans le CV. Format : "+33 1 23 45 67 89" ou "01.23.45.67.89" ou "0123456789". Si absent, laisse vide.
   - **"summary"** : Crée un résumé professionnel concis (2-3 lignes max) qui présente le candidat, ses compétences principales et son expérience clé. Sois impactant et professionnel.
   - **"languages"** : Extrais toutes les langues mentionnées dans le CV (section langues, expériences internationales, formations, etc.). Utilise les noms complets en français : "Français", "Anglais", "Allemand", "Espagnol", "Italien", etc. Si aucune langue n'est mentionnée, laisse un tableau vide [].

3. **CHAMP "poste"** :
   - C'est le TITRE DU POSTE RECHERCHÉ basé sur l'analyse des expériences passées
   - **MÉTHODE D'EXTRACTION** :
     a) Si le candidat indique un poste recherché spécifique → utilise-le
     b) Sinon, analyse toutes les expériences et déduis le titre le plus représentatif
     c) Privilégie le poste le plus récent ou celui qui reflète le mieux l'évolution de carrière
     d) Sois précis et professionnel dans le titre (évite les termes génériques)
   - Exemples : "Ingénieur Conception Mécanique", "Solution Architecte", "Data Engineer", "Développeur Full Stack", "Chef de Projet", "Consultant", "Ingénieur Génie Civil", "Product Manager"

4. **DATES ET CALCUL D'EXPÉRIENCE** :
   - **DATES D'EXPÉRIENCES** : Extrais TOUJOURS les dates de début et fin de chaque expérience
   - Format des dates : "Février 2020", "Décembre 2022", "Janvier 2018", etc.
   - Si l'expérience est en cours, utilise "En cours" pour date_fin
   - **CALCUL DE LA DURÉE** : Calcule automatiquement la durée entre date_debut et date_fin
     - Si durée < 1 an : affiche en mois (ex: "6 mois", "8 mois")
     - Si durée ≥ 1 an : affiche en années (ex: "2 ans", "1 an 3 mois", "3 ans")
   - **CALCUL EXPÉRIENCE TOTALE** : Pour le champ "expérience", trouve la date de début de l'expérience significative la plus ancienne et calcule : 2025 - année_de_début = années d'expérience
   - Exemple : si la première expérience commence en 2018 → "7 ans d'expérience"

5. **FORMATIONS** :
   - Remplis TOUS les champs : date_debut, date_fin, diplome, ecole_cursus
   - Sois précis sur le type de diplôme : Master, Bachelor, BTS, Diplôme d'Ingénieur, MBA, etc.

6. **LOGICIELS DANS LES EXPÉRIENCES** :
   - Extrais TOUS les logiciels/outils mentionnés dans chaque expérience
   - Déduis-les du contexte si nécessaire (ex: si "développement web" → ajoute HTML, CSS, JavaScript)

7. **COMPLETUDE** :
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
