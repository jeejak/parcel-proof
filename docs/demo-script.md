# Demo video script (3:30)

The rules ask for **3 to 4 minutes**. A video outside that window may be penalised or stopped, and one that does not clearly show the working solution does not reach prototype evaluation. So:

- **Aim for 3:20 to 3:40.** A re-encode can shave a fraction of a second; never cut it close to 3:00.
- **Show the real app for at least half the running time,** in a browser with the address bar (`jeejak-parcel-proof-run.web.app`) readable during every live step. Motion graphics explain; the app proves.
- **Speak.** The words below are paced at about 2.2 words a second, unhurried. Put the same words on screen as captions too: judges often watch muted.

Live URL: https://jeejak-parcel-proof-run.web.app (Firebase Hosting in front of Cloud Run)

| Time | On screen | Say |
|---|---|---|
| 0:00 to 0:20 | A dark packing photo, a blurry one, then a marketplace dispute screen (buyer data blurred) | "Every day, Indonesian online sellers lose Shopee and TikTok Shop disputes because their proof photo is useless: too dark, the label cut off, the box never shown sealed. They find out days later, when the parcel is gone and nothing can be re-shot." |
| 0:20 to 0:35 | JejakPaket on Google Play, then the ParcelProof title | "We are Jeejak. Our app JejakPaket, on Google Play, records proof for every parcel. ParcelProof is its AI layer: right after packing, it tells the seller whether those photos would win a dispute." |
| 0:35 to 1:05 | **Browser, full window, address bar visible.** Click **Good evidence**, then **Too dark** | "This is the live app, on Firebase Hosting and Google Cloud Run. Each parcel gets a verdict: Bukti kuat, strong; Perlu dicek, worth a look; or Bukti lemah, weak, with the reasons. Two checks run side by side. A free basic check measures brightness, sharpness and the barcode, with no model. Gemini reads the whole photo set." |
| 1:05 to 1:55 | Click **Wrong label on the parcel**. Press **Run Gemini live**; keep the real wait, sped up, labelled "sped up". Highlight "label reads" on the result card | "Here is the case that costs sellers money. The basic check says strong: the photos are sharp and bright. Now I run Gemini live. On the free tier that takes about twenty seconds, sped up here. Gemini 3.5 Flash-Lite reads the label: 4070-7162. The order is 4070-3162. One digit, and this parcel loses its dispute. Gemini is never told which number to expect. It only transcribes; our server compares." |
| 1:55 to 2:10 | **Upload your own photo** (one with no real address in frame). Show the basic check result and the note that it is not sent to a model | "Your own photo gets the basic check only. It is never stored and never reaches a model, because a real waybill shows the buyer's name, phone and address." |
| 2:10 to 2:20 | Scroll to **Recent live Gemini checks** | "Every live result is stored in Firestore, so any instance, and the next visitor, sees it at once." |
| 2:20 to 2:45 | **Measured, not claimed** on the page, or the numbers as graphics | "We measured it before any seller sees it. On 64 staged parcels, AI-generated images labelled by a person, every Bukti lemah verdict was right, 35 of 35, and 35 of 37 weak parcels were caught. No schema failures. Under a fifth of a US cent per parcel." |
| 2:45 to 3:15 | Architecture diagram (README on GitHub, or the graphic) | "Gemini answers six fixed fields through a response schema: structured output only. A pure Go function, not the model, decides the verdict, and sellers see fixed Indonesian text, never model text. Firebase Hosting serves it from Cloud Run in Jakarta, with the key in Secret Manager and live results in Firestore. Our production backend already has the same check on Vertex AI, switched off until shops opt in." |
| 3:15 to 3:30 | Live URL, github.com/jeejak/parcel-proof, jeejak.id | "Next: shadow mode on real shops, then the warning on the phone while the parcel is still on the table. ParcelProof, by Jeejak." |

## Words to keep exact

Each of these is checked against the eval report; do not round or rephrase them on screen or in the voice.

- **"35 of 35" and "35 of 37"**, or 100% precision and 94.6% recall. "95%" is acceptable for recall when spoken.
- **"Under a fifth of a US cent".** $0.0017 is 0.17 cents: closer to a sixth than a fifth, so never "about a fifth".
- **"64 staged parcels, AI-generated images, labelled by a person".** Not "graded" and not "photos": the labels are the shot list's intended values with three rows corrected on review (Plan 0016), and the images are generated, not photographed.
- **Model:** Gemini 3.5 Flash-Lite. **Region:** Jakarta (`asia-southeast2`).
- **Production is built, not running:** the Vertex AI check exists and is switched off until shops opt in. Never say it runs on real sellers.

## Before recording

1. Open the live URL in an incognito window and check `/health` shows `"status":"ok"`.
2. Pick the samples you will show. For **Wrong label on the parcel**, record the first live call so the wait is real; every sample checked before is served from Firestore instantly, which is right for the others but would make the live moment look canned.
3. Prepare an upload photo with no real name, phone or address anywhere in frame.
4. Browser at 100% zoom, 1920×1080, bookmarks bar hidden, no other tabs or notifications.

## In the edit

- **A timer on screen must match the result.** If you show a "live call" counter, stop it on the seconds the result card prints.
- **Label anything sped up** ("sped up") and never cut out the wait without saying so.
- **Music 18 to 20 dB under the voice**, and let it drop out entirely during the live call.
- **Captions** burned in, same words as the voice.
- **Check the export** before uploading: duration between 3:20 and 3:40, and the video plays with the sound on.

## When the paket demo console is up

Replace 2:10 to 2:20 with the Jeejak seller console at `paket-demo.jeejak.id`: the Bukti Drop list with quality badges and one order's Foto tab, about 20 seconds, and trim the architecture row to keep the total under 3:40. Say: "The same check runs inside our real seller console, on a demo shop with the same eight staged parcels."
