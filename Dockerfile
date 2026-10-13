# ParcelProof demo for Cloud Run. `gcloud run deploy --source .` builds this.
FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/server /server
COPY samples /samples
ENV SAMPLES_DIR=/samples
USER nonroot:nonroot
ENTRYPOINT ["/server"]
