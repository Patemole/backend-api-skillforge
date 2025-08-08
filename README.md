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