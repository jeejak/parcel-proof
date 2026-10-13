# Pitch deck content

The deck must use the **prescribed template** from the Hack2skill dashboard. Its slide order is not known yet; map these sections onto it and drop what the template does not ask for. Numbers are copied from [`eval`](eval); do not round them further.

## 1. Team
- Team: Jeejak. Members: Samsul Hadi (lead), Rahul Todur, Ryan Rizki. `ROLES`: one line each on what they own.
- Jeejak, by CV. Adisatya IT Consultant, Kediri, Indonesia.

## 2. Problem
- Indonesian online sellers on Shopee and TikTok Shop face buyer claims: missing item, empty parcel, wrong item. The marketplace asks the seller for proof.
- The proof often fails: too dark, label cut off, box never shown sealed. The seller learns this days later, when the parcel is gone.
- `REAL_CASE`: one real dispute a seller lost, with the amount, if you have one you can share.

## 3. Solution
- ParcelProof checks the packing photos right after packing and returns Bukti kuat / Perlu dicek / Bukti lemah with reasons.
- Part of JejakPaket (live on Google Play), which already records resi scans, packing photos and handover video.
- Screenshot: the wrong-label sample, basic check "kuat" vs Gemini "lemah" (reads NLIDAP40707162, order is NLIDAP40703162).

## 4. How it uses Google AI
- Gemini 3.5 Flash-Lite, multimodal, structured output with `responseSchema`: six fields, nothing else.
- The model observes, the server decides. The model is never told the expected tracking number.
- Production path: Gemini on Vertex AI. Demo: Gemini API on staged samples.

## 5. Architecture
- Cloud Run (Jakarta), Secret Manager, Gemini. Diagram from the README.
- Free basic check in Go (brightness, sharpness, barcode) beside Gemini.

## 6. Results (measured, not claimed)
- 64 graded staged parcels, 37 weak: precision 100% (35/35), recall 94.6% (35/37), 0 schema failures.
- $0.0017 per parcel at the Vertex AI list price; 2.1 s median. `MEDIUM` resolution: $0.0010.
- Honest limit: the set is generated images; real photos are measured next in shadow mode.

## 7. Business impact
- Fewer lost disputes for sellers: proof that holds up, caught while the parcel can still be re-packed.
- Paid-tier feature in JejakPaket (Pro Lite and up): cloud sync already ships; ParcelProof adds the reason to upgrade.
- Cost per check is a fraction of a cent, so it fits inside the subscription.
- `TRACTION`: installs, active shops or paying tenants, if you can share them.

## 8. Privacy and trust
- Real waybills show buyer details, so the demo sends Gemini staged photos only; uploads get the basic check only.
- In production: opt-in per shop, only records after opt-in, privacy policy names Google Cloud as processor first.

## 9. Roadmap
- Shadow mode on opted-in shops, precision re-measured on real photos.
- The warning on the phone at packing time ("photo too dark, retake").
- Claim triage: read the buyer's claim and point to the photo that answers it.
- The Jeejak Super App: products, finance and orders, each with its own AI layer.

## 10. Links
- Live: `LIVE_URL`
- Code: github.com/jeejak/parcel-proof
- App: play.google.com/store/apps/details?id=id.jeejak.paket
- Site: jeejak.id
