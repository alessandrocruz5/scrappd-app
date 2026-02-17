# Scrappd ML Service

Background removal service using BiRefNet model.

## Quick Start with Makefile
```bash
# Setup (first time only)
make setup
source venv/bin/activate
make install

# Development
make dev          # Run with auto-reload
make test         # Test the service
make test-image IMG=path/to/image.jpg  # Test with specific image

# Docker
make docker-build    # Build image
make docker-run      # Run container
make docker-logs     # View logs
make docker-test     # Test Dockerized service
make docker-stop     # Stop container

# Cleanup
make clean          # Clean generated files
make clean-all      # Clean everything including venv
```

## Common Commands
```bash
make help           # Show all available commands
make info           # Show service information
```

## Setup
```bash
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

## Run
```bash
uvicorn app.main:app --reload --host 0.0.0.0 --port 8000
```

## Test
```bash
# In another terminal
python test_service.py path/to/test/image.jpg
```

## API Endpoints

- `GET /` - Health check
- `GET /health` - Detailed health check
- `POST /process` - Remove background from image
- `GET /stats` - Service statistics

## Example Usage
```bash
curl -X POST "http://localhost:8000/process" \
  -F "file=@test.jpg" \
  --output result.png
```

## Deploy (CPU-only)

```bash
cd scrappd-ml-service

# 1. Build image
# Note: docker buildx is not required on this machine.
docker build --platform linux/amd64 \
  -t asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-ml:v13 .

# 2. Push to Artifact Registry
docker push asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-ml:v13

# 3. Deploy Cloud Run (CPU-only)
gcloud run deploy scrappd-ml \
  --project=scrappd-prod \
  --region=asia-southeast1 \
  --image=asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-ml:v13 \
  --platform=managed \
  --cpu=8 \
  --memory=8Gi \
  --gpu=0 \
  --timeout=300 \
  --concurrency=2 \
  --min-instances=1 \
  --max-instances=3 \
  --no-allow-unauthenticated \
  --cpu-boost \
  --set-env-vars="ENVIRONMENT=production,MODEL_NAME=birefnet-general-lite"

# 4. Verify resources (must NOT include nvidia.com/gpu)
gcloud run services describe scrappd-ml \
  --project=scrappd-prod \
  --region=asia-southeast1 \
  --format='yaml(spec.template.spec.containers[0].resources.limits,spec.template.spec.containerConcurrency,status.latestReadyRevisionName)'
```

## Notes

- This service is private (`--no-allow-unauthenticated`), so direct `curl` without auth returns `403`.
- For authenticated health checks:

```bash
TOKEN=$(gcloud auth print-identity-token)
curl -H "Authorization: Bearer ${TOKEN}" \
  https://scrappd-ml-755228083251.asia-southeast1.run.app/health
```

- To change model without a full redeploy:

```bash
gcloud run services update scrappd-ml \
  --project=scrappd-prod \
  --region=asia-southeast1 \
  --update-env-vars="MODEL_NAME=birefnet-general-lite"
```
