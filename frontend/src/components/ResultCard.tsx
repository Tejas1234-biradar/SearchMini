import type { SearchResult } from '../lib/types';

interface ResultCardProps {
  result: SearchResult;
}

function normalizeUrl(url: string): string {
  if (!url) return '#';
  return url.startsWith('http://') || url.startsWith('https://')
    ? url
    : `https://${url}`;
}

function titleFromUrl(url: string): string | null {
  try {
    const parsed = new URL(normalizeUrl(url));
    const parts = parsed.pathname.split('/').filter(Boolean);

    if (parts.length === 0) return parsed.hostname;

    const lastSegment = decodeURIComponent(parts[parts.length - 1]);

    return lastSegment
      .replace(/[-_]+/g, ' ')
      .replace(/\s+/g, ' ')
      .trim();
  } catch {
    return null;
  }
}

export default function ResultCard({ result }: ResultCardProps) {
  const finalUrl = normalizeUrl(result.url);
  const displayTitle =
    titleFromUrl(result.url) || result.title || 'Untitled page';

  return (
    <article className="space-y-3 rounded-3xl border border-outline bg-surface-muted p-6">
      <div className="flex items-center gap-2 text-xs text-on-surface-variant">
        <span className="material-symbols-outlined text-sm">language</span>
        <a
          href={finalUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="truncate max-w-md hover:underline"
        >
          {result.url}
        </a>
      </div>

      <a
        href={finalUrl}
        target="_blank"
        rel="noopener noreferrer"
        className="block text-2xl font-semibold text-primary hover:underline"
      >
        {displayTitle}
      </a>

      <div className="flex gap-4 mt-1 text-[0.65rem] text-outline uppercase font-bold tracking-tight">
        <span>Score: {result.score.toFixed(4)}</span>
        <span>PR: {result.pagerank.toFixed(6)}</span>
      </div>
    </article>
  );
}