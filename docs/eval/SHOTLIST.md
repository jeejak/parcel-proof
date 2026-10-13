# Plan 0016 Phase 0: staged eval set, shot list

64 records, 279 photos, 70 waybills to print. Intended verdicts: 13 kuat, 37 lemah, 14 perlu_dicek. Plan 0016 D8 needs at least 60 records and at least 30 lemah.

Generated with `labels.csv` and `waybills.pdf` from one record table, so the three always agree. If you change a record, change it in all three (or ask for a regeneration).

## Rules for every record

1. **Nothing personal in frame, ever.** Only the printed dummy waybills from `waybills.pdf`. No real waybill, receipt, address, phone screen, face, name tag, or document anywhere in the photo, background included. Phase 0 may run on the Gemini Developer API free tier (Plan 0016 D11), whose terms let Google use what is sent and let human reviewers read it, and say: "Do not submit sensitive, confidential, or personal information." Shoot on a plain table against a plain wall.
2. **One folder per record**, named exactly the record id: `R01/`, `R02/`, ... inside this directory.
3. **Photos named in shooting order**: `01.jpg`, `02.jpg`, ... File-name order stands in for `order_index`.
4. **Cut the grey strip off every waybill before sticking it.** The strip carries the record id for you; it must never be in a photo, because the model could read it as a resi.
5. **Phone camera, default settings, no flash** unless a record says otherwise. Shoot portrait or landscape, either is fine.
6. **What the model receives** (Plan 0016 D2): all photos when a record has 4 or fewer; when it has more, only the first 2 and the last 2. Photos marked *[middle photo: not sent to the model]* are there to exercise that selection. Nothing decisive may appear only in a middle photo, and the expected labels below describe the photos that are sent.
7. **Expected values describe what a careful person sees** in the sent photos. If a shot does not come out as described (for example a "not legible" label turns out readable), keep the photo and tell me: the label gets corrected, the photo does not get reshot to match.
8. **Koli**: `koli` is the record's koli count (what the seller entered). `koli_seen` is how many boxes appear in the photos. They differ only in the koli-mismatch records.

## Summary

| ID | Category | Verdict | Photos | Koli (record / shown) | Courier | Waybill to stick |
|---|---|---|---|---|---|---|
| R01 | good | kuat | 4 | 1 / 1 | SPX Express | `SPXID5774637064` |
| R02 | good | kuat | 3 | 1 / 1 | J&T Express | `JP3304452940` |
| R03 | good | kuat | 6 | 1 / 1 | JNE | `16394029077943` |
| R04 | good | kuat | 4 | 2 / 2 | SiCepat | `004546157682` x2 |
| R05 | good | kuat | 8 | 1 / 1 | AnterAja | `100094904108537` |
| R06 | good | kuat | 5 | 1 / 1 | Ninja Xpress | `NLIDAP24457712` |
| R07 | good | kuat | 8 | 3 / 3 | SPX Express | `SPXID7077071791` x3 |
| R08 | good | kuat | 3 | 1 / 1 | J&T Express | `JP7023650806` |
| R09 | good | kuat | 7 | 2 / 2 | JNE | `16957813511813` x2 |
| R10 | good | kuat | 4 | 1 / 1 | SiCepat | `000601062432` |
| R11 | good | kuat | 6 | 1 / 1 | AnterAja | `100090383112466` |
| R12 | good | kuat | 4 | 2 / 2 | Ninja Xpress | `NLIDAP06777400` x2 |
| R13 | good | kuat | 5 | 1 / 1 | SPX Express | `SPXID0364861316` |
| R14 | dark | lemah | 4 | 1 / 1 | J&T Express | `JP1417626322` |
| R15 | dark | lemah | 3 | 1 / 1 | JNE | `16668362572636` |
| R16 | dark | lemah | 5 | 2 / 2 | SiCepat | `000749411473` x2 |
| R17 | dark | lemah | 4 | 1 / 1 | AnterAja | `100087397827811` |
| R18 | dark | lemah | 6 | 1 / 1 | Ninja Xpress | `NLIDAP50717085` |
| R19 | dark (very) | lemah | 4 | 1 / 1 | SPX Express | `SPXID8215646133` |
| R20 | blurry | lemah | 4 | 1 / 1 | J&T Express | `JP7300356993` |
| R21 | blurry | lemah | 3 | 1 / 1 | JNE | `16040602833884` |
| R22 | blurry | lemah | 6 | 1 / 1 | SiCepat | `008099837975` |
| R23 | blurry | lemah | 4 | 2 / 2 | AnterAja | `100027213015412` x2 |
| R24 | blurry | lemah | 5 | 1 / 1 | Ninja Xpress | `NLIDAP02913539` |
| R25 | blurry (heavy) | lemah | 4 | 1 / 1 | SPX Express | `SPXID6121710922` |
| R26 | obstructed (hand over label) | lemah | 4 | 1 / 1 | J&T Express | `JP2360516641` |
| R27 | obstructed (hand over label) | lemah | 5 | 1 / 1 | JNE | `16969357560531` |
| R28 | obstructed (hand over label) | lemah | 3 | 1 / 1 | SiCepat | `003660936850` |
| R29 | obstructed (bag in front) | lemah | 4 | 1 / 1 | AnterAja | `100086502635260` |
| R30 | obstructed (hand over top) | lemah | 4 | 1 / 1 | Ninja Xpress | `NLIDAP60670202` |
| R31 | no label | lemah | 4 | 1 / 1 | SPX Express | none |
| R32 | no label | lemah | 3 | 1 / 1 | J&T Express | none |
| R33 | no label | lemah | 5 | 1 / 1 | JNE | none |
| R34 | no label | lemah | 6 | 1 / 1 | SiCepat | none |
| R35 | no label | lemah | 4 | 1 / 1 | AnterAja | none |
| R36 | label with a different resi | lemah | 4 | 1 / 1 | Ninja Xpress | `NLIDAP40707162` (different resi) |
| R37 | label with a different resi | lemah | 6 | 1 / 1 | SPX Express | `SPXID5691455256` (different resi) |
| R38 | label with a different resi | lemah | 4 | 1 / 1 | J&T Express | `JP5323818140` (different resi) |
| R39 | label with a different resi | lemah | 5 | 2 / 2 | JNE | `16461569763934` (different resi) x2 |
| R40 | label with a different resi | lemah | 3 | 1 / 1 | SiCepat | `002614465563` (different resi) |
| R41 | label with a different resi | lemah | 5 | 1 / 1 | AnterAja | `100086918238302` (different resi) |
| R42 | item never shown | lemah | 3 | 1 / 1 | Ninja Xpress | `NLIDAP02435707` |
| R43 | item never shown | lemah | 4 | 1 / 1 | SPX Express | `SPXID2357655346` |
| R44 | item never shown | lemah | 3 | 1 / 1 | J&T Express | `JP0603620399` |
| R45 | item never shown | lemah | 4 | 1 / 1 | JNE | `16209183936407` |
| R46 | item never shown (empty box) | lemah | 4 | 1 / 1 | SiCepat | `000545533084` |
| R47 | label unreadable (faint) | perlu_dicek | 4 | 1 / 1 | AnterAja | `100024551317043` (faint) |
| R48 | label unreadable (smudge) | perlu_dicek | 4 | 1 / 1 | Ninja Xpress | `NLIDAP46824805` |
| R49 | label unreadable (far) | perlu_dicek | 5 | 1 / 1 | SPX Express | `SPXID6035155568` |
| R50 | label unreadable (fold) | perlu_dicek | 3 | 1 / 1 | J&T Express | `JP4391399192` |
| R51 | never sealed | perlu_dicek | 4 | 1 / 1 | JNE | `16682892392397` |
| R52 | never sealed | perlu_dicek | 3 | 1 / 1 | SiCepat | `008332694013` |
| R53 | never sealed | perlu_dicek | 6 | 1 / 1 | AnterAja | `100058430910112` |
| R54 | never sealed | perlu_dicek | 4 | 1 / 1 | Ninja Xpress | `NLIDAP20665438` |
| R55 | koli mismatch (record 2, 1 box shown) | perlu_dicek | 4 | 2 / 1 | SPX Express | `SPXID2255969683` |
| R56 | koli mismatch (record 3, 2 boxes shown) | perlu_dicek | 4 | 3 / 2 | J&T Express | `JP8462516675` x2 |
| R57 | koli mismatch (record 1, 2 boxes shown) | perlu_dicek | 4 | 1 / 2 | JNE | `16263696923681` x2 |
| R58 | koli mismatch (record 1, 3 boxes shown) | perlu_dicek | 5 | 1 / 3 | SiCepat | `009686108937` x3 |
| R59 | mixed: never sealed + koli mismatch | perlu_dicek | 4 | 2 / 1 | AnterAja | `100055719824119` |
| R60 | mixed: label unreadable (faint) + never sealed | perlu_dicek | 4 | 1 / 1 | Ninja Xpress | `NLIDAP55984903` (faint) |
| R61 | mixed: no label + never sealed | lemah | 4 | 1 / 1 | SPX Express | none |
| R62 | mixed: dark + koli mismatch | lemah | 4 | 2 / 1 | J&T Express | `JP3854585781` |
| R63 | mixed: different resi + item never shown | lemah | 3 | 1 / 1 | JNE | `16916597221586` (different resi) |
| R64 | mixed: blurry + never sealed | lemah | 4 | 1 / 1 | SiCepat | `004930549425` |

