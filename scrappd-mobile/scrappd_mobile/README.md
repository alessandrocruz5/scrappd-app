# scrappd_mobile

A new Flutter project.

## Getting Started

This project is a starting point for a Flutter application.

A few resources to get you started if this is your first Flutter project:

- [Lab: Write your first Flutter app](https://docs.flutter.dev/get-started/codelab)
- [Cookbook: Useful Flutter samples](https://docs.flutter.dev/cookbook)

For help getting started with Flutter development, view the
[online documentation](https://docs.flutter.dev/), which offers tutorials,
samples, guidance on mobile development, and a full API reference.

## Deploying to Cloud

### ML Service (CPU-only)
```bash
cd ../scrappd-ml-service

docker build --platform linux/amd64 \
  -t asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-ml:v13 .
docker push asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-ml:v13

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
```

### API Service
```bash
cd ../backend

docker build --platform linux/amd64 \
  -t asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-api:v6 .
docker push asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-api:v6
```
