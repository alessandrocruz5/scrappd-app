import { QueryClient } from '@tanstack/react-query';

// Single shared React Query client. Tuned conservatively for mobile: don't
// hammer the network by refetching on every focus; let screens opt in.
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 30_000,
    },
  },
});
