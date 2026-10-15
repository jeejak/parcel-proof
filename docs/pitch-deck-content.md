# Pitch deck content

Mapped to the AI Builder Cup 2026 template, slide by slide. Paste each block into its slide, then submit the deck as .pptx (File > Download > Microsoft PowerPoint). Numbers come from [`eval`](eval); keep them as written.

## Slide 1: Team details

- **Team name:** Jeejak
- **Team leader name:** Samsul Hadi
- **Problem statement:** Retail & Commerce: Intelligent Customer and Business Experiences

Team: Samsul Hadi (lead, product and engineering), Cena Azzahra, Ryan Rizki (software engineer).

## Slide 2: template note

Leave as the template has it.

## Slide 3: Brief about the idea

**ParcelProof: will this packing photo win the dispute?**

- Indonesian online sellers lose Shopee and TikTok Shop disputes because their proof photo is useless: too dark, the shipping label cut off, the box never shown sealed. They find out days later, when nothing can be re-shot.
- ParcelProof checks the packing photos right after packing and returns **Bukti kuat** (strong), **Perlu dicek** (worth a look) or **Bukti lemah** (weak), with the reason.
- Gemini reads the photos; our own code decides the verdict and compares the tracking number, so the model can't write a tracking number into a seller's proof.
- It is the AI layer of JejakPaket, our app on Google Play that already records resi scans, packing photos and handover video.

## Slide 4: Opportunities

**How is it different from existing ideas?**
- Recording apps store proof. ParcelProof tells the seller, before the parcel leaves, whether that proof will hold up, and why.
- Built on a product sellers can already install (JejakPaket on Google Play), not a standalone demo.

**How does it solve the problem?**
- It catches weak proof while the parcel is still on the table, when re-shooting costs seconds instead of a lost dispute.
- Reasons are specific: dark, blurry, label not visible, label number doesn't match the order, item not shown, box not shown sealed, wrong parcel count.

**USP**
- Measured before trusted: 100% precision and 94.6% recall on weak evidence, against targets set before the run.
- The model observes, the server decides: fixed JSON schema in, deterministic verdict out, fixed Indonesian text to the seller. Gemini can't write a tracking number into a seller's proof.
- Privacy by design: real photos only after a shop opts in; a free check that sends no photo anywhere runs beside Gemini.

## Slide 5: List of features

- Verdict per parcel: Bukti kuat / Perlu dicek / Bukti lemah, with reason codes
- Tracking-number check: Gemini reads the label, the server compares it with the order
- Photo quality: dark, blurry, obstructed
- Completeness: item shown before packing, box shown sealed, parcel count (koli)
- Free basic check: brightness, sharpness and shipping barcode with no model and no cost
- Structured output only: six fields in a JSON schema; invalid answers are discarded, never repaired
- Cost controls: four photos per parcel at 1024 px, optional lower media resolution (45% fewer input tokens)
- Offline eval tool: precision, recall, cost and latency on a labelled set before any release

## Slide 6: Process flow diagram

Draw left to right:

1. **Packer** scans the resi and takes packing photos in JejakPaket
2. **Photos sync** to Jeejak (paid plans)
3. **ParcelProof** selects up to four photos (first two, last two) and downscales them
4. **Gemini** answers six fields in a JSON schema
5. **Decide** (our code) compares the tracking number, applies the rules, returns verdict + reasons
6. **Seller** sees the badge; weak proof means re-shoot while the parcel is still there

Side branch at step 3: the **basic check** (brightness, sharpness, barcode) with no model.

## Slide 7: Wireframes (optional)

Skip, or reuse a screenshot from Slide 11.

## Slide 8: Architecture diagram

```
Browser ──HTTPS──> Firebase Hosting (web.app) ──rewrite──> Cloud Run: parcel-proof (asia-southeast2, Jakarta)
                     ├─> Basic check (Go: pixels + Code 128 barcode, no model)
                     ├─> Gemini API: gemini-3.5-flash-lite (staged samples only, responseSchema)
                     └─> Decide (Go): verdict + reason codes ──> Browser
Secret Manager (gemini-api-key) ··> Cloud Run
Cloud Run <──> Cloud Firestore (live results + history)

Jeejak backend (built, merged): job over synced photos ──> Gemini on Vertex AI ──> verdict ──> seller console badge.
Running on a demo shop at paket-demo.jeejak.id with the 8 staged parcels; real shops only after they opt in.
```

