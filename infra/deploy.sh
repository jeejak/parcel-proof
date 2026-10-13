#!/usr/bin/env bash
# Deploy ParcelProof to Cloud Run.
#
#   infra/deploy.sh <gcp-project> [region]
#
# Needs: gcloud logged in; billing linked on <gcp-project> (Cloud Run,
# Cloud Build and Artifact Registry need it even inside the free tier);
# the run, cloudbuild, artifactregistry and secretmanager APIs enabled;
# and optionally a Secret Manager secret `gemini-api-key` readable by the
# project's default compute service account.
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

export CLOUDSDK_CORE_DISABLE_PROMPTS=1

# First deploy on a new project: the source-deploy image repository, and
# the build role new projects no longer grant the default service account.
number=$(gcloud projects describe "$project" --format 'value(projectNumber)')
gcloud artifacts repositories describe cloud-run-source-deploy --project "$project" --location "$region" >/dev/null 2>&1 ||
  gcloud artifacts repositories create cloud-run-source-deploy --project "$project" --location "$region" --repository-format docker
gcloud projects add-iam-policy-binding "$project" --condition=None --quiet \
  --member "serviceAccount:${number}-compute@developer.gserviceaccount.com" --role roles/run.builder >/dev/null

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
