// state.js — single source of truth for client-side state.
//
// Every module imports `state` rather than holding its own copies. That's
// the DRY win of pulling state out of the modules that mutate it: you can
// reason about everything the page knows by reading this one file.
//
// Persistence: progress and chat history are mirrored to localStorage so a
// page refresh restores them. We use try/catch around every storage call
// because Safari private mode (and Firefox with cookies blocked) throws on
// localStorage writes.

// state is the live, mutable global. It is exported by reference — modules
// read and mutate it directly. Keeping it in one place beats threading a
// dozen "context" props through callbacks.
export const state = {
  // Catalog summaries from /api/lessons, plus the currently-active full
  // Lesson once fetched.
  lessons: [],
  currentLesson: null,

  // Set<string> of completed lesson ids, mirrored to localStorage.
  progress: new Set(loadProgress()),

  // Chat sub-state.
  chat: {
    status: null,                   // {available, provider, model, hint}
    histories: loadChatHistories(), // lessonId → [{role, content}, ...]
    pending: false,                 // true while a request is in flight
  },
};

// ───────────────────────── progress (completed lessons) ─────────────────────────

// loadProgress reads the completed-lessons array from localStorage. Returns
// [] if nothing is stored or the storage is unreadable (Safari private mode
// throws on access, malformed JSON, etc.).
function loadProgress() {
  try { return JSON.parse(localStorage.getItem('go-tut-progress') || '[]'); }
  catch { return []; }
}

// saveProgress persists state.progress to localStorage. Silent on failure
// because progress is a nice-to-have and we'd rather keep the UI alive than
// throw an exception into the user's face.
export function saveProgress() {
  try {
    localStorage.setItem('go-tut-progress', JSON.stringify([...state.progress]));
  } catch (e) {
    console.warn('progress save failed:', e);
  }
}

// ───────────────────────── chat history (per-lesson threads) ─────────────────────────

// loadChatHistories returns the per-lesson chat map from localStorage. Bad
// JSON → empty map.
function loadChatHistories() {
  try { return JSON.parse(localStorage.getItem('go-tut-chat-history') || '{}'); }
  catch { return {}; }
}

// saveChatHistories serializes every per-lesson thread. localStorage has a
// ~5 MB quota per origin which is plenty for this app; if a user somehow
// blows past it (extremely long conversations across many lessons) we drop
// the save silently rather than crashing the UI.
export function saveChatHistories() {
  try {
    localStorage.setItem('go-tut-chat-history', JSON.stringify(state.chat.histories));
  } catch (e) {
    console.warn('chat history save failed:', e);
  }
}

// clearAllChatHistories drops every thread. Exported for future "reset
// everything" flows (the sidebar's Reset button currently only clears
// progress).
export function clearAllChatHistories() {
  state.chat.histories = {};
  saveChatHistories();
}
