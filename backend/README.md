## Deploy
Make sure ml is deployed first if there are changes

ML_SERVICE_URL=$(gcloud run services describe scrappd-ml \
  --region=asia-southeast1 \
  --format='value(status.url)')
echo "ML Service URL: $ML_SERVICE_URL"

PROJECT_NUMBER=$(gcloud projects describe scrappd-prod --format='value(projectNumber)')
gcloud run services add-iam-policy-binding scrappd-ml \
  --region=asia-southeast1 \
  --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
  --role="roles/run.invoker"

  cd ../backend
# Add the idtoken dependency
go get google.golang.org/api/idtoken
# Rebuild with updated ml_client.go
# Build/push new image (bump vN)
docker build --platform linux/amd64 -t scrappd-api:v10 .
docker tag scrappd-api:v10 asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-api:v10
docker push asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-api:v10

gcloud run deploy scrappd-api \
--image=asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-api:v10 \
--region=asia-southeast1 \
--platform=managed \
--memory=512Mi \
--cpu=1 \
--allow-unauthenticated \
--set-secrets="DATABASE_URL=DATABASE_URL:latest,JWT_ACCESS_SECRET=JWT_ACCESS_SECRET:latest,JWT_REFRESH_SECRET=JWT_REFRESH_SECRET:latest,REDIS_URL=REDIS_URL:latest,STORAGE_ACCESS_KEY_ID=R2_ACCESS_KEY_ID:latest,STORAGE_SECRET_ACCESS_KEY=R2_SECRET_ACCESS_KEY:latest,STORAGE_BUCKET_NAME=R2_BUCKET_NAME:latest,STORAGE_ENDPOINT=R2_ENDPOINT:latest" \
--set-env-vars="ML_SERVICE_URL=${ML_SERVICE_URL},ENVIRONMENT=production,BYPASS_USAGE_LIMITS=true"