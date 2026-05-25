// api.js — the single place that talks to the backend.
//
// Every other module imports these helpers instead of calling fetch()
// directly, so URL paths and request shapes only need to change in one
// file. The functions intentionally stay thin: HTTP error → thrown Error,
// payload → the parsed JSON body, no retry logic.

// fetchLessons returns the catalog summary list ([{id, title, category}]).
// Throws on non-2xx.
export async function fetchLessons() {
  const res = await fetch('/api/lessons');
  if (!res.ok) throw new Error(`lessons: ${res.status}`);
  return res.json();
}

// fetchLesson returns one full lesson (description + starter code + notes).
// Throws on non-2xx; the lesson id is URL-encoded for safety.
export async function fetchLesson(id) {
  const res = await fetch(`/api/lessons/${encodeURIComponent(id)}`);
  if (!res.ok) throw new Error(`lesson ${id}: ${res.status}`);
  return res.json();
}

// runCode POSTs the editor buffer to /api/run and returns the
// {stdout, stderr, error, duration} envelope. The backend always replies
// HTTP 200, surfacing compile/runtime failures in the body, so we don't
// throw on res.ok being false here — the caller renders the result either
// way.
export async function runCode(code) {
  const res = await fetch('/api/run', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code }),
  });
  return res.json();
}

// fetchChatStatus returns {available, provider, model, hint} so the UI can
// show "AI chat enabled (gemini-2.5-flash)" or an "no key configured" hint
// before the user types anything. Failures collapse to an unavailable
// snapshot rather than throwing — chat is an optional feature and the rest
// of the app should keep working when the backend is unreachable.
export async function fetchChatStatus() {
  try {
    const res = await fetch('/api/chat/status');
    return res.json();
  } catch {
    return { available: false, hint: 'Could not reach /api/chat/status.' };
  }
}

// sendChat posts one conversation turn (lesson id + current editor code +
// full history) and returns the {message, error, retryAfter} envelope from
// the server.
export async function sendChat(lessonId, code, messages) {
  const res = await fetch('/api/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ lessonId, code, messages }),
  });
  const data = await res.json().catch(() => ({}));
  if (res.status === 429) {
    return {
      error: data.error || 'rate limit exceeded',
      retryAfter: data.retryAfter,
    };
  }
  if (!res.ok && !data.error) {
    return { error: `chat: ${res.status}` };
  }
  return data;
}
