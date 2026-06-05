import { CropperScreen } from '@/cropper/cropper-screen';

// The instant shape cropper is the app's centerpiece: pick a shape, aim it at a
// pattern, and snap a transparent cutout (client-side Skia). Implementation
// lives in src/cropper.
export default function CropperRoute() {
  return <CropperScreen />;
}
