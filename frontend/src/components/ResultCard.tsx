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
      <div className="text-xs text-on-surface-variant">
        {result.url}
      </div>

      <a
        href={finalUrl}
        target="_blank"
        rel="noopener noreferrer"
        className="block text-2xl font-semibold text-primary hover:underline"
      >
        {displayTitle}
      </a>
    </article>
  );
}