## Slide 9: Technologies

- **Google AI:** Gemini 3.5 Flash-Lite, multimodal, structured output (`responseSchema`); Gemini API for the public demo; Vertex AI in the Jeejak seller console
- **Google Cloud:** Firebase Hosting, Cloud Run (asia-southeast2, Jakarta), Cloud Build, Artifact Registry, Secret Manager, Cloud Firestore
- **Backend:** Go, Google API client, gozxing for barcodes
- **Product:** JejakPaket (Flutter, Google Play), Jeejak console (Go + React), Postgres, object storage
- **Quality:** offline eval harness, Go race tests, gitleaks secret scanning, govulncheck, GitHub Actions

## Slide 10: Estimated implementation cost (optional)

- Gemini 3.5 Flash-Lite at the Vertex AI list price: **$0.0017 per parcel** (measured: 4,440 input and 72 output tokens per parcel). Re-run on Vertex AI: $0.0018 per parcel at the list price
- At `MEDIUM` media resolution: **$0.0010 per parcel**
- Example, 100 parcels a day for 30 days: about **$5** a month, or **$3** at `MEDIUM`
- Basic check: **$0**
- Hosting: one Cloud Run instance always up (at most 2) plus Firestore, paid from Google Cloud free-trial credit

## Slide 11: Snapshots of the prototype

Screenshots from https://jeejak-parcel-proof-run.web.app:

1. The sample gallery and the "Measured, not claimed" numbers
2. **Wrong label on the parcel:** basic check says Bukti kuat; Gemini says Bukti lemah, reading NLIDAP40707162 while the order is NLIDAP40703162
3. **Too dark:** both checks say Bukti lemah
4. Upload: your own photo gets the basic check only and is never sent to a model

## Slide 12: Prototype performance report / benchmarking

64 staged parcels labelled by a person (37 weak), Gemini 3.5 Flash-Lite, prompt 2026-10-07.1:

| Measure | Result | Target |
|---|---|---|
| "Bukti lemah" precision | 100.0% (35/35) | at least 90% |
| "Bukti lemah" recall | 94.6% (35/37) | at least 90% |
| "unsure" + schema failures | 0.0% (0/64) | under 15% |
| Latency p50 / p95 | 2.08 s / 4.38 s | |
| Cost per parcel | $0.0017 | |

At `MEDIUM` resolution: precision 97.2%, recall 94.6%, $0.0010. Basic check on held-out parcels: dark 4/4, blurry 4/4, no false alarm on 21 good parcels, 0.22 s, $0.

Caveat: the parcels are AI-generated images of staged packing, not real orders; real photos are measured next, in shadow mode.

## Slide 13: Additional details / future development

- Done: the check runs inside the Jeejak seller console on Vertex AI, on a demo shop with the 8 staged parcels; 48 of 48 fields and 8 of 8 verdicts match the free-tier eval
- Shadow mode on opted-in shops, with precision re-measured on real photos
- The warning on the phone at packing time: "photo too dark, retake"
- Claim triage: read the buyer's claim and point to the photo that answers it
- English across Jeejak's site, app and console (planned)
- The Jeejak Super App: products, finance and orders, each with its own AI layer

## Slide 14: Links

- **GitHub public repository:** https://github.com/jeejak/parcel-proof
- **Demo video link (3 minutes):** `VIDEO_URL`
- **Final product link:** https://jeejak-parcel-proof-run.web.app

## Spare slide (template's blank slide after Links): Showcase

**Showcase: the check inside the Jeejak seller console**

Two screenshots from the demo shop: Bukti Drop with a verdict badge on each parcel, and order PP2R36's Foto tab ("Bukti lemah · Nomor resi di label beda dengan pesanan").

- Bukti Drop (left): every packed parcel gets a verdict badge. Order Foto tab (right): the reason in fixed Indonesian text; advice only, it never changes the order.
- Gemini 3.5 Flash-Lite on Vertex AI, on a demo shop holding only the 8 staged parcels (synthetic, no seller data). 8 of 8 verdicts match the free-tier eval.

## Last slide

Closing slide: leave as the template has it.
