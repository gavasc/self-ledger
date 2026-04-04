import type { HandleClientError } from '@sveltejs/kit';

export const handleError: HandleClientError = ({ error, event }) => {
    console.error('SvelteKit client error:', error, 'event:', event.url.pathname);
    return { message: 'Internal Error' };
};
