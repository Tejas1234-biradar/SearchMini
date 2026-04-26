import type { RandomResponse, SearchApiResponse, SuggestionResponse } from './types';

const apiBase = '/api';

async function parseJsonResponse<T>(response: Response, fallbackMessage: string): Promise<T> {
  const raw = await response.text();

  if (!response.ok) {
    throw new Error(`${fallbackMessage} (${response.status})`);
  }

  try {
    return JSON.parse(raw) as T;
  } catch {
    throw new Error(`${fallbackMessage}: invalid JSON response`);
  }
}

export async function fetchSuggestions(q: string): Promise<string[]> {
  const response = await fetch(`${apiBase}/suggestions?q=${encodeURIComponent(q)}`);
  const data = await parseJsonResponse<SuggestionResponse>(response, 'Failed to fetch suggestions');
  return data.suggestions || [];
}

export async function fetchSearchResults(query: string, page = 1): Promise<SearchApiResponse> {
  const response = await fetch(`${apiBase}/search?q=${encodeURIComponent(query)}&page=${page}`);
  return parseJsonResponse<SearchApiResponse>(response, 'Failed to fetch search results');
}

export async function fetchRandomUrl(): Promise<string | null> {
  const response = await fetch(`${apiBase}/random`);
  if (!response.ok) return null;
  const data = await parseJsonResponse<RandomResponse>(response, 'Failed to fetch random URL');
  return data.url || null;
}
