import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export interface HealthStatus {
  status: string;
  timestamp: string;
  uptime: number;
  connections: number;
  total_requests: number;
}

export interface ConnectionsResponse {
  count: number;
  client_ids: string[];
  timestamp: string;
}

export interface MessageRequest {
  client_id?: string;
  message: string;
  type?: string;
}

export interface HelloResponse {
  message: string;
  timestamp: string;
  server_info: {
    version: string;
    stage: string;
  };
}

export const apiService = {
  // Health check endpoints
  async ping(): Promise<{ message: string; timestamp: string; server: string }> {
    const response = await api.get('/ping');
    return response.data;
  },

  async getHealth(): Promise<HealthStatus> {
    const response = await api.get('/health');
    return response.data;
  },

  // API endpoints
  async getHello(): Promise<HelloResponse> {
    const response = await api.get('/api/v1/hello');
    return response.data;
  },

  async getWelcome(): Promise<any> {
    const response = await api.get('/api/v1/welcome');
    return response.data;
  },

  // Connections
  async getConnections(): Promise<ConnectionsResponse> {
    const response = await api.get('/api/v1/connections');
    return response.data;
  },

  // Send message via POST
  async sendMessage(messageData: MessageRequest): Promise<any> {
    const response = await api.post('/api/v1/message', messageData);
    return response.data;
  },

  // Server info (aggregated from ping response)
  async getServerInfo(): Promise<{ version: string; message: string; endpoints: Record<string, string> }> {
    const response = await api.get('/');
    return response.data;
  },
};