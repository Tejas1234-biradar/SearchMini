import type { SearchResult } from '../lib/types';

interface ResultCardProps {
  result: SearchResult;
}

function titleFromUrl(url: string): string | null {
  try {
    const fullUrl = url.startsWith('http') ? url : `https://${url}`;
    const parsed = new URL(fullUrl);
    const parts = parsed.pathname.split('/').filter(Boolean);
    if (parts.length === 0) return null;

    const lastSegment = decodeURIComponent(parts[parts.length - 1]);
    const normalized = lastSegment
      .replace(/[-_]+/g, ' ')
      .replace(/\s+/g, ' ')
      .trim();

    return normalized || null;
  } catch {
    return null;
  }
}

export default function ResultCard({ result }: ResultCardProps) {
  const displayTitle = titleFromUrl(result.url) || result.title || 'Untitled page';

  return (
    <article className="space-y-3 rounded-3xl border border-outline bg-surface-muted p-6 transition hover:border-surface-border/80 hover:bg-surface-strong">
      <div className="flex flex-wrap items-center gap-2 text-xs text-on-surface-variant">
        <span className="material-symbols-outlined text-sm">language</span>
        <span className="truncate max-w-full break-words">{result.url}</span>
      </div>
      <a
        href={result.url}
        target="_blank"
        rel="noreferrer"
        className="block text-2xl font-semibold text-primary transition hover:underline"
      >
        {displayTitle}
      </a>
      <p className="text-sm leading-7 text-on-surface-variant">{result.summary_text || result.description || 'No description available.'}</p>
      <div className="flex flex-wrap gap-4 text-[0.7rem] uppercase tracking-[0.18em] text-outline">
        <span>Score: {result.score.toFixed(4)}</span>
        <span>PageRank: {result.pagerank.toFixed(6)}</span>
      </div>
    </article>
  );
}
