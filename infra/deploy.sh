#!/usr/bin/env bash
# Deploy ParcelProof to Cloud Run.
#
#   infra/deploy.sh <gcp-project> [region]
#
# Needs: gcloud logged in, billing linked on <gcp-project> (Cloud Run,
# Cloud Build and Artifact Registry need it even inside the free tier),
# and optionally a Secret Manager secret `gemini-api-key`.
#
# Billing trap: create the Gemini API key in a separate Google Cloud
# project with NO billing account. A key from a billing-enabled project
# is on the paid tier and is charged. The free tier may use what is sent
# to improve Google's products, which is why the demo sends it staged
# sample photos only.
set -euo pipefail

project=${1:?usage: infra/deploy.sh <gcp-project> [region]}
region=${2:-asia-southeast2}
service=parcel-proof
root=$(git rev-parse --show-toplevel)

secret_flag=()
if gcloud secrets describe gemini-api-key --project "$project" >/dev/null 2>&1; then
  secret_flag=(--set-secrets GEMINI_API_KEY=gemini-api-key:latest)
else
  echo "no gemini-api-key secret: the demo serves the results the eval recorded" >&2
fi

gcloud run deploy "$service" \
  --project "$project" --region "$region" --source "$root" \
  --allow-unauthenticated --max-instances 2 --memory 512Mi --cpu 1 \
  --concurrency 20 --timeout 90 --cpu-boost \
  --set-env-vars GEMINI_MODEL=gemini-3.5-flash-lite \
  ${secret_flag[@]+"${secret_flag[@]}"}

url=$(gcloud run services describe "$service" --project "$project" --region "$region" --format 'value(status.url)')
echo "deployed: $url"
curl -fsS "$url/health" && echo
