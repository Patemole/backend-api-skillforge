package nuextract

import "fmt"

// GetExtractionPrompt retourne le prompt pour l'extraction et la structuration des données CV
func GetExtractionPrompt(nuextractJSON string) string {
	return GetExtractionPromptWithLanguage(nuextractJSON, "fr")
}

// GetExtractionPromptWithLanguage retourne le prompt pour l'extraction selon la langue
func GetExtractionPromptWithLanguage(nuextractJSON string, language string) string {
	if language == "en" {
		return GetExtractionPromptEnglish(nuextractJSON)
	}
	return GetExtractionPromptFrench(nuextractJSON)
}

// GetExtractionPromptFrench retourne le prompt français pour l'extraction
func GetExtractionPromptFrench(nuextractJSON string) string {
	return `Tu es un expert RH spécialisé dans l'analyse de CV. Je souhaite que tu analyses le dictionnaire JSON de l'extraction de CV que je te fournis en input et que tu extraies TOUTES les informations pertinentes sous la forme d'un dictionnaire structuré, pouvant être enregistré en JSON, selon le modèle suivant :

NE CHANGE SURTOUT PAS LES CLÉS DE CE DICTIONNAIRE, CAR IL DOIT ÊTRE UTILISÉ AUTREMENT PAR LA SUITE.

{
  "prenom": "PRÉNOM DU CANDIDAT - OBLIGATOIRE. Extrais le prénom du candidat depuis le CV. Ce champ doit TOUJOURS être rempli sauf cas exceptionnel où le prénom n'est vraiment pas mentionné dans le CV. Cherche dans l'en-tête, la signature, ou toute mention du nom complet du candidat.",
  "nom": "Nom de famille du candidat. Si il n'est pas explicitement présent dans le CV, laisse ce champ vide (\"\"). Ne l'invente pas.",
  "email": "Adresse email du candidat. Si elle n'est pas explicitement présente dans le CV, laisse ce champ vide (\"\"). Ne l'invente pas.",
  "phone": "Numéro de téléphone du candidat. Si il n'est pas explicitement présent dans le CV, laisse ce champ vide (\"\"). Ne l'invente pas.",
  "summary": "Résumé professionnel en 2-3 lignes maximum présentant le candidat, ses compétences clés et son expérience principale. Sois concis mais impactant pour donner une vision d'ensemble du profil.",
  "age": "Si l'âge n'est pas explicitement mentionné dans le CV, laisse ce champ vide (\"\"). Ne l'estime pas.",
  "poste": "TITRE DU POSTE RECHERCHÉ - Analyse les expériences passées et déduis le titre de poste le plus approprié en etant precis si il a un domaine d'activite precis. Si le candidat cherche un poste spécifique, utilise-le. Sinon, déduis du poste le plus récent ou le plus représentatif de son profil. IMPORTANT : Supprime les mots 'alternant', 'stagiaire', 'stage', 'apprenti', 'apprentissage' du titre. Exemples : 'Ingénieur Conception Mécanique', 'Solution Architecte', 'Data Engineer', 'Chef de Projet', 'Développeur Full Stack'",
  "diplome": "Formation principale (nom de l'école d'ingénieur, de commerce ou du M2)",
  "expérience": "Calcule l'expérience totale en années : trouve la date de début de l'expérience la plus ancienne et soustrais de l'année actuelle (2025). Si aucune date n'est disponible, laisse vide.",
  "mobilité": "Position géographique recherchée si précisée.",
  "disponibilité": "",
  "permis_B": "",
  "hobbies": ["Liste des centres d'intérêts"],
  "languages": ["Liste des langues parlées (ex: Français, Anglais, Allemand, Espagnol, etc.)"],
  "secteurs_activites": ["Liste des secteurs d'activités dans lesquels le candidat a travaillé (ex: Automobile, Aéronautique, Informatique, Finance, Santé, etc.)"],
  "domaines_expertise": ["Liste des domaines de compétences d'expertise du candidat (ex: Développement Web, Data Science, Gestion de Projet, Marketing Digital, etc.)"],
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
      "date_debut": "OBLIGATOIRE - Date de début au format MM/YY (ex: 02/20, 09/19)",
      "date_fin": "OBLIGATOIRE - Date de fin au format MM/YY (ex: 12/22, 08/20). Si l'expérience est en cours, utilise 'En cours'",
      "entreprise": "OBLIGATOIRE - Nom de l'entreprise",
      "detail_entreprise": "OBLIGATOIRE - Description de l'entreprise en 1-2 phrases : secteur d'activité, taille, spécialité, position sur le marché. Exemple : 'Startup spécialisée dans l'intelligence artificielle et le machine learning, comptant 50 employés et leader dans l'analyse prédictive pour le secteur bancaire'",
      "durée": "OBLIGATOIRE - Durée calculée automatiquement (ex: 2 ans, 6 mois, 1 an 3 mois). Si inférieur à 1 an, affiche en mois. Si supérieur ou égal à 1 an, affiche en années.",
      "poste": "OBLIGATOIRE - Titre du poste occupé (ex: Ingénieur Conception, Développeur Senior, Chef de Projet, etc.) - CRITIQUE : Supprime OBLIGATOIREMENT les mots 'alternant', 'stagiaire', 'stage', 'apprenti', 'apprentissage' du titre. Exemples de transformation : 'Stagiaire Développeur' → 'Développeur', 'Alternant Ingénieur' → 'Ingénieur', 'Stagiaire de Recherche' → 'Chercheur'",
      "secteur": "Secteur d'activité de l'expérience en 1-2 mots. Extrais le secteur depuis les informations de l'entreprise, du projet ou du contexte. Exemples : Agroalimentaire, Automobile, Banque, Bâtiments, Biomédical, Chimie, Conseil, Défense, Énergie, Environnement, Ferroviaire, Grande distribution, Infrastructure, Logistique, Métallurgie / Sidérurgie, Naval, Nucléaire, Oil & Gas, Pétrochimie, Pharmaceutique, Santé, Secteur public, Télécommunications, IRVE, Photovoltaïque, Traitement des eaux, Revalorisation énergétique, Hydroélectricité, ENR (Énergies renouvelables), Énergie éolienne, Biogaz, Education, Ressources Humaines. Si le secteur n'est pas identifiable, laisse ce champ vide (\"\").",
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
   - **PRÉNOM** : Le champ "prenom" est OBLIGATOIRE et doit TOUJOURS être rempli. Cherche le prénom dans l'en-tête, la signature, ou toute mention du nom complet. Ne laisse ce champ vide que dans des cas exceptionnels où le prénom n'est vraiment pas mentionné.
   - Extrais TOUTES les expériences professionnelles (stages, CDI, CDD, alternances, etc.) - NE PAS EN OUBLIER UNE SEULE
   - Pour chaque formation : date_debut, date_fin, diplome ET ecole_cursus sont OBLIGATOIRES
   - Pour chaque expérience : date_debut, date_fin, entreprise, detail_entreprise, durée, poste ET logiciels sont OBLIGATOIRES
   - **CRITIQUE** : Relis le CV plusieurs fois pour être sûr d'avoir extrait TOUTES les expériences mentionnées

2. **NOUVEAUX CHAMPS** :
   - **"phone"** : Extrais le numéro de téléphone s'il est présent dans le CV. Format : "+33 1 23 45 67 89" ou "01.23.45.67.89" ou "0123456789". Si absent, laisse vide.
   - **"nom"** : Extrais le nom de famille du candidat s'il est présent dans le CV. Si absent, laisse vide.
   - **"summary"** : Crée un résumé professionnel concis (2-3 lignes max) qui présente le candidat, ses compétences principales et son expérience clé. Sois impactant et professionnel.
   - **"languages"** : Extrais toutes les langues mentionnées dans le CV (section langues, expériences internationales, formations, etc.). Utilise les noms complets en français : "Français", "Anglais", "Allemand", "Espagnol", "Italien", etc. Si aucune langue n'est mentionnée, laisse un tableau vide [].
   - **"secteurs_activites"** : Extrais tous les secteurs d'activités dans lesquels le candidat a travaillé (ex: "Automobile", "Aéronautique", "Informatique", "Finance", "Santé", "Énergie", "Télécommunications", etc.). Analyse les expériences professionnelles pour identifier les secteurs.
   - **"domaines_expertise"** : Extrais les domaines de compétences d'expertise du candidat (ex: "Développement Web", "Data Science", "Gestion de Projet", "Marketing Digital", "Conception Mécanique", "Intelligence Artificielle", etc.). Base-toi sur les compétences techniques et les expériences.
   - **"detail_entreprise"** : Pour chaque expérience, ajoute une description de l'entreprise en 1-2 phrases incluant : secteur d'activité, taille (startup, PME, grand groupe), spécialité, position sur le marché. Sois précis et informatif.

3. **CHAMP "poste"** :
   - C'est le TITRE DU POSTE RECHERCHÉ basé sur l'analyse des expériences passées
   - **MÉTHODE D'EXTRACTION** :
     a) Si le candidat indique un poste recherché spécifique → utilise-le
     b) Sinon, analyse toutes les expériences et déduis le titre le plus représentatif
     c) Privilégie le poste le plus récent ou celui qui reflète le mieux l'évolution de carrière
     d) Sois précis et professionnel dans le titre (évite les termes génériques)
   - **IMPORTANT** : Supprime TOUJOURS les mots "alternant", "stagiaire", "stage", "apprenti", "apprentissage" du titre du poste
   - Exemples : "Ingénieur Conception Mécanique", "Solution Architecte", "Data Engineer", "Développeur Full Stack", "Chef de Projet", "Consultant", "Ingénieur Génie Civil", "Product Manager"

4. **DATES ET CALCUL D'EXPÉRIENCE** :
  - **DATES D'EXPÉRIENCES** : Extrais TOUJOURS les dates de début et fin de chaque expérience
  - Format des dates : MM/YY strictement (ex: 02/20, 12/22). AUCUN autre format n'est accepté
  - Si l'expérience est en cours, utilise "En cours" pour date_fin (pas de MM/YY)
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

**RÈGLE CRITIQUE POUR LES TERMES DE STAGE/ALTERNANCE** :
- Supprime TOUJOURS les mots "alternant", "stagiaire", "stage", "apprenti", "apprentissage" des titres de postes (champ "poste" principal et champ "poste" des expériences)
- Conserve TOUTES les expériences (stages, alternances, etc.) mais traite-les comme des expériences professionnelles normales
- Reformule les titres pour qu'ils soient professionnels sans mentionner le statut
- **EXEMPLES DE TRANSFORMATION OBLIGATOIRES** :
  * "Stagiaire Développeur" → "Développeur"
  * "Alternant Ingénieur" → "Ingénieur" 
  * "Stagiaire de Recherche" → "Chercheur"
  * "Apprenti Data Analyst" → "Data Analyst"
  * "Stage Marketing" → "Marketing"
  * "Alternance Commercial" → "Commercial"

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

// GetExtractionPromptEnglish retourne le prompt anglais pour l'extraction
func GetExtractionPromptEnglish(nuextractJSON string) string {
	return `You are an HR expert specialized in CV analysis. I want you to analyze the CV extraction JSON dictionary that I provide as input and extract ALL relevant information in the form of a structured dictionary, which can be saved as JSON, according to the following model:

DO NOT CHANGE THE KEYS OF THIS DICTIONARY, AS IT WILL BE USED OTHERWISE LATER.

{
  "prenom": "CANDIDATE'S FIRST NAME - MANDATORY. Extract the candidate's first name from the CV. This field must ALWAYS be filled except in exceptional cases where the first name is really not mentioned in the CV. Look in the header, signature, or any mention of the candidate's full name.",
  "nom": "Candidate's last name. If it is not explicitly present in the CV, leave this field empty (\"\"). Do not invent it.",
  "email": "Candidate's email address. If it is not explicitly present in the CV, leave this field empty (\"\"). Do not invent it.",
  "phone": "Candidate's phone number. If it is not explicitly present in the CV, leave this field empty (\"\"). Do not invent it.",
  "summary": "Professional summary in 2-3 lines maximum presenting the candidate, their key skills and main experience. Be concise but impactful to give an overview of the profile.",
  "age": "If age is not explicitly mentioned in the CV, leave this field empty (\"\"). Do not estimate it.",
  "poste": "JOB TITLE SOUGHT - Analyze past experiences and deduce the most appropriate job title, being precise if they have a specific field of activity. If the candidate is looking for a specific position, use it. Otherwise, deduce from the most recent or most representative position of their profile. IMPORTANT: Remove the words 'intern', 'trainee', 'internship', 'apprentice', 'apprenticeship' from the title. Examples: 'Mechanical Design Engineer', 'Solution Architect', 'Data Engineer', 'Project Manager', 'Full Stack Developer'",
  "diplome": "Main education (name of engineering school, business school or M2)",
  "expérience": "Calculate total experience in years: find the start date of the oldest experience and subtract from the current year (2025). If no date is available, leave empty.",
  "mobilité": "Geographic location sought if specified.",
  "disponibilité": "",
  "permis_B": "",
  "hobbies": ["List of interests"],
  "languages": ["List of spoken languages (e.g.: French, English, German, Spanish, etc.)"],
  "secteurs_activites": ["List of business sectors in which the candidate has worked (e.g.: Automotive, Aerospace, IT, Finance, Health, etc.)"],
  "domaines_expertise": ["List of candidate's expertise domains (e.g.: Web Development, Data Science, Project Management, Digital Marketing, etc.)"],
  "formations": [
    {
      "date_debut": "MANDATORY - Start year (e.g.: 2020, 2018-2019)",
      "date_fin": "MANDATORY - End year (e.g.: 2022, 2020-2021)",
      "diplome": "MANDATORY - Precise type of degree (e.g.: Master in Mechanical Engineering, Engineering Degree, Bachelor in Computer Science, BTS Commerce, Medical Degree, MBA, etc.)",
      "ecole_cursus": "MANDATORY - Full name of school/university (e.g.: École Centrale Paris, Pierre and Marie Curie University, HEC Paris, etc.)"
    }
  ],
  "expériences": [
    {
      "date_debut": "MANDATORY - Start date in month and year format (e.g.: February 2020, January 2018, September 2019)",
      "date_fin": "MANDATORY - End date in month and year format (e.g.: December 2022, August 2020, In progress). If the experience is ongoing, use 'In progress'",
      "entreprise": "MANDATORY - Company name",
      "detail_entreprise": "MANDATORY - Company description in 1-2 sentences: business sector, size, specialty, market position. Example: 'Startup specialized in artificial intelligence and machine learning, with 50 employees and leader in predictive analysis for the banking sector'",
      "durée": "MANDATORY - Automatically calculated duration (e.g.: 2 years, 6 months, 1 year 3 months). If less than 1 year, display in months. If greater than or equal to 1 year, display in years.",
      "poste": "MANDATORY - Job title held (e.g.: Design Engineer, Senior Developer, Project Manager, etc.) - CRITICAL: MANDATORY remove the words 'intern', 'trainee', 'internship', 'apprentice', 'apprenticeship' from the title. Examples of transformation: 'Intern Developer' → 'Developer', 'Apprentice Engineer' → 'Engineer', 'Research Intern' → 'Researcher'",
      "secteur": "Business sector of the experience in 1-2 words. Extract the sector from the company information, project, or context. Examples: Food industry, Automotive, Banking, Construction, Biomedical, Chemistry, Consulting, Defense, Energy, Environment, Railway, Retail, Infrastructure, Logistics, Metallurgy / Steel industry, Naval, Nuclear, Oil & Gas, Petrochemical, Pharmaceutical, Healthcare, Public sector, Telecommunications, IRVE, Photovoltaics, Water treatment, Energy recovery, Hydroelectricity, Renewable Energy (ENR), Wind energy, Biogas, Education, Human Resources. If the sector is not identifiable, leave this field empty (\"\").",
      "contexte": "Summarize the experience succinctly to present the project carried out in one sentence.",
      "projet": "Here, expand as much as possible the objectives/projects of this experience and reformulate to make it as long as possible, in the form of a title, without showing the candidate's name.",
      "logiciels": ["MANDATORY - Extract ALL software/tools mentioned in this experience (e.g.: SolidWorks, Python, React, AWS, Docker, etc.) - even if they are not explicitly listed, deduce them from the context"],
      "réalisations": [
        "List the missions carried out, reformulated to provide maximum detail. Add as many elements as possible by reformulating them to be as long as possible."
      ],
      "AI_suggest": ["If you can deduce relevant elements not present in the CV. The suggestions must be specific and adapted to each experience, relevant for recruiters, their number must vary according to experiences, without redundancy between them. Don't put them systematically: it must seem natural."]
    }
  ],
  "logiciels": [
    {
      "logiciel": "",
      "level": "Estimate the level between: Beginner, Intermediate, Advanced, Expert.",
      "temps_utilisation": "Estimate the usage time in months."
    }
  ]
}

CRITICAL INSTRUCTIONS:

1. **MANDATORY EXTRACTIONS**:
   - **FIRST NAME**: The "prenom" field is MANDATORY and must ALWAYS be filled. Look for the first name in the header, signature, or any mention of the full name. Do not leave this field empty except in exceptional cases where the first name is really not mentioned.
   - Extract ALL professional experiences (internships, permanent contracts, fixed-term contracts, apprenticeships, etc.) - DO NOT MISS A SINGLE ONE
   - For each education: date_debut, date_fin, diplome AND ecole_cursus are MANDATORY
   - For each experience: date_debut, date_fin, entreprise, detail_entreprise, durée, poste AND logiciels are MANDATORY
   - **CRITICAL**: Reread the CV several times to make sure you have extracted ALL mentioned experiences

2. **NEW FIELDS**:
   - **"phone"**: Extract the phone number if present in the CV. Format: "+33 1 23 45 67 89" or "01.23.45.67.89" or "0123456789". If absent, leave empty.
   - **"nom"**: Extract the candidate's last name if present in the CV. If absent, leave empty.
   - **"summary"**: Create a concise professional summary (2-3 lines max) that presents the candidate, their main skills and key experience. Be impactful and professional.
   - **"languages"**: Extract all languages mentioned in the CV (language section, international experiences, education, etc.). Use full names in English: "French", "English", "German", "Spanish", "Italian", etc. If no language is mentioned, leave an empty array [].
   - **"secteurs_activites"**: Extract all business sectors in which the candidate has worked (e.g.: "Automotive", "Aerospace", "IT", "Finance", "Health", "Energy", "Telecommunications", etc.). Analyze professional experiences to identify sectors.
   - **"domaines_expertise"**: Extract the candidate's expertise domains (e.g.: "Web Development", "Data Science", "Project Management", "Digital Marketing", "Mechanical Design", "Artificial Intelligence", etc.). Base yourself on technical skills and experiences.
   - **"detail_entreprise"**: For each experience, add a company description in 1-2 sentences including: business sector, size (startup, SME, large group), specialty, market position. Be precise and informative.

3. **"poste" FIELD**:
   - This is the JOB TITLE SOUGHT based on the analysis of past experiences
   - **EXTRACTION METHOD**:
     a) If the candidate indicates a specific sought position → use it
     b) Otherwise, analyze all experiences and deduce the most representative title
     c) Prioritize the most recent position or the one that best reflects career evolution
     d) Be precise and professional in the title (avoid generic terms)
   - **IMPORTANT**: ALWAYS remove the words "intern", "trainee", "internship", "apprentice", "apprenticeship" from the job title
   - Examples: "Mechanical Design Engineer", "Solution Architect", "Data Engineer", "Full Stack Developer", "Project Manager", "Consultant", "Civil Engineer", "Product Manager"