Records with more than 4 photos (selection test, decisive shots in the last two): R03, R05, R06, R07, R09, R11, R13, R16, R18, R22, R24, R27, R33, R34, R37, R39, R41, R49, R53, R58.

Records where `unsure` is the honest answer: R19, R25, R30 (3 of 64, 5%). D8 counts every record the model answers `unsure` on, even when unsure is correct, so these are kept few on purpose; see the note at the end.

## Records

### R01: good (kuat)

- **Record resi:** `SPXID5774637064`  **Courier:** SPX Express  **Koli:** 1  **Item:** a mug
- **Waybill:** stick the slip printed `SPXID5774637064`.
- **Photos:**
  01. A mug laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A mug inside the open box, shot top-down, contents clearly visible.
  03. The box fully taped, label on top, whole box in frame, label visible.
  04. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `SPXID5774637064`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **kuat**

### R02: good (kuat)

- **Record resi:** `JP3304452940`  **Courier:** J&T Express  **Koli:** 1  **Item:** a folded T-shirt
- **Waybill:** stick the slip printed `JP3304452940`.
- **Photos:**
  01. A folded T-shirt inside the open box, shot top-down, contents clearly visible.
  02. The box fully taped, label on top, whole box in frame, label visible.
  03. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `JP3304452940`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **kuat**

### R03: good (kuat)

- **Record resi:** `16394029077943`  **Courier:** JNE  **Koli:** 1  **Item:** a paperback book
- **Waybill:** stick the slip printed `16394029077943`.
- **Photos:**
  01. A paperback book laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A paperback book inside the open box, shot top-down, contents clearly visible.
  03. Same open box from a 45 degree angle, item still visible. [middle photo: not sent to the model]
  04. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  05. The box fully taped, label on top, whole box in frame, label visible.
  06. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `16394029077943`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **kuat**

### R04: good (kuat)

- **Record resi:** `004546157682`  **Courier:** SiCepat  **Koli:** 2  **Item:** a water bottle
- **Waybill:** stick the slip printed `004546157682` on each of the 2 boxes.
- **Photos:**
  01. A water bottle laid out on the table next to 2 open, empty boxes side by side. Good light, whole item in frame.
  02. A water bottle inside 2 open boxes side by side, shot top-down, contents clearly visible.
  03. All 2 boxes, side by side, fully taped, labels on top, whole boxes in frame, labels visible.
  04. Close-up of one label (it reads Koli 1/2): fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `004546157682`, package_sealed `yes`, item_visible `yes`, koli_seen `2` -> **kuat**

### R05: good (kuat)

- **Record resi:** `100094904108537`  **Courier:** AnterAja  **Koli:** 1  **Item:** a toy car
- **Waybill:** stick the slip printed `100094904108537`.
- **Photos:**
  01. A toy car laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A toy car inside the open box, shot top-down, contents clearly visible.
  03. Same open box from a 45 degree angle, item still visible. [middle photo: not sent to the model]
  04. Padding (bubble wrap or paper) around the item, box still open, item partly visible. [middle photo: not sent to the model]
  05. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  06. The box fully taped, from a 45 degree angle, tape visible across the top seam. [middle photo: not sent to the model]
  07. The box fully taped, label on top, whole box in frame, label visible.
  08. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `100094904108537`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **kuat**

### R06: good (kuat)

- **Record resi:** `NLIDAP24457712`  **Courier:** Ninja Xpress  **Koli:** 1  **Item:** three small cosmetics bottles
- **Waybill:** stick the slip printed `NLIDAP24457712`.
- **Photos:**
  01. Three small cosmetics bottles laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. Three small cosmetics bottles inside the open box, shot top-down, contents clearly visible.
  03. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  04. The box fully taped, label on top, whole box in frame, label visible.
  05. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `NLIDAP24457712`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **kuat**

### R07: good (kuat)

