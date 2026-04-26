import { ChangeEvent, FormEvent, useState, KeyboardEvent, useEffect, useRef } from 'react';

interface SearchBarProps {
  value: string;
  placeholder?: string;
  loading?: boolean;
  suggestions: string[];
  onChange: (value: string) => void;
  onSearch: () => void;
  onSelectSuggestion: (value: string) => void;
}

export default function SearchBar({
  value,
  placeholder = 'Search the web...',
  loading,
  suggestions,
  onChange,
  onSearch,
  onSelectSuggestion
}: SearchBarProps) {
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const [isFocused, setIsFocused] = useState(false);
  const listRef = useRef<HTMLDivElement>(null);

  // Reset selected index when suggestions change
  useEffect(() => {
    setSelectedIndex(-1);
  }, [suggestions, value]);

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (suggestions.length === 0) return;

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      setSelectedIndex((prev) => (prev < suggestions.length - 1 ? prev + 1 : 0));
    } else if (e.key === 'ArrowUp') {
      e.preventDefault();
      setSelectedIndex((prev) => (prev > 0 ? prev - 1 : suggestions.length - 1));
    } else if (e.key === 'Enter') {
      if (selectedIndex >= 0 && selectedIndex < suggestions.length) {
        e.preventDefault();
        onSelectSuggestion(suggestions[selectedIndex]);
      }
    } else if (e.key === 'Escape') {
      setIsFocused(false);
      setSelectedIndex(-1);
    }
  };

  const highlightMatch = (text: string, query: string) => {
    if (!query) return text;
    const lowerText = text.toLowerCase();
    const lowerQuery = query.toLowerCase();
    
    // For prefix match
    if (lowerText.startsWith(lowerQuery)) {
      return (
        <>
          <span className="font-semibold text-primary">{text.slice(0, query.length)}</span>
          {text.slice(query.length)}
        </>
      );
    }
    
    // For substring match
    const index = lowerText.indexOf(lowerQuery);
    if (index >= 0) {
      return (
        <>
          {text.slice(0, index)}
          <span className="font-semibold text-primary">{text.slice(index, index + query.length)}</span>
          {text.slice(index + query.length)}
        </>
      );
    }
    
    return text;
  };

  const showSuggestions = isFocused && suggestions.length > 0;

  return (
    <div className="mx-auto w-full max-w-3xl px-2 sm:px-0">
      <form
        onSubmit={(event: FormEvent) => {
          event.preventDefault();
          onSearch();
        }}
        className="relative"
      >
        <div className={`flex min-w-0 items-center gap-3 border border-outline bg-surface-muted px-4 py-3 shadow-xl transition focus-within:border-primary/30 ${showSuggestions ? 'rounded-t-3xl rounded-b-none' : 'rounded-full hover:border-surface-border/80'}`}>
          <span className="material-symbols-outlined text-outline">search</span>
          <input
            value={value}
            onChange={(event: ChangeEvent<HTMLInputElement>) => onChange(event.target.value)}
            onKeyDown={handleKeyDown}
            onFocus={() => setIsFocused(true)}
            onBlur={() => {
              // Delay hiding suggestions to allow clicking them
              setTimeout(() => setIsFocused(false), 200);
            }}
            className="min-w-0 w-full bg-transparent border-none text-base text-on-surface outline-none placeholder:text-on-surface-variant"
            placeholder={placeholder}
            aria-label="Search query"
            autoComplete="off"
            role="combobox"
            aria-expanded={showSuggestions}
            aria-controls="search-suggestions"
            aria-activedescendant={selectedIndex >= 0 ? `suggestion-${selectedIndex}` : undefined}
          />
          <div className="flex items-center gap-3">
            {loading ? (
              <span className="material-symbols-outlined animate-spin text-sm text-on-surface-variant">progress_activity</span>
            ) : null}
            <button
              type="submit"
              className="flex-shrink-0 rounded-full bg-primary px-4 py-2 text-sm font-semibold text-on-primary transition hover:opacity-90"
            >
              Search
            </button>
          </div>
        </div>
        {showSuggestions && (
          <div 
            id="search-suggestions"
            ref={listRef}
            className="absolute left-0 right-0 z-20 w-full rounded-b-3xl border border-t-0 border-outline bg-surface pb-2 pt-1 shadow-2xl overflow-hidden"
            role="listbox"
          >
            {suggestions.map((item, index) => {
              const isSelected = index === selectedIndex;
              return (
                <button
                  type="button"
                  key={item}
                  id={`suggestion-${index}`}
                  role="option"
                  aria-selected={isSelected}
                  onClick={() => onSelectSuggestion(item)}
                  onMouseEnter={() => setSelectedIndex(index)}
                  className={`flex w-full items-center gap-3 px-5 py-3 text-left transition ${
                    isSelected ? 'bg-surface-border/30 text-primary' : 'text-on-surface hover:bg-surface-border/20'
                  }`}
                >
                  <span className="material-symbols-outlined text-sm text-outline opacity-70">search</span>
                  <span className="flex-1 truncate">{highlightMatch(item, value)}</span>
                </button>
              );
            })}
          </div>
        )}
      </form>
    </div>
  );
}
