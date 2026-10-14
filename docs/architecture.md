# Architecture

## Demo (this repo, on Cloud Run)

```mermaid
sequenceDiagram
  participant U as Browser
  participant S as Cloud Run (cmd/server)
  participant B as Basic check (Go)
  participant G as Gemini API
  participant F as Firestore
  U->>S: POST /api/check/sample/{id}
  S->>B: brightness, sharpness, Code 128 barcode
  B-->>S: observation (unsure where it cannot see)
  S->>F: latest live result for this sample?
  S-->>U: basic verdict + stored live (or recorded) Gemini verdict, no model call
  U->>S: POST /api/check/sample/{id}/live ("Run Gemini live")
  S->>G: up to 4 photos (1024 px) + koli + courier, responseSchema
  G-->>S: six fields as JSON
  S->>S: Decide(observation, resi, koli) -> verdict + reasons
  S->>F: save to live_latest/{sample} and live_history
  S-->>U: live verdict
```

- Uploads: `POST /api/check/upload` runs the basic check only. Photos are read into memory, checked and dropped; nothing is stored and no model is called.
- `/health` answers without calling a model.
- If Gemini fails or the free tier is busy, the sample shows the result the eval recorded for the same photos, labelled as such.
- Cloud Firestore (Native mode, asia-southeast2) keeps live results: `live_latest/{sample}` holds the newest per sample, `live_history` every live call. Any instance serves a stored live result before the recorded one, warm-up reads the store instead of spending Gemini quota, and `GET /api/history` lists recent live checks. If Firestore is unreachable the page falls back to memory and the recorded results.
- Cloud Run keeps one instance up (`--min-instances 1`, at most 2), so the page never waits on a cold start. The one wait left is the Gemini call itself on the free tier.

## Production (Jeejak backend, private)

```mermaid
flowchart LR
  P[JejakPaket app] -->|sync photos| O[(Object storage)]
  P -->|records| DB[(Postgres)]
  J[Background job] --> DB
  J --> O
  J -->|"Vertex AI, opted-in paid shops only"| V[Gemini on Vertex AI]
  J -->|verdict row| DB
  C[Seller console] -->|evidence API + quality badge| DB
```

The job uses this repo's `evidencecheck` package: photo selection, prompt, schema and `Decide`. It runs only for shops that opted in, records only captured after the opt-in, once per photo set, prompt version, model and media resolution, under a per-shop daily cap. Off until the privacy policy names Google Cloud as a processor.

## Request to Gemini

- System instruction and response schema: [`evidencecheck/checker.go`](../evidencecheck/checker.go).
- User text: koli count and courier name only. No tracking number, recipient, operator or shop.
- Photos: the first two and the last two packing photos by order, downscaled to 1024 px JPEG ([`evidencecheck/photos.go`](../evidencecheck/photos.go)).
- `temperature: 0`, `responseMimeType: application/json`.
