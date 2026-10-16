# Demo video script (3:40, hard maximum 3:45)

> Re-record from this version. The 2026-10-15 cut (`parcelproof-demo-voice.mp4`, 3:35) predates the deck's "hold up in a dispute" wording, the competitor line at 0:20 and the pilot line at the close.

The rules ask for **3 to 4 minutes**. A video outside that window may be penalised or stopped, and one that does not clearly show the working solution does not reach prototype evaluation. So:

- **Aim for 3:35 to 3:42, never above 3:45.** The rules allow up to 4:00, but 3:45 is our own ceiling so a re-encode or a slow intro can never push it over. Never below 3:20.
- **Show the real app for at least half the running time,** in a browser with the address bar (`jeejak-parcel-proof-run.web.app`) readable during every live step. Motion graphics explain; the app proves.
- **Speak.** The words below are paced at about 2.2 words a second, unhurried. Put the same words on screen as captions too: judges often watch muted.

Live URL: https://jeejak-parcel-proof-run.web.app (Firebase Hosting in front of Cloud Run)

| Time | On screen | Say |
|---|---|---|
| 0:00 to 0:20 | A dark packing photo, a blurry one, then a marketplace dispute screen (buyer data blurred) | "Indonesian online sellers can lose Shopee and TikTok Shop disputes when their proof photo is useless: too dark, the label cut off, the box never shown sealed. They find out days later, when the parcel is gone and nothing can be re-shot." |
| 0:20 to 0:40 | The JejakPaket phone on jeejak.id stepping through **Scan, Foto, Video, PDF** (the tabs advance every 4 seconds; record the hero, or use stills `20` to `23`), ending on the dispute PDF, then the ParcelProof title | "We are Jeejak. Our app JejakPaket, on Google Play, records proof for every parcel. Other apps record proof too; ParcelProof checks it. Right after packing, it tells the seller whether those photos would hold up in a dispute." |
| 0:40 to 1:10 | **Browser, full window, address bar visible.** Click **Good evidence**, then **Too dark** | "This is the live app, on Firebase Hosting and Google Cloud Run. Each parcel gets a verdict: Bukti kuat, strong; Perlu dicek, worth a look; or Bukti lemah, weak, with the reasons. Two checks run side by side. A free basic check measures brightness, sharpness and the barcode, with no model. Gemini reads the whole photo set." |
| 1:10 to 1:50 | Click **Wrong label on the parcel**, press **Run Gemini live**; it answers at once. Highlight the **21.8 s** on the result card, then "label reads" | "Here is the case that costs sellers money. The basic check says strong: the photos are sharp and bright. Gemini disagrees. This result came from a live call that took 21.8 seconds on the free tier, and Firestore keeps it, so it comes back at once. Gemini 3.5 Flash-Lite reads the label: 4070-7162. The order is 4070-3162. One digit, and this proof no longer matches the order. Gemini is never told which number to expect. It only transcribes; our server compares." |
| 1:50 to 2:05 | **Upload your own photo** (one with no real address in frame). Show the basic check result and the note that it is not sent to a model | "Your own photo gets the basic check only. It is never stored and never reaches a model, because a real waybill shows the buyer's name, phone and address." |
| 2:05 to 2:30 | `paket-demo.jeejak.id`, already logged in: **Bukti Drop** with the quality badges, then **Sudah dikirim**, order **PP2R36**, **Foto** tab: "Bukti lemah · Nomor resi di label beda dengan pesanan" | "The same check runs inside our real seller console, on Gemini through Vertex AI. Here it is on a demo shop that holds only the same eight staged parcels. The seller sees the verdict next to the photos, in fixed Indonesian text. It is advice only and never changes the order." |
| 2:30 to 2:55 | **Measured, not claimed** on the page, or the numbers as graphics | "We measured it before any seller sees it. On 64 staged parcels, AI-generated images labelled by a person, every Bukti lemah verdict was right, 35 of 35, and 35 of 37 weak parcels were caught. No schema failures. Under a fifth of a US cent per parcel." |
| 2:55 to 3:25 | Architecture diagram (the deck's architecture slide with both lanes, or the README on GitHub) | "Gemini answers six fixed fields through a response schema. A pure Go function, not the model, decides the verdict. Firebase Hosting serves it from Cloud Run in Jakarta, with the key in Secret Manager and live results in Firestore. In our product backend it runs on Vertex AI, on the demo shop only; real shops get it after they opt in." |
| 3:25 to 3:42 | Live URL, github.com/jeejak/parcel-proof, jeejak.id; hold the end card until 3:42 | "Next: a four-week pilot with five to ten sellers, in shadow mode first, then the warning on the phone while the parcel is still on the table. ParcelProof, by Jeejak." |

The voice is about 435 words, roughly 3:18 at 2.2 words a second. The other 20 seconds or so are clicks, page loads and the console navigation, played at real speed with no voice, which lands the cut near 3:40. Do not fill those gaps by talking faster. If the edit runs over 3:45, trim silent page loads first, never the 21.8 s card or the measured numbers; if it comes in under 3:35, hold the result cards on screen longer.

**Why the wrong-label call is not filmed live any more.** All eight samples already have a live result stored in Firestore (R36 ran on 2026-10-14 and took 21.8 s). Each new Cloud Run instance loads those results at startup, so **Run Gemini live** returns the stored result at once and shows its original latency. So the voice says the call happened earlier and quotes the number on the card. Never act out a wait, and never call an instant answer "live right now".

## Words to keep exact

Each of these is checked against the eval report; do not round or rephrase them on screen or in the voice.

- **"35 of 35" and "35 of 37"**, or 100% precision and 94.6% recall. "95%" is acceptable for recall when spoken.
- **"Hold up in a dispute", never "win the dispute".** The deck makes no promise about dispute outcomes (Cena's research: the complaint data do not isolate packing failures), so the voice must not either.
- **"Under a fifth of a US cent".** $0.0017 is 0.17 cents: closer to a sixth than a fifth, so never "about a fifth".
- **"64 staged parcels, AI-generated images, labelled by a person".** Not "graded" and not "photos": the labels are the shot list's intended values with three rows corrected on review (Plan 0016), and the images are generated, not photographed.
- **Model:** Gemini 3.5 Flash-Lite. **Region:** Jakarta (`asia-southeast2`).
- **"21.8 seconds"** for the wrong-label call: it is the number printed on that result card. The eval median is about 2 seconds; do not quote both as if they were one call.
- **The product backend runs it on a demo shop only:** Vertex AI checks the eight staged parcels, and no real shop is opted in. Never say it runs on real sellers.

## Before recording

1. Open the live URL in an incognito window and check `/health` shows `"status":"ok"`.
2. Open `/api/history` and check that **R36 Wrong label on the parcel** is there at 21832 ms. If it is missing, someone cleared Firestore and the next press makes a new call: then the wait is real and the voice must quote the new number instead.
3. Prepare an upload photo with no real name, phone or address anywhere in frame.
4. Browser at 100% zoom, 1920×1080, bookmarks bar hidden, no other tabs or notifications.
5. Log in to `paket-demo.jeejak.id` as the demo owner **before** you start recording, so the login form and the password are never on screen. Check that Bukti Drop shows the badges on all eight parcels.

## In the edit

- **The PDF carries no AI verdict.** Show it as what the seller sends in a dispute, never as ParcelProof output.

- **No timer and no fake wait** for the wrong-label result: it answers at once, and the card's 21.8 s is the real figure.
- **Label anything sped up** ("sped up"), such as a scroll or a page load you shortened.
- **Music 18 to 20 dB under the voice.**
- **Captions say "labelled", never "graded"** (version 1.0 had "graded" in the measured-results caption).
- **Captions** burned in, same words as the voice.
- **Check the export** before uploading: duration between 3:35 and 3:45, and the video plays with the sound on.

## If the console is down

The 2:00 to 2:25 segment needs the demo server and its tunnel. If `paket-demo.jeejak.id` does not load on the day, use the stills in its place (`01` to `05` in the folder below) with the same voice, and do not describe them as live. As a last resort, cut the segment and show **Recent live Gemini checks** on the ParcelProof page for 10 seconds: "Every live result is stored in Firestore, so any instance, and the next visitor, sees it at once." Then hold the result cards longer so the total stays above 3:20.

Stills for the deck or the edit are in `AIBuilderCup/screenshots/` in the Jeejak workspace (not in this repo).
