# Guide Frontend pour le endpoint `/extract` asynchrone

## 🎯 Résumé

L'endpoint `/extract` est maintenant **asynchrone** pour éviter les timeouts Koyeb (limite de 60s). Le frontend doit récupérer le résultat via **polling** ou **WebSockets**.

## 📊 Comparaison des méthodes

### 1. Polling HTTP (méthode actuelle)

**Avantages:**
- ✅ Simple à implémenter (pas de WebSockets)
- ✅ Fonctionne partout
- ✅ Pas de connexion persistante

**Inconvénients:**
- ⚠️ ~30 requêtes/job (poll toutes les 2s pendant 1 min)
- ⚠️ Consommation réseau inutile
- ⚠️ Légère surcharge DB (négligeable si <100 jobs/min)

**Quand utiliser:** Pour un MVP ou si le trafic est faible.

### 2. WebSockets (Supabase Realtime) - BONNE SOLUTION

**Avantages:**
- ✅ Push temps réel (instantané)
- ✅ 1 seule connexion → pas de surcharge
- ✅ Efficient pour nombreuses utilisateurs simultanés
- ✅ Supabase Realtime inclus par défaut

**Inconvénients:**
- ⚠️ Nécessite une lib WS côté frontend
- ⚠️ Connexion persistante

**Quand utiliser:** Production avec trafic modéré/élevé.

### 3. Server-Sent Events (SSE) - ALTERNATIVE

**Avantages:**
- ✅ Push temps réel
- ✅ Simple côté serveur
- ✅ Reconnexion automatique

**Inconvénients:**
- ⚠️ N'a pas besoin de coding maintenant (on peut le coder plus tard)

**Quand utiliser:** Si vous voulez éviter WebSockets.

---

## 💰 Calcul de la surcharge polling

**Scénario basique:**
- 100 jobs/extractions par jour
- ~30 requêtes/job (poll toutes les 2s pendant 1 min)
- Total: 3 000 requêtes/jour = 0.003 req/seconde

**Impact:**
- 🟢 Négligeable pour DB (<1% charge)
- 🟢 Network: ~500 KB/jour (négligeable)
- 🟢 Serveur: aucune charge supplémentaire

**Conclusion:** Le polling est OK pour 99% des cas. Passer à WebSockets si vous avez >1000 utilisateurs simultanés.

---

## 📡 Endpoints

### 1. POST `/extract` - Créer un job d'extraction

**Request:**
```typescript
const formData = new FormData();
formData.append('file', cvFile);          // Fichier CV (PDF, DOCX, TXT)
formData.append('language', 'fr');        // 'fr', 'en', 'de', 'sp', 'it', 'pr'
formData.append('generationMode', 'fast'); // 'fast' ou 'detailed'
formData.append('user_id', userId);       // UUID utilisateur (OBLIGATOIRE)

const response = await fetch('https://marginal-rickie-patemole-14667741.koyeb.app/extract', {
  method: 'POST',
  body: formData
});

const data = await response.json();
```

**Response (HTTP 202 Accepted):**
```json
{
  "success": true,
  "job_id": 42,
  "status": "pending",
  "type": "extract_cv"
}
```

**Erreurs possibles:**
- `400 Bad Request`: `{"success": false, "error": "file not provided"}` ou `{"success": false, "error": "invalid user_id"}`
- `500 Internal Server Error`: `{"success": false, "error": "Échec de la création du job", "error_code": "JOB_CREATION_FAILED"}`

---

### 2. GET `/jobs/:id/status` - Récupérer le statut d'un job

**Request:**
```typescript
const response = await fetch(`https://marginal-rickie-patemole-14667741.koyeb.app/jobs/${jobId}/status`);

const data = await response.json();
```

**Response (succès):**
```json
{
  "success": true,
  "status": "done",        // "pending" | "processing" | "done" | "failed"
  "result": {              // Données extraites du CV (JSON), null si pas terminé
    "nom": "Dupont",
    "prenom": "Jean",
    "email": "jean.dupont@example.com",
    // ... autres champs extraits
  },
  "error": null,           // null ou message d'erreur si status='failed'
  "job_id": "42"
}
```

**Statuts possibles:**
- `pending`: Job créé mais pas encore traité par le worker
- `processing`: Worker en train de traiter le CV
- `done`: Traitement terminé avec succès
- `failed`: Échec du traitement (erreur dans `error`)

---

## 💡 Code Frontend TypeScript/React

```typescript
import { useState } from 'react';

