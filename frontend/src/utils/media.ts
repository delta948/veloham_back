import { MEDIA_BASE } from '../api/client';

export function listingImageUrl(imageUrl?: string) {
  if (!imageUrl) {
    return 'https://images.unsplash.com/photo-1485965120184-e220f721d03e?q=80&w=1200&auto=format&fit=crop';
  }
  if (imageUrl.startsWith('http://') || imageUrl.startsWith('https://')) {
    return imageUrl;
  }
  return `${MEDIA_BASE}${imageUrl}`;
}
