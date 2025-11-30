from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    # Service settings
    service_name: str = "Scrappd ML Service"
    version: str = "1.0.0"
    debug: bool = False
    
    # Model settings
    model_name: str = "birefnet-general"
    
    # Image processing settings
    max_image_size: int = 10 * 1024 * 1024  # 10MB
    allowed_formats: list = ["image/jpeg", "image/png", "image/jpg"]
    
    class Config:
        env_file = ".env"

settings = Settings()