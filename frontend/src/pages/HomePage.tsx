import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ThemeToggle from '../components/ThemeToggle';
import SearchBar from '../components/SearchBar';
import { fetchSuggestions } from '../lib/api';

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

  return (
    <div className="min-h-screen bg-surface text-on-surface">
      <header className="sticky top-0 z-50 border-b border-surface-border/60 bg-surface/95 backdrop-blur-xl">
        <nav className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4">
          <a href="/" className="text-xl font-black tracking-tight text-primary">searchMini</a>
          <div>
            <ThemeToggle />
          </div>
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
            onSelectSuggestion={(value) => {
              setQuery(value);
              navigate(`/search?q=${encodeURIComponent(value)}`);
            }}
          />
        </div>
      </main>

      <footer className="border-t border-surface-border/60 bg-surface-container-low px-6 py-6 text-sm text-secondary">
        <div className="mx-auto flex max-w-7xl items-center justify-between">
          <div>United Kingdom</div>
          <div className="text-xs uppercase tracking-[0.18em]">Search with intention</div>
        </div>
      </footer>
    </div>
  );
}
