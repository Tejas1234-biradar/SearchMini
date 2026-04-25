import type { RandomResponse, SearchApiResponse, SuggestionResponse } from '@searchmini/shared';

const apiBase = '/api';

export async function fetchSuggestions(q: string): Promise<string[]> {
  const response = await fetch(`${apiBase}/suggestions?q=${encodeURIComponent(q)}`);
  const data = (await response.json()) as SuggestionResponse;
  return data.suggestions || [];
}

export async function fetchSearchResults(query: string, page = 1): Promise<SearchApiResponse> {
  const response = await fetch(`${apiBase}/search?q=${encodeURIComponent(query)}&page=${page}`);
  return (await response.json()) as SearchApiResponse;
}

export async function fetchRandomUrl(): Promise<string | null> {
  const response = await fetch(`${apiBase}/random`);
  if (!response.ok) {
    return null;
  }
  const data = (await response.json()) as RandomResponse;
  return data.url || null;
}
