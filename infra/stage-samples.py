#!/usr/bin/env python3
"""Stage the demo's sample parcels (Plan 0016 D12).

Copies the chosen records' photos from the synthetic eval set and writes
samples.json with each record's labels and the Gemini observation the
Phase 0 eval recorded for it. Staged photos only: dummy waybills, no
real buyer data.

    stage-samples.py <eval-set-dir> <results.jsonl> <out-dir>
"""
import csv, json, os, shutil, sys

SAMPLES = [
    ("R01", "Good evidence"),
    ("R14", "Too dark"),
    ("R20", "Blurry"),
    ("R26", "Label covered"),
    ("R36", "Wrong label on the parcel"),
    ("R42", "Item never shown"),
    ("R51", "Box not shown sealed"),
    ("R55", "Two parcels, one shown"),
]

def main(evalset, results, out):
    labels = {r["record_id"]: r for r in csv.DictReader(open(os.path.join(evalset, "labels.csv")))}
    recorded = {}
    for line in open(results):
        d = json.loads(line)
        recorded[d["record_id"]] = d
    os.makedirs(out, exist_ok=True)
    samples = []
    for rid, title in SAMPLES:
        lab, rec = labels[rid], recorded[rid]
        src = os.path.join(evalset, rid)
        photos = sorted(f for f in os.listdir(src) if f.lower().endswith((".jpg", ".jpeg")))
        os.makedirs(os.path.join(out, rid), exist_ok=True)
        for f in photos:
            shutil.copyfile(os.path.join(src, f), os.path.join(out, rid, f))
        samples.append({
            "id": rid, "title": title, "resi": lab["resi"], "koli": int(lab["koli"]),
            "courier": lab["courier"], "expected": lab["human_verdict"], "photos": photos,
            "recorded": {"observation": rec["observation"], "model": rec["model"],
                         "prompt_version": rec["prompt_version"]},
        })
    json.dump(samples, open(os.path.join(out, "samples.json"), "w"), indent=1)
    print(f"staged {len(samples)} samples into {out}")

if __name__ == "__main__":
    if len(sys.argv) != 4:
        sys.exit(__doc__)
    main(*sys.argv[1:])