- **Record resi:** `SPXID7077071791`  **Courier:** SPX Express  **Koli:** 3  **Item:** a pair of sandals
- **Waybill:** stick the slip printed `SPXID7077071791` on each of the 3 boxes.
- **Photos:**
  01. A pair of sandals laid out on the table next to 3 open, empty boxes side by side. Good light, whole item in frame.
  02. A pair of sandals inside 3 open boxes side by side, shot top-down, contents clearly visible.
  03. Same open boxes from a 45 degree angle, item still visible. [middle photo: not sent to the model]
  04. Padding (bubble wrap or paper) around the item, boxes still open, item partly visible. [middle photo: not sent to the model]
  05. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  06. All 3 boxes, side by side, fully taped, from a 45 degree angle, tape visible across the top seam. [middle photo: not sent to the model]
  07. All 3 boxes, side by side, fully taped, labels on top, whole boxes in frame, labels visible.
  08. Close-up of one label (it reads Koli 1/3): fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `SPXID7077071791`, package_sealed `yes`, item_visible `yes`, koli_seen `3` -> **kuat**

### R08: good (kuat)

- **Record resi:** `JP7023650806`  **Courier:** J&T Express  **Koli:** 1  **Item:** a phone case in its retail box
- **Waybill:** stick the slip printed `JP7023650806`.
- **Photos:**
  01. A phone case in its retail box inside the open box, shot top-down, contents clearly visible.
  02. The box fully taped, label on top, whole box in frame, label visible.
  03. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `JP7023650806`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **kuat**

### R09: good (kuat)

- **Record resi:** `16957813511813`  **Courier:** JNE  **Koli:** 2  **Item:** a scarf
- **Waybill:** stick the slip printed `16957813511813` on each of the 2 boxes.
- **Photos:**
  01. A scarf laid out on the table next to 2 open, empty boxes side by side. Good light, whole item in frame.
  02. A scarf inside 2 open boxes side by side, shot top-down, contents clearly visible.
  03. Same open boxes from a 45 degree angle, item still visible. [middle photo: not sent to the model]
  04. Padding (bubble wrap or paper) around the item, boxes still open, item partly visible. [middle photo: not sent to the model]
  05. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  06. All 2 boxes, side by side, fully taped, labels on top, whole boxes in frame, labels visible.
  07. Close-up of one label (it reads Koli 1/2): fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `16957813511813`, package_sealed `yes`, item_visible `yes`, koli_seen `2` -> **kuat**

### R10: good (kuat)

- **Record resi:** `000601062432`  **Courier:** SiCepat  **Koli:** 1  **Item:** a pencil case
- **Waybill:** stick the slip printed `000601062432`.
- **Photos:**
  01. A pencil case laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A pencil case inside the open box, shot top-down, contents clearly visible.
  03. The box fully taped, label on top, whole box in frame, label visible.
  04. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `000601062432`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **kuat**

### R11: good (kuat)

- **Record resi:** `100090383112466`  **Courier:** AnterAja  **Koli:** 1  **Item:** a mug
- **Waybill:** stick the slip printed `100090383112466`.
- **Photos:**
  01. A mug laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A mug inside the open box, shot top-down, contents clearly visible.
  03. Same open box from a 45 degree angle, item still visible. [middle photo: not sent to the model]
  04. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  05. The box fully taped, label on top, whole box in frame, label visible.
  06. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `100090383112466`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **kuat**

### R12: good (kuat)

- **Record resi:** `NLIDAP06777400`  **Courier:** Ninja Xpress  **Koli:** 2  **Item:** a folded T-shirt
- **Waybill:** stick the slip printed `NLIDAP06777400` on each of the 2 boxes.
- **Photos:**
  01. A folded T-shirt laid out on the table next to 2 open, empty boxes side by side. Good light, whole item in frame.
  02. A folded T-shirt inside 2 open boxes side by side, shot top-down, contents clearly visible.
  03. All 2 boxes, side by side, fully taped, labels on top, whole boxes in frame, labels visible.
  04. Close-up of one label (it reads Koli 1/2): fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `NLIDAP06777400`, package_sealed `yes`, item_visible `yes`, koli_seen `2` -> **kuat**

### R13: good (kuat)

- **Record resi:** `SPXID0364861316`  **Courier:** SPX Express  **Koli:** 1  **Item:** a paperback book
- **Waybill:** stick the slip printed `SPXID0364861316`.
- **Photos:**
  01. A paperback book laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A paperback book inside the open box, shot top-down, contents clearly visible.
  03. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  04. The box fully taped, label on top, whole box in frame, label visible.
  05. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `SPXID0364861316`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **kuat**

### R14: dark (lemah)

- **Record resi:** `JP1417626322`  **Courier:** J&T Express  **Koli:** 1  **Item:** a water bottle
- **Waybill:** stick the slip printed `JP1417626322`.
- **Photos:**
  01. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. A water bottle laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. A water bottle inside the open box, shot top-down, contents clearly visible.
  03. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. The box fully taped, label on top, whole box in frame, label visible.
  04. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. Close-up of the label in the same dim light: the label's white rectangle is visible, the resi number is NOT legible.
- **Expected:** image_usable `dark`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R15: dark (lemah)

- **Record resi:** `16668362572636`  **Courier:** JNE  **Koli:** 1  **Item:** a toy car
- **Waybill:** stick the slip printed `16668362572636`.
- **Photos:**
  01. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. A toy car inside the open box, shot top-down, contents clearly visible.
  02. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. The box fully taped, label on top, whole box in frame, label visible.
  03. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. Close-up of the label in the same dim light: the label's white rectangle is visible, the resi number is NOT legible.
- **Expected:** image_usable `dark`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R16: dark (lemah)

- **Record resi:** `000749411473`  **Courier:** SiCepat  **Koli:** 2  **Item:** three small cosmetics bottles
- **Waybill:** stick the slip printed `000749411473` on each of the 2 boxes.
- **Photos:**
  01. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. Three small cosmetics bottles laid out on the table next to 2 open, empty boxes side by side. Good light, whole item in frame.
  02. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. Three small cosmetics bottles inside 2 open boxes side by side, shot top-down, contents clearly visible.
  03. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  04. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. All 2 boxes, side by side, fully taped, labels on top, whole boxes in frame, labels visible.
  05. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. Close-up of the label in the same dim light: the label's white rectangle is visible, the resi number is NOT legible.
- **Expected:** image_usable `dark`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `2` -> **lemah**

### R17: dark (lemah)

- **Record resi:** `100087397827811`  **Courier:** AnterAja  **Koli:** 1  **Item:** a pair of sandals
- **Waybill:** stick the slip printed `100087397827811`.
- **Photos:**
  01. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. A pair of sandals laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. A pair of sandals inside the open box, shot top-down, contents clearly visible.
  03. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. The box fully taped, label on top, whole box in frame, label visible.
  04. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. Close-up of the label in the same dim light: the label's white rectangle is visible, the resi number is NOT legible.
