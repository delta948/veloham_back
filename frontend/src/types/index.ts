export type User = {
  id: string;
  username: string;
  email?: string;
  phone?: string;
  avatar_url?: string;
  city?: string;
  contact?: string;
  role: 'user' | 'admin';
  is_blocked: boolean;
  blocked_reason?: string;
  rating: number;
  created_at: string;
};

export type ListingImage = {
  id: string;
  listing_id: string;
  image_url: string;
  sort_order: number;
};

export type Listing = {
  id: string;
  user_id: string;
  user: User;
  title: string;
  description: string;
  price: number;
  initial_price: number;
  previous_price?: number;
  last_price_change_at?: string;
  city: string;
  brand?: string;
  category: string;
  bike_type?: string;
  condition: string;
  frame_size?: string;
  frame_size_text?: string;
  rider_height_min?: number;
  rider_height_max?: number;
  recommended_height_min?: number;
  recommended_height_max?: number;
  deal_type: string;
  labels: string[];
  is_urgent?: boolean;
  is_bargain?: boolean;
  is_exchange?: boolean;
  extra_payment_from_me?: boolean;
  extra_payment_from_buyer?: boolean;
  status: 'active' | 'sold' | 'hidden';
  views: number;
  images: ListingImage[];
  build_card?: BuildCard;
  match_preference?: MatchPreference;
  created_at: string;
  updated_at: string;
};

export type PriceHistoryItem = { old_price: number; new_price: number; change_amount: number; change_percent: number; changed_at: string };
export type PriceHistory = {
  listing_id: string; initial_price: number; current_price: number; minimum_price: number; maximum_price: number;
  minimum_price_30_days: number; total_change: number; total_change_percent: number; history: PriceHistoryItem[];
};

export type Notification = { id: string; listing_id: string; type: string; message: string; link: string; is_read: boolean; created_at: string };

export type AdminPriceHistory = { id: string; listing_id: string; old_price: number; new_price: number; changed_at: string; changed_by: string; ip_address: string; suspicious: boolean; suspicious_reason?: string; change_count: number; listing: Listing; changed_by_user: User };
export type UserBlockEvent = { id: string; user_id: string; admin_id: string; action: 'blocked' | 'unblocked'; reason: string; created_at: string; user: User; admin: User };
export type ListingPlacement = { id: string; user_id: string; listing_id?: string; kind: string; amount: number; currency: string; status: string; provider?: string; provider_payment_id?: string; created_at: string; paid_at?: string };
export type AdminStats = { users: number; listings: number; reports: number; blocked_users: number; pending_payments: number; paid_payments: number };

export type BuildCard = {
  id?: string;
  listing_id?: string;
  frame?: string;
  size?: string;
  fork?: string;
  wheels?: string;
  hubs?: string;
  tires?: string;
  handlebar?: string;
  stem?: string;
  saddle?: string;
  cranks?: string;
  bottom_bracket?: string;
  chain?: string;
  cog?: string;
  brakes?: string;
  weight?: string;
  frame_condition?: string;
  defects?: string;
  documents?: boolean;
};

export type MatchPreference = {
  id?: string;
  listing_id?: string;
  exchange_enabled: boolean;
  wants?: string;
  categories?: string;
  min_price?: number;
  max_price?: number;
  can_add_cash?: boolean;
  max_cash_add?: number;
  same_city_only?: boolean;
};

export type Favorite = {
  id: string;
  user_id: string;
  listing_id: string;
  listing: Listing;
  created_at: string;
};

export type Chat = {
  id: string;
  buyer_id: string;
  seller_id: string;
  listing_id: string;
  buyer: User;
  seller: User;
  listing: Listing;
  created_at: string;
};

export type Message = {
  id: string;
  chat_id: string;
  sender_id: string;
  sender: User;
  text: string;
  created_at: string;
  is_read: boolean;
};

export type Review = {
  id: string;
  seller_id: string;
  author_id: string;
  listing_id: string;
  rating: number;
  text: string;
  author: User;
  created_at: string;
};

export type Report = {
  id: string;
  reporter_id: string;
  listing_id?: string;
  seller_id?: string;
  reason: string;
  text?: string;
  status: string;
  created_at: string;
  reporter?: User;
  listing?: Listing;
};

export type WantedRequest = {
  id: string;
  user_id: string;
  user: User;
  title: string;
  category: string;
  min_budget: number;
  max_budget: number;
  budget_min?: number;
  budget_max?: number;
  city: string;
  frame_size?: string;
  rider_height: number;
  height?: number;
  preferred_bike_type?: string;
  description?: string;
  status: 'active' | 'closed';
  offers?: WantedOffer[];
  created_at: string;
  updated_at: string;
};

export type WantedOffer = {
  id: string;
  wanted_id: string;
  seller_id: string;
  listing_id: string;
  message: string;
  seller: User;
  listing: Listing;
  created_at: string;
};
