export interface SearchResult {
  url: string;
  title: string;
  description: string;
  summary_text: string;
  last_crawled: string;
  tfidf_weight: number;
  pagerank: number;
  score: number;
}

export interface SearchApiResponse {
  total: number;
  page: number;
  results: SearchResult[];
}

export interface SuggestionResponse {
  suggestions: string[];
}

export interface RandomResponse {
  url: string;
}

export interface StatsResponse {
  status: string;
  pages: number;
}
