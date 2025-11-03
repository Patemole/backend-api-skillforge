package nuextract

// GetExtractionPromptProduction retourne le prompt de production pour l'extraction et la structuration des données CV
func GetExtractionPromptProduction(nuextractJSON string) string {
	return GetExtractionPromptProductionWithLanguage(nuextractJSON, "fr")
}

// GetExtractionPromptProductionWithLanguage retourne le prompt de production pour l'extraction selon la langue
func GetExtractionPromptProductionWithLanguage(nuextractJSON string, language string) string {
	switch language {
	case "en":
		return GetExtractionPromptProductionEnglish(nuextractJSON)
	case "pr":
		return GetExtractionPromptProductionPortuguese(nuextractJSON)
	case "de":
		return GetExtractionPromptProductionGerman(nuextractJSON)
	case "sp":
		return GetExtractionPromptProductionSpanish(nuextractJSON)
	case "it":
		return GetExtractionPromptProductionItalian(nuextractJSON)
	default:
		return GetExtractionPromptProductionFrench(nuextractJSON)
	}
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
    "competence_fonctionnelle": ["Liste des compétences fonctionnelles mentionnées dans le CV. Ces compétences peuvent être présentes dans les expériences, les soft skills, ou dans le résumé. Extrais les compétences fonctionnelles comme : 'Planification, coordination', 'Gestion de projet', 'Évaluation et comparaison de technologies', 'Reporting et analyse, communication interne', 'Prise de décision et recommandation stratégique', 'Pilotage de projet en autonomie', 'Analyse fonctionnelle et technique', 'Veille technologique', 'Interactions avec des parties prenantes techniques', 'Comparaison et sélection de solutions techniques', 'Synthèse et reporting', 'Maîtrise des outils d'amélioration continue (Lean, Six Sigma)', 'Optimisation de la production', 'Efficacité industrielle', 'Étude des non-conformités, diagnostic et résolution de problèmes', 'Autonomie et prise d'initiative', 'Communication et collaboration interne', 'Adaptabilité du discours, gestion du changement', 'Esprit d'analyse et de synthèse', 'Leadership d'équipe', 'Gestion budgétaire', 'Négociation commerciale', 'Formation et développement d'équipe', 'Stratégie et vision', 'Innovation et créativité', 'Relation client', 'Gestion de la qualité', 'Organisation et planification', 'Résolution de conflits', etc. Si aucune compétence fonctionnelle n'est mentionnée, laisse un tableau vide []."],
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
		"secteur": "Secteur d'activité de l'expérience en 1-2 mots. Extrais le secteur depuis les informations de l'entreprise, du projet ou du contexte. Exemples : Agroalimentaire, Automobile, Banque, Bâtiments, Biomédical, Chimie, Conseil, Défense, Énergie, Environnement, Ferroviaire, Grande distribution, Infrastructure, Logistique, Métallurgie / Sidérurgie, Naval, Nucléaire, Oil & Gas, Pétrochimie, Pharmaceutique, Santé, Secteur public, Télécommunications, IRVE, Photovoltaïque, Traitement des eaux, Revalorisation énergétique, Hydroélectricité, ENR (Énergies renouvelables), Énergie éolienne, Biogaz, Education, Ressources Humaines. Si le secteur n'est pas identifiable, laisse ce champ vide (\"\").",
		"contexte": "Résume l'expérience succinctement pour présenter le projet réalisé en une phrase.",
		"projet": "ÉTOFFE ce champ en créant une description fluide et connectée du projet/expérience, comme dans l'exemple concurrent. Utilise TOUTES les informations disponibles : entreprise, secteur, missions, réalisations, équipements, technologies, clients, projets mentionnés. Crée des phrases liées qui racontent une histoire cohérente, pas des bullet points séparés. IMPORTANT : Utilise des mots et formulations DIFFÉRENTS de ceux utilisés dans les réalisations - évite la répétition des mêmes termes. STYLE NARRATIF : Raconte le projet comme une histoire avec contexte, objectifs et résultats. LIMITE : Maximum 2 phrases pour garder la fluidité. INTERDICTION ABSOLUE : Ne reprends JAMAIS mot pour mot les missions des réalisations. Le projet doit raconter l'histoire du projet, pas lister les tâches. LANGUE : TOUT en français, à la troisième personne. STYLE : Utilise des noms d'action (Réalisation de..., Participation à..., Coordination de..., Rédaction de..., Supervision de..., Mise en service de..., etc.) au lieu de 'Il a fait...'. Exemple de style concurrent : 'Gestion de projets d'électrolyse alcaline haute pression et développement de nouveaux prototypes pour la production d'hydrogène. Ce projet inclut la sélection et la spécification des équipements, la rédaction de la documentation technique détaillée, le suivi des commandes et la conformité technique des installations.'",
		"projets_name": ["Liste des noms de projets mentionnés pour cette expérience, si présents. Extrais les noms de projets spécifiques mentionnés dans le contexte, les réalisations ou ailleurs dans la description de l'expérience. Souvent il y a un seul projet par expérience, mais il peut y en avoir plusieurs. Exemples : 'Projet de réhabilitation d'un immeuble de bureaux pour WEWORK', 'Projet Reine des Neiges', 'Projet de réhabilitation d'un immeuble de bureaux pour AXA - Boulevard des Italiens', etc. Si aucun nom de projet n'est mentionné, laisse un tableau vide []."],
		"result": "Phrase courte qui explique le résultat de l'expérience. Extrais ou déduis une phrase concise qui résume les résultats obtenus ou les accomplissements clés de cette expérience. Exemples : 'Recommandations et identification de Next Step', 'Location de technologie au sein des laboratoires de recherche pour des essais', 'Mise en place du projet et de l'équipe en charge du projet, identification des causes et d'un mode d'évaluation, planification des essais', etc. Si aucun résultat n'est mentionné, laisse ce champ vide (\"\").",
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
      - **"competence_fonctionnelle"** : Extrais les compétences fonctionnelles mentionnées dans le CV. Ces compétences peuvent être présentes dans les expériences, les soft skills, ou dans le résumé. Chercher dans toutes les sections du CV pour identifier ces compétences fonctionnelles. Exemples : "Planification, coordination", "Gestion de projet", "Évaluation et comparaison de technologies", "Reporting et analyse, communication interne", "Prise de décision et recommandation stratégique", "Pilotage de projet en autonomie", "Analyse fonctionnelle et technique", "Veille technologique", "Interactions avec des parties prenantes techniques", "Comparaison et sélection de solutions techniques", "Synthèse et reporting", "Maîtrise des outils d'amélioration continue (Lean, Six Sigma)", "Optimisation de la production", "Efficacité industrielle", "Étude des non-conformités, diagnostic et résolution de problèmes", "Autonomie et prise d'initiative", "Communication et collaboration interne", "Adaptabilité du discours, gestion du changement", "Esprit d'analyse et de synthèse", "Leadership d'équipe", "Gestion budgétaire", "Négociation commerciale", "Formation et développement d'équipe", "Stratégie et vision", "Innovation et créativité", "Relation client", "Gestion de la qualité", "Organisation et planification", "Résolution de conflits", etc. Si aucune compétence fonctionnelle n'est mentionnée, laisse un tableau vide [].
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
  "competence_fonctionnelle": ["List of functional competencies mentioned in the CV. These competencies can be present in experiences, soft skills, or in the resume. Extract functional competencies such as: 'Planning, coordination', 'Project management', 'Technology evaluation and comparison', 'Reporting and analysis, internal communication', 'Decision-making and strategic recommendation', 'Autonomous project management', 'Functional and technical analysis', 'Technology watch', 'Interactions with technical stakeholders', 'Technical solution comparison and selection', 'Synthesis and reporting', 'Continuous improvement tools mastery (Lean, Six Sigma)', 'Production optimization', 'Industrial efficiency', 'Non-conformity study, diagnosis and problem resolution', 'Autonomy and initiative', 'Internal communication and collaboration', 'Discourse adaptability, change management', 'Analysis and synthesis mindset', 'Team leadership', 'Budget management', 'Commercial negotiation', 'Team training and development', 'Strategy and vision', 'Innovation and creativity', 'Customer relations', 'Quality management', 'Organization and planning', 'Conflict resolution', etc. If no functional competency is mentioned, leave an empty array []."],
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
      "secteur": "Business sector of the experience in 1-2 words. Extract the sector from the company information, project, or context. Examples: Food industry, Automotive, Banking, Construction, Biomedical, Chemistry, Consulting, Defense, Energy, Environment, Railway, Retail, Infrastructure, Logistics, Metallurgy / Steel industry, Naval, Nuclear, Oil & Gas, Petrochemical, Pharmaceutical, Healthcare, Public sector, Telecommunications, IRVE, Photovoltaics, Water treatment, Energy recovery, Hydroelectricity, Renewable Energy (ENR), Wind energy, Biogas, Education, Human Resources. If the sector is not identifiable, leave this field empty (\"\").",
      "contexte": "Summarize the experience succinctly to present the project carried out in one sentence.",
      "projet": "EXPAND this field by creating a fluid and connected description of the project/experience, like in the competitor example. Use ALL available information: company, sector, missions, achievements, equipment, technologies, clients, mentioned projects. Create linked sentences that tell a coherent story, not separate bullet points. IMPORTANT: Use words and formulations DIFFERENT from those used in achievements - avoid repetition of the same terms. NARRATIVE STYLE: Tell the project as a story with context, objectives and results. LIMIT: Maximum 2 sentences to maintain fluidity. ABSOLUTE PROHIBITION: NEVER reproduce word for word the missions from achievements. The project must tell the story of the project, not list tasks. LANGUAGE: EVERYTHING in English, in the third person. STYLE: Use action nouns (Realization of..., Participation in..., Coordination of..., Writing of..., Supervision of..., Commissioning of..., etc.) instead of 'He did...'. Example of competitor style: 'Management of high-pressure alkaline electrolysis projects and development of new prototypes for hydrogen production. This project includes equipment selection and specification, detailed technical documentation writing, order tracking and technical compliance of installations.'",
      "projets_name": ["List of project names mentioned for this experience, if present. Extract specific project names mentioned in the context, achievements or elsewhere in the experience description. Often there is only one project per experience, but there can be several. Examples: 'Office building rehabilitation project for WEWORK', 'Reine des Neiges Project', 'Office building rehabilitation project for AXA - Boulevard des Italiens', etc. If no project name is mentioned, leave an empty array []."],
      "result": "Short phrase explaining the result of the experience. Extract or deduce a concise phrase that summarizes the results achieved or key accomplishments of this experience. Examples: 'Recommendations and identification of Next Step', 'Technology rental within research laboratories for trials', 'Project and project team setup, identification of causes and evaluation method, test planning', etc. If no result is mentioned, leave this field empty (\"\").",
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
   - **"competence_fonctionnelle"**: Extract functional competencies mentioned in the CV. These competencies can be present in experiences, soft skills, or in the resume. Search in all CV sections to identify these functional competencies. Examples: "Planning, coordination", "Project management", "Technology evaluation and comparison", "Reporting and analysis, internal communication", "Decision-making and strategic recommendation", "Autonomous project management", "Functional and technical analysis", "Technology watch", "Interactions with technical stakeholders", "Technical solution comparison and selection", "Synthesis and reporting", "Continuous improvement tools mastery (Lean, Six Sigma)", "Production optimization", "Industrial efficiency", "Non-conformity study, diagnosis and problem resolution", "Autonomy and initiative", "Internal communication and collaboration", "Discourse adaptability, change management", "Analysis and synthesis mindset", "Team leadership", "Budget management", "Commercial negotiation", "Team training and development", "Strategy and vision", "Innovation and creativity", "Customer relations", "Quality management", "Organization and planning", "Conflict resolution", etc. If no functional competency is mentioned, leave an empty array [].
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

// GetExtractionPromptProductionPortuguese retorna o prompt de produção em português para a extração
func GetExtractionPromptProductionPortuguese(nuextractJSON string) string {
	return `Você é um especialista de RH, especializado na análise de currículos (CV). Quero que você analise o dicionário JSON da extração de CV que forneço como input e extraia TODAS as informações relevantes na forma de um dicionário estruturado, que possa ser salvo em JSON, de acordo com o modelo a seguir:
  
  NÃO ALTERE DE FORMA ALGUMA AS CHAVES DESTE DICIONÁRIO, POIS ELE SERÁ UTILIZADO POSTERIORMENTE.
  
  {
	"prenom": "NOME PRÓPRIO DO CANDIDATO – OBRIGATÓRIO. Extraia o primeiro nome do candidato a partir do CV. Este campo deve SEMPRE ser preenchido, exceto em casos excepcionais em que o primeiro nome realmente não esteja mencionado no CV. Procure no cabeçalho, na assinatura ou em qualquer menção ao nome completo do candidato.",
	"nom": "Sobrenome do candidato. Se não estiver explicitamente presente no CV, deixe este campo vazio (\"\"). Não invente.",
	"email": "Endereço de e-mail do candidato. Se não estiver explicitamente presente no CV, deixe este campo vazio (\"\"). Não invente.",
	"phone": "Número de telefone do candidato. Se não estiver explicitamente presente no CV, deixe este campo vazio (\"\"). Não invente.",
    "summary": "Resumo profissional em no máximo 2–3 linhas apresentando o candidato, suas competências-chave e experiência principal. Seja conciso, mas impactante, para dar uma visão geral do perfil. ESTILO: Use um estilo profissional como 'Mais de X anos de experiência em [área], incluindo [atividades específicas] para [indústrias]. Experiência em [ambientes] com histórico comprovado de colaboração com [tipos de empresas].'",
	"age": "Se a idade não estiver explicitamente mencionada no CV, deixe este campo vazio (\"\"). Não estime.",
	"poste": "TÍTULO DO CARGO PRETENDIDO – Analise as experiências anteriores e deduza o título de cargo mais apropriado, sendo preciso se houver um domínio de atuação específico. Se o candidato busca um cargo específico, use-o. Caso contrário, deduza a partir do cargo mais recente ou mais representativo do seu perfil. Exemplos: 'Engenheiro de Concepção Mecânica', 'Solution Architect', 'Data Engineer', 'Gestor de Projetos', 'Desenvolvedor Full Stack'",
	"diplome": "Formação principal (nome da escola de engenharia, de negócios ou do mestrado M2).",
	"expérience": "Calcule a experiência total em anos: encontre a data de início da experiência mais antiga e subtraia do ano atual (2025). Se nenhuma data estiver disponível, deixe em branco.",
	"mobilité": "Localização geográfica desejada, se especificada.",
	"disponibilité": "",
	"permis_B": "",
    "hobbies": ["Lista de interesses"],
    "languages": [
      {
        "language": "Nome do idioma (ex.: Francês, Inglês, Alemão, Espanhol, etc.)",
        "level": "Nível QECR (ex.: A1, A2, B1, B2, C1, C2, Nativo, Fluente, Intermédio, Iniciante, etc.)"
      }
    ],
    "certifications": ["Lista de certificações mencionadas no CV (ex.: BOSIET, NEBOSH, IOSH, OSHA, API, ASME, ISO 9001, Six Sigma, PMP, Initial training on HV-LV Electrical Installations, Electrical accreditation certificate according to the C18-510 standard, National first-aid diploma, BCTE Certificate, Qess Certification, ATEX Certificate, etc.). Não invente – extraia apenas as que são explicitamente mencionadas."],
    "technical_skills": ["Lista de competências técnicas detalhadas e precisas mencionadas no CV. Extraia competências técnicas completas, não apenas palavras-chave. Exemplos: 'Construction, QA/QC, Pre-Commissioning, Commissioning, hook-up, Start-up, Maintenance', 'Smart Relays & Soft Starters', 'Instrument calibrations and loop checks', 'Knowledge of PCS, PDCS & DCS system operation', 'Primary and secondary earthing system, support & cable tray installation, instrument and junction boxes installation, cable pulling dressing & termination, Circuits Control', 'HV/MV/LV, firefighting, CCTV & security and access control systems', 'Check list A&B, static testing (Pressure testing, resistance checks), calibration, loop check, cause and effects (loop function check and multilevel trip sequence)', 'Validations, documents reviews, and investigations (Documents : URS, FAT, FDS, CSV, IQ, OQ, and PQ)', 'Good experience in oil and gas industry from wellhead installations to storage and transfer', 'Ability to apply safety rules – hazardous areas - explosive atmosphere (LEL), Indices of protection - protection methods using on hazardous area', 'Ability to understand and use work permits system to control/coordinate and communicate all tasks', 'Ability to do risk assessments before starting work', etc."],
    "competence_fonctionnelle": ["Lista de competências funcionais mencionadas no CV. Essas competências podem estar presentes nas experiências, soft skills ou no resumo. Extraia competências funcionais como: 'Planejamento, coordenação', 'Gestão de projetos', 'Avaliação e comparação de tecnologias', 'Relatórios e análise, comunicação interna', 'Tomada de decisão e recomendação estratégica', 'Gestão autônoma de projetos', 'Análise funcional e técnica', 'Veille tecnológica', 'Interações com partes interessadas técnicas', 'Comparação e seleção de soluções técnicas', 'Síntese e relatórios', 'Domínio de ferramentas de melhoria contínua (Lean, Six Sigma)', 'Otimização da produção', 'Eficiência industrial', 'Estudo de não conformidades, diagnóstico e resolução de problemas', 'Autonomia e iniciativa', 'Comunicação e colaboração interna', 'Adaptabilidade do discurso, gestão da mudança', 'Espírito de análise e síntese', 'Liderança de equipe', 'Gestão orçamentária', 'Negociação comercial', 'Formação e desenvolvimento de equipe', 'Estratégia e visão', 'Inovação e criatividade', 'Relacionamento com clientes', 'Gestão da qualidade', 'Organização e planejamento', 'Resolução de conflitos', etc. Se nenhuma competência funcional for mencionada, deixe um array vazio []."],
     "secteurs_activites": ["Lista dos setores de atividade em que o candidato trabalhou (ex.: Petróleo e Gás, Petroquímica, Energia, Automotivo, Aeronáutico, TI, Finanças, Saúde, etc.)."],
    "domaines_expertise": ["Lista dos domínios de especialização do candidato (ex.: Desenvolvimento Web, Data Science, Gestão de Projetos, Marketing Digital, etc.)."],
	"formations": [
	  {
		"date_debut": "OBRIGATÓRIO – Ano de início (ex.: 2020, 2018-2019)",
		"date_fin": "OBRIGATÓRIO – Ano de conclusão (ex.: 2022, 2020-2021)",
		"diplome": "OBRIGATÓRIO – Tipo exato de diploma (ex.: Mestrado em Engenharia Mecânica, Diploma de Engenheiro, Bacharelado em Informática, BTS Comércio, Diploma de Medicina, MBA, etc.)",
		"ecole_cursus": "OBRIGATÓRIO – Nome completo da escola/universidade (ex.: École Centrale Paris, Université Pierre et Marie Curie, HEC Paris, etc.)"
	  }
	],
    "expériences": [
	  {
        "date_debut": "OBRIGATÓRIO – Data de início no formato MM/YY (ex.: 02/20, 09/19)",
        "date_fin": "OBRIGATÓRIO – Data de término no formato MM/YY (ex.: 12/22, 08/20). Se a experiência estiver em andamento, use 'Em curso'",
		"entreprise": "OBRIGATÓRIO – Nome da empresa",
		"detail_entreprise": "OBRIGATÓRIO – Descrição da empresa em 1–2 frases: setor de atividade, porte, especialidade, posição no mercado. Exemplo: 'Startup especializada em inteligência artificial e machine learning, com 50 colaboradores e líder em análise preditiva para o setor bancário'",
		"durée": "OBRIGATÓRIO - Duração calculada automaticamente (ex.: 2 anos, 6 meses, 1 ano e 3 meses). Se < 1 ano, exibir em meses. Se ≥ 1 ano, exibir em anos.",
		"poste": "OBRIGATÓRIO - Título do cargo exercido (ex.: Engenheiro de Concepção, Desenvolvedor Sênior, Gestor de Projetos, etc.)",
		"secteur": "Setor de atividade da experiência em 1-2 palavras. Extraia o setor a partir das informações da empresa, do projeto ou do contexto. Exemplos: Agroalimentar, Automotivo, Bancário, Construção, Biomédico, Química, Consultoria, Defesa, Energia, Meio ambiente, Ferroviário, Grande distribuição, Infraestrutura, Logística, Metalurgia / Siderurgia, Naval, Nuclear, Petróleo e Gás, Petroquímica, Farmacêutico, Saúde, Setor público, Telecomunicações, IRVE, Fotovoltaico, Tratamento de águas, Revalorização energética, Hidrelétrica, Energias Renováveis (ENR), Energia eólica, Biogás, Educação, Recursos Humanos. Se o setor não for identificável, deixe este campo vazio (\"\").",
		"contexte": "Resuma a experiência de forma sucinta para apresentar o projeto realizado em uma frase.",
		"projet": "APROFUNDE este campo criando uma descrição fluida e conectada do projeto/experiência, como no exemplo do concorrente. Use TODAS as informações disponíveis: empresa, setor, missões, realizações, equipamentos, tecnologias, clientes, projetos mencionados. Crie frases encadeadas que contem uma história coerente, não tópicos separados. IMPORTANTE: Utilize palavras e formulações DIFERENTES das usadas nas realizações – evite repetição dos mesmos termos. ESTILO NARRATIVO: Conte o projeto como uma história com contexto, objetivos e resultados. LIMITE: Máximo de 2 frases para manter a fluidez. PROIBIÇÃO ABSOLUTA: NUNCA reproduza palavra por palavra as missões das realizações. O projeto deve contar a história, não listar tarefas. LÍNGUA: TUDO em português, na terceira pessoa. ESTILO: Use nomes de ação (Realização de..., Participação em..., Coordenação de..., Redação de..., Supervisão de..., Comissionamento de..., etc.) em vez de 'Ele/ela fez...'. Exemplo de estilo do concorrente: 'Gestão de projetos de eletrólise alcalina de alta pressão e desenvolvimento de novos protótipos para produção de hidrogénio. Este projeto inclui a seleção e especificação de equipamentos, a elaboração de documentação técnica detalhada, o acompanhamento de pedidos e a conformidade técnica das instalações.'",
		"projets_name": ["Lista dos nomes de projetos mencionados para esta experiência, se presentes. Extraia nomes específicos de projetos mencionados no contexto, realizações ou em outro lugar na descrição da experiência. Frequentemente há apenas um projeto por experiência, mas pode haver vários. Exemplos: 'Projeto de reabilitação de edifício de escritórios para WEWORK', 'Projeto Reine des Neiges', 'Projeto de reabilitação de edifício de escritórios para AXA - Boulevard des Italiens', etc. Se nenhum nome de projeto for mencionado, deixe um array vazio []."],
		"result": "Frase curta que explica o resultado da experiência. Extraia ou deduza uma frase concisa que resuma os resultados obtidos ou os principais feitos desta experiência. Exemplos: 'Recomendações e identificação de Next Step', 'Locação de tecnologia nos laboratórios de pesquisa para testes', 'Implementação do projeto e da equipe responsável pelo projeto, identificação das causas e de um método de avaliação, planejamento dos testes', etc. Se nenhum resultado for mencionado, deixe este campo vazio (\"\").",
		"logiciels": ["OBRIGATÓRIO – Extraia TODOS os softwares/ferramentas mencionados nesta experiência (ex.: SolidWorks, Python, React, AWS, Docker, etc.) – mesmo que não estejam explicitamente listados, deduza pelo contexto"],
		"réalisations": [
		  "Liste TODAS as missões/realizações mencionadas no CV para esta experiência. CORRESPONDÊNCIA EXATA OBRIGATÓRIA: se o CV tiver 6 tópicos, você deve ter 6 tópicos. Se tiver 8, você deve ter 8. Nunca resuma – liste tudo o que está escrito no CV, mesmo que muito detalhado. INCLUA TODOS os detalhes técnicos: equipamentos específicos (ESDV, Control valves, PSV, flowmeters, etc.), documentos técnicos (hook-up drawings, loop diagrams, wiring diagrams, etc.), cálculos específicos (CO2 snuffing, Static mixer, etc.), tipos de instrumentos (pressure & temperature & level transmitters, etc.). Reproduza TODOS os detalhes técnicos mencionados no CV original. LÍNGUA: TUDO em português, na terceira pessoa (ele/ela, o/a engenheiro/a, etc.)."
		],
		"AI_suggest": ["Se puder deduzir elementos relevantes não presentes no CV. As sugestões devem ser específicas e adaptadas a cada experiência, relevantes para recrutadores; a quantidade deve variar conforme a experiência, sem redundâncias. Não inclua sistematicamente: deve soar natural."]
	  }
	],
	"logiciels": [
	  {
		"logiciel": "",
		"level": "Estime o nível entre: Iniciante, Intermédio, Avançado, Expert.",
		"temps_utilisation": "Estime o tempo de utilização em meses."
	  }
	]
  }
  
  INSTRUÇÕES CRÍTICAS:
  
  1. **EXTRAÇÕES OBRIGATÓRIAS**:
	 - Extraia TODAS as experiências profissionais (estágios, contratos permanentes, temporários, alternância, etc.) – NÃO DEIXE NENHUMA DE FORA
	 - Para cada formação: date_debut, date_fin, diplome E ecole_cursus são OBRIGATÓRIOS
	 - Para cada experiência: date_debut, date_fin, entreprise, detail_entreprise, durée, poste E logiciels são OBRIGATÓRIOS
	 - **CRÍTICO**: Releia o CV várias vezes para garantir que TODAS as experiências mencionadas foram extraídas
	 - **EXAUSTIVIDADE DAS REALIZAÇÕES**: Para cada experiência, liste TODAS as missões/realizações mencionadas no CV, mesmo que numerosas. CORRESPONDÊNCIA EXATA OBRIGATÓRIA: se o CV tiver 6 tópicos, você deve ter 6; se tiver 8, você deve ter 8. Nunca resuma as realizações. INCLUA TODOS os detalhes técnicos específicos: equipamentos (ESDV, Control valves, PSV, flowmeters), documentos (hook-up drawings, loop diagrams, wiring diagrams), cálculos (CO2 snuffing, Static mixer), instrumentos (pressure & temperature & level transmitters), etc.
	 - **APROFUNDAMENTO DO CAMPO "PROJET"**: Para cada experiência, aprofunde o campo "projet" criando uma descrição fluida e conectada como no exemplo do concorrente. Use TODAS as informações disponíveis: empresa, setor, missões, realizações, equipamentos, tecnologias, clientes, projetos mencionados. Crie frases conectadas que contem uma história coerente, não tópicos soltos. IMPORTANTE: Use palavras e formulações DIFERENTES das utilizadas nas realizações – evite repetições. ESTILO NARRATIVO: Conte com contexto, objetivos e resultados. LIMITE: Máximo 2 frases. PROIBIÇÃO ABSOLUTA: Nunca reproduza literalmente as missões das realizações. LÍNGUA: TUDO em português, na terceira pessoa. ESTILO: Nomes de ação (Realização de..., Participação em..., etc.).
  
  2. **NOVOS CAMPOS**:
	 - **"phone"**: Extraia o número de telefone se presente no CV. Formatos: "+33 1 23 45 67 89" ou "01.23.45.67.89" ou "0123456789". Se ausente, deixe vazio.
	 - **"nom"**: Extraia o sobrenome do candidato se presente. Se ausente, deixe vazio.
	 - **"summary"**: Crie um resumo profissional conciso (máx. 2–3 linhas) que apresente o candidato, suas principais competências e experiência-chave. Seja impactante e profissional.
     - **"languages"**: Extraia todos os idiomas mencionados no CV (seção de idiomas, experiências internacionais, formações, etc.) com seus níveis do QECR. Use nomes completos em português: "Francês", "Inglês", "Alemão", "Espanhol", "Italiano", etc. Para o nível, use QECR (A1, A2, B1, B2, C1, C2) ou termos como "Nativo", "Fluente", "Intermédio", "Iniciante" se o nível não for especificado. **IMPORTANTE**: Se o nível não for mencionado, deixe o campo "level" vazio (""). Exemplos: {"language": "Francês", "level": "Nativo"}, {"language": "Inglês", "level": "B2"}, {"language": "Alemão", "level": ""}. Se nenhum idioma for mencionado, deixe um array vazio [].
      - **"certifications"**: Extraia TODAS as certificações mencionadas no CV (seções de certificações, formações, educação, experiências, etc.). Não invente – extraia apenas as explicitamente mencionadas. Exemplos: "BOSIET", "NEBOSH", "IOSH", "OSHA", "API", "ASME", "ISO 9001", "Six Sigma", "PMP", "Initial training on HV-LV Electrical Installations", "Electrical accreditation certificate according to the C18-510 standard", "National first-aid diploma", "BCTE Certificate", "Qess Certification", "ATEX Certificate", etc. Se nenhuma for mencionada, deixe [].
      - **"technical_skills"**: Extraia competências técnicas detalhadas e precisas mencionadas no CV. Não se limite a palavras-chave – extraia competências completas e detalhadas. Procure nas seções "Technical Skills", "Compétences", "Skills", experiências profissionais, etc. Se não houver, deixe [].
      - **"secteurs_activites"**: Extraia todos os setores em que o candidato trabalhou (ex.: "Petróleo e Gás", "Petroquímica", "Energia", "Automotivo", "Aeronáutico", "TI", "Finanças", "Saúde", "Telecomunicações", etc.). Identifique a partir das experiências.
     - **"domaines_expertise"**: Extraia os domínios de especialização do candidato (ex.: "Desenvolvimento Web", "Data Science", "Gestão de Projetos", "Marketing Digital", "Conceção Mecânica", "Inteligência Artificial", etc.). Baseie-se nas competências técnicas e experiências.
     - **"detail_entreprise"**: Para cada experiência, adicione uma descrição da empresa em 1–2 frases incluindo: setor, porte (startup, PME, grande grupo), especialidade, posição no mercado. Seja preciso e informativo.
  
  3. **CAMPO "poste"**:
	 - Este é o TÍTULO DO CARGO PRETENDIDO com base na análise das experiências passadas.
	 - **MÉTODO DE EXTRAÇÃO**:
	   a) Se o candidato indicar um cargo pretendido específico → use-o  
	   b) Caso contrário, analise todas as experiências e deduza o título mais representativo  
	   c) Priorize o cargo mais recente ou o que melhor reflita a evolução de carreira  
	   d) Seja preciso e profissional no título (evite termos genéricos)
	 - Exemplos: "Engenheiro de Concepção Mecânica", "Solution Architect", "Data Engineer", "Desenvolvedor Full Stack", "Gestor de Projetos", "Consultor", "Engenheiro Civil", "Product Manager"
  
  4. **DATAS E CÁLCULO DE EXPERIÊNCIA**:
     - **DATAS DAS EXPERIÊNCIAS**: SEMPRE extraia as datas de início e término de cada experiência.
     - Formato de datas: estritamente MM/YY (ex.: 02/20, 12/22). NENHUM outro formato é aceito.
     - Se a experiência estiver em andamento, use "Em curso" para date_fin (não MM/YY).
	 - **CÁLCULO DA DURAÇÃO**: Calcule automaticamente a duração entre date_debut e date_fin.
	   - Se a duração < 1 ano: exiba em meses (ex.: "6 meses", "8 meses")
	   - Se a duração ≥ 1 ano: exiba em anos (ex.: "2 anos", "1 ano e 3 meses", "3 anos")
	 - **CÁLCULO DA EXPERIÊNCIA TOTAL**: Para o campo "expérience", encontre o ano de início da experiência significativa mais antiga e calcule: 2025 – ano_de_início = anos de experiência.
	 - Exemplo: se a primeira experiência começa em 2018 → "7 anos de experiência".
  
  5. **FORMAÇÕES**:
	 - Preencha TODOS os campos: date_debut, date_fin, diplome, ecole_cursus.
	 - Seja preciso quanto ao tipo de diploma: Mestrado, Bacharelado, BTS, Diploma de Engenheiro, MBA, etc.
  
  6. **SOFTWARES NAS EXPERIÊNCIAS**:
	 - Extraia TODOS os softwares/ferramentas mencionados em cada experiência.
	 - Deduzia-os pelo contexto quando necessário (ex.: se "desenvolvimento web" → acrescente HTML, CSS, JavaScript).
  
  7. **COMPLETUDE**:
	 - Não deixe NENHUMA experiência de fora – mesmo estágios curtos, missões pontuais, projetos.
	 - Não deixe NENHUMA formação de fora.
	 - Analise TODO o conteúdo do CV.
	 - **VERIFICAÇÃO**: Conte o número de experiências mencionadas no CV e garanta ter extraído o mesmo número.
	 - **REALIZAÇÕES COMPLETAS**: Para cada experiência, extraia TODAS as missões/realizações mencionadas no CV, mesmo que muito numerosas. CORRESPONDÊNCIA EXATA OBRIGATÓRIA: 6 tópicos no CV → 6 tópicos aqui; 8 → 8. Nunca resuma o conteúdo. INCLUA TODOS os detalhes técnicos: equipamentos específicos (ESDV, Control valves, PSV, flowmeters, etc.), documentos técnicos (hook-up drawings, loop diagrams, wiring diagrams, junction boxes, equipments layouts, I/O list, Instrument list, Alarm and Set point list), cálculos específicos (CO2 snuffing, Static mixer, etc.), tipos de instrumentos (pressure & temperature & level transmitters, etc.). Reproduza TODOS os detalhes técnicos do CV original.
  
  NB: Não exiba o tipo de contrato (ex.: Estágio, Alternância, Contrato Permanente, Contrato a Termo...) nas experiências.

  **REGRA CRÍTICA PARA TERMOS DE ESTÁGIO/ALTERNÂNCIA**:
	- SEMPRE remova as palavras "alternant", "stagiaire", "stage", "apprenti", "apprentissage" dos títulos de cargos (campo "poste" principal e campo "poste" nas experiências).
	- Mantenha TODAS as experiências (estágios, alternâncias, etc.), mas trate-as como experiências profissionais normais.
	- Reformule os títulos para que sejam profissionais, sem mencionar o estatuto.
	- **EXEMPLOS DE TRANSFORMAÇÃO OBRIGATÓRIA**:
  * "Stagiaire Développeur" → "Développeur"
  * "Alternant Ingénieur" → "Ingénieur" 
  * "Stagiaire de Recherche" → "Chercheur"
  * "Apprenti Data Analyst" → "Data Analyst"
  * "Stage Marketing" → "Marketing"
  * "Alternance Commercial" → "Commercial"
  
  Acrescente o máximo de informações possível analisando o CV e deduzindo elementos que não estejam necessariamente presentes, como faria um especialista de RH.
  
  A saída deve respeitar EXATAMENTE o modelo acima. Se uma informação não estiver presente e você não puder estimá-la, deixe o campo vazio (string vazia "").
  
  Aqui está o JSON de extração para analisar:
  
  ` + nuextractJSON + `
  
  **ÚLTIMA INSTRUÇÃO CRÍTICA**: 
  - Extraia TODAS as experiências e TODAS as suas realizações sem exceção
  - Nunca resuma o conteúdo, mesmo que seja muito longo
  - Garanta que cada experiência tenha todas as suas missões/realizações listadas
  - Se uma experiência tiver muitos detalhes no CV, reproduza TODOS esses detalhes
  - **CORRESPONDÊNCIA EXATA OBRIGATÓRIA**: Se o CV tiver 6 tópicos, você deve ter 6 tópicos. Se tiver 8, você deve ter 8. NUNCA resuma – reproduza TODOS os tópicos do CV original
  - **DETALHES TÉCNICOS OBRIGATÓRIOS**: Inclua TODOS os detalhes técnicos mencionados: equipamentos específicos (ESDV, Control valves, PSV, flowmeters, etc.), documentos técnicos (hook-up drawings, loop diagrams, wiring diagrams, junction boxes, equipments layouts, I/O list, Instrument list, Alarm and Set point list), cálculos específicos (CO2 snuffing, Static mixer, etc.), tipos de instrumentos (pressure & temperature & level transmitters, etc.)
  - **DISTINÇÃO PROJETO/REALIZAÇÕES**: O campo "projet" deve ser um parágrafo narrativo fluido que conte a história do projeto (máximo 2 frases). O campo "réalisations" deve ser uma lista detalhada de missões. NUNCA repita entre os dois campos – use palavras diferentes. PROIBIÇÃO ABSOLUTA: Nunca reproduza literalmente as missões das realizações no projeto. LÍNGUA: TUDO em português, na terceira pessoa. ESTILO: Use nomes de ação (Realização de..., Participação em..., Coordenação de..., Redação de..., Supervisão de..., Comissionamento de..., etc.) em vez de 'Ele/ela fez...'.
  - **OBJETIVO**: Documento o mais completo possível, sem resumo – reproduza TODOS os detalhes técnicos do CV original
  
  Responda APENAS com o JSON estruturado, sem texto antes ou depois.`
}

// GetExtractionPromptProductionSpanish devuelve el prompt de producción en español para la extracción
func GetExtractionPromptProductionSpanish(nuextractJSON string) string {
	return `Eres un experto de RR. HH. especializado en análisis de CV. Quiero que analices el diccionario JSON de extracción de CV que te proporciono como input y que extraigas TODA la información pertinente en forma de un diccionario estructurado, que pueda guardarse en JSON, según el siguiente modelo:
  
  NO CAMBIES BAJO NINGÚN CONCEPTO LAS CLAVES DE ESTE DICCIONARIO, YA QUE SE UTILIZARÁ MÁS ADELANTE.
  
  {
	"prenom": "NOMBRE DE PILA DEL CANDIDATO - OBLIGATORIO. Extrae el nombre de pila del candidato desde el CV. Este campo debe SIEMPRE estar cumplimentado salvo casos excepcionales en los que el nombre realmente no esté mencionado en el CV. Busca en el encabezado, la firma o cualquier mención del nombre completo del candidato.",
	"nom": "Apellido del candidato. Si no está explícitamente presente en el CV, deja este campo vacío (\"\"). No lo inventes.",
	"email": "Dirección de email del candidato. Si no está explícitamente presente en el CV, deja este campo vacío (\"\"). No lo inventes.",
	"phone": "Número de teléfono del candidato. Si no está explícitamente presente en el CV, deja este campo vacío (\"\"). No lo inventes.",
    "summary": "Resumen profesional en 2-3 líneas como máximo que presente al candidato, sus competencias clave y su experiencia principal. Sé conciso pero impactante para dar una visión global del perfil. ESTILO: Usa un estilo profesional como 'Más de X años de experiencia en [área], incluyendo [actividades específicas] para [industrias]. Experiencia en [entornos] con un historial probado de colaboración con [tipos de empresas].'",
	"age": "Si la edad no está explícitamente mencionada en el CV, deja este campo vacío (\"\"). No la estimes.",
	"poste": "TÍTULO DEL PUESTO BUSCADO - Analiza las experiencias pasadas y deduce el título de puesto más apropiado, siendo preciso si tiene un dominio de actividad específico. Si el candidato busca un puesto concreto, utilízalo. De lo contrario, dedúcelo a partir del puesto más reciente o más representativo de su perfil. Ejemplos: 'Ingeniero de Diseño Mecánico', 'Solution Architect', 'Data Engineer', 'Jefe de Proyecto', 'Desarrollador Full Stack'",
	"diplome": "Formación principal (nombre de la escuela de ingeniería, de negocios o del M2)",
	"expérience": "Calcula la experiencia total en años: encuentra la fecha de inicio de la experiencia más antigua y réstala del año actual (2025). Si no hay ninguna fecha disponible, deja vacío.",
	"mobilité": "Ubicación geográfica deseada si se especifica.",
	"disponibilité": "",
	"permis_B": "",
    "hobbies": ["Lista de intereses"],
    "languages": [
      {
        "language": "Nombre del idioma (p. ej.: Francés, Inglés, Alemán, Español, etc.)",
        "level": "Nivel MCER (p. ej.: A1, A2, B1, B2, C1, C2, Nativo, Fluido, Intermedio, Principiante, etc.)"
      }
    ],
    "certifications": ["Lista de certificaciones mencionadas en el CV (p. ej.: BOSIET, NEBOSH, IOSH, OSHA, API, ASME, ISO 9001, Six Sigma, PMP, Initial training on HV-LV Electrical Installations, Electrical accreditation certificate according to the C18-510 standard, National first-aid diploma, BCTE Certificate, Qess Certification, ATEX Certificate, etc.). No inventes: extrae únicamente las mencionadas explícitamente."],
    "technical_skills": ["Lista de competencias técnicas detalladas y precisas mencionadas en el CV. Extrae competencias técnicas completas, no solo palabras clave. Ejemplos: 'Construction, QA/QC, Pre-Commissioning, Commissioning, hook-up, Start-up, Maintenance', 'Smart Relays & Soft Starters', 'Instrument calibrations and loop checks', 'Knowledge of PCS, PDCS & DCS system operation', 'Primary and secondary earthing system, support & cable tray installation, instrument and junction boxes installation, cable pulling dressing & termination, Circuits Control', 'HV/MV/LV, firefighting, CCTV & security and access control systems', 'Check list A&B, static testing (Pressure testing, resistance checks), calibration, loop check, cause and effects (loop function check and multilevel trip sequence)', 'Validations, documents reviews, and investigations (Documents : URS, FAT, FDS, CSV, IQ, OQ, and PQ)', 'Good experience in oil and gas industry from wellhead installations to storage and transfer', 'Ability to apply safety rules – hazardous areas - explosive atmosphere (LEL), Indices of protection - protection methods using on hazardous area', 'Ability to understand and use work permits system to control/coordinate and communicate all tasks', 'Ability to do risk assessments before starting work', etc."],
    "competence_fonctionnelle": ["Lista de competencias funcionales mencionadas en el CV. Estas competencias pueden estar presentes en las experiencias, soft skills o en el resumen. Extrae competencias funcionales como: 'Planificación, coordinación', 'Gestión de proyectos', 'Evaluación y comparación de tecnologías', 'Informes y análisis, comunicación interna', 'Toma de decisiones y recomendación estratégica', 'Gestión autónoma de proyectos', 'Análisis funcional y técnico', 'Vigilancia tecnológica', 'Interacciones con partes interesadas técnicas', 'Comparación y selección de soluciones técnicas', 'Síntesis e informes', 'Dominio de herramientas de mejora continua (Lean, Six Sigma)', 'Optimización de la producción', 'Eficiencia industrial', 'Estudio de no conformidades, diagnóstico y resolución de problemas', 'Autonomía e iniciativa', 'Comunicación y colaboración interna', 'Adaptabilidad del discurso, gestión del cambio', 'Espíritu de análisis y síntesis', 'Liderazgo de equipo', 'Gestión presupuestaria', 'Negociación comercial', 'Formación y desarrollo de equipo', 'Estrategia y visión', 'Innovación y creatividad', 'Relación con clientes', 'Gestión de la calidad', 'Organización y planificación', 'Resolución de conflictos', etc. Si no se menciona ninguna competencia funcional, deja un array vacío []."],
     "secteurs_activites": ["Lista de sectores de actividad en los que ha trabajado el candidato (p. ej.: Petróleo y Gas, Petroquímica, Energía, Automoción, Aeroespacial, Informática, Finanzas, Salud, etc.)."],
    "domaines_expertise": ["Lista de dominios de especialización del candidato (p. ej.: Desarrollo Web, Ciencia de Datos, Gestión de Proyectos, Marketing Digital, etc.)."],
	"formations": [
	  {
		"date_debut": "OBLIGATORIO - Año de inicio (p. ej.: 2020, 2018-2019)",
		"date_fin": "OBLIGATORIO - Año de fin (p. ej.: 2022, 2020-2021)",
		"diplome": "OBLIGATORIO - Tipo de título preciso (p. ej.: Máster en Ingeniería Mecánica, Título de Ingeniero, Grado en Informática, BTS Comercio, Título de Medicina, MBA, etc.)",
		"ecole_cursus": "OBLIGATORIO - Nombre completo de la escuela/universidad (p. ej.: École Centrale Paris, Université Pierre et Marie Curie, HEC Paris, etc.)"
	  }
	],
    "expériences": [
	  {
        "date_debut": "OBLIGATORIO - Fecha de inicio en formato MM/AA (p. ej.: 02/20, 09/19)",
        "date_fin": "OBLIGATORIO - Fecha de fin en formato MM/AA (p. ej.: 12/22, 08/20). Si la experiencia está en curso, utiliza 'En curso'",
		"entreprise": "OBLIGATORIO - Nombre de la empresa",
		"detail_entreprise": "OBLIGATORIO - Descripción de la empresa en 1-2 frases: sector de actividad, tamaño, especialidad, posición en el mercado. Ejemplo: 'Startup especializada en inteligencia artificial y machine learning, con 50 empleados y referente en análisis predictivo para el sector bancario'",
		"durée": "OBLIGATORIO - Duración calculada automáticamente (p. ej.: 2 años, 6 meses, 1 año y 3 meses). Si es inferior a 1 año, muestra en meses. Si es mayor o igual a 1 año, muestra en años.",
		"poste": "OBLIGATORIO - Título del puesto desempeñado (p. ej.: Ingeniero de Diseño, Desarrollador Senior, Jefe de Proyecto, etc.)",
		"secteur": "Sector de actividad de la experiencia en 1-2 palabras. Extrae el sector desde la información de la empresa, del proyecto o del contexto. Ejemplos: Agroalimentario, Automoción, Banca, Construcción, Biomédico, Química, Consultoría, Defensa, Energía, Medio ambiente, Ferroviario, Gran distribución, Infraestructura, Logística, Metalurgia / Siderurgia, Naval, Nuclear, Petróleo y Gas, Petroquímica, Farmacéutico, Salud, Sector público, Telecomunicaciones, IRVE, Fotovoltaico, Tratamiento de aguas, Revalorización energética, Hidroeléctrica, Energías Renovables (ENR), Energía eólica, Biogás, Educación, Recursos Humanos. Si el sector no es identificable, deja este campo vacío (\"\").",
		"contexte": "Resume la experiencia de forma sucinta para presentar el proyecto realizado en una frase.",
		"projet": "AMPLÍA este campo creando una descripción fluida y conectada del proyecto/experiencia, como en el ejemplo del competidor. Utiliza TODA la información disponible: empresa, sector, misiones, logros, equipos, tecnologías, clientes, proyectos mencionados. Crea frases enlazadas que cuenten una historia coherente, no viñetas separadas. IMPORTANTE: Utiliza palabras y formulaciones DIFERENTES a las usadas en los logros; evita la repetición de los mismos términos. ESTILO NARRATIVO: Cuenta el proyecto como una historia con contexto, objetivos y resultados. LÍMITE: Máximo 2 frases para mantener la fluidez. PROHIBICIÓN ABSOLUTA: JAMÁS reproduzcas palabra por palabra las misiones de los logros. El proyecto debe contar la historia del proyecto, no listar tareas. IDIOMA: TODO en español, en tercera persona. ESTILO: Utiliza sustantivos de acción (Realización de..., Participación en..., Coordinación de..., Redacción de..., Supervisión de..., Puesta en marcha de..., etc.) en lugar de 'Él/ella hizo...'. Ejemplo de estilo del competidor: 'Gestión de proyectos de electrólisis alcalina de alta presión y desarrollo de nuevos prototipos para la producción de hidrógeno. Este proyecto incluye la selección y especificación de equipos, la redacción de documentación técnica detallada, el seguimiento de pedidos y la conformidad técnica de las instalaciones.'",
		"projets_name": ["Lista de nombres de proyectos mencionados para esta experiencia, si están presentes. Extrae nombres específicos de proyectos mencionados en el contexto, logros o en otro lugar en la descripción de la experiencia. A menudo hay un solo proyecto por experiencia, pero puede haber varios. Ejemplos: 'Proyecto de rehabilitación de edificio de oficinas para WEWORK', 'Proyecto Reine des Neiges', 'Proyecto de rehabilitación de edificio de oficinas para AXA - Boulevard des Italiens', etc. Si no se menciona ningún nombre de proyecto, deja un array vacío []."],
		"result": "Frase corta que explica el resultado de la experiencia. Extrae o deduce una frase concisa que resuma los resultados obtenidos o los logros clave de esta experiencia. Ejemplos: 'Recomendaciones e identificación de Next Step', 'Alquiler de tecnología en laboratorios de investigación para pruebas', 'Puesta en marcha del proyecto y del equipo a cargo del proyecto, identificación de causas y método de evaluación, planificación de pruebas', etc. Si no se menciona ningún resultado, deja este campo vacío (\"\").",
		"logiciels": ["OBLIGATORIO - Extrae TODAS las herramientas/softwares mencionados en esta experiencia (p. ej.: SolidWorks, Python, React, AWS, Docker, etc.) - incluso si no están listados explícitamente, dedúcelos por contexto"],
		"réalisations": [
		  "Lista TODAS las misiones/logros mencionados en el CV para esta experiencia. CORRESPONDENCIA EXACTA OBLIGATORIA: si el CV tiene 6 viñetas, debes tener 6 viñetas. Si el CV tiene 8 viñetas, debes tener 8 viñetas. No trunques NUNCA: lista todo lo que está escrito en el CV, aunque sea muy detallado. INCLUYE TODOS los detalles técnicos: equipos específicos (ESDV, válvulas de control, PSV, caudalímetros, etc.), documentos técnicos (hook-up drawings, loop diagrams, wiring diagrams, etc.), cálculos específicos (CO2 snuffing, static mixer, etc.), tipos de instrumentos (pressure & temperature & level transmitters, etc.). Reproduce TODOS los detalles técnicos mencionados en el CV original. IDIOMA: TODO en español, en tercera persona (él/ella, el/la ingeniero/a, etc.)."
		],
		"AI_suggest": ["Si puedes deducir elementos relevantes no presentes en el CV. Las sugerencias deben ser específicas y adaptadas a cada experiencia, relevantes para los reclutadores; su número debe variar según las experiencias, sin redundancias entre ellas. No las incluyas sistemáticamente: debe parecer natural."]
	  }
	],
	"logiciels": [
	  {
		"logiciel": "",
		"level": "Estima el nivel entre: Principiante, Intermedio, Avanzado, Experto.",
		"temps_utilisation": "Estima el tiempo de uso en meses."
	  }
	]
  }
  
  INSTRUCCIONES CRÍTICAS:
  
  1. **EXTRACCIONES OBLIGATORIAS**:
	 - Extrae TODAS las experiencias profesionales (prácticas, contratos indefinidos, temporales, alternancia, etc.) - NO OLVIDES NI UNA SOLA
	 - Para cada formación: date_debut, date_fin, diplome Y ecole_cursus son OBLIGATORIOS
	 - Para cada experiencia: date_debut, date_fin, entreprise, detail_entreprise, durée, poste Y logiciels son OBLIGATORIOS
	 - **CRÍTICO**: Relee el CV varias veces para asegurarte de haber extraído TODAS las experiencias mencionadas
	 - **EXHAUSTIVIDAD DE LOS LOGROS**: Para cada experiencia, lista TODAS las misiones/logros mencionados en el CV, aunque sean numerosos. CORRESPONDENCIA EXACTA OBLIGATORIA: si el CV tiene 6 viñetas, debes tener 6. Si tiene 8, debes tener 8. No trunques NUNCA los logros. INCLUYE TODOS los detalles técnicos específicos: equipos (ESDV, válvulas de control, PSV, caudalímetros), documentos (hook-up drawings, loop diagrams, wiring diagrams), cálculos (CO2 snuffing, static mixer), instrumentos (pressure & temperature & level transmitters), etc.
	 - **AMPLIACIÓN DEL CAMPO "PROJET"**: Para cada experiencia, amplía el campo "projet" creando una descripción fluida y conectada como en el ejemplo del competidor. Utiliza TODA la información disponible: empresa, sector, misiones, logros, equipos, tecnologías, clientes, proyectos mencionados. Crea frases que cuenten una historia coherente, no listas separadas. IMPORTANTE: Usa palabras y formulaciones DIFERENTES a las usadas en los logros; evita la repetición. ESTILO NARRATIVO: Cuenta el proyecto con contexto, objetivos y resultados. LÍMITE: Máximo 2 frases. PROHIBICIÓN ABSOLUTA: JAMÁS reproduzcas palabra por palabra las misiones de los logros. El proyecto no debe listar tareas. IDIOMA: TODO en español, en tercera persona. ESTILO: Usa sustantivos de acción (Réalisation de..., Participation à..., etc.). 
  
  2. **NUEVOS CAMPOS**:
	 - **"phone"**: Extrae el número de teléfono si está presente en el CV. Formato: "+33 1 23 45 67 89" o "01.23.45.67.89" o "0123456789". Si está ausente, deja vacío.
	 - **"nom"**: Extrae el apellido del candidato si está presente en el CV. Si está ausente, deja vacío.
	 - **"summary"**: Crea un resumen profesional conciso (2-3 líneas máx.) que presente al candidato, sus competencias principales y su experiencia clave. Sé impactante y profesional.
     - **"languages"**: Extrae todos los idiomas mencionados en el CV (sección de idiomas, experiencias internacionales, formaciones, etc.) con su nivel MCER. Usa los nombres completos en español: "Francés", "Inglés", "Alemán", "Español", "Italiano", etc. Para el nivel, usa los niveles del MCER (A1, A2, B1, B2, C1, C2) o términos como "Nativo", "Fluido", "Intermedio", "Principiante" si no se especifica el nivel. **IMPORTANTE**: Si no se menciona el nivel en el CV, deja el campo "level" vacío (""). Ejemplos: {"language": "Francés", "level": "Nativo"}, {"language": "Inglés", "level": "B2"}, {"language": "Alemán", "level": ""}. Si no se menciona ningún idioma, deja un array vacío [].
      - **"certifications"**: Extrae TODAS las certificaciones mencionadas en el CV (secciones de certificaciones, formaciones, educación, experiencias, etc.). No inventes: extrae solo las explícitamente mencionadas. Ejemplos: "BOSIET", "NEBOSH", "IOSH", "OSHA", "API", "ASME", "ISO 9001", "Six Sigma", "PMP", "Initial training on HV-LV Electrical Installations", "Electrical accreditation certificate according to the C18-510 standard", "National first-aid diploma", "BCTE Certificate", "Qess Certification", "ATEX Certificate", etc. Si no se menciona ninguna certificación, deja un array vacío [].
      - **"technical_skills"**: Extrae las competencias técnicas detalladas y precisas mencionadas en el CV. No te limites a palabras clave: extrae competencias completas y detalladas. Busca en las secciones "Technical Skills", "Compétences", "Skills", experiencias profesionales, etc. Ejemplos: "Construction, QA/QC, Pre-Commissioning, Commissioning, hook-up, Start-up, Maintenance", "Smart Relays & Soft Starters", ... Si no se menciona ninguna competencia técnica, deja un array vacío [].
      - **"secteurs_activites"**: Extrae todos los sectores de actividad en los que el candidato ha trabajado (p. ej.: "Petróleo y Gas", "Petroquímica", "Energía", "Automoción", "Aeroespacial", "Informática", "Finanzas", "Salud", "Telecomunicaciones", etc.). Analiza las experiencias profesionales para identificarlos.
     - **"domaines_expertise"**: Extrae los dominios de competencia de especialización del candidato (p. ej.: "Desarrollo Web", "Ciencia de Datos", "Gestión de Proyectos", "Marketing Digital", "Diseño Mecánico", "Inteligencia Artificial", etc.). Basándote en las competencias técnicas y las experiencias.
     - **"detail_entreprise"**: Para cada experiencia, añade una descripción de la empresa en 1-2 frases que incluya: sector de actividad, tamaño (startup, pyme, gran grupo), especialidad, posición en el mercado. Sé preciso e informativo.
  
  3. **CAMPO "poste"**:
	 - Es el TÍTULO DEL PUESTO BUSCADO basado en el análisis de las experiencias pasadas
	 - **MÉTODO DE EXTRACCIÓN**:
	   a) Si el candidato indica un puesto específico buscado → utilízalo
	   b) En caso contrario, analiza todas las experiencias y deduce el título más representativo
	   c) Prioriza el puesto más reciente o el que mejor refleje la evolución de su carrera
	   d) Sé preciso y profesional en el título (evita términos genéricos)
	 - Ejemplos: "Ingeniero de Diseño Mecánico", "Solution Architect", "Data Engineer", "Desarrollador Full Stack", "Jefe de Proyecto", "Consultor", "Ingeniero Civil", "Product Manager"
  
  4. **FECHAS Y CÁLCULO DE EXPERIENCIA**:
     - **FECHAS DE EXPERIENCIAS**: Extrae SIEMPRE las fechas de inicio y fin de cada experiencia
     - Formato de fechas: estrictamente MM/AA (p. ej.: 02/20, 12/22). NO se acepta otro formato
     - Si la experiencia está en curso, usa "En curso" para date_fin (sin MM/AA)
	 - **CÁLCULO DE LA DURACIÓN**: Calcula automáticamente la duración entre date_debut y date_fin
	   - Si duración < 1 año: muestra en meses (p. ej.: "6 meses", "8 meses")
	   - Si duración ≥ 1 año: muestra en años (p. ej.: "2 años", "1 año y 3 meses", "3 años")
	 - **CÁLCULO DE EXPERIENCIA TOTAL**: Para el campo "expérience", encuentra la fecha de inicio de la experiencia significativa más antigua y calcula: 2025 - año_de_inicio = años de experiencia
	 - Ejemplo: si la primera experiencia comienza en 2018 → "7 años de experiencia"
  
  5. **FORMACIONES**:
	 - Rellena TODOS los campos: date_debut, date_fin, diplome, ecole_cursus
	 - Sé preciso sobre el tipo de título: Máster, Grado/Bachelor, BTS, Título de Ingeniero, MBA, etc.
  
  6. **SOFTWARES EN LAS EXPERIENCIAS**:
	 - Extrae TODOS los softwares/herramientas mencionados en cada experiencia
	 - Dedúcelos por contexto si es necesario (p. ej.: si dice "desarrollo web" → añade HTML, CSS, JavaScript)
  
  7. **COMPLETITUD**:
	 - No dejes NINGUNA experiencia fuera: incluso prácticas cortas, misiones puntuales, proyectos
	 - No dejes NINGUNA formación fuera
	 - Analiza TODO el contenido del CV
	 - **VERIFICACIÓN**: Cuenta el número de experiencias mencionadas en el CV y asegúrate de haber extraído el mismo número
	 - **LOGROS COMPLETOS**: Para cada experiencia, extrae TODAS las misiones/logros mencionados en el CV, aunque sean muy numerosos. CORRESPONDENCIA EXACTA OBLIGATORIA: si el CV tiene 6 viñetas, debes tener 6. Si tiene 8, debes tener 8. No trunques NUNCA el contenido. INCLUYE TODOS los detalles técnicos: equipos específicos (ESDV, válvulas de control, PSV, caudalímetros, etc.), documentos técnicos (hook-up drawings, loop diagrams, wiring diagrams, junction boxes, equipments layouts, I/O list, Instrument list, Alarm and Set point list), cálculos específicos (CO2 snuffing, static mixer, etc.), tipos de instrumentos (pressure & temperature & level transmitters, etc.). Reproduce TODOS los detalles técnicos del CV original.
  
  Nota: No muestres el tipo de contrato (ej.: Prácticas, Alternancia, Contrato indefinido, Contrato temporal...) en las experiencias.

  **REGLA CRÍTICA PARA TÉRMINOS DE PRÁCTICAS/ALTERNANCIA**:
	- Elimina SIEMPRE las palabras "alternant", "stagiaire", "stage", "apprenti", "apprentissage" de los títulos de puestos (campo "poste" principal y campo "poste" de las experiencias)
	- Conserva TODAS las experiencias (prácticas, alternancias, etc.) pero trátalas como experiencias profesionales normales
	- Reformula los títulos para que sean profesionales sin mencionar el estatus
	- **EJEMPLOS DE TRANSFORMACIÓN OBLIGATORIA**:
  * "Stagiaire Développeur" → "Développeur"
  * "Alternant Ingénieur" → "Ingénieur" 
  * "Stagiaire de Recherche" → "Chercheur"
  * "Apprenti Data Analyst" → "Data Analyst"
  * "Stage Marketing" → "Marketing"
  * "Alternance Commercial" → "Commercial"
  
  Añade tanta información como sea posible analizando el CV y deduciendo elementos que no estén necesariamente presentes, como lo haría un experto de RR. HH.
  
  El output debe respetar EXACTAMENTE el modelo anterior. Si una información no está presente y no puedes estimarla, deja el campo vacío (cadena vacía "").
  
  Aquí está el JSON de extracción a analizar:
  
  ` + nuextractJSON + `
  
  **ÚLTIMA INSTRUCCIÓN CRÍTICA**: 
  - Extrae TODAS las experiencias y TODOS sus logros sin excepción
  - No trunques NUNCA el contenido, aunque sea muy largo
  - Asegúrate de que cada experiencia tenga todas sus misiones/logros listados
  - Si una experiencia tiene muchos detalles en el CV, reproduce TODOS esos detalles
  - **CORRESPONDENCIA EXACTA OBLIGATORIA**: Si el CV tiene 6 viñetas, debes tener 6. Si tiene 8, debes tener 8. No resumas NUNCA: reproduce TODAS las viñetas del CV original
  - **DETALLES TÉCNICOS OBLIGATORIOS**: Incluye TODOS los detalles técnicos mencionados: equipos específicos (ESDV, válvulas de control, PSV, caudalímetros, etc.), documentos técnicos (hook-up drawings, loop diagrams, wiring diagrams, junction boxes, equipments layouts, I/O list, Instrument list, Alarm and Set point list), cálculos específicos (CO2 snuffing, static mixer, etc.), tipos de instrumentos (pressure & temperature & level transmitters, etc.)
  - **DISTINCIÓN PROYECTO/LOGROS**: El campo "projet" debe ser un párrafo narrativo fluido que cuente la historia del proyecto (máximo 2 frases). El campo "réalisations" debe ser una lista detallada de misiones. JAMÁS repitas entre ambos campos: utiliza palabras diferentes. PROHIBICIÓN ABSOLUTA: No reproduzcas JAMÁS palabra por palabra las misiones de los logros en el proyecto. IDIOMA: TODO en español, en tercera persona. ESTILO: Usa sustantivos de acción (Realización de..., Participación en..., Coordinación de..., Redacción de..., Supervisión de..., Puesta en marcha de..., etc.) en lugar de 'Él/ella hizo...'.
  - **OBJETIVO**: Documento lo más completo posible, sin resumen: reproduce TODOS los detalles técnicos del CV original
  
  Responde ÚNICAMENTE con el JSON estructurado, sin texto antes o después.
  `
}

// GetExtractionPromptProductionItalian restituisce il prompt di produzione in italiano per l'estrazione
func GetExtractionPromptProductionItalian(nuextractJSON string) string {
	return `Sei un esperto HR specializzato nell'analisi di CV. Desidero che tu analizzi il dizionario JSON dell'estrazione del CV che ti fornisco in input e che estragga TUTTE le informazioni pertinenti sotto forma di dizionario strutturato, salvabile in JSON, secondo il seguente modello:
  
  NON CAMBIARE ASSOLUTAMENTE LE CHIAVI DI QUESTO DIZIONARIO, POICHÉ SARANNO UTILIZZATE SUCCESSIVAMENTE.
  
  {
	"prenom": "NOME DI BATTESIMO DEL CANDIDATO - OBBLIGATORIO. Estrai il nome del candidato dal CV. Questo campo deve SEMPRE essere compilato salvo casi eccezionali in cui il nome non sia realmente menzionato nel CV. Cerca nell’intestazione, nella firma o in qualsiasi riferimento al nome completo del candidato.",
	"nom": "Cognome del candidato. Se non è esplicitamente presente nel CV, lascia questo campo vuoto (\"\"). Non inventarlo.",
	"email": "Indirizzo email del candidato. Se non è esplicitamente presente nel CV, lascia questo campo vuoto (\"\"). Non inventarlo.",
	"phone": "Numero di telefono del candidato. Se non è esplicitamente presente nel CV, lascia questo campo vuoto (\"\"). Non inventarlo.",
    "summary": "Sintesi professionale in massimo 2-3 righe che presenti il candidato, le competenze chiave e l’esperienza principale. Sii conciso ma incisivo per offrire una visione d’insieme del profilo. STILE: Usa uno stile professionale come 'Oltre X anni di esperienza in [ambito], includendo [attività specifiche] per [settori]. Esperienza in [ambienti] con un comprovato track record di collaborazione con [tipi di aziende].'",
	"age": "Se l’età non è esplicitamente menzionata nel CV, lascia questo campo vuoto (\"\"). Non stimarla.",
	"poste": "TITOLO DEL RUOLO RICERCATO - Analizza le esperienze pregresse e deduci il titolo più appropriato, essendo preciso se esiste un dominio di attività specifico. Se il candidato cerca un ruolo specifico, utilizza quello. Altrimenti, deducilo dalla posizione più recente o più rappresentativa del suo profilo. Esempi: 'Ingegnere Progettazione Meccanica', 'Solution Architect', 'Data Engineer', 'Project Manager', 'Sviluppatore Full Stack'",
	"diplome": "Formazione principale (nome della scuola di ingegneria, di business o del M2)",
	"expérience": "Calcola l’esperienza totale in anni: trova la data di inizio dell’esperienza più antica e sottraila dall’anno corrente (2025). Se non è disponibile alcuna data, lascia vuoto.",
	"mobilité": "Posizione geografica desiderata se specificata.",
	"disponibilité": "",
	"permis_B": "",
    "hobbies": ["Elenco degli interessi"],
    "languages": [
      {
        "language": "Nome della lingua (es.: Francese, Inglese, Tedesco, Spagnolo, ecc.)",
        "level": "Livello QCER (es.: A1, A2, B1, B2, C1, C2, Madrelingua, Fluente, Intermedio, Principiante, ecc.)"
      }
    ],
    "certifications": ["Elenco delle certificazioni menzionate nel CV (es.: BOSIET, NEBOSH, IOSH, OSHA, API, ASME, ISO 9001, Six Sigma, PMP, Initial training on HV-LV Electrical Installations, Electrical accreditation certificate according to the C18-510 standard, National first-aid diploma, BCTE Certificate, Qess Certification, ATEX Certificate, ecc.). Non inventare - estrarre solo quelle esplicitamente menzionate."],
    "technical_skills": ["Elenco delle competenze tecniche dettagliate e precise menzionate nel CV. Estrarre competenze tecniche complete, non solo parole chiave. Esempi: 'Construction, QA/QC, Pre-Commissioning, Commissioning, hook-up, Start-up, Maintenance', 'Smart Relays & Soft Starters', 'Instrument calibrations and loop checks', 'Knowledge of PCS, PDCS & DCS system operation', 'Primary and secondary earthing system, support & cable tray installation, instrument and junction boxes installation, cable pulling dressing & termination, Circuits Control', 'HV/MV/LV, firefighting, CCTV & security and access control systems', 'Check list A&B, static testing (Pressure testing, resistance checks), calibration, loop check, cause and effects (loop function check and multilevel trip sequence)', 'Validations, documents reviews, and investigations (Documents : URS, FAT, FDS, CSV, IQ, OQ, and PQ)', 'Good experience in oil and gas industry from wellhead installations to storage and transfer', 'Ability to apply safety rules – hazardous areas - explosive atmosphere (LEL), Indices of protection - protection methods using on hazardous area', 'Ability to understand and use work permits system to control/coordinate and communicate all tasks', 'Ability to do risk assessments before starting work', ecc."],
    "competence_fonctionnelle": ["Elenco delle competenze funzionali menzionate nel CV. Queste competenze possono essere presenti nelle esperienze, soft skills o nel riassunto. Estrai competenze funzionali come: 'Pianificazione, coordinamento', 'Gestione progetti', 'Valutazione e confronto tecnologie', 'Report e analisi, comunicazione interna', 'Decision-making e raccomandazione strategica', 'Gestione autonoma progetti', 'Analisi funzionale e tecnica', 'Technology watch', 'Interazioni con stakeholder tecnici', 'Confronto e selezione soluzioni tecniche', 'Sintesi e report', 'Padronanza strumenti miglioramento continuo (Lean, Six Sigma)', 'Ottimizzazione produzione', 'Efficienza industriale', 'Studio non conformità, diagnosi e risoluzione problemi', 'Autonomia e iniziativa', 'Comunicazione e collaborazione interna', 'Adattabilità del discorso, gestione del cambiamento', 'Spirito di analisi e sintesi', 'Leadership di team', 'Gestione budget', 'Negoziazione commerciale', 'Formazione e sviluppo team', 'Strategia e visione', 'Innovazione e creatività', 'Relazione clienti', 'Gestione qualità', 'Organizzazione e pianificazione', 'Risoluzione conflitti', ecc. Se non viene menzionata alcuna competenza funzionale, lascia un array vuoto []."],
     "secteurs_activites": ["Elenco dei settori di attività in cui il candidato ha lavorato (es.: Petrolio e Gas, Petrolchimico, Energia, Automotive, Aerospazio, Informatica, Finanza, Sanità, ecc.)."],
    "domaines_expertise": ["Elenco delle aree di competenza di specializzazione del candidato (es.: Sviluppo Web, Data Science, Project Management, Marketing Digitale, ecc.)."],
	"formations": [
	  {
		"date_debut": "OBBLIGATORIO - Anno di inizio (es.: 2020, 2018-2019)",
		"date_fin": "OBBLIGATORIO - Anno di fine (es.: 2022, 2020-2021)",
		"diplome": "OBBLIGATORIO - Tipo di titolo preciso (es.: Master in Ingegneria Meccanica, Laurea in Ingegneria, Bachelor in Informatica, BTS Commercio, Laurea in Medicina, MBA, ecc.)",
		"ecole_cursus": "OBBLIGATORIO - Nome completo della scuola/università (es.: École Centrale Paris, Université Pierre et Marie Curie, HEC Paris, ecc.)"
	  }
	],
    "expériences": [
	  {
        "date_debut": "OBBLIGATORIO - Data di inizio in formato MM/YY (es.: 02/20, 09/19)",
        "date_fin": "OBBLIGATORIO - Data di fine in formato MM/YY (es.: 12/22, 08/20). Se l’esperienza è in corso, usa 'In corso'",
		"entreprise": "OBBLIGATORIO - Nome dell’azienda",
		"detail_entreprise": "OBBLIGATORIO - Descrizione dell’azienda in 1-2 frasi: settore di attività, dimensione, specializzazione, posizione di mercato. Esempio: 'Startup specializzata in intelligenza artificiale e machine learning, con 50 dipendenti e leader nell’analisi predittiva per il settore bancario'",
		"durée": "OBBLIGATORIO - Durata calcolata automaticamente (es.: 2 anni, 6 mesi, 1 anno e 3 mesi). Se inferiore a 1 anno, mostra in mesi. Se maggiore o uguale a 1 anno, mostra in anni.",
		"poste": "OBBLIGATORIO - Titolo del ruolo ricoperto (es.: Ingegnere di Progettazione, Sviluppatore Senior, Project Manager, ecc.)",
		"secteur": "Settore di attività dell'esperienza in 1-2 parole. Estrai il settore dalle informazioni dell'azienda, del progetto o del contesto. Esempi: Agroalimentare, Automobilistico, Bancario, Edilizia, Biomedico, Chimica, Consulenza, Difesa, Energia, Ambiente, Ferroviario, Grande distribuzione, Infrastrutture, Logistica, Metallurgia / Siderurgia, Navale, Nucleare, Petrolio e Gas, Petrochimica, Farmaceutico, Sanità, Settore pubblico, Telecomunicazioni, IRVE, Fotovoltaico, Trattamento delle acque, Rivalorizzazione energetica, Idroelettrica, Energie Rinnovabili (ENR), Energia eolica, Biogas, Istruzione, Risorse Umane. Se il settore non è identificabile, lascia questo campo vuoto (\"\").",
		"contexte": "Riassumi l'esperienza in modo sintetico per presentare il progetto realizzato in una frase.",
		"projet": "ARRICCHISCI questo campo creando una descrizione fluida e collegata del progetto/esperienza, come nell'esempio del concorrente. Usa TUTTE le informazioni disponibili: azienda, settore, missioni, risultati, apparecchiature, tecnologie, clienti, progetti menzionati. Crea frasi collegate che raccontino una storia coerente, non elenchi puntati separati. IMPORTANTE: Usa parole e formulazioni DIVERSE da quelle utilizzate nei risultati - evita la ripetizione degli stessi termini. STILE NARRATIVO: Racconta il progetto come una storia con contesto, obiettivi e risultati. LIMITE: Massimo 2 frasi per mantenere la fluidità. DIVIETO ASSOLUTO: Non riprendere MAI parola per parola le missioni dei risultati. Il progetto deve raccontare la storia del progetto, non elencare compiti. LINGUA: TUTTO in italiano, alla terza persona. STILE: Usa nomi d'azione (Realizzazione di..., Partecipazione a..., Coordinamento di..., Redazione di..., Supervisione di..., Messa in servizio di..., ecc.) invece di 'Ha fatto...'. Esempio di stile del concorrente: 'Gestione di progetti di elettrolisi alcalina ad alta pressione e sviluppo di nuovi prototipi per la produzione di idrogeno. Questo progetto include la selezione e la specifica delle apparecchiature, la redazione della documentazione tecnica dettagliata, il follow-up degli ordini e la conformità tecnica delle installazioni.'",
		"projets_name": ["Elenco dei nomi dei progetti menzionati per questa esperienza, se presenti. Estrai nomi specifici di progetti menzionati nel contesto, risultati o altrove nella descrizione dell'esperienza. Spesso c'è un solo progetto per esperienza, ma possono essercene diversi. Esempi: 'Progetto di riabilitazione di edificio per uffici per WEWORK', 'Progetto Reine des Neiges', 'Progetto di riabilitazione di edificio per uffici per AXA - Boulevard des Italiens', ecc. Se non viene menzionato alcun nome di progetto, lascia un array vuoto []."],
		"result": "Frase breve che spiega il risultato dell'esperienza. Estrai o deduci una frase concisa che riassuma i risultati ottenuti o i traguardi chiave di questa esperienza. Esempi: 'Raccomandazioni e identificazione di Next Step', 'Locazione di tecnologia all'interno dei laboratori di ricerca per prove', 'Implementazione del progetto e del team responsabile del progetto, identificazione delle cause e di un metodo di valutazione, pianificazione delle prove', ecc. Se non viene menzionato alcun risultato, lascia questo campo vuoto (\"\").",
		"logiciels": ["OBBLIGATORIO - Estrai TUTTI i software/strumenti menzionati in questa esperienza (es.: SolidWorks, Python, React, AWS, Docker, ecc.) - anche se non esplicitamente elencati, deducili dal contesto"],
		"réalisations": [
		  "Elenca TUTTE le missioni/realizzazioni menzionate nel CV per questa esperienza. CORRISPONDENZA ESATTA OBBLIGATORIA: se il CV ha 6 bullet point, devi averne 6. Se il CV ha 8 bullet point, devi averne 8. Non troncare MAI - elenca tutto ciò che è scritto nel CV, anche se molto dettagliato. INCLUDI TUTTI i dettagli tecnici: apparecchiature specifiche (ESDV, control valves, PSV, flowmeters, ecc.), documenti tecnici (hook-up drawings, loop diagrams, wiring diagrams, ecc.), calcoli specifici (CO2 snuffing, static mixer, ecc.), tipologie di strumenti (pressure & temperature & level transmitters, ecc.). Riproduci TUTTI i dettagli tecnici menzionati nel CV originale. LINGUA: TUTTO in italiano, alla terza persona (lui/lei, l’ingegnere, ecc.)."
		],
		"AI_suggest": ["Se puoi dedurre elementi rilevanti non presenti nel CV. I suggerimenti devono essere specifici e adattati a ciascuna esperienza, pertinenti per i recruiter; il loro numero deve variare in base alle esperienze, senza ridondanze. Non inserirli sistematicamente: deve sembrare naturale."]
	  }
	],
	"logiciels": [
	  {
		"logiciel": "",
		"level": "Stima il livello tra: Principiante, Intermedio, Avanzato, Esperto.",
		"temps_utilisation": "Stima il tempo di utilizzo in mesi."
	  }
	]
  }
  
  ISTRUZIONI CRITICHE:
  
  1. **ESTRAZIONI OBBLIGATORIE**:
	 - Estrai TUTTE le esperienze professionali (tirocini, CDI, CDD, apprendistati, ecc.) - NON DIMENTICARNE NEMMENO UNA
	 - Per ogni formazione: date_debut, date_fin, diplome E ecole_cursus sono OBBLIGATORI
	 - Per ogni esperienza: date_debut, date_fin, entreprise, detail_entreprise, durée, poste E logiciels sono OBBLIGATORI
	 - **CRITICO**: Rileggi il CV più volte per essere certo di aver estratto TUTTE le esperienze menzionate
	 - **ESAUSTIVITÀ DELLE REALIZZAZIONI**: Per ogni esperienza, elenca TUTTE le missioni/realizzazioni menzionate nel CV, anche se numerose. CORRISPONDENZA ESATTA OBBLIGATORIA: se il CV ha 6 bullet point, devi averne 6. Se ne ha 8, devi averne 8. Non troncare MAI le realizzazioni. INCLUDI TUTTI i dettagli tecnici specifici: apparecchiature (ESDV, control valves, PSV, flowmeters), documenti (hook-up drawings, loop diagrams, wiring diagrams), calcoli (CO2 snuffing, static mixer), strumenti (pressure & temperature & level transmitters), ecc.
	 - **ARRICCHIMENTO DEL CAMPO "PROJET"**: Per ogni esperienza, arricchisci il campo "projet" creando una descrizione fluida e collegata come nell’esempio del concorrente. Usa TUTTE le informazioni disponibili: azienda, settore, missioni, realizzazioni, apparecchiature, tecnologie, clienti, progetti menzionati. Crea frasi che raccontino una storia coerente, non elenchi. IMPORTANTE: Usa parole e formulazioni DIVERSE da quelle usate nelle realizzazioni - evita ripetizioni. STILE NARRATIVO: Racconta il progetto con contesto, obiettivi e risultati. LIMITE: Massimo 2 frasi. DIVIETO ASSOLUTO: Non riprodurre MAI parola per parola le missioni delle realizzazioni nel progetto. LINGUA: TUTTO in italiano, alla terza persona. STILE: Usa nomi d’azione (Réalisation de..., Participation à..., ecc.).
  
  2. **NUOVI CAMPI**:
	 - **"phone"**: Estrai il numero di telefono se presente nel CV. Formato: "+33 1 23 45 67 89" oppure "01.23.45.67.89" oppure "0123456789". Se assente, lascia vuoto.
	 - **"nom"**: Estrai il cognome del candidato se presente nel CV. Se assente, lascia vuoto.
	 - **"summary"**: Crea una sintesi professionale concisa (max 2-3 righe) che presenti il candidato, le principali competenze e l’esperienza chiave. Sii incisivo e professionale.
     - **"languages"**: Estrai tutte le lingue menzionate nel CV (sezione lingue, esperienze internazionali, formazioni, ecc.) con il relativo livello QCER. Usa i nomi completi in italiano: "Francese", "Inglese", "Tedesco", "Spagnolo", "Italiano", ecc. Per il livello, usa i livelli QCER (A1, A2, B1, B2, C1, C2) o termini come "Madrelingua", "Fluente", "Intermedio", "Principiante" se non è specificato. **IMPORTANTE**: Se il livello non è menzionato nel CV, lascia il campo "level" vuoto (""). Esempi: {"language": "Francese", "level": "Madrelingua"}, {"language": "Inglese", "level": "B2"}, {"language": "Tedesco", "level": ""}. Se non è menzionata alcuna lingua, lascia un array vuoto [].
      - **"certifications"**: Estrai TUTTE le certificazioni menzionate nel CV (sezioni certificazioni, formazione, education, esperienze, ecc.). Non inventare — estrai solo quelle esplicitamente menzionate. Esempi: "BOSIET", "NEBOSH", "IOSH", "OSHA", "API", "ASME", "ISO 9001", "Six Sigma", "PMP", "Initial training on HV-LV Electrical Installations", "Electrical accreditation certificate according to the C18-510 standard", "National first-aid diploma", "BCTE Certificate", "Qess Certification", "ATEX Certificate", ecc. Se non è menzionata alcuna certificazione, lascia [].
      - **"technical_skills"**: Estrai le competenze tecniche dettagliate e precise menzionate nel CV. Non limitarti alle parole chiave — estrai competenze complete e dettagliate. Cerca nelle sezioni "Technical Skills", "Compétences", "Skills", esperienze professionali, ecc. Esempi forniti sopra. Se non è menzionata alcuna competenza tecnica, lascia [].
      - **"secteurs_activites"**: Estrai tutti i settori di attività in cui il candidato ha lavorato (es.: "Petrolio e Gas", "Petrolchimico", "Energia", "Automotive", "Aerospazio", "Informatica", "Finanza", "Sanità", "Telecomunicazioni", ecc.). Analizza le esperienze professionali per identificarli.
     - **"domaines_expertise"**: Estrai le aree di specializzazione del candidato (es.: "Sviluppo Web", "Data Science", "Project Management", "Marketing Digitale", "Progettazione Meccanica", "Intelligenza Artificiale", ecc.). Basati su competenze tecniche ed esperienze.
     - **"detail_entreprise"**: Per ogni esperienza, aggiungi una descrizione dell’azienda in 1-2 frasi includendo: settore di attività, dimensione (startup, PMI, grande gruppo), specializzazione, posizione di mercato. Sii preciso e informativo.
  
  3. **CAMPO "poste"**:
	 - È il TITOLO DEL RUOLO RICERCATO basato sull’analisi delle esperienze pregresse
	 - **METODO DI ESTRAZIONE**:
	   a) Se il candidato indica un ruolo specifico ricercato → utilizza quello
	   b) In caso contrario, analizza tutte le esperienze e deduci il titolo più rappresentativo
	   c) Dai priorità al ruolo più recente o a quello che riflette meglio l’evoluzione di carriera
	   d) Sii preciso e professionale nel titolo (evita termini generici)
	 - Esempi: "Ingegnere Progettazione Meccanica", "Solution Architect", "Data Engineer", "Sviluppatore Full Stack", "Project Manager", "Consulente", "Ingegnere Civile", "Product Manager"
  
  4. **DATE E CALCOLO DELL’ESPERIENZA**:
     - **DATE DELLE ESPERIENZE**: Estrai SEMPRE le date di inizio e fine di ogni esperienza
     - Formato delle date: strettamente MM/YY (es.: 02/20, 12/22). NON è accettato altro formato
     - Se l’esperienza è in corso, usa "In corso" per date_fin (senza MM/YY)
	 - **CALCOLO DELLA DURATA**: Calcola automaticamente la durata tra date_debut e date_fin
	   - Se durata < 1 anno: mostra in mesi (es.: "6 mesi", "8 mesi")
	   - Se durata ≥ 1 anno: mostra in anni (es.: "2 anni", "1 anno e 3 mesi", "3 anni")
	 - **CALCOLO ESPERIENZA TOTALE**: Per il campo "expérience", trova la data di inizio dell’esperienza significativa più antica e calcola: 2025 - anno_di_inizio = anni di esperienza
	 - Esempio: se la prima esperienza inizia nel 2018 → "7 anni di esperienza"
  
  5. **FORMAZIONI**:
	 - Compila TUTTI i campi: date_debut, date_fin, diplome, ecole_cursus
	 - Sii preciso sul tipo di titolo: Master, Bachelor, BTS, Laurea in Ingegneria, MBA, ecc.
  
  6. **SOFTWARE NELLE ESPERIENZE**:
	 - Estrai TUTTI i software/strumenti menzionati in ogni esperienza
	 - Deducili dal contesto se necessario (es.: se "sviluppo web" → aggiungi HTML, CSS, JavaScript)
  
  7. **COMPLETEZZA**:
	 - Non lasciare ALCUNA esperienza fuori — anche tirocini brevi, missioni occasionali, progetti
	 - Non lasciare ALCUNA formazione fuori
	 - Analizza TUTTO il contenuto del CV
	 - **VERIFICA**: Conta il numero di esperienze menzionate nel CV e assicurati di aver estratto lo stesso numero
	 - **REALIZZAZIONI COMPLETE**: Per ogni esperienza, estrai TUTTE le missioni/realizzazioni menzionate nel CV, anche se molto numerose. CORRISPONDENZA ESATTA OBBLIGATORIA: se il CV ha 6 bullet point, devi averne 6. Se ne ha 8, devi averne 8. Non troncare MAI il contenuto. INCLUDI TUTTI i dettagli tecnici: apparecchiature specifiche (ESDV, control valves, PSV, flowmeters, ecc.), documenti tecnici (hook-up drawings, loop diagrams, wiring diagrams, junction boxes, equipments layouts, I/O list, Instrument list, Alarm and Set point list), calcoli specifici (CO2 snuffing, static mixer, ecc.), tipologie di strumenti (pressure & temperature & level transmitters, ecc.). Riproduci TUTTI i dettagli tecnici del CV originale.
  
  NB: Non far apparire il tipo di contratto (es.: Tirocinio, Apprendistato, CDI, CDD...) nelle esperienze.

  **REGOLA CRITICA PER I TERMINI TIROCINIO/APPRENDISTATO**:
	- Rimuovi SEMPRE le parole "alternant", "stagiaire", "stage", "apprenti", "apprentissage" dai titoli dei ruoli (campo "poste" principale e campo "poste" delle esperienze)
	- Conserva TUTTE le esperienze (tirocini, apprendistati, ecc.) ma trattale come esperienze professionali normali
	- Riformula i titoli affinché siano professionali senza menzionare lo status
	- **ESEMPI DI TRASFORMAZIONE OBBLIGATORIA**:
  * "Stagiaire Développeur" → "Développeur"
  * "Alternant Ingénieur" → "Ingénieur" 
  * "Stagiaire de Recherche" → "Chercheur"
  * "Apprenti Data Analyst" → "Data Analyst"
  * "Stage Marketing" → "Marketing"
  * "Alternance Commercial" → "Commercial"
  
  Aggiungi quante più informazioni possibili analizzando il CV e deducendo elementi non necessariamente presenti, come farebbe un esperto HR.
  
  L’output deve rispettare ESATTAMENTE il modello sopra. Se un’informazione non è presente e non puoi stimarla, lascia il campo vuoto (stringa vuota "").
  
  Ecco il JSON di estrazione da analizzare:
  
  ` + nuextractJSON + `
  
  **ULTIMA ISTRUZIONE CRITICA**: 
  - Estrai TUTTE le esperienze e TUTTE le loro realizzazioni senza eccezioni
  - Non troncare MAI il contenuto, anche se molto lungo
  - Assicurati che ogni esperienza abbia tutte le sue missioni/realizzazioni elencate
  - Se un’esperienza ha molti dettagli nel CV, riproduci TUTTI questi dettagli
  - **CORRISPONDENZA ESATTA OBBLIGATORIA**: Se il CV ha 6 bullet point, devi averne 6. Se ne ha 8, devi averne 8. Non riassumere MAI - riproduci TUTTI i bullet point del CV originale
  - **DETTAGLI TECNICI OBBLIGATORI**: Includi TUTTI i dettagli tecnici menzionati: apparecchiature specifiche (ESDV, control valves, PSV, flowmeters, ecc.), documenti tecnici (hook-up drawings, loop diagrams, wiring diagrams, junction boxes, equipments layouts, I/O list, Instrument list, Alarm and Set point list), calcoli specifici (CO2 snuffing, static mixer, ecc.), tipologie di strumenti (pressure & temperature & level transmitters, ecc.)
  - **DISTINZIONE PROGETTO/REALIZZAZIONI**: Il campo "projet" deve essere un paragrafo narrativo fluido che racconti la storia del progetto (massimo 2 frasi). Il campo "réalisations" deve essere un elenco dettagliato delle missioni. MAI ripetizione tra i due campi - usa parole diverse. DIVIETO ASSOLUTO: Non riprendere MAI parola per parola le missioni delle realizzazioni nel progetto. LINGUA: TUTTO in italiano, alla terza persona. STILE: Usa nomi d’azione (Realizzazione di..., Partecipazione a..., Coordinamento di..., Redazione di..., Supervisione di..., Messa in servizio di..., ecc.) invece di 'Ha fatto...'.
  - **OBIETTIVO**: Documento il più completo possibile, senza riassunto - riproduci TUTTI i dettagli tecnici del CV originale
  
  Rispondi SOLO con il JSON strutturato, senza testo prima o dopo.
  `
}

// GetExtractionPromptProductionGerman gibt das deutschsprachige Produktions-Prompt für die Extraktion zurück
func GetExtractionPromptProductionGerman(nuextractJSON string) string {
	return `Du bist ein HR-Experte, der auf die Analyse von Lebensläufen (CV) spezialisiert ist. Analysiere das JSON-Wörterbuch der CV-Extraktion, das ich dir als Input bereitstelle, und extrahiere ALLE relevanten Informationen in Form eines strukturierten Wörterbuchs, das als JSON gespeichert werden kann, gemäß folgendem Modell:
  
  ÄNDERE AUF KEINEN FALL DIE SCHLÜSSEL DIESES WÖRTERBUCHS, DA SIE SPÄTER ANDERWEITIG VERWENDET WERDEN.
  
  {
	"prenom": "VORNAME DES KANDIDATEN – PFLICHTFELD. Extrahiere den Vornamen des Kandidaten aus dem CV. Dieses Feld muss IMMER ausgefüllt sein, außer in Ausnahmefällen, in denen der Vorname wirklich nicht im CV genannt wird. Suche im Kopfbereich, in der Signatur oder in jeder Nennung des vollständigen Namens des Kandidaten.",
	"nom": "Nachname des Kandidaten. Wenn er nicht ausdrücklich im CV vorhanden ist, lasse dieses Feld leer (\"\"). Nicht erfinden.",
	"email": "E-Mail-Adresse des Kandidaten. Wenn sie nicht ausdrücklich im CV vorhanden ist, lasse dieses Feld leer (\"\"). Nicht erfinden.",
	"phone": "Telefonnummer des Kandidaten. Wenn sie nicht ausdrücklich im CV vorhanden ist, lasse dieses Feld leer (\"\"). Nicht erfinden.",
    "summary": "Berufliche Zusammenfassung in maximal 2–3 Zeilen, die den Kandidaten, seine Schlüsselkompetenzen und seine Haupterfahrung präsentiert. Sei prägnant, aber wirkungsvoll, um einen Überblick über das Profil zu geben. STIL: Verwende einen professionellen Stil wie 'Mehr als X Jahre Berufserfahrung in [Bereich], einschließlich [spezifischer Aktivitäten] für [Branchen]. Erfahrung in [Umgebungen] mit nachweislicher Zusammenarbeit mit [Unternehmensarten].'",
	"age": "Wenn das Alter nicht ausdrücklich im CV genannt ist, lasse dieses Feld leer (\"\"). Nicht schätzen.",
	"poste": "GESUCHTE FUNKTIONSBEZEICHNUNG – Analysiere die bisherigen Erfahrungen und leite die passendste Stellenbezeichnung ab; sei präzise, falls ein spezifisches Tätigkeitsfeld erkennbar ist. Wenn der Kandidat eine bestimmte Position sucht, verwende diese. Andernfalls leite sie aus der aktuellsten oder repräsentativsten Position seines Profils ab. Beispiele: 'Ingenieur Mechanische Konstruktion', 'Solution Architect', 'Data Engineer', 'Projektleiter', 'Full-Stack-Entwickler'",
	"diplome": "Höchste/maßgebliche Ausbildung (Name der Ingenieur-/Business-School oder des M2).",
	"expérience": "Berechne die Gesamterfahrung in Jahren: Finde das Startdatum der ältesten Erfahrung und ziehe es vom aktuellen Jahr (2025) ab. Wenn kein Datum verfügbar ist, leer lassen.",
	"mobilité": "Gesuchte geografische Position, sofern angegeben.",
	"disponibilité": "",
	"permis_B": "",
    "hobbies": ["Liste der Interessen"],
    "languages": [
      {
        "language": "Name der Sprache (z. B.: Französisch, Englisch, Deutsch, Spanisch, usw.)",
        "level": "GER-Niveau (z. B.: A1, A2, B1, B2, C1, C2, Muttersprachler, Fließend, Mittelstufe, Anfänger, usw.)"
      }
    ],
    "certifications": ["Liste der im CV genannten Zertifizierungen (z. B.: BOSIET, NEBOSH, IOSH, OSHA, API, ASME, ISO 9001, Six Sigma, PMP, Initial training on HV-LV Electrical Installations, Electrical accreditation certificate according to the C18-510 standard, National first-aid diploma, BCTE Certificate, Qess Certification, ATEX Certificate, usw.). Nicht erfinden – nur ausdrücklich genannte Zertifizierungen extrahieren."],
    "technical_skills": ["Liste der im CV genannten, detaillierten und präzisen technischen Kompetenzen. Vollständige technische Kompetenzen extrahieren, nicht nur Schlagwörter. Beispiele: 'Construction, QA/QC, Pre-Commissioning, Commissioning, hook-up, Start-up, Maintenance', 'Smart Relays & Soft Starters', 'Instrument calibrations and loop checks', 'Knowledge of PCS, PDCS & DCS system operation', 'Primary and secondary earthing system, support & cable tray installation, instrument and junction boxes installation, cable pulling dressing & termination, Circuits Control', 'HV/MV/LV, firefighting, CCTV & security and access control systems', 'Check list A&B, static testing (Pressure testing, resistance checks), calibration, loop check, cause and effects (loop function check and multilevel trip sequence)', 'Validations, documents reviews, and investigations (Documents: URS, FAT, FDS, CSV, IQ, OQ, and PQ)', 'Good experience in oil and gas industry from wellhead installations to storage and transfer', 'Ability to apply safety rules – hazardous areas - explosive atmosphere (LEL), Indices of protection - protection methods using on hazardous area', 'Ability to understand and use work permits system to control/coordinate and communicate all tasks', 'Ability to do risk assessments before starting work', usw."],
    "competence_fonctionnelle": ["Liste der im CV genannten funktionalen Kompetenzen. Diese Kompetenzen können in Erfahrungen, Soft Skills oder im Lebenslauf vorhanden sein. Extrahiere funktionale Kompetenzen wie: 'Planung, Koordination', 'Projektmanagement', 'Technologiebewertung und -vergleich', 'Berichterstattung und Analyse, interne Kommunikation', 'Entscheidungsfindung und strategische Empfehlung', 'Autonomes Projektmanagement', 'Funktionale und technische Analyse', 'Technologiebeobachtung', 'Interaktionen mit technischen Stakeholdern', 'Technische Lösungsvergleiche und -auswahl', 'Zusammenfassung und Berichterstattung', 'Beherrschung von Werkzeugen für kontinuierliche Verbesserung (Lean, Six Sigma)', 'Produktionsoptimierung', 'Industrielle Effizienz', 'Studie von Nichtkonformitäten, Diagnose und Problemlösung', 'Autonomie und Initiative', 'Interne Kommunikation und Zusammenarbeit', 'Anpassungsfähigkeit des Diskurses, Veränderungsmanagement', 'Analyse- und Synthesedenken', 'Teamführung', 'Budgethandhabung', 'Kommerzielle Verhandlung', 'Teamausbildung und -entwicklung', 'Strategie und Vision', 'Innovation und Kreativität', 'Kundenbeziehungen', 'Qualitätsmanagement', 'Organisation und Planung', 'Konfliktlösung', usw. Wenn keine funktionale Kompetenz erwähnt wird, lasse ein leeres Array []."],
     "secteurs_activites": ["Liste der Branchen, in denen der Kandidat gearbeitet hat (z. B.: Öl & Gas, Petrochemie, Energie, Automobil, Luft- und Raumfahrt, IT, Finanzen, Gesundheit, usw.)."],
    "domaines_expertise": ["Liste der fachlichen Kompetenzfelder/Expertisen des Kandidaten (z. B.: Webentwicklung, Data Science, Projektmanagement, Digitales Marketing, usw.)."],
	"formations": [
	  {
		"date_debut": "PFLICHT – Startjahr (z. B.: 2020, 2018-2019)",
		"date_fin": "PFLICHT – Abschlussjahr (z. B.: 2022, 2020-2021)",
		"diplome": "PFLICHT – Genaue Abschlussart (z. B.: Master Maschinenbau, Ingenieur-Diplom, Bachelor Informatik, BTS Handel, Medizinstudium, MBA, usw.)",
		"ecole_cursus": "PFLICHT – Vollständiger Name der Hochschule/Universität (z. B.: École Centrale Paris, Université Pierre et Marie Curie, HEC Paris, usw.)"
	  }
	],
    "expériences": [
	  {
        "date_debut": "PFLICHT – Startdatum im Format MM/YY (z. B.: 02/20, 09/19)",
        "date_fin": "PFLICHT – Enddatum im Format MM/YY (z. B.: 12/22, 08/20). Wenn die Erfahrung noch andauert, verwende 'Laufend'",
		"entreprise": "PFLICHT – Firmenname",
		"detail_entreprise": "PFLICHT – Unternehmensbeschreibung in 1–2 Sätzen: Branche, Größe, Spezialisierung, Marktposition. Beispiel: 'Startup, spezialisiert auf Künstliche Intelligenz und Machine Learning, mit 50 Mitarbeitenden und führend in prädiktiver Analyse für den Bankensektor'",
		"durée": "PFLICHT – Automatisch berechnete Dauer (z. B.: 2 Jahre, 6 Monate, 1 Jahr 3 Monate). Bei < 1 Jahr in Monaten anzeigen. Bei ≥ 1 Jahr in Jahren anzeigen.",
		"poste": "PFLICHT – Positionsbezeichnung (z. B.: Konstruktionsingenieur, Senior-Entwickler, Projektleiter, usw.)",
		"secteur": "Branche der Erfahrung in 1-2 Wörtern. Extrahiere die Branche aus den Unternehmensinformationen, dem Projekt oder dem Kontext. Beispiele: Lebensmittelindustrie, Automobil, Bankwesen, Bauwesen, Biomedizin, Chemie, Beratung, Verteidigung, Energie, Umwelt, Bahn, Einzelhandel, Infrastruktur, Logistik, Metallurgie / Stahlindustrie, Marine, Kernenergie, Öl & Gas, Petrochemie, Pharmazie, Gesundheitswesen, Öffentlicher Sektor, Telekommunikation, IRVE, Photovoltaik, Wasseraufbereitung, Energiegewinnung, Wasserkraft, Erneuerbare Energien (ENR), Windenergie, Biogas, Bildung, Personalwesen. Wenn die Branche nicht identifizierbar ist, lasse dieses Feld leer (\"\").",
		"contexte": "Fasse die Erfahrung knapp zusammen, um das durchgeführte Projekt in einem Satz vorzustellen.",
		"projet": "ERWEITERE dieses Feld, indem du eine flüssige, zusammenhängende Beschreibung des Projekts/der Erfahrung erstellst – wie im Wettbewerbsbeispiel. Nutze ALLE verfügbaren Informationen: Unternehmen, Branche, Aufgaben, Ergebnisse, Ausrüstungen, Technologien, Kunden, erwähnte Projekte. Bilde verbundene Sätze, die eine kohärente Geschichte erzählen, keine getrennten Aufzählungen. WICHTIG: Verwende ANDERE Wörter/Formulierungen als bei den Ergebnissen – Wiederholungen vermeiden. ERZÄHLSTIL: Das Projekt mit Kontext, Zielen und Ergebnissen erzählen. GRENZE: Maximal 2 Sätze für flüssige Lesbarkeit. ABSOLUTES VERBOT: NIEMALS die Aufgaben aus den Ergebnissen wortwörtlich wiederholen. Das Projekt muss die Geschichte schildern, nicht Aufgaben auflisten. SPRACHE: ALLES auf Deutsch, in der dritten Person. STIL: Verwende Tätigkeitsnomen (Durchführung von…, Mitwirkung an…, Koordination von…, Erstellung von…, Überwachung von…, Inbetriebnahme von…, usw.) statt 'Er/Sie hat…'. Wettbewerbsstil-Beispiel: 'Leitung von Hochdruck-Alkalielektrolyse-Projekten und Entwicklung neuer Prototypen zur Wasserstoffproduktion. Dieses Projekt umfasst die Auswahl und Spezifikation der Ausrüstung, die Erstellung detaillierter technischer Dokumentation, die Bestellverfolgung und die technische Konformität der Anlagen.'",
		"projets_name": ["Liste der für diese Erfahrung genannten Projektnamen, falls vorhanden. Extrahiere spezifische Projektnamen, die im Kontext, in den Ergebnissen oder anderswo in der Erfahrungsbeschreibung erwähnt werden. Oft gibt es nur ein Projekt pro Erfahrung, aber es können mehrere sein. Beispiele: 'Bürogebäude-Sanierungsprojekt für WEWORK', 'Projekt Reine des Neiges', 'Bürogebäude-Sanierungsprojekt für AXA - Boulevard des Italiens', usw. Wenn kein Projektname erwähnt wird, lasse ein leeres Array []."],
		"result": "Kurzer Satz, der das Ergebnis der Erfahrung erklärt. Extrahiere oder leite eine prägnante Phrase ab, die die erzielten Ergebnisse oder die Schlüsselerfolge dieser Erfahrung zusammenfasst. Beispiele: 'Empfehlungen und Identifizierung von Next Step', 'Technologievermietung innerhalb von Forschungslaboren für Versuche', 'Einrichtung des Projekts und des Projektteams, Identifizierung von Ursachen und einer Bewertungsmethode, Planung von Tests', usw. Wenn kein Ergebnis erwähnt wird, lasse dieses Feld leer (\"\").",
		"logiciels": ["PFLICHT – ALLE in dieser Erfahrung genannten Software/Tools extrahieren (z. B.: SolidWorks, Python, React, AWS, Docker, usw.) – auch wenn nicht ausdrücklich gelistet, aus dem Kontext ableiten"],
		"réalisations": [
		  "ALLE im CV genannten Aufgaben/Ergebnisse für diese Erfahrung auflisten. EXAKTE KORRESPONDENZ PFLICHT: Hat das CV 6 Bullet Points, müssen hier 6 stehen. Hat es 8, müssen 8 stehen. NIEMALS kürzen – alles auflisten, auch sehr detaillierte Angaben. ALLE technischen Details einschließen: spezifische Ausrüstungen (ESDV, Control Valves, PSV, Flowmeters, usw.), technische Dokumente (Hook-up Drawings, Loop Diagrams, Wiring Diagrams, usw.), spezifische Berechnungen (CO2 Snuffing, Static Mixer, usw.), Instrumententypen (Pressure & Temperature & Level Transmitters, usw.). ALLE im Original-CV genannten technischen Details reproduzieren. SPRACHE: ALLES auf Deutsch, in der dritten Person (er/sie, der/die Ingenieur/in, usw.)."
		],
		"AI_suggest": ["Wenn du relevante, im CV nicht vorhandene Elemente ableiten kannst. Die Vorschläge müssen spezifisch, pro Erfahrung passend und für Recruiter relevant sein; die Anzahl variiert je Erfahrung, ohne Redundanz. Nicht systematisch hinzufügen: es soll natürlich wirken."]
	  }
	],
	"logiciels": [
	  {
		"logiciel": "",
		"level": "Schätze das Niveau ein: Anfänger, Mittelstufe, Fortgeschritten, Experte.",
		"temps_utilisation": "Schätze die Nutzungsdauer in Monaten."
	  }
	]
  }
  
  KRITISCHE ANWEISUNGEN:
  
  1. **PFLICHT-EXTRAKTIONEN**:
	 - ALLE beruflichen Erfahrungen extrahieren (Praktika, unbefristete Befristete Verträge, duale Modelle usw.) – KEINE EINZIGE AUSLASSEN
	 - Für jede Ausbildung: date_debut, date_fin, diplome UND ecole_cursus sind PFLICHT
	 - Für jede Erfahrung: date_debut, date_fin, entreprise, detail_entreprise, durée, poste UND logiciels sind PFLICHT
	 - **KRITISCH**: Lies den CV mehrfach, um sicherzustellen, dass WIRKLICH ALLE genannten Erfahrungen extrahiert wurden
	 - **VOLLSTÄNDIGKEIT DER ERGEBNISSE**: Für jede Erfahrung ALLE im CV genannten Aufgaben/Ergebnisse auflisten, auch wenn sie zahlreich sind. EXAKTE KORRESPONDENZ PFLICHT: Bei 6 Bullets → 6; bei 8 Bullets → 8. Ergebnisse NIEMALS kürzen. ALLE spezifischen technischen Details einschließen: Ausrüstungen (ESDV, Control Valves, PSV, Flowmeters), Dokumente (Hook-up Drawings, Loop Diagrams, Wiring Diagrams), Berechnungen (CO2 Snuffing, Static Mixer), Instrumente (Pressure & Temperature & Level Transmitters) usw.
	 - **AUSARBEITUNG DES FELDES „PROJET“**: Für jede Erfahrung das Feld „projet“ als flüssige, verknüpfte Beschreibung wie im Wettbewerbsbeispiel ausformulieren. ALLE verfügbaren Infos nutzen: Unternehmen, Branche, Aufgaben, Ergebnisse, Ausrüstungen, Technologien, Kunden, erwähnte Projekte. Verbundene Sätze, keine Stichpunkte. WICHTIG: Andere Wörter als bei den Ergebnissen; Wiederholungen vermeiden. ERZÄHLSTIL; GRENZE 2 Sätze. ABSOLUTES VERBOT: Niemals Aufgaben aus den Ergebnissen wörtlich kopieren. SPRACHE: ALLES auf Deutsch, dritte Person. STIL: Tätigkeitsnomen (Réalisation de..., Participation à..., usw.).
  
  2. **NEUE FELDER**:
	 - **"phone"**: Telefonnummer extrahieren, falls im CV vorhanden. Formate: "+33 1 23 45 67 89" oder "01.23.45.67.89" oder "0123456789". Falls nicht vorhanden: leer lassen.
	 - **"nom"**: Nachnamen extrahieren, falls vorhanden. Falls nicht vorhanden: leer lassen.
	 - **"summary"**: Prägnante berufliche Zusammenfassung (max. 2–3 Zeilen) mit Kandidat, Hauptkompetenzen und Kernerfahrung erstellen. Professionell und wirkungsvoll.
     - **"languages"**: Alle im CV genannten Sprachen (Sprachsektion, internationale Erfahrungen, Ausbildungen, usw.) mit GER-Niveau extrahieren. Vollständige deutsche Sprachennamen verwenden: „Französisch“, „Englisch“, „Deutsch“, „Spanisch“, „Italienisch“, usw. Beim Niveau GER (A1–C2) oder Begriffe wie „Muttersprachler“, „Fließend“, „Mittelstufe“, „Anfänger“. **WICHTIG**: Wenn kein Niveau genannt ist, Feld „level“ leer lassen („“). Beispiele: {"language": "Französisch", "level": "Muttersprachler"}, {"language": "Englisch", "level": "B2"}, {"language": "Deutsch", "level": ""}. Wenn keine Sprache genannt ist: [].
      - **"certifications"**: ALLE im CV genannten Zertifizierungen extrahieren (Zertifikats-, Ausbildungs-, Erfahrungssektionen etc.). Nicht erfinden – nur ausdrücklich genannte. Beispiele: „BOSIET“, „NEBOSH“, „IOSH“, „OSHA“, „API“, „ASME“, „ISO 9001“, „Six Sigma“, „PMP“, „Initial training on HV-LV Electrical Installations“, „Electrical accreditation certificate according to the C18-510 standard“, „National first-aid diploma“, „BCTE Certificate“, „Qess Certification“, „ATEX Certificate“, usw. Wenn keine: [].
      - **"technical_skills"**: Detaillierte, präzise technische Kompetenzen extrahieren. Nicht auf Keywords beschränken – vollständige Kompetenzen. In „Technical Skills“, „Compétences“, „Skills“ und in den Erfahrungen suchen. Wenn keine: [].
      - **"secteurs_activites"**: Alle Branchen extrahieren, in denen der Kandidat gearbeitet hat (z. B.: „Öl & Gas“, „Petrochemie“, „Energie“, „Automobil“, „Luft- und Raumfahrt“, „IT“, „Finanzen“, „Gesundheit“, „Telekommunikation“, usw.). Aus den Erfahrungen ableiten.
     - **"domaines_expertise"**: Fachgebiete/Expertisen extrahieren (z. B.: „Webentwicklung“, „Data Science“, „Projektmanagement“, „Digitales Marketing“, „Mechanische Konstruktion“, „Künstliche Intelligenz“, usw.). Auf Basis von Kompetenzen und Erfahrungen.
     - **"detail_entreprise"**: Für jede Erfahrung eine Firmenbeschreibung in 1–2 Sätzen hinzufügen: Branche, Größe (Startup, KMU, Konzern), Spezialisierung, Marktposition. Präzise und informativ.
  
  3. **FELD „poste“**:
	 - Dies ist die GESUCHTE FUNKTIONSBEZEICHNUNG basierend auf der Analyse der bisherigen Erfahrungen.
	 - **EXTRAKTIONSMETHODE**:
	   a) Gibt der Kandidat eine spezifische Zielposition an → diese verwenden  
	   b) Andernfalls alle Erfahrungen analysieren und die repräsentativste Bezeichnung ableiten  
	   c) Die jüngste oder die die Laufbahnentwicklung am besten widerspiegelnde Position priorisieren  
	   d) Präzise und professionell formulieren (generische Begriffe vermeiden)
	 - Beispiele: „Ingenieur Mechanische Konstruktion“, „Solution Architect“, „Data Engineer“, „Full-Stack-Entwickler“, „Projektleiter“, „Consultant“, „Bauingenieur“, „Product Manager“
  
  4. **DATEN UND ERFAHRUNGSBERECHNUNG**:
     - **ERFAHRUNGSDATEN**: Start- und Enddaten jeder Erfahrung IMMER extrahieren.
     - Datumsformat: strikt MM/YY (z. B.: 02/20, 12/22). KEIN anderes Format zulässig.
     - Läuft die Erfahrung noch, „Laufend“ als date_fin verwenden (kein MM/YY).
	 - **DAUERBERECHNUNG**: Dauer zwischen date_debut und date_fin automatisch berechnen.
	   - Wenn Dauer < 1 Jahr: in Monaten angeben (z. B.: „6 Monate“, „8 Monate“)
	   - Wenn Dauer ≥ 1 Jahr: in Jahren angeben (z. B.: „2 Jahre“, „1 Jahr 3 Monate“, „3 Jahre“)
	 - **GESAMTERFAHRUNG**: Für „expérience“ das Startjahr der ältesten signifikanten Erfahrung finden und berechnen: 2025 – Startjahr = Jahre Erfahrung.
	 - Beispiel: Beginnt die erste Erfahrung 2018 → „7 Jahre Erfahrung“.
  
  5. **AUSBILDUNGEN**:
	 - ALLE Felder ausfüllen: date_debut, date_fin, diplome, ecole_cursus.
	 - Genau beim Abschluss-Typ: Master, Bachelor, BTS, Ingenieur-Diplom, MBA, usw.
  
  6. **SOFTWARE IN DEN ERFAHRUNGEN**:
	 - ALLE in jeder Erfahrung genannten Software/Tools extrahieren.
	 - Bei Bedarf aus dem Kontext ableiten (z. B.: bei „Webentwicklung“ → HTML, CSS, JavaScript).
  
  7. **VOLLSTÄNDIGKEIT**:
	 - KEINE Erfahrung auslassen – auch kurze Praktika, punktuelle Einsätze, Projekte.
	 - KEINE Ausbildung auslassen.
	 - DEN GESAMTEN CV analysieren.
	 - **PRÜFUNG**: Anzahl der im CV genannten Erfahrungen zählen und sicherstellen, dass dieselbe Anzahl extrahiert wurde.
	 - **VOLLSTÄNDIGE ERGEBNISSE**: Für jede Erfahrung ALLE Aufgaben/Ergebnisse extrahieren, auch wenn sehr zahlreich. EXAKTE KORRESPONDENZ PFLICHT: 6 Bullets → 6; 8 Bullets → 8. Inhalt NIEMALS kürzen. ALLE technischen Details einschließen: spezifische Ausrüstungen (ESDV, Control Valves, PSV, Flowmeters, usw.), technische Dokumente (Hook-up Drawings, Loop Diagrams, Wiring Diagrams, Junction Boxes, Equipments Layouts, I/O List, Instrument List, Alarm and Set Point List), spezifische Berechnungen (CO2 Snuffing, Static Mixer, usw.), Instrumententypen (Pressure & Temperature & Level Transmitters, usw.). ALLE technischen Details des Original-CV reproduzieren.
  
  NB: Die Vertragsart (z. B.: Praktikum, dual, unbefristet, befristet …) NICHT in den Erfahrungen anzeigen.

  **KRITISCHE REGEL FÜR PRAKTIKUM/DUALES SYSTEM**:
	- IMMER die Wörter „alternant“, „stagiaire“, „stage“, „apprenti“, „apprentissage“ aus den Positionsbezeichnungen entfernen (Hauptfeld „poste“ und „poste“ in den Erfahrungen).
	- ALLE Erfahrungen (Praktika, duale Modelle, usw.) beibehalten, aber wie reguläre Berufserfahrungen behandeln.
	- Bezeichnungen professionell umformulieren, ohne den Status zu nennen.
	- **VERPFLICHTENDE TRANSFORMATIONSBEISPIELE**:
  * "Stagiaire Développeur" → "Développeur"
  * "Alternant Ingénieur" → "Ingénieur" 
  * "Stagiaire de Recherche" → "Chercheur"
  * "Apprenti Data Analyst" → "Data Analyst"
  * "Stage Marketing" → "Marketing"
  * "Alternance Commercial" → "Commercial"
  
  Füge so viele Informationen wie möglich hinzu, indem du den CV analysierst und Elemente deduzierst, die nicht zwingend explizit genannt sind – so wie es ein HR-Experte tun würde.
  
  Die Ausgabe muss EXAKT dem obigen Modell entsprechen. Wenn eine Information nicht vorhanden ist und nicht geschätzt werden kann, das Feld leer lassen (leere Zeichenfolge "").
  
  Hier ist das zu analysierende Extraktions-JSON:
  
  ` + nuextractJSON + `
  
  **LETZTE KRITISCHE ANWEISUNG**: 
  - ALLE Erfahrungen und ALLE dazugehörigen Ergebnisse ausnahmslos extrahieren
  - Inhalt NIEMALS kürzen, auch wenn sehr lang
  - Sicherstellen, dass jede Erfahrung alle ihre Aufgaben/Ergebnisse aufgelistet hat
  - Wenn eine Erfahrung viele Details im CV hat, ALLE diese Details reproduzieren
  - **EXAKTE KORRESPONDENZ PFLICHT**: Hat der CV 6 Bullets, musst du 6 haben. Bei 8 → 8. NIEMALS zusammenfassen – ALLE Bullets des Original-CV wiedergeben
  - **PFLICHT-TECHNISCHE DETAILS**: ALLE genannten technischen Details einschließen: spezifische Ausrüstungen (ESDV, Control Valves, PSV, Flowmeters, usw.), technische Dokumente (Hook-up Drawings, Loop Diagrams, Wiring Diagrams, Junction Boxes, Equipments Layouts, I/O List, Instrument List, Alarm and Set Point List), spezifische Berechnungen (CO2 Snuffing, Static Mixer, usw.), Instrumententypen (Pressure & Temperature & Level Transmitters, usw.)
  - **ABGRENZUNG PROJEKT/ERGEBNISSE**: Das Feld „projet“ muss ein flüssiger, narrativer Absatz sein, der die Projektgeschichte erzählt (max. 2 Sätze). Das Feld „réalisations“ muss eine detaillierte Aufgabenliste sein. KEINE Wiederholung zwischen den Feldern – unterschiedliche Wörter verwenden. ABSOLUTES VERBOT: Aufgaben aus den Ergebnissen NIEMALS wortwörtlich im Projekt wiederholen. SPRACHE: ALLES auf Deutsch, in der dritten Person. STIL: Tätigkeitsnomen (Durchführung von…, Mitwirkung an…, Koordination von…, Erstellung von…, Überwachung von…, Inbetriebnahme von…, usw.) statt „Er/Sie hat …“.
  - **ZIEL**: So vollständig wie möglich, keine Zusammenfassung – ALLE technischen Details des Original-CV reproduzieren.
  
  Antworte AUSSCHLIESSLICH mit dem strukturierten JSON, ohne Text davor oder danach.`
}
