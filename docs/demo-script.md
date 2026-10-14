# Demo video script (3:00)

The submission template asks for a **3-minute** video. Record the deployed app (https://parcel-proof-545977798090.asia-southeast2.run.app) in a browser, screen plus voice.

| Time | On screen | Say |
|---|---|---|
| 0:00 to 0:20 | A dark packing photo, then a marketplace dispute screen (buyer data blurred) | "Indonesian online sellers lose Shopee and TikTok disputes because their proof photo is useless: too dark, the label cut off, the box never shown sealed. They find out days later, when nothing can be re-shot." |
| 0:20 to 0:35 | JejakPaket on Google Play, then the live URL | "Our app JejakPaket records proof for every parcel. ParcelProof is its AI layer: right after packing, it tells the seller whether the photos would hold up, using Gemini on Google Cloud." |
| 0:35 to 1:05 | Click **Good evidence**, then **Too dark** | "Each parcel gets a verdict, Bukti kuat, Perlu dicek or Bukti lemah, with reasons. A free basic check measures brightness, sharpness and the barcode with no model; Gemini reads the whole photo set." |
| 1:05 to 1:50 | Click **Wrong label on the parcel**, press **Run Gemini live**, point at "label reads", then scroll to **Recent live Gemini checks** | "The basic check says strong: the photos are sharp. Gemini reads the label, NLIDAP40707162. The order is NLIDAP40703162. One digit, and that parcel loses a dispute. Gemini is never told which number to expect; it transcribes, our server compares." |
| 1:50 to 2:15 | Scroll to **Measured, not claimed** | "We measured it before any seller sees it: on 64 graded parcels, every weak verdict was right and 95% of weak parcels were caught, at about a fifth of a cent per parcel." |
| 2:15 to 2:45 | README architecture diagram on GitHub | "Gemini answers six fixed fields in a JSON schema, and our code decides the verdict. It runs on Cloud Run in Jakarta with the key in Secret Manager, and every live result is stored in Firestore, so the next visitor sees it at once. Uploaded photos never reach a model, because a real waybill shows the buyer's address." |
| 2:45 to 3:00 | Live URL, github.com/jeejak/parcel-proof, jeejak.id | "Next: shadow mode on real shops, then the warning on the phone while the parcel is still on the table. ParcelProof by Jeejak." |

Before recording: open the live URL in an incognito window, press **Run Gemini live** once on each sample you will show (on the free tier a live call can take 30 s; after that Firestore serves the stored result instantly), and check `/health`.
