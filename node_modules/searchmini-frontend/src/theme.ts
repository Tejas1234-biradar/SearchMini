// To add a new theme:
// 1. Add the theme id and label to ThemeId and themes.
// 2. Add a matching html[data-theme="..."] block in src/index.css with CSS variable values.
// 3. Use theme tokens through Tailwind color utilities in components.
export type ThemeId = 'catppuccin' | 'mocha' | 'solarized-light' | 'quiet-light';

export interface ThemeEntry {
  id: ThemeId;
  label: string;
  description: string;
}

export const themeStorageKey = 'searchmini-theme';

export const themes: ThemeEntry[] = [
  {
    id: 'catppuccin',
    label: 'Catppuccin',
    description: 'Soft pastel and cozy tones.'
  },
  {
    id: 'mocha',
    label: 'Mocha',
    description: 'Rich dark contrast with elegant warmth.'
  },
  {
    id: 'solarized-light',
    label: 'Solarized Light',
    description: 'Warm, productive, high-readability light mode.'
  },
  {
    id: 'quiet-light',
    label: 'Quiet Light',
    description: 'Minimal, clean, editor-friendly light theme.'
  }
];

export const defaultTheme: ThemeId = 'mocha';

export const isThemeId = (value: unknown): value is ThemeId =>
  typeof value === 'string' &&
  themes.some((theme) => theme.id === value);

export function getThemeFromStorage(): ThemeId {
  if (typeof window === 'undefined') {
    return defaultTheme;
  }

  const stored = window.localStorage.getItem(themeStorageKey);
  if (isThemeId(stored)) {
    return stored;
  }

  return defaultTheme;
}
