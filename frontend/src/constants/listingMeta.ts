export const DEAL_TYPES = ['продажа', 'обмен', 'продажа или обмен'] as const;

export const LISTING_LABELS = ['срочно', 'торг', 'обмен', 'с моей доплатой', 'с вашей доплатой'] as const;

export const labelTone = (label: string) => {
  switch (label) {
    case 'срочно':
      return 'bg-danger text-white';
    case 'торг':
      return 'bg-acid text-black';
    case 'обмен':
      return 'bg-white text-black';
    case 'с моей доплатой':
      return 'border border-acid text-acid';
    case 'с вашей доплатой':
      return 'border border-danger text-danger';
    default:
      return 'border border-white/15 text-white/70';
  }
};
