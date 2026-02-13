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
    # Performance controls
    max_dim: int = 1600  # Max width/height; set 0 to disable
    max_pixels: int = 2_000_000  # Max total pixels; set 0 to disable
    # Output controls
    trim_transparent: bool = True  # Crop transparent borders
    alpha_threshold: int = 5  # 0-255; pixels <= threshold are treated as transparent
    
    class Config:
        env_file = ".env"

settings = Settings()
