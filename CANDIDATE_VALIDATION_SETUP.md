# Configuration de l'endpoint Candidate Validation

## 📋 Variables d'environnement requises

Ajoutez ces variables à votre fichier `.env` :

```bash
# Resend API Key (pour l'envoi d'emails)
RESEND_API_KEY=your_resend_api_key_here

# Email d'envoi par défaut (optionnel)
RESEND_FROM_EMAIL=noreply@getskillforge.app
```

## 🔑 Obtenir une clé API Resend

1. Allez sur [https://resend.com/](https://resend.com/)
2. Créez un compte ou connectez-vous
3. Allez dans "API Keys" dans votre dashboard
4. Créez une nouvelle clé API
5. Copiez la clé et ajoutez-la à votre `.env`

## 🧪 Test de l'endpoint

### 1. Démarrer le serveur
```bash
go run ./cmd/server
```

### 2. Exécuter le test
```bash
./test_candidate_validation.sh
```

### 3. Test manuel avec curl
```bash
curl -X POST http://localhost:8081/candidate-validation \
  -H "Content-Type: application/json" \
  -d @test_candidate_validation.json
```

## 📧 Format de l'email envoyé

L'email sera envoyé à l'adresse `inviter_email` avec :
- **Sujet** : "✅ Dossier de compétences validé - [Nom du candidat]"
- **Contenu** : Template HTML avec les informations du candidat
- **Lien** : URL directe vers le dossier du candidat

## 🔒 Sécurité

- Validation des UUIDs
- Validation des adresses email
- Validation des URLs
- Logs d'audit pour chaque tentative d'envoi
- Vérification que le pourcentage de completion est à 100%

## 📝 Logs

Les logs incluent :
- Tentatives d'envoi d'email (succès/échec)
- Détails des erreurs
- IDs de notification
- Timestamps

## 🚨 Gestion d'erreurs

L'endpoint gère :
- Erreurs de validation des données
- Erreurs d'API Resend
- Erreurs de format de date
- Erreurs de réseau

## 📊 Réponses

### Succès (200)
```json
{
  "success": true,
  "message": "Notification envoyée avec succès",
  "notification_id": "uuid-de-la-notification"
}
```

### Erreur (400/500)
```json
{
  "success": false,
  "message": "Description de l'erreur",
  "error": "Détails de l'erreur"
}
```
