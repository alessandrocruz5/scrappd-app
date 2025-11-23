from rembg import remove, new_session
from PIL import Image
import os
import time

def compare_models():
    """
    Process all images with both U2Net (default) and BiRefNet for comparison
    """
    input_folder = "test_images"
    output_folder = "output"
    
    # Create subfolders for each model
    u2net_folder = os.path.join(output_folder, "u2net")
    birefnet_folder = os.path.join(output_folder, "birefnet")
    os.makedirs(u2net_folder, exist_ok=True)
    os.makedirs(birefnet_folder, exist_ok=True)
    
    # Get all image files
    image_files = [f for f in os.listdir(input_folder) 
                   if f.lower().endswith(('.png', '.jpg', '.jpeg'))]
    
    print(f"Found {len(image_files)} images to process\n")
    print("=" * 60)
    
    # Create BiRefNet session (downloads model on first run)
    print("Loading BiRefNet model...")
    birefnet_session = new_session("birefnet-general")
    print("✓ BiRefNet loaded\n")
    print("=" * 60)
    
    u2net_total = 0
    birefnet_total = 0
    
    for idx, filename in enumerate(image_files, 1):
        input_path = os.path.join(input_folder, filename)
        name_without_ext = os.path.splitext(filename)[0]
        
        print(f"\n[{idx}/{len(image_files)}] Processing: {filename}")
        print("-" * 60)
        
        # Open image once
        input_image = Image.open(input_path)
        
        # Test U2Net (default)
        print("  U2Net (default)...", end=" ", flush=True)
        start = time.time()
        u2net_output = remove(input_image)
        u2net_time = time.time() - start
        u2net_total += u2net_time
        u2net_path = os.path.join(u2net_folder, f"{name_without_ext}.png")
        u2net_output.save(u2net_path)
        print(f"✓ {u2net_time:.2f}s")
        
        # Test BiRefNet
        print("  BiRefNet...", end=" ", flush=True)
        start = time.time()
        birefnet_output = remove(input_image, session=birefnet_session)
        birefnet_time = time.time() - start
        birefnet_total += birefnet_time
        birefnet_path = os.path.join(birefnet_folder, f"{name_without_ext}.png")
        birefnet_output.save(birefnet_path)
        print(f"✓ {birefnet_time:.2f}s")
        
        print(f"  Speed difference: BiRefNet is {u2net_time/birefnet_time:.1f}x " + 
              ("faster" if birefnet_time < u2net_time else "slower"))
    
    print("\n" + "=" * 60)
    print("\n📊 FINAL RESULTS:")
    print(f"\nTotal images processed: {len(image_files)}")
    print(f"\nU2Net (default):")
    print(f"  Total time: {u2net_total:.2f}s")
    print(f"  Average: {u2net_total/len(image_files):.2f}s per image")
    print(f"\nBiRefNet:")
    print(f"  Total time: {birefnet_total:.2f}s")
    print(f"  Average: {birefnet_total/len(image_files):.2f}s per image")
    print(f"\n📁 Results saved to:")
    print(f"  U2Net: {u2net_folder}/")
    print(f"  BiRefNet: {birefnet_folder}/")
    print("\n💡 Now compare the outputs side-by-side!")

if __name__ == "__main__":
    compare_models()