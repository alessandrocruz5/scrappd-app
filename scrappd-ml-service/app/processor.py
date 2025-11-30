from rembg import remove, new_session
from PIL import Image
import io
import time
from app.config import settings

class BackgroundRemover:
    def __init__(self):
        """Initialize the background remover with BiRefNet model"""
        print(f"Loading model: {settings.model_name}...")
        self.session = new_session(settings.model_name)
        print("✓ Model loaded successfully")
    
    def process_image(self, image_bytes: bytes) -> tuple[bytes, float]:
        """
        Remove background from image bytes
        
        Args:
            image_bytes: Input image as bytes
            
        Returns:
            tuple: (processed_image_bytes, processing_time)
        """
        start_time = time.time()
        
        # Open image from bytes
        input_image = Image.open(io.BytesIO(image_bytes))
        
        # Remove background
        output_image = remove(input_image, session=self.session)
        
        # Convert back to bytes
        output_buffer = io.BytesIO()
        output_image.save(output_buffer, format="PNG")
        output_bytes = output_buffer.getvalue()
        
        processing_time = time.time() - start_time
        
        return output_bytes, processing_time

# Global instance (loaded once when service starts)
background_remover = BackgroundRemover()