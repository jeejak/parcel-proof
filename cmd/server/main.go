// Command evidencecheck-demo is the public prototype of the Plan 0016
// evidence quality check (D12), built for the AI Builder Cup 2026. It
// runs on Cloud Run and shows two checks side by side:
//
//   - the basic check: brightness, sharpness and the shipping barcode,
//     measured in Go with no model and no cost;
//   - the Gemini check, on bundled staged sample parcels only.
//
// Privacy rule (Plan 0016 D11, D12): Gemini sees only the staged
// samples, whose photos carry dummy waybills. A photo a visitor uploads
// gets the basic check only, is never stored, and is never sent to any
// model, because a real waybill carries a buyer's name and address and
// the free Gemini tier may use its inputs to improve Google's products.
//
// Configuration (environment):
//
//	PORT            listen port (Cloud Run sets it; default 8080)
//	SAMPLES_DIR     staged samples: samples.json + <id>/NN.jpg (default ./samples)
//	GEMINI_API_KEY  optional; with it, a sample's first check calls Gemini live
//	GEMINI_MODEL    default gemini-3.5-flash-lite
//	FIRESTORE_PROJECT optional; stores live results in Firestore so every
//	                instance can serve them, and lists recent live checks
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	var store liveStore
	if project := os.Getenv("FIRESTORE_PROJECT"); project != "" {
		fs, err := newFirestoreStore(context.Background(), project)
		if err != nil {
			logger.Error("firestore unavailable; live results stay in memory", "err", err)
		} else {
			store = fs
		}
	}
	srv, err := newServer(serverConfig{
		Store:      store,
		SamplesDir: envOr("SAMPLES_DIR", "samples"),
		APIKey:     os.Getenv("GEMINI_API_KEY"),
		Model:      envOr("GEMINI_MODEL", "gemini-3.5-flash-lite"),
		Logger:     logger,
	})
	if err != nil {
		logger.Error("demo failed to start", "err", err)
		os.Exit(1)
	}
	go srv.warm(context.Background())
	addr := ":" + envOr("PORT", "8080")
	logger.Info("evidencecheck demo listening", "addr", addr, "samples", len(srv.samples), "gemini_live", srv.live != nil, "firestore", store != nil)
	hs := &http.Server{
		Addr:              addr,
		Handler:           srv.routes(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      90 * time.Second,
	}
	if err := hs.ListenAndServe(); err != nil {
		logger.Error("demo stopped", "err", err)
		os.Exit(1)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
