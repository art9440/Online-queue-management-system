import { QueryClient } from '@tanstack/react-query';

export const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            retry: 3,
            staleTime: 10 * 1000,
            cacheTime: 300_000,
            refetchOnWindowsFocus: false
        }
    }
})