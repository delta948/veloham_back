import { api } from './client';
import type { Listing, PriceHistory } from '../types';

export type ListingFilters = {
  search?: string;
  category?: string;
  categories?: string;
  brand?: string;
  bike_type?: string;
  city?: string;
  condition?: string;
  frame_size?: string;
  rider_height?: string;
  deal_type?: string;
  labels?: string;
  label?: string;
  min_price?: string;
  max_price?: string;
  sort?: string;
  price_reduced?: boolean;
};

export const getListings = async (filters: ListingFilters = {}) => {
  const { data } = await api.get<Listing[]>('/listings', { params: filters });
  return data;
};

export const getListing = async (id: string) => {
  const { data } = await api.get<Listing>(`/listings/${id}`);
  return data;
};

export const getPriceHistory = async (id: string) => {
  const { data } = await api.get<PriceHistory>(`/listings/${id}/price-history`);
  return data;
};

export const saveListing = async (form: FormData, id?: string) => {
  const { data } = id
    ? await api.put<Listing>(`/listings/${id}`, form)
    : await api.post<Listing>('/listings', form);
  return id && 'listing' in (data as unknown as Record<string, unknown>) ? (data as unknown as { listing: Listing }).listing : data;
};
