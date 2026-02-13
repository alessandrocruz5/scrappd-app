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

# Deploying to cloud
docker build --platform linux/amd64 -t scrappd-api:v6 .
docker tag scrappd-api:v6 asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-api:v6
docker push asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-api:v6
gcloud run deploy scrappd-ml \
  --image=asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-ml:v5 \
  --region=asia-southeast1 \
  --platform=managed \
  --memory=16Gi \
  --cpu=4 \
  --timeout=300 \
  --concurrency=1 \
  --min-instances=0 \
  --max-instances=3 \
  --no-allow-unauthenticated \
  --set-env-vars="ENVIRONMENT=production"