- **Expected:** image_usable `dark`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R18: dark (lemah)

- **Record resi:** `NLIDAP50717085`  **Courier:** Ninja Xpress  **Koli:** 1  **Item:** a phone case in its retail box
- **Waybill:** stick the slip printed `NLIDAP50717085`.
- **Photos:**
  01. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. A phone case in its retail box laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. A phone case in its retail box inside the open box, shot top-down, contents clearly visible.
  03. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. Same open box from a 45 degree angle, item still visible. [middle photo: not sent to the model]
  04. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  05. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. The box fully taped, label on top, whole box in frame, label visible.
  06. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. Close-up of the label in the same dim light: the label's white rectangle is visible, the resi number is NOT legible.
- **Expected:** image_usable `dark`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R19: dark (very) (lemah)

- **Record resi:** `SPXID8215646133`  **Courier:** SPX Express  **Koli:** 1  **Item:** a scarf
- **Waybill:** stick the slip printed `SPXID8215646133`.
- **Note:** Deliberate honest-unsure record: unsure is the correct answer for three fields.
- **Photos:**
  01. VERY DARK: lights off, curtains closed, no other light. You can barely tell there is a box at all. (Framing as for: A scarf laid out on the table next to the open, empty box. Good light, whole item in frame.)
  02. VERY DARK: lights off, curtains closed, no other light. You can barely tell there is a box at all. (Framing as for: A scarf inside the open box, shot top-down, contents clearly visible.)
  03. VERY DARK: lights off, curtains closed, no other light. You can barely tell there is a box at all. (Framing as for: The box fully taped, label on top, whole box in frame, label visible.)
  04. VERY DARK: lights off, curtains closed, no other light. You can barely tell there is a box at all. (Framing as for: Close-up of label: fill about half the frame, resi number sharp and readable.)
- **Expected:** image_usable `dark`, label_visible `unsure`, label_resi_read `(empty: null)`, package_sealed `unsure`, item_visible `unsure`, koli_seen `(empty: null)` -> **lemah**

### R20: blurry (lemah)

- **Record resi:** `JP7300356993`  **Courier:** J&T Express  **Koli:** 1  **Item:** a pencil case
- **Waybill:** stick the slip printed `JP7300356993`.
- **Photos:**
  01. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. A pencil case laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. A pencil case inside the open box, shot top-down, contents clearly visible.
  03. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. The box fully taped, label on top, whole box in frame, label visible.
  04. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. Close-up of the label, also swept: label visible, resi number NOT legible.
- **Expected:** image_usable `blurry`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R21: blurry (lemah)

- **Record resi:** `16040602833884`  **Courier:** JNE  **Koli:** 1  **Item:** a mug
- **Waybill:** stick the slip printed `16040602833884`.
- **Photos:**
  01. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. A mug inside the open box, shot top-down, contents clearly visible.
  02. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. The box fully taped, label on top, whole box in frame, label visible.
  03. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. Close-up of the label, also swept: label visible, resi number NOT legible.
- **Expected:** image_usable `blurry`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R22: blurry (lemah)

- **Record resi:** `008099837975`  **Courier:** SiCepat  **Koli:** 1  **Item:** a folded T-shirt
- **Waybill:** stick the slip printed `008099837975`.
- **Photos:**
  01. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. A folded T-shirt laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. A folded T-shirt inside the open box, shot top-down, contents clearly visible.
  03. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. Same open box from a 45 degree angle, item still visible. [middle photo: not sent to the model]
  04. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  05. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. The box fully taped, label on top, whole box in frame, label visible.
  06. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. Close-up of the label, also swept: label visible, resi number NOT legible.
- **Expected:** image_usable `blurry`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R23: blurry (lemah)

- **Record resi:** `100027213015412`  **Courier:** AnterAja  **Koli:** 2  **Item:** a paperback book
- **Waybill:** stick the slip printed `100027213015412` on each of the 2 boxes.
- **Photos:**
  01. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. A paperback book laid out on the table next to 2 open, empty boxes side by side. Good light, whole item in frame.
  02. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. A paperback book inside 2 open boxes side by side, shot top-down, contents clearly visible.
  03. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. All 2 boxes, side by side, fully taped, labels on top, whole boxes in frame, labels visible.
  04. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. Close-up of the label, also swept: label visible, resi number NOT legible.
- **Expected:** image_usable `blurry`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `2` -> **lemah**

### R24: blurry (lemah)

- **Record resi:** `NLIDAP02913539`  **Courier:** Ninja Xpress  **Koli:** 1  **Item:** a water bottle
- **Waybill:** stick the slip printed `NLIDAP02913539`.
- **Photos:**
  01. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. A water bottle laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. A water bottle inside the open box, shot top-down, contents clearly visible.
  03. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  04. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. The box fully taped, label on top, whole box in frame, label visible.
  05. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. Close-up of the label, also swept: label visible, resi number NOT legible.
- **Expected:** image_usable `blurry`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R25: blurry (heavy) (lemah)

- **Record resi:** `SPXID6121710922`  **Courier:** SPX Express  **Koli:** 1  **Item:** a toy car
- **Waybill:** stick the slip printed `SPXID6121710922`.
- **Note:** Deliberate honest-unsure record.
- **Photos:**
  01. HEAVY BLUR: fast sweep, so heavy that you cannot tell whether tape is on or what is in the box. (Framing as for: A toy car laid out on the table next to the open, empty box. Good light, whole item in frame.)
  02. HEAVY BLUR: fast sweep, so heavy that you cannot tell whether tape is on or what is in the box. (Framing as for: A toy car inside the open box, shot top-down, contents clearly visible.)
  03. HEAVY BLUR: fast sweep, so heavy that you cannot tell whether tape is on or what is in the box. (Framing as for: The box fully taped, label on top, whole box in frame, label visible.)
  04. HEAVY BLUR: fast sweep, so heavy that you cannot tell whether tape is on or what is in the box. (Framing as for: Close-up of the label, heavy sweep: a white rectangle, nothing legible.)
- **Expected:** image_usable `blurry`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `unsure`, item_visible `unsure`, koli_seen `1` -> **lemah**

### R26: obstructed (hand over label) (lemah)

