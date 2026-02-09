from rembg import remove, new_session
from PIL import Image, ImageOps
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
        # Apply EXIF orientation to avoid unexpected large decode paths
        input_image = ImageOps.exif_transpose(input_image)

        # Convert to RGB to keep the model path consistent
        if input_image.mode not in ("RGB", "RGBA"):
            input_image = input_image.convert("RGB")

        # Downscale large images for speed
        max_dim = settings.max_dim
        max_pixels = settings.max_pixels
        w, h = input_image.size
        if (max_dim and max_dim > 0) or (max_pixels and max_pixels > 0):
            scale_dim = 1.0
            scale_px = 1.0
            if max_dim and max_dim > 0:
                scale_dim = min(1.0, max_dim / max(w, h))
            if max_pixels and max_pixels > 0:
                scale_px = min(1.0, (max_pixels / (w * h)) ** 0.5)
            scale = min(scale_dim, scale_px)
            if scale < 1.0:
                new_w = max(1, int(w * scale))
                new_h = max(1, int(h * scale))
                input_image = input_image.resize((new_w, new_h), resample=Image.BILINEAR)
        
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
