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
docker build --platform linux/amd64 -t scrappd-api:v14 .
docker tag scrappd-api:v14 asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-api:v14
docker push asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-api:v14

gcloud run deploy scrappd-api \
    --image=asia-southeast1-docker.pkg.dev/scrappd-prod/scrappd-repo/scrappd-api:v14 \
    --region=asia-southeast1 \
    --platform=managed \
    --memory=512Mi \
    --cpu=1 \
    --allow-unauthenticated \
    --set-secrets="DATABASE_URL=DATABASE_URL:latest,JWT_ACCESS_SECRET=JWT_ACCESS_SECRET:latest,JWT_REFRESH_SECRET=JWT_REFRESH_SECRET:latest,REDIS_URL=REDIS_URL:latest,STORAGE_ACCESS_KEY_ID=R2_ACCESS_KEY_ID:latest,STORAGE_SECRET_ACCESS_KEY=R2_SECRET_ACCESS_KEY:latest,STORAGE_BUCKET_NAME=R2_BUCKET_NAME:latest,STORAGE_ENDPOINT=R2_ENDPOINT:latest,INTERNAL_TASK_SECRET=INTERNAL_TASK_SECRET:latest" \
    --set-env-vars="ML_SERVICE_URL=${ML_SERVICE_URL},ENVIRONMENT=production,BYPASS_USAGE_LIMITS=true,CLOUD_TASKS_ENABLED=true,CLOUD_TASKS_PROJECT_ID=scrappd-
  prod,CLOUD_TASKS_LOCATION=asia-southeast1,CLOUD_TASKS_QUEUE_ID=scrappd-items,CLOUD_TASKS_SERVICE_URL=https://scrappd-api-j6bicsikba-
  as.a.run.app,CLOUD_TASKS_SERVICE_ACCOUNT_EMAIL=scrappd-tasks@scrappd-prod.iam.gserviceaccount.com"

gcloud run services update scrappd-api \
    --project=scrappd-prod \
    --region=asia-southeast1 \
    --set-env-vars=CLOUD_TASKS_ENABLED=true,CLOUD_TASKS_QUEUE_ID=scrappd-items,CLOUD_TASKS_SERVICE_URL=https://scrappd-api-j6bicsikba-as.a.run.app,CLOUD_TASKS_SERVICE_ACCOUNT_EMAIL=scrappd-tasks@scrappd-prod.iam.gserviceaccount.com,INTERNAL_TASK_SECRET=828c3aafd761e53755e4fd49dc02add3647164d262af77c4bf32346521f4c108

  curl -i -X POST https://scrappd-api-j6bicsikba-as.a.run.app/api/v1/items \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiM2E1ZWJmMjktMGRkMS00ZDI3LWEwNzAtYTY3MzUxMjQwN2Q0IiwiZW1haWwiOiJhbGVzc2FuZHJvcmFmYWVsY3J1ekBnbWFpbC5jb20iLCJ1c2VybmFtZSI6ImFwYWNydXoiLCJzdWJzY3JpcHRpb25fdGllciI6ImZyZWUiLCJpc3MiOiJzY3JhcHBkLWFwaSIsInN1YiI6IjNhNWViZjI5LTBkZDEtNGQyNy1hMDcwLWE2NzM1MTI0MDdkNCIsImV4cCI6MTc3MDczOTkyNywibmJmIjoxNzcwNzM5MDI3LCJpYXQiOjE3NzA3MzkwMjd9.Y2VTz10VifIphKHkVWzZLWd8fFD47P_yY9uTYOgXYpc" \
    -F "image=@/home/apa/Documents/Code/test_images/test1.jpg"

    curl -s -X POST https://scrappd-api-j6bicsikba-as.a.run.app/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"alessandrorafaelcruz@gmail.com","password":"test1234"}'

    gcloud run services update-traffic scrappd-api \
    --project=scrappd-prod \
    --region=asia-southeast1 \
    --to-revisions=scrappd-api-00031-bxc=100