export const CATEGORIES = [
  'Велосипеды целиком',
  'Рамы / фреймсеты',
  'Колёса',
  'Трансмиссия',
  'Руль и управление',
  'Тормоза',
  'Седло и посадка',
  'Аксессуары',
  'Экипировка',
  'Вело-услуги',
  'Вело-события'
];

export const BIKE_TYPES = ['Fixed Gear', 'Road Bike', 'MTB', 'BMX'];

export const CATEGORY_ROUTES: Record<string, string> = {
  'Велосипеды целиком': '/bikes',
  'Рамы / фреймсеты': '/framesets',
  'Колёса': '/wheels',
  'Трансмиссия': '/drivetrain',
  'Руль и управление': '/cockpit',
  'Тормоза': '/brakes',
  'Седло и посадка': '/fit',
  'Аксессуары': '/accessories',
  'Экипировка': '/gear',
  'Вело-услуги': '/services',
  'Вело-события': '/community'
};

export function categoryPath(category: string) {
  return CATEGORY_ROUTES[category] ?? `/market?category=${encodeURIComponent(category)}`;
}
