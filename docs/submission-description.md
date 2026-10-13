# Brief description

**ParcelProof by Jeejak** checks, right after packing, whether an Indonesian online seller's packing photos would hold up as proof in a Shopee or TikTok Shop dispute.

Gemini 3.5 Flash-Lite reads up to four packing photos and answers six fixed fields in a JSON schema: is the photo usable, is the shipping label visible, what tracking number it reads, is the box sealed, is the item visible, how many parcels it sees. Our server, not the model, turns that into a verdict (Bukti kuat, Perlu dicek, Bukti lemah) with reasons, and compares the tracking number itself, so the model is never told what to expect and can never put a wrong number in a seller's document. A free basic check (brightness, sharpness, barcode, no model) runs beside it.

Measured on 64 graded staged parcels before any seller sees a verdict: 100% precision and 94.6% recall on weak evidence, no schema failures, about $0.0017 per parcel at the Vertex AI price. Deployed on Cloud Run (Jakarta) with Secret Manager; the demo sends only staged sample photos to Gemini.

ParcelProof is the AI layer of Jeejak's JejakPaket app (on Google Play), which records resi scans, packing photos and handover video for sellers.

Live: https://parcel-proof-545977798090.asia-southeast2.run.app. Code: github.com/jeejak/parcel-proof (Apache-2.0).