- **Record resi:** `JP2360516641`  **Courier:** J&T Express  **Koli:** 1  **Item:** three small cosmetics bottles
- **Waybill:** stick the slip printed `JP2360516641`.
- **Photos:**
  01. Three small cosmetics bottles laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. Three small cosmetics bottles inside the open box, shot top-down, contents clearly visible.
  03. OBSTRUCTED: your hand is in front of the label in the last two photos. The box fully taped, label on top but covered by your flat hand. Whole box in frame. The label is never visible.
  04. OBSTRUCTED: your hand is in front of the label in the last two photos. Box top close-up with your hand still covering the label completely. Only the hand and tape are visible.
- **Expected:** image_usable `obstructed`, label_visible `no`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R27: obstructed (hand over label) (lemah)

- **Record resi:** `16969357560531`  **Courier:** JNE  **Koli:** 1  **Item:** a pair of sandals
- **Waybill:** stick the slip printed `16969357560531`.
- **Photos:**
  01. A pair of sandals laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A pair of sandals inside the open box, shot top-down, contents clearly visible.
  03. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  04. OBSTRUCTED: your hand is in front of the label in the last two photos. The box fully taped, label on top but covered by your flat hand. Whole box in frame. The label is never visible.
  05. OBSTRUCTED: your hand is in front of the label in the last two photos. Box top close-up with your hand still covering the label completely. Only the hand and tape are visible.
- **Expected:** image_usable `obstructed`, label_visible `no`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R28: obstructed (hand over label) (lemah)

- **Record resi:** `003660936850`  **Courier:** SiCepat  **Koli:** 1  **Item:** a phone case in its retail box
- **Waybill:** stick the slip printed `003660936850`.
- **Photos:**
  01. A phone case in its retail box inside the open box, shot top-down, contents clearly visible.
  02. OBSTRUCTED: your hand is in front of the label in the last two photos. The box fully taped, label on top but covered by your flat hand. Whole box in frame. The label is never visible.
  03. OBSTRUCTED: your hand is in front of the label in the last two photos. Box top close-up with your hand still covering the label completely. Only the hand and tape are visible.
- **Expected:** image_usable `obstructed`, label_visible `no`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R29: obstructed (bag in front) (lemah)

- **Record resi:** `100086502635260`  **Courier:** AnterAja  **Koli:** 1  **Item:** a scarf
- **Waybill:** stick the slip printed `100086502635260`.
- **Photos:**
  01. OBSTRUCTED: a plastic bag or cloth hangs in front of the lens and covers about half of every photo. A scarf laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. OBSTRUCTED: a plastic bag or cloth hangs in front of the lens and covers about half of every photo. A scarf inside the open box, shot top-down, contents clearly visible.
  03. OBSTRUCTED: a plastic bag or cloth hangs in front of the lens and covers about half of every photo. The box fully taped, label on top, whole box in frame, label visible.
  04. OBSTRUCTED: a plastic bag or cloth hangs in front of the lens and covers about half of every photo. Close-up of the label: the bag still covers the edge of the frame, but the resi number is fully visible and readable.
- **Expected:** image_usable `obstructed`, label_visible `yes`, label_resi_read `100086502635260`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R30: obstructed (hand over top) (lemah)

- **Record resi:** `NLIDAP60670202`  **Courier:** Ninja Xpress  **Koli:** 1  **Item:** a pencil case
- **Waybill:** stick the slip printed `NLIDAP60670202`.
- **Note:** Deliberate honest-unsure record: the seal is hidden, so unsure is correct.
- **Photos:**
  01. A pencil case laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A pencil case inside the open box, shot top-down, contents clearly visible.
  03. OBSTRUCTED: the label is stuck on the SIDE of the box. Your flat hand covers the whole top of the box, so you cannot tell if it is taped. Whole box in frame.
  04. Close-up of the label on the side of the box: resi number sharp and readable. Hand still on top, top seam not visible.
- **Expected:** image_usable `obstructed`, label_visible `yes`, label_resi_read `NLIDAP60670202`, package_sealed `unsure`, item_visible `yes`, koli_seen `1` -> **lemah**

### R31: no label (lemah)

- **Record resi:** `SPXID3949960018`  **Courier:** SPX Express  **Koli:** 1  **Item:** a mug
- **Waybill:** none. Do not stick any label on this box.
- **Photos:**
  01. A mug laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A mug inside the open box, shot top-down, contents clearly visible.
  03. The box fully taped. No label is stuck anywhere. Whole box in frame.
  04. The taped box from a 45 degree angle, top and two sides visible, no label anywhere.
- **Expected:** image_usable `ok`, label_visible `no`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R32: no label (lemah)

- **Record resi:** `JP8025991984`  **Courier:** J&T Express  **Koli:** 1  **Item:** a folded T-shirt
- **Waybill:** none. Do not stick any label on this box.
- **Photos:**
  01. A folded T-shirt inside the open box, shot top-down, contents clearly visible.
  02. The box fully taped, shot as if the label were underneath: no label is visible anywhere. Do not stick one on. Whole box in frame.
  03. The taped box from a 45 degree angle, top and two sides visible, still no label in frame.
- **Expected:** image_usable `ok`, label_visible `no`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R33: no label (lemah)

- **Record resi:** `16729649968129`  **Courier:** JNE  **Koli:** 1  **Item:** a paperback book
- **Waybill:** none. Do not stick any label on this box.
- **Photos:**
  01. A paperback book laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A paperback book inside the open box, shot top-down, contents clearly visible.
  03. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  04. The box fully taped. No label is stuck anywhere. Whole box in frame.
  05. The taped box from a 45 degree angle, top and two sides visible, no label anywhere.
- **Expected:** image_usable `ok`, label_visible `no`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R34: no label (lemah)

- **Record resi:** `000444751464`  **Courier:** SiCepat  **Koli:** 1  **Item:** a water bottle
- **Waybill:** none. Do not stick any label on this box.
- **Photos:**
  01. A water bottle laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A water bottle inside the open box, shot top-down, contents clearly visible.
  03. Same open box from a 45 degree angle, item still visible. [middle photo: not sent to the model]
  04. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  05. The box fully taped. No label is stuck anywhere. Whole box in frame.
  06. The taped box from a 45 degree angle, top and two sides visible, no label anywhere.
- **Expected:** image_usable `ok`, label_visible `no`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R35: no label (lemah)

- **Record resi:** `100097011564743`  **Courier:** AnterAja  **Koli:** 1  **Item:** a toy car
- **Waybill:** none. Do not stick any label on this box.
- **Photos:**
  01. A toy car laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A toy car inside the open box, shot top-down, contents clearly visible.
  03. The box fully taped, shot as if the label were underneath: no label is visible anywhere. Do not stick one on. Whole box in frame.
  04. The taped box from a 45 degree angle, top and two sides visible, still no label in frame.
- **Expected:** image_usable `ok`, label_visible `no`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R36: label with a different resi (lemah)

