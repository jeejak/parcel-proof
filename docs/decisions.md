# Design decisions

Code comments cite these by label (for example "D2"). They come from Jeejak's internal plan for this feature.

- **D1 Advisory.** A verdict never blocks a capture, a sync or a dispute PDF. It is a badge.
- **D2 What is sent.** Up to four packing photos per parcel (first two and last two by order), downscaled to 1024 px, plus koli count and courier. Never videos, handover or receipt photos, the tracking number, recipient, operator, device or shop.
- **D3 Structured output only.** A fixed six-field schema. A response that does not fit is discarded. "unsure" never produces a warning.
- **D4 The server decides.** `Decide` turns observations into the verdict. The model is never told the expected tracking number; the server compares.
- **D5 Off unless configured, opt-in per shop.**
- **D6 Once per record, bounded.** Idempotent on (record, photo set, prompt version, model, media resolution), with a per-shop daily cap.
- **D7 Additive API change.** An optional `quality` object on existing evidence responses.
- **D8 Measured before trusted.** Precision of "lemah" at least 90%, recall at least 90%, "unsure" plus schema failures under 15%, on at least 60 graded records (30 weak). See [`eval`](eval).
- **D9 Nothing extra retained.** Downscaled copies are never written.
- **D10 Few dependencies.** Gemini over REST with the Google API client; one extra module (gozxing) for barcodes.
- **D11 Free tier for staged photos only.** The Gemini API free tier may use inputs to improve Google's products, so it only ever sees staged photos with dummy waybills. Real seller photos go to Vertex AI.
- **D12 Public demo.** Gemini on the staged samples only; uploads get the basic check only.
- **D13 Basic check.** Darkness, blur and the barcode without a model, at no cost and with no new data processor.
