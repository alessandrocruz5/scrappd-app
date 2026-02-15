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

## Deploy

cd scrappd-ml-service
docker build -t scrappd-ml:v9 .

# 2. Tag and push to Artifact Registry
docker tag scrappd-ml:v9 asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-ml:v9
docker push asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-ml:v9

# 3. Deploy ML Service (GPU-accelerated)
gcloud run deploy scrappd-ml \
  --image=asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-ml:v9 \
  --region=asia-southeast1 \
  --platform=managed \
  --memory=16Gi \
  --cpu=4 \
  --gpu=1 \
  --gpu-type=nvidia-l4 \
  --timeout=300 \
  --concurrency=4 \
  --min-instances=1 \
  --max-instances=1 \
  --no-allow-unauthenticated \
  --no-gpu-zonal-redundancy \
  --set-env-vars="ENVIRONMENT=production"

### GPU Deployment Notes
- GPU instances require `--gpu=1 --gpu-type=nvidia-l4`
- L4 GPUs on Cloud Run are available in: us-central1, europe-west4, asia-southeast1
- GPU instances have longer cold starts (~20-60s) but dramatically faster inference
- With GPU, concurrency can be increased from 1 to 4+
- Consider `--min-instances=1` in production to avoid GPU cold starts (at additional cost)