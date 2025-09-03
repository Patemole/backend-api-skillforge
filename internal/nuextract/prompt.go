package nuextract

import "fmt"

// GetExtractionPrompt retourne le prompt pour l'extraction et la structuration des données CV
func GetExtractionPrompt(nuextractJSON string) string {
	return `Je souhaite que tu analyses le dictionnaire JSON de l'extraction de CV que je te fournis en input et que tu extraies toutes les informations pertinentes sous la forme d'un dictionnaire structuré, pouvant être enregistré en JSON, selon le modèle suivant :

NE CHANGE SURTOUT PAS LES CLÉS DE CE DICTIONNAIRE, CAR IL DOIT ÊTRE UTILISÉ AUTREMENT PAR LA SUITE.

{
  "prenom": "",
  "age": "Si ce n'est pas explicite, estimer à partir de la date de naissance si disponible.",
  "poste": "",
  "diplome": "Formation (nom de l'école d'ingénieur, de commerce ou du M2)",
  "expérience": "Ne prends en compte que les expériences pertinentes en cumulant leur durée respective.",
  "mobilité": "Position géographique recherchée si précisée.",
  "disponibilité": "",
  "permis_B": "",
  "hobbies": ["Liste des centres d'intérêts"],
  "formations": [
    {
      "date_debut": "",
      "date_fin": "",
      "diplome": "",
      "ecole_cursus": ""
    }
  ],
  "expériences": [
    {
      "entreprise": "",
      "durée": "",
      "poste": "",
      "contexte": "Résume l'expérience succinctement pour présenter le projet réalisé en une phrase.",
      "projet": "Ici, étoffe autant que possible les objectifs / projets de cette expérience et reformule pour rendre cela le plus long possible, sous forme de titre, sans faire apparaître le nom du candidat.",
      "logiciels": [""],
      "réalisations": [
        "Liste les missions réalisées, reformulées pour apporter un maximum de détails. Ajoute autant d'éléments que possible en les reformulant pour qu'ils soient le plus long possible. ],
       "AI_suggest": [Si tu peux déduire des éléments pertinents non présents dans le CV. Les suggestions doivent être spécifiques et adaptées à chaque expérience, pertinentes pour les recruteurs, leur nombre doit varier selon les expériences, sans redondance entre elles. N'en mets pas systématiquement : cela doit paraître naturel."]
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
