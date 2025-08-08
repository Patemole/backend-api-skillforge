# Go + Supabase starter

## Lancer en local
```bash
cp .env.example .env
go run ./cmd/server



tester avec un payload test dans payload.json avec un requete CURL:
curl -X POST -H "Content-Type: application/json" -d @payload.json http://localhost:8081/jobs

tester le endpoint de polling 
sur un job 4 deja en base et pour le test qui se trouve en "processing"

et tester en changeant sur le job a "done" pour voir si on recupere bien les informations necessaires
curl http://localhost:8081/jobs/4/status

Si a pending ou processing
{"job_id":5}Mac:backend-api-skillforgcurl http://localhost:8081/jobs/5/status
{"error":null,"result":null,"status":"pending"}Mac:backend-api-skillforge lucy-cto$ 

Si a Done
{"job_id":4}Mac:backend-api-skillforgcurl http://localhost:8081/jobs/4/status
{"error":null,"result":{"file_url":"https://gksurcxmvvdvjcrssair.supabase.co/storage/v1/object/public/generated-documents/4.docx?"},"status":"done"}Mac:backend-api-skillforge lucy-cto$ 


Mac:backend-api-skillforge lucy-cto$ curl -X POST -F "file=@/Users/lucy-cto/Desktop/CV_Abdelkader_DAADAA.pdf" http://localhost:8081/extract
{"result":{"Profil candidat":{"Titre du poste":"Ingénieur génie civil - management de projet","Résumé professionnel":"Ingénieur en génie civil passionné par le bâtiment, avec une expérience solide en conduite de travaux et en coordination sur des projets d'envergure. Autonome, adaptable et orienté solutions, j'interviens efficacement dans la gestion des interfaces, la planification et la résolution des aléas chantier. Curieux et investi, je cherche à contribuer activement à des projets ambitieux au sein d'équipes engagées.","Expérience":null,"Disponibilité":null,"Langues":{"Français":"Niveau avancé","Anglais":"Niveau moyen","Espagnol":null}},"Experiences":[{"Poste":"Responsable d'opération MOEX/OPC","Entreprise":"BEXCONSULT Clichy","Date de début":"2025-01","Date de fin":null,"Description":"Pilotage opérationnel du chantier de la phase EXE à la réception. Ordonnancement, pilotage et coordination des différents corps d'état. Suivi de l'avancement des travaux, gestion des points bloquants et relances des entreprises. Animation des réunions de chantier et rédaction des comptes rendus. Contrôle de la qualité d'exécution, du respect des délais et du budget. Interface entre la maîtrise d'ouvrage, les entreprises et les autres intervenants (BET, CSPS...). Suivi des situations de travaux et participation à la gestion financière du projet.","Tak skills":[]},{"Poste":"Stage Ingénieur OPC-G","Entreprise":"Egis groupement Keiros Paris","Date de début":"2023-10","Date de fin":"2024-01","Description":"Surveillance de la progression des travaux. Mise à jour du planning et élaboration du planning sur 3 semaines. Participation à la coordination entre les différents acteurs du chantier. Suivi des rendus. Rédaction du rapport de visite hebdomadaire.","Tak skills":[]},{"Poste":"Chargé d'affaire Réhabilitation","Entreprise":"France Bâtiment Peinture Paris","Date de début":"2023-06","Date de fin":"2023-08","Description":"Elaboration des procédures d'exécution. Etablissement des devis et suivi financier des chantiers. Approvisionnement des matériaux. Analyse des documents marchés (CCTP, CCAP, acte d'engagement) et préparation des documents PPSPS.","Tak skills":[]},{"Poste":"Ingénieur Travaux","Entreprise":"ARCHISUD","Date de début":"2019-09","Date de fin":"2022-07","Description":"Elaborer les procédures d'exécution et préparer les documents nécessaires. Planification des travaux sur Ms project. Établir les devis et assurer le suivi financier des chantiers, y compris l'approvisionnement des matériaux. Coordonner les travaux entre les corps d'état techniques et architecturaux pour garantir leur bonne intégration. Suivre l'avancement des sous-traitants, vérifier l'exécution des tâches. Réaliser les OPR, gérer les réserves et contrôler la qualité et la conformité des travaux. Assurer le suivi budgétaire pour respecter les coûts prévus.","Tak skills":[]},{"Poste":"Conducteur travaux","Entreprise":"L'Etoile Immobilière Tunisie","Date de début":"2017-02","Date de fin":"2018-02","Description":"Suivi des travaux et contrôle de la qualité des ouvrages exécutés. Coordination opérationnelle de l'activité des sous-traitants. Gestion du planning avec ajustements en fonction de l'activité.","Tak skills":[]}],"Compétences techniques":["Contrôle qualité","Normes de construction","Planification","Réception d'ouvrages","Management d'équipe","Gestion de projets"],"Compétences comportementales":["Adaptabilité","Esprit d'équipe","Gestion du stress","Sens des responsabilités","Rigueur"],"Certifications":["QHSE ISO (9001/19001/14001/45001)","Revit Structure"]},"completionTokens":1081,"promptTokens":1840,"totalTokens":
Mac:backend-api-skillforge lucy-cto$ 




tester le endpoint extract:

curl -X POST -F "file=@/Users/lucy-cto/Desktop/CV_Abdelkader_DAADAA.pdf" http://localhost:8081/extract