interface ExtractResult {
  nom?: string;
  prenom?: string;
  email?: string;
  // ... autres champs
}

const extractCV = async (
  file: File, 
  language: string = 'fr',
  generationMode: 'fast' | 'detailed' = 'fast',
  userId: string
): Promise<ExtractResult> => {
  // 1. Upload le CV et créer le job
  const formData = new FormData();
  formData.append('file', file);
  formData.append('language', language);
  formData.append('generationMode', generationMode);
  formData.append('user_id', userId);

  const uploadResponse = await fetch('https://marginal-rickie-patemole-14667741.koyeb.app/extract', {
    method: 'POST',
    body: formData
  });

  if (!uploadResponse.ok) {
    throw new Error(`Upload failed: ${await uploadResponse.text()}`);
  }

  const { job_id } = await uploadResponse.json();

  // 2. Polling jusqu'à ce que le job soit terminé
  const maxAttempts = 180; // 6 min max (180 * 2s) - OpenAI peut prendre jusqu'à 4-5 min
  let attempts = 0;

  while (attempts < maxAttempts) {
    await new Promise(resolve => setTimeout(resolve, 2000)); // Attendre 2s

    const statusResponse = await fetch(
      `https://marginal-rickie-patemole-14667741.koyeb.app/jobs/${job_id}/status`
    );

    if (!statusResponse.ok) {
      throw new Error(`Status check failed: ${await statusResponse.text()}`);
    }

    const { status, result, error } = await statusResponse.json();

    if (status === 'done') {
      return result;
    } else if (status === 'failed') {
      throw new Error(error || 'Extraction failed');
    } else if (status === 'pending' || status === 'processing') {
      attempts++;
      continue; // Continue polling
    }
  }

  throw new Error('Timeout: extraction took too long');
};

// Exemple d'utilisation dans un composant React
const MyComponent = () => {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<ExtractResult | null>(null);

  const handleFileUpload = async (file: File) => {
    setLoading(true);
    try {
      const data = await extractCV(file, 'fr', 'fast', 'user-uuid-here');
      setResult(data);
    } catch (error) {
      console.error('Erreur:', error);
      alert(error.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      {loading && <p>⏳ Extraction en cours...</p>}
      {result && <pre>{JSON.stringify(result, null, 2)}</pre>}
    </div>
  );
};
```

---

## 🎨 Exemple avec progression

Si vous voulez afficher une progression pendant l'extraction :

```typescript
const extractCVWithProgress = async (
  file: File,
  language: string,
  generationMode: string,
  userId: string,
  onStatusChange?: (status: string) => void
): Promise<ExtractResult> => {
  // Upload
  // ... (même code que ci-dessus)
  const { job_id } = await uploadResponse.json();

  // Polling avec callback de progression
  while (attempts < maxAttempts) {
    const { status, result, error } = await checkStatus(job_id);

    onStatusChange?.(status); // Callback pour mettre à jour l'UI

    if (status === 'done') return result;
    if (status === 'failed') throw new Error(error);

    await sleep(2000);
    attempts++;
  }
};
```

---

## 📝 Notes importantes

1. **Timing**: Le worker poll la table toutes les 3 secondes, donc comptez 3-6s minimum avant que le statut passe à `processing`.

2. **Timeout**: Si vous attendez plus de 2-3 minutes, il y a probablement un problème (timeout ou erreur silencieuse).

3. **user_id obligatoire**: Il faut fournir un UUID valide pour l'utilisateur. Sans ça, le backend utilise `00000000-0000-0000-0000-000000000000`.

4. **Format du fichier**: PDF, DOCX, TXT sont supportés.

5. **Result**: Le champ `result` dans la réponse est un JSON contenant toutes les données extraites du CV (voir `SCHEMA_REPONSE_EXTRACT.md`).

---

## 🚀 Test rapide avec curl

```bash
# 1. Créer un job
curl -X POST -F "file=@/path/to/cv.pdf" \
  -F "language=fr" \
  -F "generationMode=fast" \
  -F "user_id=123e4567-e89b-12d3-a456-426614174000" \
  https://marginal-rickie-patemole-14667741.koyeb.app/extract

# Response: {"success":true,"job_id":42,"status":"pending","type":"extract_cv"}

# 2. Vérifier le statut (polling)
curl https://marginal-rickie-patemole-14667741.koyeb.app/jobs/42/status

# Response (pending): {"success":true,"status":"pending","result":null,"error":null,"job_id":"42"}
# Response (done): {"success":true,"status":"done","result":{...},"error":null,"job_id":"42"}
```

