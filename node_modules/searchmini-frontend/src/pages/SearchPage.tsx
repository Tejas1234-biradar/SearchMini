import { useEffect, useMemo, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { fetchSearchResults, fetchSuggestions } from '../lib/api';
import SearchBar from '../components/SearchBar';
import ResultCard from '../components/ResultCard';
import ThemeToggle from '../components/ThemeToggle';
import type { SearchResult } from '@searchmini/shared';

function useQuery() {
  return new URLSearchParams(useLocation().search);
}

export default function SearchPage() {
  const queryParams = useQuery();
  const query = queryParams.get('q') ?? '';
  const navigate = useNavigate();
  const [searchTerm, setSearchTerm] = useState(query);
  const [results, setResults] = useState<SearchResult[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [suggestionLoading, setSuggestionLoading] = useState(false);

  useEffect(() => {
    setSearchTerm(query);
  }, [query]);

  useEffect(() => {
    if (!query) {
      setResults([]);
      setTotal(0);
      setError(null);
      return;
    }

    setLoading(true);
    setError(null);

    fetchSearchResults(query)
      .then((data) => {
        setResults(data.results || []);
        setTotal(data.total || 0);
      })
      .catch((err) => {
        setError(err?.message || 'Search request failed');
      })
      .finally(() => setLoading(false));
  }, [query]);

  useEffect(() => {
    if (!searchTerm || searchTerm.length < 2) {
      setSuggestions([]);
      return;
    }

    const timer = window.setTimeout(async () => {
      setSuggestionLoading(true);
      const items = await fetchSuggestions(searchTerm);
      setSuggestions(items);
      setSuggestionLoading(false);
    }, 250);

    return () => window.clearTimeout(timer);
  }, [searchTerm]);

  const handleSearch = () => {
    const trimmed = searchTerm.trim();
    if (trimmed) {
      navigate(`/search?q=${encodeURIComponent(trimmed)}`);
    }
  };

  const topResult = useMemo(() => results[0], [results]);

  return (
    <div className="min-h-screen bg-surface text-on-surface">
      <header className="sticky top-0 z-50 border-b border-surface-border/60 bg-surface/95 backdrop-blur-xl">
        <div className="mx-auto flex max-w-7xl flex-col gap-4 px-6 py-4 md:flex-row md:items-center md:justify-between">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
            <a href="/" className="text-2xl font-black text-primary">searchMini</a>
            <div className="rounded-full border border-outline bg-surface-muted px-2 py-1 text-[0.75rem] uppercase tracking-[0.25em] text-secondary">
              Search
            </div>
          </div>
          <ThemeToggle />
        </div>
      </header>

      <main className="mx-auto max-w-7xl space-y-8 px-6 py-8">
        <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <h1 className="text-4xl font-black tracking-tight text-primary sm:text-5xl">Results for</h1>
            <p className="mt-2 text-sm text-secondary">Explore the intelligence behind searchMini.</p>
          </div>
          <SearchBar
            value={searchTerm}
            loading={suggestionLoading}
            suggestions={suggestions}
            placeholder="Search again..."
            onChange={setSearchTerm}
            onSearch={handleSearch}
            onLucky={() => navigate(`/search?q=${encodeURIComponent(searchTerm)}`)}
            onSelectSuggestion={(value) => {
              setSearchTerm(value);
              navigate(`/search?q=${encodeURIComponent(value)}`);
            }}
          />
        </div>

        <section className="grid gap-10 lg:grid-cols-[2fr_1fr]">
          <div className="space-y-6">
            <div className="rounded-3xl border border-surface-border bg-surface-muted p-6">
              <p className="text-sm text-secondary">About {total.toLocaleString()} results</p>
              <h2 className="mt-3 text-3xl font-black tracking-tight text-primary">{query}</h2>
            </div>

            {loading ? (
              <div className="rounded-3xl border border-surface-border bg-surface-muted p-6 text-center text-secondary">Searching...</div>
            ) : error ? (
              <div className="rounded-3xl border border-red-500/20 bg-surface-strong/90 p-6 text-red-200">{error}</div>
            ) : results.length === 0 ? (
              <div className="rounded-3xl border border-surface-border bg-surface-muted p-6 text-secondary">No results found for “{query}”.</div>
            ) : (
              results.map((result) => <ResultCard key={result.url} result={result} />)
            )}
          </div>

          <aside className="hidden lg:block">
            {topResult ? (
              <div className="space-y-6 rounded-3xl border border-surface-border bg-surface-muted p-6">
                <div>
                  <h3 className="text-lg font-semibold text-primary">Top result summary</h3>
                  <p className="mt-3 text-sm leading-7 text-secondary">{topResult.summary_text || topResult.description || 'This result is ranked highest for the current query.'}</p>
                </div>
                <div className="rounded-3xl border border-outline bg-surface-strong p-5">
                  <div className="mb-3 flex items-center justify-between text-xs uppercase tracking-[0.25em] text-secondary">
                    <span>Source</span>
                    <a
                      href={topResult.url}
                      target="_blank"
                      rel="noreferrer"
                      className="text-primary hover:underline"
                    >
                      {(() => {
                        try {
                          return new URL(topResult.url).hostname;
                        } catch {
                          return topResult.url;
                        }
                      })()}
                    </a>
                  </div>
                  <div className="flex flex-col gap-3 text-sm text-on-surface-variant">
                    <div className="rounded-2xl bg-surface-border/10 p-4">
                      <span className="block text-[0.7rem] uppercase tracking-[0.25em] text-secondary">Score</span>
                      <span className="mt-1 block text-lg font-semibold text-primary">{topResult.score.toFixed(4)}</span>
                    </div>
                    <div className="rounded-2xl bg-surface-border/10 p-4">
                      <span className="block text-[0.7rem] uppercase tracking-[0.25em] text-secondary">PageRank</span>
                      <span className="mt-1 block text-lg font-semibold text-primary">{topResult.pagerank.toFixed(6)}</span>
                    </div>
                  </div>
                </div>
              </div>
            ) : null}
          </aside>
        </section>
      </main>
    </div>
  );
}
