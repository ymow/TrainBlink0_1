import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiService, type MessageRequest } from '../services/api';

export function useServerHealth() {
  return useQuery({
    queryKey: ['health'],
    queryFn: apiService.getHealth,
    refetchInterval: 5000, // Poll every 5 seconds
    retry: 3,
  });
}

export function usePing() {
  return useQuery({
    queryKey: ['ping'],
    queryFn: apiService.ping,
    refetchInterval: 10000, // Poll every 10 seconds
    retry: 3,
  });
}

export function useConnections() {
  return useQuery({
    queryKey: ['connections'],
    queryFn: apiService.getConnections,
    refetchInterval: 3000, // Poll every 3 seconds for real-time updates
    retry: 3,
  });
}

export function useHello() {
  return useQuery({
    queryKey: ['hello'],
    queryFn: apiService.getHello,
    retry: 2,
  });
}

export function useWelcome() {
  return useQuery({
    queryKey: ['welcome'],
    queryFn: apiService.getWelcome,
    retry: 2,
  });
}

export function useServerInfo() {
  return useQuery({
    queryKey: ['server-info'],
    queryFn: apiService.getServerInfo,
    retry: 2,
    staleTime: 5 * 60 * 1000, // 5 minutes
  });
}

export function useSendMessage() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (messageData: MessageRequest) => apiService.sendMessage(messageData),
    onSuccess: () => {
      // Invalidate connections to refresh the count
      queryClient.invalidateQueries({ queryKey: ['connections'] });
    },
  });
}