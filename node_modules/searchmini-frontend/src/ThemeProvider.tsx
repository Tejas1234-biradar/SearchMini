import { createContext, useContext, useEffect, useMemo, useState } from 'react';
import { defaultTheme, getThemeFromStorage, themeStorageKey, themes, ThemeId } from './theme';

interface ThemeContextValue {
  theme: ThemeId;
  themes: typeof themes;
  setTheme: (theme: ThemeId) => void;
}

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined);

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<ThemeId>(getThemeFromStorage);

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    window.localStorage.setItem(themeStorageKey, theme);
  }, [theme]);

  const value = useMemo(
    () => ({ theme, themes, setTheme: setThemeState }),
    [theme]
  );

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>;
}

export function useTheme() {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within ThemeProvider');
  }
  return context;
}
