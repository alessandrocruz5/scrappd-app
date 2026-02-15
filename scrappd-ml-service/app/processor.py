from rembg import remove, new_session
from PIL import Image, ImageOps
import io
import time
import threading
from app.config import settings

class BackgroundRemover:
    def __init__(self):
        """Initialize the background remover (lazy model loading)"""
        self.session = None
        self._lock = threading.Lock()
        self._loading = False
        self._loaded = False

    def load_model(self):
        """Load the model in the background"""
        self._loading = True
        print(f"Loading model: {settings.model_name}...")
        self.session = new_session(settings.model_name)
        self._loaded = True
        self._loading = False
        print("✓ Model loaded successfully")

    @property
    def is_ready(self) -> bool:
        return self._loaded
    
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

        # Ensure alpha channel for trimming
        if output_image.mode != "RGBA":
            output_image = output_image.convert("RGBA")

        # Trim transparent borders to shrink output dimensions
        if settings.trim_transparent:
            alpha = output_image.split()[-1]
            threshold = max(0, min(255, settings.alpha_threshold))
            if threshold > 0:
                alpha = alpha.point(lambda p: 255 if p > threshold else 0)
            bbox = alpha.getbbox()
            if bbox:
                output_image = output_image.crop(bbox)
        
        # Convert back to bytes
        output_buffer = io.BytesIO()
        output_image.save(output_buffer, format="PNG")
        output_bytes = output_buffer.getvalue()
        
        processing_time = time.time() - start_time
        
        return output_bytes, processing_time

# Global instance (model loaded asynchronously after server starts)
background_remover = BackgroundRemover()
