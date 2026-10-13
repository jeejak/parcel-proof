# Demo video script (3:40)

Record the **deployed** app (https://parcel-proof-545977798090.asia-southeast2.run.app) in a browser, screen plus voice. No slides except where marked. Keep it under 4:00.

| Time | On screen | Say |
|---|---|---|
| 0:00 to 0:20 | A packing photo that is too dark, then a marketplace dispute screen (yours, buyer data blurred) | "Indonesian online sellers lose Shopee and TikTok disputes because their proof photo turns out to be useless: too dark, the label cut off, the box never shown sealed. They find out days later, when nothing can be re-shot." |
| 0:20 to 0:40 | JejakPaket on Google Play, then https://parcel-proof-545977798090.asia-southeast2.run.app | "Our app JejakPaket records proof for every parcel. ParcelProof is its AI layer: right after packing, it tells the seller whether the photos would hold up, using Gemini on Google Cloud." |
| 0:40 to 1:20 | Click **Good evidence**, then **Too dark** | "Each parcel gets a verdict, Bukti kuat, Perlu dicek or Bukti lemah, with reasons. Two checks run side by side: a basic check that measures brightness, sharpness and the barcode with no model and no cost, and Gemini, which reads the whole photo set." |
| 1:20 to 2:10 | Click **Wrong label on the parcel**, then **Run Gemini live**; point at "label reads" | "Here the basic check says strong: the photos are sharp. Gemini reads the tracking number on the label, NLIDAP40707162, while the order is NLIDAP40703162. One digit. That parcel would lose a dispute. Gemini is never told which number to expect; it transcribes, and our server compares, so a wrong number can never reach a document the seller sends." |
| 2:10 to 2:40 | Click **Box not shown sealed** and **Item never shown** | "Gemini answers six fixed fields in a JSON schema. Our server decides the verdict. Sellers see fixed Indonesian text, never model output." |
| 2:40 to 3:05 | Scroll to **Measured, not claimed** | "Before any seller sees this, we measured it on 64 graded parcels: every 'weak' verdict was right, 95% of weak parcels were caught, no schema failures, about a fifth of a cent per parcel on Vertex AI." |
| 3:05 to 3:25 | README architecture diagram (GitHub) | "It runs on Cloud Run in Jakarta, with the key in Secret Manager. Uploaded photos get the basic check only and never reach a model, because a real waybill shows the buyer's address. In production the same package runs on Vertex AI for shops that opt in." |
| 3:25 to 3:40 | https://parcel-proof-545977798090.asia-southeast2.run.app, `github.com/jeejak/parcel-proof`, jeejak.id | "Next: shadow mode on real shops, then the warning on the phone while the parcel is still on the table. ParcelProof by Jeejak." |

Checklist before recording: open https://parcel-proof-545977798090.asia-southeast2.run.app in an incognito window; press **Run Gemini live** once on each sample you will show so the live result is warm; check `/health`.
