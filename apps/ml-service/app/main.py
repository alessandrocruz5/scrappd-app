from fastapi import FastAPI, File, UploadFile, HTTPException
from fastapi.responses import Response
import time

from app.config import settings
from app.models import ProcessResponse, HealthResponse
from app.processor import background_remover

app = FastAPI(
    title=settings.service_name,
    version=settings.version,
    debug=settings.debug
)

@app.get("/", response_model=HealthResponse)
async def root():
    """Root endpoint - health check"""
    return {
        "status": "healthy",
        "service": settings.service_name,
        "version": settings.version,
        "model": settings.model_name
    }

@app.get("/health", response_model=HealthResponse)
async def health_check():
    """Health check endpoint"""
    return {
        "status": "healthy",
        "service": settings.service_name,
        "version": settings.version,
        "model": settings.model_name
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
    # Validate file type
    if file.content_type not in settings.allowed_formats:
        raise HTTPException(
            status_code=400,
            detail=f"Invalid file type. Allowed: {', '.join(settings.allowed_formats)}"
        )
    
    # Read file
    image_bytes = await file.read()
    
    # Validate file size
    if len(image_bytes) > settings.max_image_size:
        raise HTTPException(
            status_code=400,
            detail=f"File too large. Max size: {settings.max_image_size / 1024 / 1024}MB"
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
                "X-Original-Filename": file.filename
            }
        )
    
    except Exception as e:
        raise HTTPException(
            status_code=500,
            detail=f"Processing failed: {str(e)}"
        )

@app.get("/stats")
async def get_stats():
    """Get service statistics"""
    return {
        "model": settings.model_name,
        "max_image_size_mb": settings.max_image_size / 1024 / 1024,
        "allowed_formats": settings.allowed_formats
    }