# Brief description

**ParcelProof by Jeejak** checks, right after packing, whether an Indonesian online seller's packing photos would hold up as proof in a Shopee or TikTok Shop dispute.

Gemini 3.5 Flash-Lite reads up to four photos and answers six fields in a JSON schema. Our server, not the model, turns that into a verdict (Bukti kuat, Perlu dicek, Bukti lemah) with reasons and compares the tracking number itself, so the model can't write a wrong number into a seller's proof. A free basic check (brightness, sharpness, barcode) runs beside it.

On 64 graded staged parcels: 100% precision and 94.6% recall on weak evidence, about $0.0017 per parcel.

Firebase Hosting serves the page and forwards to Cloud Run in Jakarta, which holds the Gemini key in Secret Manager. Cloud Firestore stores every live Gemini result, so any instance shows it at once. ParcelProof is the AI layer of JejakPaket, our app on Google Play.

Live: https://jeejak-parcel-proof-run.web.app
Code: github.com/jeejak/parcel-proof