- **Record resi:** `NLIDAP40703162`  **Courier:** Ninja Xpress  **Koli:** 1  **Item:** three small cosmetics bottles
- **Waybill:** stick the slip printed `NLIDAP40707162`. This is deliberately NOT the record's resi (one digit different).
- **Photos:**
  01. Three small cosmetics bottles laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. Three small cosmetics bottles inside the open box, shot top-down, contents clearly visible.
  03. The box fully taped, label on top, whole box in frame, label visible.
  04. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `NLIDAP40707162`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R37: label with a different resi (lemah)

- **Record resi:** `SPXID5691405256`  **Courier:** SPX Express  **Koli:** 1  **Item:** a pair of sandals
- **Waybill:** stick the slip printed `SPXID5691455256`. This is deliberately NOT the record's resi (one digit different).
- **Photos:**
  01. A pair of sandals laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A pair of sandals inside the open box, shot top-down, contents clearly visible.
  03. Same open box from a 45 degree angle, item still visible. [middle photo: not sent to the model]
  04. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  05. The box fully taped, label on top, whole box in frame, label visible.
  06. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `SPXID5691455256`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R38: label with a different resi (lemah)

- **Record resi:** `JP5323808140`  **Courier:** J&T Express  **Koli:** 1  **Item:** a phone case in its retail box
- **Waybill:** stick the slip printed `JP5323818140`. This is deliberately NOT the record's resi (one digit different).
- **Photos:**
  01. A phone case in its retail box laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A phone case in its retail box inside the open box, shot top-down, contents clearly visible.
  03. The box fully taped, label on top, whole box in frame, label visible.
  04. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `JP5323818140`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R39: label with a different resi (lemah)

- **Record resi:** `16814311775001`  **Courier:** JNE  **Koli:** 2  **Item:** a scarf
- **Waybill:** stick the slip printed `16461569763934`. This is deliberately NOT the record's resi (a different number).
- **Photos:**
  01. A scarf laid out on the table next to 2 open, empty boxes side by side. Good light, whole item in frame.
  02. A scarf inside 2 open boxes side by side, shot top-down, contents clearly visible.
  03. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  04. All 2 boxes, side by side, fully taped, labels on top, whole boxes in frame, labels visible.
  05. Close-up of one label (it reads Koli 1/2): fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `16461569763934`, package_sealed `yes`, item_visible `yes`, koli_seen `2` -> **lemah**

### R40: label with a different resi (lemah)

- **Record resi:** `002322651203`  **Courier:** SiCepat  **Koli:** 1  **Item:** a pencil case
- **Waybill:** stick the slip printed `002614465563`. This is deliberately NOT the record's resi (a different number).
- **Photos:**
  01. A pencil case inside the open box, shot top-down, contents clearly visible.
  02. The box fully taped, label on top, whole box in frame, label visible.
  03. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `002614465563`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R41: label with a different resi (lemah)

- **Record resi:** `100053708050786`  **Courier:** AnterAja  **Koli:** 1  **Item:** a mug
- **Waybill:** stick the slip printed `100086918238302`. This is deliberately NOT the record's resi (a different number).
- **Photos:**
  01. A mug laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A mug inside the open box, shot top-down, contents clearly visible.
  03. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  04. The box fully taped, label on top, whole box in frame, label visible.
  05. Close-up of label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `100086918238302`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R42: item never shown (lemah)

- **Record resi:** `NLIDAP02435707`  **Courier:** Ninja Xpress  **Koli:** 1  **Item:** a folded T-shirt
- **Waybill:** stick the slip printed `NLIDAP02435707`.
- **Photos:**
  01. The box ALREADY fully taped (first photo; never show it open), label on top, whole box in frame.
  02. The taped box from a 45 degree angle, tape across the top seam visible, label visible.
  03. Close-up of the label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `NLIDAP02435707`, package_sealed `yes`, item_visible `no`, koli_seen `1` -> **lemah**

### R43: item never shown (lemah)

- **Record resi:** `SPXID2357655346`  **Courier:** SPX Express  **Koli:** 1  **Item:** a paperback book
- **Waybill:** stick the slip printed `SPXID2357655346`.
- **Photos:**
  01. The box ALREADY fully taped (first photo; never show it open), label on top, whole box in frame.
  02. The taped box from a 45 degree angle, tape across the top seam visible, label visible.
  03. Close-up of the label: fill about half the frame, resi number sharp and readable.
  04. The taped box from the other side, label still visible.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `SPXID2357655346`, package_sealed `yes`, item_visible `no`, koli_seen `1` -> **lemah**

### R44: item never shown (lemah)

- **Record resi:** `JP0603620399`  **Courier:** J&T Express  **Koli:** 1  **Item:** a water bottle
- **Waybill:** stick the slip printed `JP0603620399`.
- **Photos:**
  01. The box ALREADY fully taped (first photo; never show it open), label on top, whole box in frame.
  02. The taped box from a 45 degree angle, tape across the top seam visible, label visible.
  03. Close-up of the label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `JP0603620399`, package_sealed `yes`, item_visible `no`, koli_seen `1` -> **lemah**

### R45: item never shown (lemah)

- **Record resi:** `16209183936407`  **Courier:** JNE  **Koli:** 1  **Item:** a toy car
- **Waybill:** stick the slip printed `16209183936407`.
- **Photos:**
  01. The box ALREADY fully taped (first photo; never show it open), label on top, whole box in frame.
  02. The taped box from a 45 degree angle, tape across the top seam visible, label visible.
  03. Close-up of the label: fill about half the frame, resi number sharp and readable.
  04. The taped box from the other side, label still visible.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `16209183936407`, package_sealed `yes`, item_visible `no`, koli_seen `1` -> **lemah**

### R46: item never shown (empty box) (lemah)

- **Record resi:** `000545533084`  **Courier:** SiCepat  **Koli:** 1  **Item:** three small cosmetics bottles
- **Waybill:** stick the slip printed `000545533084`.
- **Note:** Empty box shipped: the 'paket kosong' dispute case.
- **Photos:**
  01. The open box, completely EMPTY, top-down. Nothing in it.
  02. Flaps folded on the empty box, tape going on.
  03. The box fully taped, label on top, whole box in frame, label visible.
  04. Close-up of the label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `000545533084`, package_sealed `yes`, item_visible `no`, koli_seen `1` -> **lemah**

### R47: label unreadable (faint) (perlu_dicek)

- **Record resi:** `100024551317043`  **Courier:** AnterAja  **Koli:** 1  **Item:** a pair of sandals
- **Waybill:** stick the FAINT slip printed `100024551317043` (light grey on purpose).
- **Photos:**
  01. A pair of sandals laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A pair of sandals inside the open box, shot top-down, contents clearly visible.
  03. The box fully taped, label on top, whole box in frame, label visible.
  04. Close-up of the FAINT label (printed light grey on purpose): the label is clearly there, but the resi number is NOT legible. Do not boost contrast.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **perlu_dicek**

