package nuextract

// GetExtractionPromptProduction retourne le prompt de production pour l'extraction et la structuration des données CV
func GetExtractionPromptProduction(nuextractJSON string) string {
	return GetExtractionPromptProductionWithLanguage(nuextractJSON, "fr")
}

// GetExtractionPromptProductionWithLanguage retourne le prompt de production pour l'extraction selon la langue
func GetExtractionPromptProductionWithLanguage(nuextractJSON string, language string) string {
	if language == "en" {
		return GetExtractionPromptProductionEnglish(nuextractJSON)
	}
	return GetExtractionPromptProductionFrench(nuextractJSON)
}

// GetExtractionPromptProductionFrench retourne le prompt français de production pour l'extraction
func GetExtractionPromptProductionFrench(nuextractJSON string) string {
	return `Tu es un expert RH spécialisé dans l'analyse de CV. Je souhaite que tu analyses le dictionnaire JSON de l'extraction de CV que je te fournis en input et que tu extraies TOUTES les informations pertinentes sous la forme d'un dictionnaire structuré, pouvant être enregistré en JSON, selon le modèle suivant :
  
  NE CHANGE SURTOUT PAS LES CLÉS DE CE DICTIONNAIRE, CAR IL DOIT ÊTRE UTILISÉ AUTREMENT PAR LA SUITE.
  
  {
	"prenom": "PRÉNOM DU CANDIDAT - OBLIGATOIRE. Extrais le prénom du candidat depuis le CV. Ce champ doit TOUJOURS être rempli sauf cas exceptionnel où le prénom n'est vraiment pas mentionné dans le CV. Cherche dans l'en-tête, la signature, ou toute mention du nom complet du candidat.",
	"nom": "Nom de famille du candidat. Si il n'est pas explicitement présent dans le CV, laisse ce champ vide (\"\"). Ne l'invente pas.",
	"email": "Adresse email du candidat. Si elle n'est pas explicitement présente dans le CV, laisse ce champ vide (\"\"). Ne l'invente pas.",
	"phone": "Numéro de téléphone du candidat. Si il n'est pas explicitement présent dans le CV, laisse ce champ vide (\"\"). Ne l'invente pas.",
    "summary": "Résumé professionnel en 2-3 lignes maximum présentant le candidat, ses compétences clés et son expérience principale. Sois concis mais impactant pour donner une vision d'ensemble du profil. STYLE : Utilise un style professionnel comme 'Plus de X années d'expérience dans [domaine], incluant [activités spécifiques] pour [industries]. Expérience en [environnements] avec un historique prouvé de collaboration avec [types d'entreprises].'",
	"age": "Si l'âge n'est pas explicitement mentionné dans le CV, laisse ce champ vide (\"\"). Ne l'estime pas.",
	"poste": "TITRE DU POSTE RECHERCHÉ - Analyse les expériences passées et déduis le titre de poste le plus approprié en etant precis si il a un domaine d'activite precis. Si le candidat cherche un poste spécifique, utilise-le. Sinon, déduis du poste le plus récent ou le plus représentatif de son profil. Exemples : 'Ingénieur Conception Mécanique', 'Solution Architecte', 'Data Engineer', 'Chef de Projet', 'Développeur Full Stack'",
	"diplome": "Formation principale (nom de l'école d'ingénieur, de commerce ou du M2)",
	"expérience": "Calcule l'expérience totale en années : trouve la date de début de l'expérience la plus ancienne et soustrais de l'année actuelle (2025). Si aucune date n'est disponible, laisse vide.",
	"mobilité": "Position géographique recherchée si précisée.",
	"disponibilité": "",
	"permis_B": "",
    "hobbies": ["Liste des centres d'intérêts"],
    "languages": [
      {
        "language": "Nom de la langue (ex: Français, Anglais, Allemand, Espagnol, etc.)",
        "level": "Niveau CECR (ex: A1, A2, B1, B2, C1, C2, Natif, Courant, Intermédiaire, Débutant, etc.)"
      }
    ],
    "certifications": ["Liste des certifications mentionnées dans le CV (ex: BOSIET, NEBOSH, IOSH, OSHA, API, ASME, ISO 9001, Six Sigma, PMP, Initial training on HV-LV Electrical Installations, Electrical accreditation certificate according to the C18-510 standard, National first-aid diploma, BCTE Certificate, Qess Certification, ATEX Certificate, etc.). Ne pas inventer - extraire uniquement celles explicitement mentionnées."],
    "technical_skills": ["Liste des compétences techniques détaillées et précises mentionnées dans le CV. Extraire les compétences techniques complètes, pas seulement des mots-clés. Exemples : 'Construction, QA/QC, Pre-Commissioning, Commissioning, hook-up, Start-up, Maintenance', 'Smart Relays & Soft Starters', 'Instrument calibrations and loop checks', 'Knowledge of PCS, PDCS & DCS system operation', 'Primary and secondary earthing system, support & cable tray installation, instrument and junction boxes installation, cable pulling dressing & termination, Circuits Control', 'HV/MV/LV, firefighting, CCTV & security and access control systems', 'Check list A&B, static testing (Pressure testing, resistance checks), calibration, loop check, cause and effects (loop function check and multilevel trip sequence)', 'Validations, documents reviews, and investigations (Documents : URS, FAT, FDS, CSV, IQ, OQ, and PQ)', 'Good experience in oil and gas industry from wellhead installations to storage and transfer', 'Ability to apply safety rules – hazardous areas - explosive atmosphere (LEL), Indices of protection - protection methods using on hazardous area', 'Ability to understand and use work permits system to control/coordinate and communicate all tasks', 'Ability to do risk assessments before starting work', etc."],
     "secteurs_activites": ["Liste des secteurs d'activités dans lesquels le candidat a travaillé (ex: Pétrole et Gaz, Pétrochimie, Énergie, Automobile, Aéronautique, Informatique, Finance, Santé, etc.)"],
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
		"poste": "OBLIGATOIRE - Titre du poste occupé (ex: Ingénieur Conception, Développeur Senior, Chef de Projet, etc.)",
		"contexte": "Résume l'expérience succinctement pour présenter le projet réalisé en une phrase.",
		"projet": "ÉTOFFE ce champ en créant une description fluide et connectée du projet/expérience, comme dans l'exemple concurrent. Utilise TOUTES les informations disponibles : entreprise, secteur, missions, réalisations, équipements, technologies, clients, projets mentionnés. Crée des phrases liées qui racontent une histoire cohérente, pas des bullet points séparés. IMPORTANT : Utilise des mots et formulations DIFFÉRENTS de ceux utilisés dans les réalisations - évite la répétition des mêmes termes. STYLE NARRATIF : Raconte le projet comme une histoire avec contexte, objectifs et résultats. LIMITE : Maximum 2 phrases pour garder la fluidité. INTERDICTION ABSOLUE : Ne reprends JAMAIS mot pour mot les missions des réalisations. Le projet doit raconter l'histoire du projet, pas lister les tâches. LANGUE : TOUT en français, à la troisième personne. STYLE : Utilise des noms d'action (Réalisation de..., Participation à..., Coordination de..., Rédaction de..., Supervision de..., Mise en service de..., etc.) au lieu de 'Il a fait...'. Exemple de style concurrent : 'Gestion de projets d'électrolyse alcaline haute pression et développement de nouveaux prototypes pour la production d'hydrogène. Ce projet inclut la sélection et la spécification des équipements, la rédaction de la documentation technique détaillée, le suivi des commandes et la conformité technique des installations.'",
		"logiciels": ["OBLIGATOIRE - Extrais TOUS les logiciels/outils mentionnés dans cette expérience (ex: SolidWorks, Python, React, AWS, Docker, etc.) - même s'ils ne sont pas explicitement listés, déduis-les du contexte"],
		"réalisations": [
		  "Liste TOUTES les missions/réalisations mentionnées dans le CV pour cette expérience. CORRESPONDANCE EXACTE OBLIGATOIRE : si le CV a 6 bullet points, tu dois avoir 6 bullet points. Si le CV a 8 bullet points, tu dois avoir 8 bullet points. Ne tronque JAMAIS - liste tout ce qui est écrit dans le CV, même si c'est très détaillé. INCLUS TOUS les détails techniques : équipements spécifiques (ESDV, Control valves, PSV, flowmeters, etc.), documents techniques (hook-up drawings, loop diagrams, wiring diagrams, etc.), calculs spécifiques (CO2 snuffing, Static mixer, etc.), types d'instruments (pressure & temperature & level transmitters, etc.). Reproduis TOUS les détails techniques mentionnés dans le CV original. LANGUE : TOUT en français, à la troisième personne (il/elle, l'ingénieur, etc.)."
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
	 - Pour chaque expérience : date_debut, date_fin, entreprise, detail_entreprise, durée, poste ET logiciels sont OBLIGATOIRES
	 - **CRITIQUE** : Relis le CV plusieurs fois pour être sûr d'avoir extrait TOUTES les expériences mentionnées
	 - **EXHAUSTIVITÉ DES RÉALISATIONS** : Pour chaque expérience, liste TOUTES les missions/réalisations mentionnées dans le CV, même si elles sont nombreuses. CORRESPONDANCE EXACTE OBLIGATOIRE : si le CV a 6 bullet points, tu dois avoir 6 bullet points. Si le CV a 8 bullet points, tu dois avoir 8 bullet points. Ne tronque JAMAIS les réalisations. INCLUS TOUS les détails techniques spécifiques : équipements (ESDV, Control valves, PSV, flowmeters), documents (hook-up drawings, loop diagrams, wiring diagrams), calculs (CO2 snuffing, Static mixer), instruments (pressure & temperature & level transmitters), etc.
	 - **ÉTOFFEMENT DU CHAMP "PROJET"** : Pour chaque expérience, étoffe le champ "projet" en créant une description fluide et connectée comme dans l'exemple concurrent. Utilise TOUTES les informations disponibles : entreprise, secteur, missions, réalisations, équipements, technologies, clients, projets mentionnés. Crée des phrases liées qui racontent une histoire cohérente, pas des bullet points séparés. IMPORTANT : Utilise des mots et formulations DIFFÉRENTS de ceux utilisés dans les réalisations - évite la répétition des mêmes termes. STYLE NARRATIF : Raconte le projet comme une histoire avec contexte, objectifs et résultats. LIMITE : Maximum 2 phrases pour garder la fluidité. INTERDICTION ABSOLUE : Ne reprends JAMAIS mot pour mot les missions des réalisations. Le projet doit raconter l'histoire du projet, pas lister les tâches. LANGUE : TOUT en français, à la troisième personne. STYLE : Utilise des noms d'action (Réalisation de..., Participation à..., Coordination de..., Rédaction de..., Supervision de..., Mise en service de..., etc.) au lieu de 'Il a fait...'. Suis le style de l'exemple concurrent avec des descriptions fluides et engageantes.
  
  2. **NOUVEAUX CHAMPS** :
	 - **"phone"** : Extrais le numéro de téléphone s'il est présent dans le CV. Format : "+33 1 23 45 67 89" ou "01.23.45.67.89" ou "0123456789". Si absent, laisse vide.
	 - **"nom"** : Extrais le nom de famille du candidat s'il est présent dans le CV. Si absent, laisse vide.
	 - **"summary"** : Crée un résumé professionnel concis (2-3 lignes max) qui présente le candidat, ses compétences principales et son expérience clé. Sois impactant et professionnel.
     - **"languages"** : Extrais toutes les langues mentionnées dans le CV (section langues, expériences internationales, formations, etc.) avec leur niveau CECR. Utilise les noms complets en français : "Français", "Anglais", "Allemand", "Espagnol", "Italien", etc. Pour le niveau, utilise les niveaux CECR (A1, A2, B1, B2, C1, C2) ou des termes comme "Natif", "Courant", "Intermédiaire", "Débutant" si le niveau CECR n'est pas spécifié. **IMPORTANT** : Si le niveau n'est pas mentionné dans le CV, laisse le champ "level" vide (""). Exemples : {"language": "Français", "level": "Natif"}, {"language": "Anglais", "level": "B2"}, {"language": "Allemand", "level": ""}. Si aucune langue n'est mentionnée, laisse un tableau vide [].
      - **"certifications"** : Extrais TOUTES les certifications mentionnées dans le CV (sections certifications, formations, education, expériences, etc.). Ne pas inventer - extraire uniquement celles explicitement mentionnées. Exemples : "BOSIET", "NEBOSH", "IOSH", "OSHA", "API", "ASME", "ISO 9001", "Six Sigma", "PMP", "Initial training on HV-LV Electrical Installations", "Electrical accreditation certificate according to the C18-510 standard", "National first-aid diploma", "BCTE Certificate", "Qess Certification", "ATEX Certificate", etc. Si aucune certification n'est mentionnée, laisse un tableau vide [].
      - **"technical_skills"** : Extrais les compétences techniques détaillées et précises mentionnées dans le CV. Ne pas se limiter aux mots-clés - extraire les compétences complètes et détaillées. Chercher dans les sections "Technical Skills", "Compétences", "Skills", expériences professionnelles, etc. Exemples : "Construction, QA/QC, Pre-Commissioning, Commissioning, hook-up, Start-up, Maintenance", "Smart Relays & Soft Starters", "Instrument calibrations and loop checks", "Knowledge of PCS, PDCS & DCS system operation", "Primary and secondary earthing system, support & cable tray installation, instrument and junction boxes installation, cable pulling dressing & termination, Circuits Control", "HV/MV/LV, firefighting, CCTV & security and access control systems", "Check list A&B, static testing (Pressure testing, resistance checks), calibration, loop check, cause and effects (loop function check and multilevel trip sequence)", "Validations, documents reviews, and investigations (Documents : URS, FAT, FDS, CSV, IQ, OQ, and PQ)", "Good experience in oil and gas industry from wellhead installations to storage and transfer", "Ability to apply safety rules – hazardous areas - explosive atmosphere (LEL), Indices of protection - protection methods using on hazardous area", "Ability to understand and use work permits system to control/coordinate and communicate all tasks", "Ability to do risk assessments before starting work", etc. Si aucune compétence technique n'est mentionnée, laisse un tableau vide [].
      - **"secteurs_activites"** : Extrais tous les secteurs d'activités dans lesquels le candidat a travaillé (ex: "Pétrole et Gaz", "Pétrochimie", "Énergie", "Automobile", "Aéronautique", "Informatique", "Finance", "Santé", "Télécommunications", etc.). Analyse les expériences professionnelles pour identifier les secteurs.
     - **"domaines_expertise"** : Extrais les domaines de compétences d'expertise du candidat (ex: "Développement Web", "Data Science", "Gestion de Projet", "Marketing Digital", "Conception Mécanique", "Intelligence Artificielle", etc.). Base-toi sur les compétences techniques et les expériences.
     - **"detail_entreprise"** : Pour chaque expérience, ajoute une description de l'entreprise en 1-2 phrases incluant : secteur d'activité, taille (startup, PME, grand groupe), spécialité, position sur le marché. Sois précis et informatif.
  
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
	 - **RÉALISATIONS COMPLÈTES** : Pour chaque expérience, extrais TOUTES les missions/réalisations mentionnées dans le CV, même si elles sont très nombreuses. CORRESPONDANCE EXACTE OBLIGATOIRE : si le CV a 6 bullet points, tu dois avoir 6 bullet points. Si le CV a 8 bullet points, tu dois avoir 8 bullet points. Ne tronque JAMAIS le contenu. INCLUS TOUS les détails techniques : équipements spécifiques (ESDV, Control valves, PSV, flowmeters, etc.), documents techniques (hook-up drawings, loop diagrams, wiring diagrams, junction boxes, equipments layouts, I/O list, Instrument list, Alarm and Set point list), calculs spécifiques (CO2 snuffing, Static mixer, etc.), types d'instruments (pressure & temperature & level transmitters, etc.). Reproduis TOUS les détails techniques du CV original.
  
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
  
  **DERNIÈRE INSTRUCTION CRITIQUE** : 
  - Extrais TOUTES les expériences et TOUTES leurs réalisations sans exception
  - Ne tronque JAMAIS le contenu, même si c'est très long
  - Assure-toi que chaque expérience a toutes ses missions/réalisations listées
  - Si une expérience a beaucoup de détails dans le CV, reproduis TOUS ces détails
  - **CORRESPONDANCE EXACTE OBLIGATOIRE** : Si le CV a 6 bullet points, tu dois avoir 6 bullet points. Si le CV a 8 bullet points, tu dois avoir 8 bullet points. Ne résume JAMAIS - reproduis TOUS les bullet points du CV original
  - **DÉTAILS TECHNIQUES OBLIGATOIRES** : Inclus TOUS les détails techniques mentionnés : équipements spécifiques (ESDV, Control valves, PSV, flowmeters, etc.), documents techniques (hook-up drawings, loop diagrams, wiring diagrams, junction boxes, equipments layouts, I/O list, Instrument list, Alarm and Set point list), calculs spécifiques (CO2 snuffing, Static mixer, etc.), types d'instruments (pressure & temperature & level transmitters, etc.)
  - **DISTINCTION PROJET/RÉALISATIONS** : Le champ "projet" doit être un paragraphe narratif fluide qui raconte l'histoire du projet (maximum 2 phrases). Le champ "réalisations" doit être une liste détaillée des missions. JAMAIS de répétition entre les deux champs - utilise des mots différents. INTERDICTION ABSOLUE : Ne reprends JAMAIS mot pour mot les missions des réalisations dans le projet. LANGUE : TOUT en français, à la troisième personne. STYLE : Utilise des noms d'action (Réalisation de..., Participation à..., Coordination de..., Rédaction de..., Supervision de..., Mise en service de..., etc.) au lieu de 'Il a fait...'.
  - **OBJECTIF** : Document le plus complet possible, pas de résumé - reproduis TOUS les détails techniques du CV original
  
  Réponds UNIQUEMENT avec le JSON structuré, sans texte avant ou après.
  `
}

