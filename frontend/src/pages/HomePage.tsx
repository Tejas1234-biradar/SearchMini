import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ThemeToggle from '../components/ThemeToggle';
import SearchBar from '../components/SearchBar';
import { fetchRandomUrl, fetchSuggestions } from '../lib/api';

export default function HomePage() {
  const navigate = useNavigate();
  const [query, setQuery] = useState('');
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);

  const shouldShowSuggestions = useMemo(() => query.length >= 2 && suggestions.length > 0, [query, suggestions]);

  useEffect(() => {
    if (!query || query.length < 2) {
      setSuggestions([]);
      return;
    }

    const timer = window.setTimeout(async () => {
      setLoading(true);
      const result = await fetchSuggestions(query);
      setSuggestions(result);
      setLoading(false);
    }, 250);

    return () => window.clearTimeout(timer);
  }, [query]);

  const search = () => {
    const trimmed = query.trim();
    if (trimmed) {
      navigate(`/search?q=${encodeURIComponent(trimmed)}`);
    }
  };

  const goLucky = async () => {
    if (!query.trim()) {
      const url = await fetchRandomUrl();
      if (url) {
        window.location.href = url;
      }
      return;
    }

    navigate(`/search?q=${encodeURIComponent(query.trim())}`);
  };

  return (
    <div className="min-h-screen bg-surface text-on-surface">
      <header className="sticky top-0 z-50 border-b border-surface-border/60 bg-surface/95 backdrop-blur-xl">
        <nav className="mx-auto flex max-w-7xl items-center justify-end px-6 py-4 text-sm text-secondary">
          <ThemeToggle />
        </nav>
      </header>

      <main className="flex min-h-[calc(100vh-84px)] flex-col items-center justify-center px-6 py-10 text-center">
        <div className="mb-8">
          <h1 className="text-6xl font-black tracking-tight text-primary sm:text-7xl">searchMini</h1>
        </div>
        <div className="w-full px-4">
          <SearchBar
            value={query}
            loading={loading}
            suggestions={shouldShowSuggestions ? suggestions : []}
            placeholder="Search the web..."
            onChange={setQuery}
            onSearch={search}
            onLucky={goLucky}
            onSelectSuggestion={(value) => {
              setQuery(value);
              navigate(`/search?q=${encodeURIComponent(value)}`);
            }}
          />
        </div>
        <div className="mt-8 flex flex-wrap justify-center gap-3">
          <button onClick={goLucky} className="rounded-full border border-outline px-6 py-3 text-sm font-semibold transition hover:border-primary/30">
            I'm Feeling Lucky
          </button>
        </div>
      </main>
    </div>
  );
}