### R48: label unreadable (smudge) (perlu_dicek)

- **Record resi:** `NLIDAP46824805`  **Courier:** Ninja Xpress  **Koli:** 1  **Item:** a phone case in its retail box
- **Waybill:** stick the slip printed `NLIDAP46824805`.
- **Photos:**
  01. A phone case in its retail box laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A phone case in its retail box inside the open box, shot top-down, contents clearly visible.
  03. The box fully taped, label on top, whole box in frame, label visible.
  04. Close-up of the label after you crumpled it and rubbed a wet thumb across the number before sticking: label visible, number NOT legible.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **perlu_dicek**

### R49: label unreadable (far) (perlu_dicek)

- **Record resi:** `SPXID6035155568`  **Courier:** SPX Express  **Koli:** 1  **Item:** a scarf
- **Waybill:** stick the slip printed `SPXID6035155568`.
- **Photos:**
  01. A scarf laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A scarf inside the open box, shot top-down, contents clearly visible.
  03. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  04. The box fully taped with the label on top, shot from about 1.5 m away so the whole table is in frame. The label is visible but tiny.
  05. The same box from 1.5 m at a slight angle. No close-up anywhere: the resi number must NOT be legible in any photo.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **perlu_dicek**

### R50: label unreadable (fold) (perlu_dicek)

- **Record resi:** `JP4391399192`  **Courier:** J&T Express  **Koli:** 1  **Item:** a pencil case
- **Waybill:** stick the slip printed `JP4391399192`.
- **Photos:**
  01. A pencil case inside the open box, shot top-down, contents clearly visible.
  02. The box fully taped, label on top, whole box in frame, label visible.
  03. Close-up of the label stuck FOLDED over the box edge, so the number wraps around the corner and no single view shows it whole. Number NOT legible.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **perlu_dicek**

### R51: never sealed (perlu_dicek)

- **Record resi:** `16682892392397`  **Courier:** JNE  **Koli:** 1  **Item:** a mug
- **Waybill:** stick the slip printed `16682892392397`.
- **Photos:**
  01. A mug laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A mug inside the open box, shot top-down, contents clearly visible.
  03. The box with its flaps STANDING OPEN and no tape at all, item visible inside. The label is stuck on the SIDE of the box, visible.
  04. Close-up of the label on the side of the open box: resi number sharp and readable. Flaps still open in the background.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `16682892392397`, package_sealed `no`, item_visible `yes`, koli_seen `1` -> **perlu_dicek**

### R52: never sealed (perlu_dicek)

- **Record resi:** `008332694013`  **Courier:** SiCepat  **Koli:** 1  **Item:** a folded T-shirt
- **Waybill:** stick the slip printed `008332694013`.
- **Photos:**
  01. A folded T-shirt inside the open box, shot top-down, contents clearly visible.
  02. The box with its flaps STANDING OPEN and no tape at all, item visible inside. The label is stuck on the SIDE of the box, visible.
  03. Close-up of the label on the side of the open box: resi number sharp and readable. Flaps still open in the background.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `008332694013`, package_sealed `no`, item_visible `yes`, koli_seen `1` -> **perlu_dicek**

### R53: never sealed (perlu_dicek)

- **Record resi:** `100058430910112`  **Courier:** AnterAja  **Koli:** 1  **Item:** a paperback book
- **Waybill:** stick the slip printed `100058430910112`.
- **Photos:**
  01. A paperback book laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A paperback book inside the open box, shot top-down, contents clearly visible.
  03. Same open box from a 45 degree angle, item still visible. [middle photo: not sent to the model]
  04. The open box from a 45 degree angle, flaps still standing, no tape anywhere. [middle photo: not sent to the model]
  05. The box with its flaps STANDING OPEN and no tape at all, item visible inside. The label is stuck on the SIDE of the box, visible.
  06. Close-up of the label on the side of the open box: resi number sharp and readable. Flaps still open in the background.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `100058430910112`, package_sealed `no`, item_visible `yes`, koli_seen `1` -> **perlu_dicek**

### R54: never sealed (perlu_dicek)

- **Record resi:** `NLIDAP20665438`  **Courier:** Ninja Xpress  **Koli:** 1  **Item:** a water bottle
- **Waybill:** stick the slip printed `NLIDAP20665438`.
- **Photos:**
  01. A water bottle laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A water bottle inside the open box, shot top-down, contents clearly visible.
  03. The box with its flaps STANDING OPEN and no tape at all, item visible inside. The label is stuck on the SIDE of the box, visible.
  04. Close-up of the label on the side of the open box: resi number sharp and readable. Flaps still open in the background.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `NLIDAP20665438`, package_sealed `no`, item_visible `yes`, koli_seen `1` -> **perlu_dicek**

### R55: koli mismatch (record 2, 1 box shown) (perlu_dicek)

- **Record resi:** `SPXID2255969683`  **Courier:** SPX Express  **Koli:** 2  **Item:** a toy car
- **Waybill:** stick the slip printed `SPXID2255969683`.
- **Boxes in frame:** 1 (the record says 2; the mismatch is the point).
- **Photos:**
  01. A toy car laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A toy car inside the open box, shot top-down, contents clearly visible.
  03. The box fully taped, label on top, whole box in frame, label visible.
  04. Close-up of label (it reads Koli 1/2): fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `SPXID2255969683`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **perlu_dicek**

### R56: koli mismatch (record 3, 2 boxes shown) (perlu_dicek)

- **Record resi:** `JP8462516675`  **Courier:** J&T Express  **Koli:** 3  **Item:** three small cosmetics bottles
- **Waybill:** stick the slip printed `JP8462516675` on each of the 2 boxes.
- **Boxes in frame:** 2 (the record says 3; the mismatch is the point).
- **Photos:**
  01. Three small cosmetics bottles laid out on the table next to 2 open, empty boxes side by side. Good light, whole item in frame.
  02. Three small cosmetics bottles inside 2 open boxes side by side, shot top-down, contents clearly visible.
  03. All 2 boxes, side by side, fully taped, labels on top, whole boxes in frame, labels visible.
  04. Close-up of one label (it reads Koli 1/3): fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `JP8462516675`, package_sealed `yes`, item_visible `yes`, koli_seen `2` -> **perlu_dicek**

### R57: koli mismatch (record 1, 2 boxes shown) (perlu_dicek)