// GetExtractionPromptProductionEnglish retourne le prompt anglais de production pour l'extraction
func GetExtractionPromptProductionEnglish(nuextractJSON string) string {
	return `You are an HR expert specialized in CV analysis. I want you to analyze the CV extraction JSON dictionary that I provide as input and extract ALL relevant information in the form of a structured dictionary, which can be saved as JSON, according to the following model:

DO NOT CHANGE THE KEYS OF THIS DICTIONARY, AS IT WILL BE USED OTHERWISE LATER.

{
  "prenom": "CANDIDATE'S FIRST NAME - MANDATORY. Extract the candidate's first name from the CV. This field must ALWAYS be filled except in exceptional cases where the first name is really not mentioned in the CV. Look in the header, signature, or any mention of the candidate's full name.",
  "nom": "Candidate's last name. If it is not explicitly present in the CV, leave this field empty (\"\"). Do not invent it.",
  "email": "Candidate's email address. If it is not explicitly present in the CV, leave this field empty (\"\"). Do not invent it.",
  "phone": "Candidate's phone number. If it is not explicitly present in the CV, leave this field empty (\"\"). Do not invent it.",
  "summary": "Professional summary in 2-3 lines maximum presenting the candidate, their key skills and main experience. Be concise but impactful to give an overview of the profile. STYLE: Use a professional style like 'More than X years of working experience in [field], including [specific activities] for [industries]. Experienced in [environments] with a proven track record of collaboration with [types of companies].'",
  "age": "If age is not explicitly mentioned in the CV, leave this field empty (\"\"). Do not estimate it.",
  "poste": "JOB TITLE SOUGHT - Analyze past experiences and deduce the most appropriate job title, being precise if they have a specific field of activity. If the candidate is looking for a specific position, use it. Otherwise, deduce from the most recent or most representative position of their profile. Examples: 'Mechanical Design Engineer', 'Solution Architect', 'Data Engineer', 'Project Manager', 'Full Stack Developer'",
  "diplome": "Main education (name of engineering school, business school or M2)",
  "expérience": "Calculate total experience in years: find the start date of the oldest experience and subtract from the current year (2025). If no date is available, leave empty.",
  "mobilité": "Geographic location sought if specified.",
  "disponibilité": "",
  "permis_B": "",
  "hobbies": ["List of interests"],
  "languages": [
    {
      "language": "Language name (e.g.: French, English, German, Spanish, etc.)",
      "level": "CEFR level (e.g.: A1, A2, B1, B2, C1, C2, Native, Fluent, Intermediate, Beginner, etc.)"
    }
  ],
  "certifications": ["List of certifications mentioned in the CV (e.g.: BOSIET, NEBOSH, IOSH, OSHA, API, ASME, ISO 9001, Six Sigma, PMP, Initial training on HV-LV Electrical Installations, Electrical accreditation certificate according to the C18-510 standard, National first-aid diploma, BCTE Certificate, Qess Certification, ATEX Certificate, etc.). Do not invent - extract only those explicitly mentioned."],
  "technical_skills": ["List of detailed and precise technical skills mentioned in the CV. Extract complete technical skills, not just keywords. Examples: 'Construction, QA/QC, Pre-Commissioning, Commissioning, hook-up, Start-up, Maintenance', 'Smart Relays & Soft Starters', 'Instrument calibrations and loop checks', 'Knowledge of PCS, PDCS & DCS system operation', 'Primary and secondary earthing system, support & cable tray installation, instrument and junction boxes installation, cable pulling dressing & termination, Circuits Control', 'HV/MV/LV, firefighting, CCTV & security and access control systems', 'Check list A&B, static testing (Pressure testing, resistance checks), calibration, loop check, cause and effects (loop function check and multilevel trip sequence)', 'Validations, documents reviews, and investigations (Documents : URS, FAT, FDS, CSV, IQ, OQ, and PQ)', 'Good experience in oil and gas industry from wellhead installations to storage and transfer', 'Ability to apply safety rules – hazardous areas - explosive atmosphere (LEL), Indices of protection - protection methods using on hazardous area', 'Ability to understand and use work permits system to control/coordinate and communicate all tasks', 'Ability to do risk assessments before starting work', etc."],
  "secteurs_activites": ["List of business sectors in which the candidate has worked (e.g.: Oil & Gas, Petrochemical, Energy, Automotive, Aerospace, IT, Finance, Health, etc.)"],
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
      "date_debut": "MANDATORY - Start date in MM/YY format (e.g.: 02/20, 09/19)",
      "date_fin": "MANDATORY - End date in MM/YY format (e.g.: 12/22, 08/20). If the experience is ongoing, use 'In progress'",
      "entreprise": "MANDATORY - Company name",
      "detail_entreprise": "MANDATORY - Company description in 1-2 sentences: business sector, size, specialty, market position. Example: 'Startup specialized in artificial intelligence and machine learning, with 50 employees and leader in predictive analysis for the banking sector'",
      "durée": "MANDATORY - Automatically calculated duration (e.g.: 2 years, 6 months, 1 year 3 months). If less than 1 year, display in months. If greater than or equal to 1 year, display in years.",
      "poste": "MANDATORY - Job title held (e.g.: Design Engineer, Senior Developer, Project Manager, etc.)",
      "contexte": "Summarize the experience succinctly to present the project carried out in one sentence.",
      "projet": "EXPAND this field by creating a fluid and connected description of the project/experience, like in the competitor example. Use ALL available information: company, sector, missions, achievements, equipment, technologies, clients, mentioned projects. Create linked sentences that tell a coherent story, not separate bullet points. IMPORTANT: Use words and formulations DIFFERENT from those used in achievements - avoid repetition of the same terms. NARRATIVE STYLE: Tell the project as a story with context, objectives and results. LIMIT: Maximum 2 sentences to maintain fluidity. ABSOLUTE PROHIBITION: NEVER reproduce word for word the missions from achievements. The project must tell the story of the project, not list tasks. LANGUAGE: EVERYTHING in English, in the third person. STYLE: Use action nouns (Realization of..., Participation in..., Coordination of..., Writing of..., Supervision of..., Commissioning of..., etc.) instead of 'He did...'. Example of competitor style: 'Management of high-pressure alkaline electrolysis projects and development of new prototypes for hydrogen production. This project includes equipment selection and specification, detailed technical documentation writing, order tracking and technical compliance of installations.'",
      "logiciels": ["MANDATORY - Extract ALL software/tools mentioned in this experience (e.g.: SolidWorks, Python, React, AWS, Docker, etc.) - even if they are not explicitly listed, deduce them from the context"],
      "réalisations": [
        "List ALL missions/achievements mentioned in the CV for this experience. EXACT CORRESPONDENCE MANDATORY: if the CV has 6 bullet points, you must have 6 bullet points. If the CV has 8 bullet points, you must have 8 bullet points. Never truncate - list everything written in the CV, even if very detailed. INCLUDE ALL technical details: specific equipment (ESDV, Control valves, PSV, flowmeters, etc.), technical documents (hook-up drawings, loop diagrams, wiring diagrams, etc.), specific calculations (CO2 snuffing, Static mixer, etc.), instrument types (pressure & temperature & level transmitters, etc.). Reproduce ALL technical details mentioned in the original CV. LANGUAGE: EVERYTHING in English, in the third person (he/she, the engineer, etc.)."
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
   - Extract ALL professional experiences (internships, permanent contracts, fixed-term contracts, apprenticeships, etc.) - DO NOT MISS A SINGLE ONE
   - For each education: date_debut, date_fin, diplome AND ecole_cursus are MANDATORY
   - For each experience: date_debut, date_fin, entreprise, detail_entreprise, durée, poste AND logiciels are MANDATORY
   - **CRITICAL**: Reread the CV several times to make sure you have extracted ALL mentioned experiences
   - **EXHAUSTIVENESS OF ACHIEVEMENTS**: For each experience, list ALL missions/achievements mentioned in the CV, even if numerous. EXACT CORRESPONDENCE MANDATORY: if the CV has 6 bullet points, you must have 6 bullet points. If the CV has 8 bullet points, you must have 8 bullet points. Never truncate achievements. INCLUDE ALL specific technical details: equipment (ESDV, Control valves, PSV, flowmeters), documents (hook-up drawings, loop diagrams, wiring diagrams), calculations (CO2 snuffing, Static mixer), instruments (pressure & temperature & level transmitters), etc.
   - **EXPANSION OF "PROJECT" FIELD**: For each experience, expand the "projet" field by creating a fluid and connected description like in the competitor example. Use ALL available information: company, sector, missions, achievements, equipment, technologies, clients, mentioned projects. Create linked sentences that tell a coherent story, not separate bullet points. IMPORTANT: Use words and formulations DIFFERENT from those used in achievements - avoid repetition of the same terms. NARRATIVE STYLE: Tell the project as a story with context, objectives and results. LIMIT: Maximum 2 sentences to maintain fluidity. ABSOLUTE PROHIBITION: NEVER reproduce word for word the missions from achievements. The project must tell the story of the project, not list tasks. LANGUAGE: EVERYTHING in English, in the third person. STYLE: Use action nouns (Realization of..., Participation in..., Coordination of..., Writing of..., Supervision of..., Commissioning of..., etc.) instead of 'He did...'. Follow the competitor example style with fluid and engaging descriptions.

2. **NEW FIELDS**:
   - **"phone"**: Extract the phone number if present in the CV. Format: "+33 1 23 45 67 89" or "01.23.45.67.89" or "0123456789". If absent, leave empty.
   - **"nom"**: Extract the candidate's last name if present in the CV. If absent, leave empty.
   - **"summary"**: Create a concise professional summary (2-3 lines max) that presents the candidate, their main skills and key experience. Be impactful and professional.
   - **"languages"**: Extract all languages mentioned in the CV (language section, international experiences, education, etc.) with their CEFR level. Use full names in English: "French", "English", "German", "Spanish", "Italian", etc. For the level, use CEFR levels (A1, A2, B1, B2, C1, C2) or terms like "Native", "Fluent", "Intermediate", "Beginner" if CEFR level is not specified. **IMPORTANT**: If the level is not mentioned in the CV, leave the "level" field empty (""). Examples: {"language": "French", "level": "Native"}, {"language": "English", "level": "B2"}, {"language": "German", "level": ""}. If no language is mentioned, leave an empty array [].
   - **"certifications"**: Extract ALL certifications mentioned in the CV (certifications sections, education, experiences, etc.). Do not invent - extract only those explicitly mentioned. Examples: "BOSIET", "NEBOSH", "IOSH", "OSHA", "API", "ASME", "ISO 9001", "Six Sigma", "PMP", "Initial training on HV-LV Electrical Installations", "Electrical accreditation certificate according to the C18-510 standard", "National first-aid diploma", "BCTE Certificate", "Qess Certification", "ATEX Certificate", etc. If no certifications are mentioned, leave an empty array [].
   - **"technical_skills"**: Extract detailed and precise technical skills mentioned in the CV. Do not limit yourself to keywords - extract complete and detailed skills. Look in "Technical Skills", "Skills", "Compétences" sections, professional experiences, etc. Examples: "Construction, QA/QC, Pre-Commissioning, Commissioning, hook-up, Start-up, Maintenance", "Smart Relays & Soft Starters", "Instrument calibrations and loop checks", "Knowledge of PCS, PDCS & DCS system operation", "Primary and secondary earthing system, support & cable tray installation, instrument and junction boxes installation, cable pulling dressing & termination, Circuits Control", "HV/MV/LV, firefighting, CCTV & security and access control systems", "Check list A&B, static testing (Pressure testing, resistance checks), calibration, loop check, cause and effects (loop function check and multilevel trip sequence)", "Validations, documents reviews, and investigations (Documents : URS, FAT, FDS, CSV, IQ, OQ, and PQ)", "Good experience in oil and gas industry from wellhead installations to storage and transfer", "Ability to apply safety rules – hazardous areas - explosive atmosphere (LEL), Indices of protection - protection methods using on hazardous area", "Ability to understand and use work permits system to control/coordinate and communicate all tasks", "Ability to do risk assessments before starting work", etc. If no technical skills are mentioned, leave an empty array [].
   - **"secteurs_activites"**: Extract all business sectors in which the candidate has worked (e.g.: "Oil & Gas", "Petrochemical", "Energy", "Automotive", "Aerospace", "IT", "Finance", "Health", "Telecommunications", etc.). Analyze professional experiences to identify sectors.
   - **"domaines_expertise"**: Extract the candidate's expertise domains (e.g.: "Web Development", "Data Science", "Project Management", "Digital Marketing", "Mechanical Design", "Artificial Intelligence", etc.). Base yourself on technical skills and experiences.
   - **"detail_entreprise"**: For each experience, add a company description in 1-2 sentences including: business sector, size (startup, SME, large group), specialty, market position. Be precise and informative.

3. **"poste" FIELD**:
   - This is the JOB TITLE SOUGHT based on the analysis of past experiences
   - **EXTRACTION METHOD**:
     a) If the candidate indicates a specific sought position → use it
     b) Otherwise, analyze all experiences and deduce the most representative title
     c) Prioritize the most recent position or the one that best reflects career evolution
     d) Be precise and professional in the title (avoid generic terms)
   - Examples: "Mechanical Design Engineer", "Solution Architect", "Data Engineer", "Full Stack Developer", "Project Manager", "Consultant", "Civil Engineer", "Product Manager"

4. **DATES AND EXPERIENCE CALCULATION**:
   - **EXPERIENCE DATES**: ALWAYS extract start and end dates of each experience
   - Date format: strictly MM/YY (e.g.: 02/20, 12/22). NO other format is accepted
   - If the experience is ongoing, use "In progress" for date_fin (not MM/YY)
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
   - **COMPLETE ACHIEVEMENTS**: For each experience, extract ALL missions/achievements mentioned in the CV, even if very numerous. EXACT CORRESPONDENCE MANDATORY: if the CV has 6 bullet points, you must have 6 bullet points. If the CV has 8 bullet points, you must have 8 bullet points. Never truncate content. INCLUDE ALL technical details: specific equipment (ESDV, Control valves, PSV, flowmeters, etc.), technical documents (hook-up drawings, loop diagrams, wiring diagrams, junction boxes, equipments layouts, I/O list, Instrument list, Alarm and Set point list), specific calculations (CO2 snuffing, Static mixer, etc.), instrument types (pressure & temperature & level transmitters, etc.). Reproduce ALL technical details from the original CV.

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

Here is the extraction JSON to analyze:

` + nuextractJSON + `

**LAST CRITICAL INSTRUCTION**: 
- Extract ALL experiences and ALL their achievements without exception
- Never truncate content, even if very long
- Make sure each experience has all its missions/achievements listed
- If an experience has many details in the CV, reproduce ALL these details
- **EXACT CORRESPONDENCE MANDATORY**: If the CV has 6 bullet points, you must have 6 bullet points. If the CV has 8 bullet points, you must have 8 bullet points. Never summarize - reproduce ALL bullet points from the original CV
- **MANDATORY TECHNICAL DETAILS**: Include ALL mentioned technical details: specific equipment (ESDV, Control valves, PSV, flowmeters, etc.), technical documents (hook-up drawings, loop diagrams, wiring diagrams, junction boxes, equipments layouts, I/O list, Instrument list, Alarm and Set point list), specific calculations (CO2 snuffing, Static mixer, etc.), instrument types (pressure & temperature & level transmitters, etc.)
- **PROJECT/ACHIEVEMENTS DISTINCTION**: The "projet" field must be a fluid narrative paragraph that tells the story of the project (maximum 2 sentences). The "réalisations" field must be a detailed list of missions. NEVER repetition between the two fields - use different words. ABSOLUTE PROHIBITION: NEVER reproduce word for word the missions from achievements in the project. LANGUAGE: EVERYTHING in English, in the third person. STYLE: Use action nouns (Realization of..., Participation in..., Coordination of..., Writing of..., Supervision of..., Commissioning of..., etc.) instead of 'He did...'.
- **OBJECTIVE**: Most complete document possible, no summary - reproduce ALL technical details from the original CV

Respond ONLY with the structured JSON, without text before or after.
`
}
