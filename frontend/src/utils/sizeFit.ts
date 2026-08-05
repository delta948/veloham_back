export type BikeType = 'fixed' | 'road' | 'mtb' | 'bmx';

export const BIKE_TYPES: { value: BikeType; label: string }[] = [
  { value: 'fixed', label: 'Fixed Gear' },
  { value: 'road', label: 'Шоссе' },
  { value: 'mtb', label: 'MTB' },
  { value: 'bmx', label: 'BMX' }
];

export const FRAME_SIZE_OPTIONS: Record<BikeType, string[]> = {
  fixed: ['XS', '46', '47', '48', '49', 'S', '50', '51', '52', 'M', '53', '54', 'L', '55', '56', '57', 'XL', '58', '59', '60', 'XXL', '61+'],
  road: ['XS', '46', '47', '48', '49', 'S', '50', '51', '52', 'M', '53', '54', 'L', '55', '56', '57', 'XL', '58', '59', '60', 'XXL', '61+'],
  mtb: ['XS', 'S', 'M', 'L', 'XL'],
  bmx: ['18"', '19"', '20"', '20.5"', '20.75"', '21"', '21.25"+']
};

export type SizeRecommendation = {
  label: string;
  frameSizes: string[];
};

export function recommendFrameSize(height: number, bikeType: BikeType): SizeRecommendation | null {
  if (!Number.isFinite(height) || height < 120) return null;

  if (bikeType === 'fixed' || bikeType === 'road') {
    if (height < 150) return { label: 'XS / 46-49', frameSizes: ['XS', '46', '47', '48', '49'] };
    if (height <= 160) return { label: 'XS / 46-49', frameSizes: ['XS', '46', '47', '48', '49'] };
    if (height <= 168) return { label: 'S / 50-52', frameSizes: ['S', '50', '51', '52'] };
    if (height <= 176) return { label: 'M / 52-54', frameSizes: ['M', '52', '53', '54'] };
    if (height <= 184) return { label: 'L / 55-57', frameSizes: ['L', '55', '56', '57'] };
    if (height <= 192) return { label: 'XL / 58-60', frameSizes: ['XL', '58', '59', '60'] };
    return { label: 'XXL / 61+', frameSizes: ['XXL', '61+'] };
  }

  if (bikeType === 'mtb') {
    if (height < 150) return { label: 'XS', frameSizes: ['XS'] };
    if (height <= 160) return { label: 'XS', frameSizes: ['XS'] };
    if (height <= 170) return { label: 'S', frameSizes: ['S'] };
    if (height <= 180) return { label: 'M', frameSizes: ['M'] };
    if (height <= 190) return { label: 'L', frameSizes: ['L'] };
    return { label: 'XL', frameSizes: ['XL'] };
  }

  if (height < 150) return { label: '18-19"', frameSizes: ['18"', '19"'] };
  if (height <= 165) return { label: '20"', frameSizes: ['20"'] };
  if (height <= 175) return { label: '20.5"', frameSizes: ['20.5"'] };
  if (height <= 185) return { label: '20.75-21"', frameSizes: ['20.75"', '21"'] };
  return { label: '21.25"+', frameSizes: ['21.25"+'] };
}

export function bikeTypeLabel(value?: string) {
  return BIKE_TYPES.find((item) => item.value === value)?.label ?? value ?? '';
}
