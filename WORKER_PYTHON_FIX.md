# Fix Worker Python - Ignorer les jobs extract_cv

## 🎯 Problème

Le worker Python cherche `template_url` dans tous les jobs `pending`, y compris les jobs `extract_cv` qui n'ont pas ce champ.

## ✅ Solution

Modifier le worker Python pour **ignorer les jobs de type `extract_cv`** car ils sont traités par le worker Go.

## 📝 Code à modifier dans `/worker-generation-word-skillforge/leafdocx/leafdocx/worker.py`

### ❌ Avant (ligne 223 environ)

```python
# Récupère tous les jobs pending sans filtre de type
jobs = supabase_client.table('jobs').select('*').eq('status', 'pending').execute()

for job in jobs.data:
    payload = job.get('payload', {})
    template_url = payload['template_url']  # ❌ Va crasher sur extract_cv
    # ...
```

### ✅ Après

```python
# Récupère uniquement les jobs de génération Word (exclut extract_cv)
jobs = (supabase_client.table('jobs')
        .select('*')
        .eq('status', 'pending')
        .neq('type', 'extract_cv')  # 🎯 Exclut les jobs extract_cv
        .execute()
)

for job in jobs.data:
    payload = job.get('payload', {})
    template_url = payload.get('template_url')  # Use .get() for safety
    
    # Vérifier que le job est bien un job de génération Word
    if 'template_url' not in payload:
        print(f"⚠️  Job {job['id']} n'a pas template_url, skip")
        continue
    
    # ... traitement normal
```

## 🔍 Explication

- **Worker Go**: Traite uniquement les jobs `type='extract_cv'`
- **Worker Python**: Traite uniquement les jobs `type != 'extract_cv'` (donc `generate_word_doc`, etc.)

## 📊 Types de jobs

| Type | Worker | Payload contient |
|------|--------|------------------|
| `extract_cv` | Go (extract_cv_worker.go) | `filename`, `file_base64`, `language`, `generationMode` |
| `generate_word_doc` | Python (worker.py) | `template_url`, `competence_dossier`, etc. |
| Autres types | Selon leur handler | Variables |

---

## 🧪 Test

Après le fix, vérifier que :

1. ✅ Les jobs `extract_cv` ne sont plus récupérés par le worker Python
2. ✅ Les jobs `generate_word_doc` sont traités normalement
3. ✅ Les deux workers peuvent tourner en parallèle sans conflit