- **Record resi:** `16263696923681`  **Courier:** JNE  **Koli:** 1  **Item:** a pair of sandals
- **Waybill:** stick the slip printed `16263696923681` on each of the 2 boxes.
- **Boxes in frame:** 2 (the record says 1; the mismatch is the point).
- **Photos:**
  01. A pair of sandals laid out on the table next to 2 open, empty boxes side by side. Good light, whole item in frame.
  02. A pair of sandals inside 2 open boxes side by side, shot top-down, contents clearly visible.
  03. All 2 boxes, side by side, fully taped, labels on top, whole boxes in frame, labels visible.
  04. Close-up of one label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `16263696923681`, package_sealed `yes`, item_visible `yes`, koli_seen `2` -> **perlu_dicek**

### R58: koli mismatch (record 1, 3 boxes shown) (perlu_dicek)

- **Record resi:** `009686108937`  **Courier:** SiCepat  **Koli:** 1  **Item:** a phone case in its retail box
- **Waybill:** stick the slip printed `009686108937` on each of the 3 boxes.
- **Boxes in frame:** 3 (the record says 1; the mismatch is the point).
- **Photos:**
  01. A phone case in its retail box laid out on the table next to 3 open, empty boxes side by side. Good light, whole item in frame.
  02. A phone case in its retail box inside 3 open boxes side by side, shot top-down, contents clearly visible.
  03. Flaps folded, tape going on. A hand in frame is fine. [middle photo: not sent to the model]
  04. All 3 boxes, side by side, fully taped, labels on top, whole boxes in frame, labels visible.
  05. Close-up of one label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `009686108937`, package_sealed `yes`, item_visible `yes`, koli_seen `3` -> **perlu_dicek**

### R59: mixed: never sealed + koli mismatch (perlu_dicek)

- **Record resi:** `100055719824119`  **Courier:** AnterAja  **Koli:** 2  **Item:** a scarf
- **Waybill:** stick the slip printed `100055719824119`.
- **Boxes in frame:** 1 (the record says 2; the mismatch is the point).
- **Photos:**
  01. A scarf laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A scarf inside the open box, shot top-down, contents clearly visible.
  03. The box with its flaps STANDING OPEN and no tape at all, item visible inside. The label is stuck on the SIDE of the box, visible.
  04. Close-up of the label on the side of the open box: resi number sharp and readable. Flaps still open in the background.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `100055719824119`, package_sealed `no`, item_visible `yes`, koli_seen `1` -> **perlu_dicek**

### R60: mixed: label unreadable (faint) + never sealed (perlu_dicek)

- **Record resi:** `NLIDAP55984903`  **Courier:** Ninja Xpress  **Koli:** 1  **Item:** a pencil case
- **Waybill:** stick the FAINT slip printed `NLIDAP55984903` (light grey on purpose).
- **Photos:**
  01. A pencil case laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A pencil case inside the open box, shot top-down, contents clearly visible.
  03. The box with its flaps STANDING OPEN and no tape at all, item visible inside. The label is stuck on the SIDE of the box, visible.
  04. Close-up of the label on the side of the open box: it is the FAINT label: resi number NOT legible. Flaps still open in the background.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `no`, item_visible `yes`, koli_seen `1` -> **perlu_dicek**

### R61: mixed: no label + never sealed (lemah)

- **Record resi:** `SPXID6146385167`  **Courier:** SPX Express  **Koli:** 1  **Item:** a mug
- **Waybill:** none. Do not stick any label on this box.
- **Photos:**
  01. A mug laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. A mug inside the open box, shot top-down, contents clearly visible.
  03. The box with its flaps STANDING OPEN and no tape at all, item visible inside. No label anywhere.
  04. The open box from the side: no label anywhere. Flaps still open in the background.
- **Expected:** image_usable `ok`, label_visible `no`, label_resi_read `(empty: null)`, package_sealed `no`, item_visible `yes`, koli_seen `1` -> **lemah**

### R62: mixed: dark + koli mismatch (lemah)

- **Record resi:** `JP3854585781`  **Courier:** J&T Express  **Koli:** 2  **Item:** a folded T-shirt
- **Waybill:** stick the slip printed `JP3854585781`.
- **Boxes in frame:** 1 (the record says 2; the mismatch is the point).
- **Photos:**
  01. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. A folded T-shirt laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. A folded T-shirt inside the open box, shot top-down, contents clearly visible.
  03. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. The box fully taped, label on top, whole box in frame, label visible.
  04. DARK: room lights off, only weak light from a doorway about 2 m away. Box, tape and item must still be recognisable as shapes. Close-up of the label in the same dim light: the label's white rectangle is visible, the resi number is NOT legible.
- **Expected:** image_usable `dark`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `yes`, item_visible `yes`, koli_seen `1` -> **lemah**

### R63: mixed: different resi + item never shown (lemah)

- **Record resi:** `16350319748525`  **Courier:** JNE  **Koli:** 1  **Item:** a paperback book
- **Waybill:** stick the slip printed `16916597221586`. This is deliberately NOT the record's resi (a different number).
- **Photos:**
  01. The box ALREADY fully taped (first photo; never show it open), label on top, whole box in frame.
  02. The taped box from a 45 degree angle, tape across the top seam visible, label visible.
  03. Close-up of the label: fill about half the frame, resi number sharp and readable.
- **Expected:** image_usable `ok`, label_visible `yes`, label_resi_read `16916597221586`, package_sealed `yes`, item_visible `no`, koli_seen `1` -> **lemah**

### R64: mixed: blurry + never sealed (lemah)

- **Record resi:** `004930549425`  **Courier:** SiCepat  **Koli:** 1  **Item:** a water bottle
- **Waybill:** stick the slip printed `004930549425`.
- **Photos:**
  01. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. A water bottle laid out on the table next to the open, empty box. Good light, whole item in frame.
  02. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. A water bottle inside the open box, shot top-down, contents clearly visible.
  03. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. The box with its flaps STANDING OPEN and no tape at all, item visible inside. The label is stuck on the SIDE of the box, visible.
  04. BLURRY: sweep the phone sideways as you press the shutter. Box, tape and item still recognisable as shapes. Close-up of the label on the side of the open box: resi number NOT legible. Flaps still open in the background.
- **Expected:** image_usable `blurry`, label_visible `yes`, label_resi_read `(empty: null)`, package_sealed `no`, item_visible `yes`, koli_seen `1` -> **lemah**

## Note on the D8 unsure target

The eval counts a record toward the 15 percent unsure ceiling whenever the model answers `unsure` on any field, including when `unsure` is the right answer. A set that honestly needed unsure on many records would fail D8 by construction. The degraded records above are therefore staged so a careful person can still answer definitely (dim but recognisable, smeared but recognisable), and only a few records need an honest unsure. Whether D8 should count only unsure-where-the-human-was-sure is a plan question, not something this set can settle.
