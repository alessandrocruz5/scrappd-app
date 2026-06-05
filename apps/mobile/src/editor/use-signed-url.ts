// Signs a private 'cutouts' Storage key into a temporary URL for display.
// Mirrors CutoutThumb's signing effect but returns the raw URL so the editor
// canvas can feed it to either an <Image> or a Skia image loader.

import { useEffect, useState } from 'react';

import { signCutoutUrl } from '@/lib/storage';

export function useSignedUrl(key: string | null | undefined): string | null {
  const [url, setUrl] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    const pending = key ? signCutoutUrl(key) : Promise.resolve(null);
    pending.then((signed) => {
      if (active) setUrl(signed);
    });
    return () => {
      active = false;
    };
  }, [key]);

  return url;
}
