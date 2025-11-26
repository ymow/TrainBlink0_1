export interface Station {
  id: string;
  place_id: string;
  name: string;
  name_en: string;
  type: 'TRA' | 'MRT' | 'HSR';
  latitude: number;
  longitude: number;
  radius: number;
  address: string;
  city: string;
  country: string;
  lines: string[];
}

export interface User {
  id: string;
  anonymous_id: string;
  station?: Station;
  connected_at: string;
}

export interface Message {
  id: string;
  user_id: string;
  content: string;
  timestamp: string;
  room_id?: string;
}

export interface ChatRoom {
  id: string;
  station_id: string;
  name: string;
  user_count: number;
  created_at: string;
}

export interface ServerStats {
  total_stations: number;
  active_sessions: number;
  total_messages: number;
  online_users: number;
}