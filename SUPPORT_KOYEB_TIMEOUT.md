# Configuration Koyeb pour éviter les 504 Gateway Timeout

## Problème
L'endpoint `/extract` peut prendre plusieurs minutes (jusqu'à 10 minutes) car il utilise OpenAI. Par défaut, les proxies de Koyeb ont un timeout de 60 secondes, ce qui provoque des erreurs 504.

## Solutions

### Solution 1 : Augmenter le timeout Koyeb (RECOMMANDÉ)

Dans les **Settings** de votre service Koyeb, ajoutez cette variable d'environnement :

```
KOYEB_DEPLOYMENT_PROXY_TIMEOUT=600
```

Cette variable indique à Koyeb d'attendre jusqu'à 10 minutes avant de couper la requête.

### Solution 2 : Vérifier la configuration CORS

Assurez-vous que la variable d'environnement suivante est bien définie :

```
ALLOWED_ORIGINS=https://getskillforge.app
```

### Solution 3 : Vérifier les logs

Les logs devraient montrer :
```
⚡ Mode de génération: fast, modèle sélectionné: gpt-5-mini
⚡ Override du modèle: gpt-5 -> gpt-5-mini
🤖 Modèle OpenAI utilisé: gpt-5-mini
```

Si vous ne voyez pas ces logs, cela signifie que le timeout se produit avant même que le backend n'ait le temps de traiter la requête.

## Variables d'environnement recommandées sur Koyeb

```env
# CORS
ALLOWED_ORIGINS=https://getskillforge.app

# Timeout
KOYEB_DEPLOYMENT_PROXY_TIMEOUT=600

# API Keys
OPENAI_API_KEY=votre_cle_openai
NUEXTRACT_PROJECT_ID=votre_project_id
NUEXTRACT_API_KEY=votre_api_key
```

## Test

Après avoir configuré ces variables, redéployez le service et testez avec un mode "fast" pour vérifier que gpt-5-mini est utilisé et que le timeout n'est plus atteint.

