# Scrappd ML Service

Background removal service using BiRefNet model.

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