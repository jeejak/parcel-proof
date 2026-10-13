# Architecture

## Demo (this repo, on Cloud Run)

```mermaid
sequenceDiagram
  participant U as Browser
  participant S as Cloud Run (cmd/server)
  participant B as Basic check (Go)
  participant G as Gemini API
  U->>S: POST /api/check/sample/{id}
  S->>B: brightness, sharpness, Code 128 barcode
  B-->>S: observation (unsure where it cannot see)
  S-->>U: basic verdict + recorded Gemini verdict (instant)
  U->>S: POST /api/check/sample/{id}/live ("Run Gemini live")
  S->>G: up to 4 photos (1024 px) + koli + courier, responseSchema
  G-->>S: six fields as JSON
  S->>S: Decide(observation, resi, koli) -> verdict + reasons
  S-->>U: live verdict (cached once per sample per instance)
```

- Uploads: `POST /api/check/upload` runs the basic check only. Photos are read into memory, checked and dropped; nothing is stored and no model is called.
- `/health` answers without calling a model.
- If Gemini fails or the free tier is busy, the sample shows the result the eval recorded for the same photos, labelled as such.

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
