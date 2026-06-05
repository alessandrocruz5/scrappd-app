from pydantic import BaseModel

class ProcessResponse(BaseModel):
    success: bool
    message: str
    processing_time: float
    
class HealthResponse(BaseModel):
    status: str
    service: str
    version: str
    model: str