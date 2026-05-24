// markdown.js — safe markdown subset for assistant chat replies.
//
// We render a deliberately tiny subset:
//   - ```lang fenced code blocks      (preserves indentation, escapes HTML)
//   - `inline code`
//   - **bold**
//   - bare http(s):// URLs  → anchor tags with rel="noopener noreferrer"
//
// Everything else is HTML-escaped, so untrusted model output cannot inject
// scripts or break out of the bubble. There is intentionally NO support
// for arbitrary HTML, images, headings, lists, etc. — the chat surface is
// short messages with code, not a full document.

// escapeHTML turns a raw string into a safe HTML-escaped representation.
// Exported so other modules can build small fragments safely.
export function escapeHTML(s) {
  return s.replace(/&/g, '&amp;')
          .replace(/</g, '&lt;')
          .replace(/>/g, '&gt;')
          .replace(/"/g, '&quot;');
}

// renderMarkdown converts a markdown-flavored string into an HTML fragment
// using the safe subset described above. The output is suitable for
// element.innerHTML; every interpolation point goes through escapeHTML.
export function renderMarkdown(text) {
  // First pass: split on fenced code blocks. We do this BEFORE inline
  // processing so backticks inside ``` blocks don't get treated as
  // inline-code markers.
  const parts = [];
  const fence = /```([\w-]*)\n?([\s\S]*?)```/g;
  let last = 0, m;
  while ((m = fence.exec(text)) !== null) {
    if (m.index > last) parts.push({ kind: 'text', body: text.slice(last, m.index) });
    parts.push({ kind: 'code', lang: m[1] || '', body: m[2] });
    last = fence.lastIndex;
  }
  if (last < text.length) parts.push({ kind: 'text', body: text.slice(last) });

  // Second pass: render each chunk. Code chunks stay literal; text chunks
  // get the inline-markdown transformations applied AFTER escaping.
  return parts.map(p => {
    if (p.kind === 'code') {
      const lang = escapeHTML(p.lang);
      const body = escapeHTML(p.body.replace(/\n$/, ''));
      return `<pre><code class="lang-${lang}">${body}</code></pre>`;
    }
    let h = escapeHTML(p.body);
    h = h.replace(/`([^`\n]+)`/g, '<code>$1</code>');
    h = h.replace(/\*\*([^*\n]+)\*\*/g, '<strong>$1</strong>');
    h = h.replace(/\bhttps?:\/\/[^\s<)]+/g, (url) =>
      `<a href="${url}" target="_blank" rel="noopener noreferrer">${url}</a>`);
    return h;
  }).join('');
}
