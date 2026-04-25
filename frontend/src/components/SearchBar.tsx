import { ChangeEvent, FormEvent } from 'react';

interface SearchBarProps {
  value: string;
  placeholder?: string;
  loading?: boolean;
  suggestions: string[];
  onChange: (value: string) => void;
  onSearch: () => void;
  onLucky: () => void;
  onSelectSuggestion: (value: string) => void;
}

export default function SearchBar({
  value,
  placeholder = 'Search the web...',
  loading,
  suggestions,
  onChange,
  onSearch,
  onLucky,
  onSelectSuggestion
}: SearchBarProps) {
  return (
    <div className="mx-auto w-full max-w-3xl px-2 sm:px-0">
      <form
        onSubmit={(event: FormEvent) => {
          event.preventDefault();
          onSearch();
        }}
        className="relative"
      >
        <div className="flex min-w-0 items-center gap-3 rounded-full border border-outline bg-surface-muted px-4 py-3 shadow-xl transition hover:border-surface-border/80 focus-within:border-primary/30">
          <span className="material-symbols-outlined text-outline">search</span>
          <input
            value={value}
            onChange={(event: ChangeEvent<HTMLInputElement>) => onChange(event.target.value)}
            className="min-w-0 w-full bg-transparent border-none text-base text-on-surface outline-none placeholder:text-on-surface-variant"
            placeholder={placeholder}
            aria-label="Search query"
          />
          <div className="flex items-center gap-3">
            {loading ? (
              <span className="text-sm text-on-surface-variant">Loading...</span>
            ) : null}
            <button type="button" onClick={onLucky} className="flex-shrink-0 rounded-full px-3 py-2 text-sm text-on-surface transition hover:bg-surface-border">
              Lucky
            </button>
          </div>
        </div>
        {suggestions.length > 0 && (
          <div className="absolute z-20 mt-2 w-full rounded-3xl border border-outline bg-surface px-2 py-2 shadow-xl">
            {suggestions.map((item) => (
              <button
                type="button"
                key={item}
                onClick={() => onSelectSuggestion(item)}
                className="flex w-full items-center gap-3 rounded-2xl px-4 py-3 text-left text-on-surface transition hover:bg-surface-border"
              >
                <span className="material-symbols-outlined text-sm text-outline">history</span>
                {item}
              </button>
            ))}
          </div>
        )}
      </form>
    </div>
  );
}
