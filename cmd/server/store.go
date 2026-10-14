package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	firestore "google.golang.org/api/firestore/v1"
	"google.golang.org/api/googleapi"
)

// liveStore keeps live Gemini results beyond one Cloud Run instance, so
// a fresh instance shows the latest live verdict for each sample at once
// instead of waiting on the model, and the page can list recent checks.
type liveStore interface {
	Latest(ctx context.Context, sampleID string) (checkResult, bool, error)
	Save(ctx context.Context, sampleID, title string, res checkResult) error
	Recent(ctx context.Context, n int) ([]historyEntry, error)
}

// historyEntry is one stored live check, for the page's history list.
type historyEntry struct {
	SampleID string      `json:"sample_id"`
	Title    string      `json:"title"`
	Result   checkResult `json:"result"`
}

// firestoreStore stores live results in Cloud Firestore through its REST
// API: live_latest/{sample} holds the newest result per sample, and
// live_history holds every live check. Each document keeps the result as
// JSON in one field, so the shape lives in one place (checkResult).
type firestoreStore struct {
	docs     *firestore.ProjectsDatabasesDocumentsService
	database string // projects/<p>/databases/(default)
}

func newFirestoreStore(ctx context.Context, project string) (*firestoreStore, error) {
	svc, err := firestore.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("firestore: %w", err)
	}
	return &firestoreStore{docs: svc.Projects.Databases.Documents, database: "projects/" + project + "/databases/(default)"}, nil
}

func (f *firestoreStore) root() string { return f.database + "/documents" }

func (f *firestoreStore) Latest(ctx context.Context, sampleID string) (checkResult, bool, error) {
	doc, err := f.docs.Get(f.root() + "/live_latest/" + sampleID).Context(ctx).Do()
	if err != nil {
		var gerr *googleapi.Error
		if errors.As(err, &gerr) && gerr.Code == 404 {
			return checkResult{}, false, nil
		}
		return checkResult{}, false, err
	}
	res, err := decodeResult(doc)
	return res, err == nil, err
}

func (f *firestoreStore) Save(ctx context.Context, sampleID, title string, res checkResult) error {
	raw, err := json.Marshal(res)
	if err != nil {
		return err
	}
	at := res.CheckedAt
	if at == "" {
		at = time.Now().UTC().Format(time.RFC3339)
	}
	fields := map[string]firestore.Value{
		"sample":    {StringValue: sampleID},
		"title":     {StringValue: title},
		"result":    {StringValue: string(raw)},
		"checkedAt": {TimestampValue: at},
	}
	if _, err := f.docs.Patch(f.root()+"/live_latest/"+sampleID, &firestore.Document{Fields: fields}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("save latest: %w", err)
	}
	if _, err := f.docs.CreateDocument(f.root(), "live_history", &firestore.Document{Fields: fields}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("save history: %w", err)
	}
	return nil
}

func (f *firestoreStore) Recent(ctx context.Context, n int) ([]historyEntry, error) {
	resp, err := f.docs.List(f.root(), "live_history").OrderBy("checkedAt desc").PageSize(int64(n)).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	out := []historyEntry{}
	for _, d := range resp.Documents {
		res, err := decodeResult(d)
		if err != nil {
			continue
		}
		out = append(out, historyEntry{SampleID: d.Fields["sample"].StringValue, Title: d.Fields["title"].StringValue, Result: res})
	}
	return out, nil
}

func decodeResult(doc *firestore.Document) (checkResult, error) {
	var res checkResult
	v, ok := doc.Fields["result"]
	if !ok {
		return res, fmt.Errorf("document has no result")
	}
	err := json.Unmarshal([]byte(v.StringValue), &res)
	return res, err
}
