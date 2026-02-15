from contextlib import asynccontextmanager
import threading

from fastapi import FastAPI, File, UploadFile, HTTPException
from fastapi.responses import Response

from app.config import settings
from app.models import HealthResponse
from app.processor import background_remover


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Load the ML model in a background thread so the server can accept
    connections (and pass Cloud Run TCP startup probes) immediately."""
    thread = threading.Thread(target=background_remover.load, daemon=True)
    thread.start()
    yield


app = FastAPI(
    title=settings.service_name,
    version=settings.version,
    debug=settings.debug,
    lifespan=lifespan,
)


@app.get("/", response_model=HealthResponse)
async def root():
    """Root endpoint - health check"""
    return {
        "status": "healthy" if background_remover.ready else "starting",
        "service": settings.service_name,
        "version": settings.version,
        "model": settings.model_name,
    }


@app.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy" if background_remover.ready else "starting",
        "service": settings.service_name,
        "version": settings.version,
        "model": settings.model_name,
    }


@app.post("/process")
async def remove_background(file: UploadFile = File(...)):
    """
    Remove background from uploaded image

    Args:
        file: Uploaded image file (JPEG or PNG)

    Returns:
        Processed image with transparent background (PNG)
    """
    if not background_remover.ready:
        raise HTTPException(
            status_code=503,
            detail="Model is still loading, please retry shortly",
        )

    # Validate file type
    if file.content_type not in settings.allowed_formats:
        raise HTTPException(
            status_code=400,
            detail=f"Invalid file type. Allowed: {', '.join(settings.allowed_formats)}",
        )

    # Read file
    image_bytes = await file.read()

    # Validate file size
    if len(image_bytes) > settings.max_image_size:
        raise HTTPException(
            status_code=400,
            detail=f"File too large. Max size: {settings.max_image_size / 1024 / 1024}MB",
        )

    try:
        # Process image
        processed_bytes, processing_time = background_remover.process_image(image_bytes)

        # Return processed image
        return Response(
            content=processed_bytes,
            media_type="image/png",
            headers={
                "X-Processing-Time": str(processing_time),
                "X-Original-Filename": file.filename,
            },
        )

    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Processing failed: {str(e)}",
        )


@app.get("/stats")
async def get_stats():
    """Get service statistics"""
    return {
        "model": settings.model_name,
        "model_ready": background_remover.ready,
        "max_image_size_mb": settings.max_image_size / 1024 / 1024,
        "allowed_formats": settings.allowed_formats,
    }