4. **DATES AND EXPERIENCE CALCULATION**:
   - **EXPERIENCE DATES**: ALWAYS extract start and end dates of each experience
   - Date format: "February 2020", "December 2022", "January 2018", etc.
   - If the experience is ongoing, use "In progress" for date_fin
   - **DURATION CALCULATION**: Automatically calculate the duration between date_debut and date_fin
     - If duration < 1 year: display in months (e.g.: "6 months", "8 months")
     - If duration ≥ 1 year: display in years (e.g.: "2 years", "1 year 3 months", "3 years")
   - **TOTAL EXPERIENCE CALCULATION**: For the "expérience" field, find the start date of the oldest significant experience and calculate: 2025 - start_year = years of experience
     - Example: if the first experience starts in 2018 → "7 years of experience"

5. **EDUCATION**:
   - Fill ALL fields: date_debut, date_fin, diplome, ecole_cursus
   - Be precise on the degree type: Master, Bachelor, BTS, Engineering Degree, MBA, etc.

6. **SOFTWARE IN EXPERIENCES**:
   - Extract ALL software/tools mentioned in each experience
   - Deduce them from context if necessary (e.g.: if "web development" → add HTML, CSS, JavaScript)

7. **COMPLETENESS**:
   - Do not leave ANY experience aside - even short internships, occasional missions, projects
   - Do not leave ANY education aside
   - Analyze ALL CV content
   - **VERIFICATION**: Count the number of experiences mentioned in the CV and make sure you have extracted the same number

