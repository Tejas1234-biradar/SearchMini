import type { ThemeId } from '../theme';
import { useTheme } from '../ThemeProvider';

export default function ThemeToggle() {
  const { theme, themes, setTheme } = useTheme();

  return (
    <div className="flex items-center gap-2">
      <label htmlFor="theme-selector" className="sr-only">
        Select theme
      </label>
      <select
        id="theme-selector"
        value={theme}
        onChange={(event) => setTheme(event.target.value as ThemeId)}
        className="rounded-full border border-outline bg-surface px-3 py-2 text-sm font-medium text-on-surface transition hover:border-primary/40 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
      >
        {themes.map((option) => (
          <option key={option.id} value={option.id}>
            {option.label}
          </option>
        ))}
      </select>
    </div>
  );
}
