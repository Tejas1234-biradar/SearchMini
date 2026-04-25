import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ThemeToggle from '../components/ThemeToggle';
import SearchBar from '../components/SearchBar';
import { fetchRandomUrl, fetchSuggestions } from '../lib/api';

const languages = ['Français', 'Español', 'Deutsch'];

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
        <nav className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4 text-sm text-secondary">
          <div className="flex items-center gap-6">
            <a href="#" className="transition hover:text-primary">About</a>
            <a href="#" className="transition hover:text-primary">Store</a>
          </div>
          <div className="flex items-center gap-4">
            <ThemeToggle />
            <a href="#" className="transition hover:text-primary">Gmail</a>
            <a href="#" className="transition hover:text-primary">Images</a>
            <button className="rounded-full p-2 text-primary transition hover:bg-surface-border">
              <span className="material-symbols-outlined">apps</span>
            </button>
            <div className="h-8 w-8 overflow-hidden rounded-full border border-outline">
              <img
                src="https://lh3.googleusercontent.com/aida-public/AB6AXuDvf1ysbQ1R1oF6DPQvkyvF3U0lfxB0S43tEE8vwXWEv6T9LeCZruM4V8q7RCYlXV_bPGnlY6JRIFhlnrp4dCYgMADWdEUgJKvxe0rVbxJyIE5X6SUj47NvWqRIxw8bJrmkuOXcZXZKUTMDDQK8E3uXXiAB4L5dxwDSVq5XwHQ7HAsTiFnQ2lU367VV2oOtN3sfwwcARgHEw9G5DJhyAneS-Wz6qM-aqx0B8cP-WFUjaYJxcts_0VlhJJeibrn3E0EJdKchz6nH9_Q"
                alt="User profile"
                className="h-full w-full object-cover"
              />
            </div>
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
            onLucky={goLucky}
            onSelectSuggestion={(value) => {
              setQuery(value);
              navigate(`/search?q=${encodeURIComponent(value)}`);
            }}
          />
        </div>
        <div className="mt-8 flex flex-wrap justify-center gap-3">
          <button onClick={search} className="rounded-full border border-outline px-6 py-3 text-sm font-semibold transition hover:border-primary/30">
            searchMini Search
          </button>
          <button onClick={goLucky} className="rounded-full border border-outline px-6 py-3 text-sm font-semibold transition hover:border-primary/30">
            I'm Feeling Lucky
          </button>
        </div>
        <div className="mt-8 text-sm text-secondary">
          <span>searchMini offered in:</span>
          {languages.map((language) => (
            <a key={language} href="#" className="ml-4 text-primary hover:underline">
              {language}
            </a>
          ))}
        </div>
      </main>

      <footer className="border-t border-surface-border/60 bg-surface-container-low px-6 py-6 text-sm text-secondary">
        <div className="mx-auto flex max-w-7xl flex-col gap-4 md:flex-row md:justify-between">
          <div>United Kingdom</div>
          <div className="flex flex-wrap gap-4">
            <a href="#" className="transition hover:text-primary">Advertising</a>
            <a href="#" className="transition hover:text-primary">Business</a>
            <a href="#" className="transition hover:text-primary">How Search works</a>
          </div>
        </div>
      </footer>
    </div>
  );
}