NB: Do not show the contract type (e.g.: Internship, Apprenticeship, Permanent Contract, Fixed-term Contract...) in experiences.

**CRITICAL RULE FOR INTERNSHIP/APPRENTICESHIP TERMS**:
- ALWAYS remove the words "intern", "trainee", "internship", "apprentice", "apprenticeship" from job titles (main "poste" field and "poste" field in experiences)
- Keep ALL experiences (internships, apprenticeships, etc.) but treat them as normal professional experiences
- Reformulate titles to be professional without mentioning the status
- **MANDATORY TRANSFORMATION EXAMPLES**:
  * "Intern Developer" → "Developer"
  * "Apprentice Engineer" → "Engineer" 
  * "Research Intern" → "Researcher"
  * "Apprentice Data Analyst" → "Data Analyst"
  * "Marketing Intern" → "Marketing"
  * "Sales Apprentice" → "Sales"

Add as much information as possible by analyzing the CV and deducing elements that are not necessarily present, as an HR expert would do.

The output must respect EXACTLY the model above. If information is not present and you cannot estimate it, leave the field empty (empty string "").

**IMPORTANT**: All extracted content must be in English. Translate any French terms, company names, job titles, and descriptions to English while maintaining accuracy and professional terminology.

Here is the extraction JSON to analyze:

` + nuextractJSON + `

Respond ONLY with the structured JSON, without text before or after.
`
}
