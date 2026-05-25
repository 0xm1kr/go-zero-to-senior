// Chat panel: open/close, send/receive, per-lesson history, render messages.
//
// Each lesson gets its own conversation thread, persisted to localStorage
// via saveChatHistories(). Switching lessons hides the current thread and
// shows that lesson's history; refreshing the page restores everything.
// The "trash" button in the header clears just the active thread.

import { $ } from './dom.js';
import { state, saveChatHistories } from './state.js';
import { fetchChatStatus, sendChat } from './api.js';
import { renderMarkdown } from './markdown.js';
import { getCurrentCode } from './playground.js';

// initChat wires up the chat panel's open/close/send controls and fetches
// the LLM provider status so we know whether to enable the input. Safe to
// call exactly once at startup.
export async function initChat() {
  state.chat.status = await fetchChatStatus();
  updateMeta();

  $('#ask-fab').addEventListener('click', open);
  $('#chat-close').addEventListener('click', close);
  $('#chat-clear').addEventListener('click', clearThread);
  $('#chat-form').addEventListener('submit', (e) => {
    e.preventDefault();
    send();
  });
  $('#chat-input').addEventListener('keydown', (e) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      send();
    }
  });
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && $('#chat-panel').getAttribute('data-open') === 'true') {
      close();
    }
  });
}

// syncToLesson is called by app.js whenever the active lesson changes so
// the chat panel updates its title (e.g. "About: Goroutines") and renders
// that lesson's persisted thread.
export function syncToLesson(lesson) {
  $('#chat-context').textContent = lesson ? `About: ${lesson.title}` : 'About: —';
  renderMessages();
}

// open slides the chat panel in and hides the floating "Ask AI" button.
// The focus is delayed slightly so the CSS slide-in animation has time to
// run before the input scrolls into view.
function open() {
  $('#chat-panel').setAttribute('data-open', 'true');
  $('#ask-fab').hidden = true;
  renderMessages();
  setTimeout(() => $('#chat-input').focus(), 250);
}

// close hides the panel and brings back the "Ask AI" FAB.
function close() {
  $('#chat-panel').setAttribute('data-open', 'false');
  $('#ask-fab').hidden = false;
}

// clearThread wipes the conversation history for the active lesson and
// persists the change immediately. No-op if no lesson is selected.
function clearThread() {
  if (!state.currentLesson) return;
  state.chat.histories[state.currentLesson.id] = [];
  saveChatHistories();
  renderMessages();
}

// currentHistory returns the message array for the active lesson,
// lazily creating an empty one on first access so callers can push to it
// without a null check.
function currentHistory() {
  const id = state.currentLesson?.id;
  if (!id) return [];
  if (!state.chat.histories[id]) state.chat.histories[id] = [];
  return state.chat.histories[id];
}

// renderMessages rebuilds the chat-messages container from scratch based
// on current state (status, history, pending flag). Cheap enough to call
// on every state change.
function renderMessages() {
  const container = $('#chat-messages');
  container.innerHTML = '';

  const status = state.chat.status;
  if (!status?.available) {
    appendMessage(container, 'system',
      'AI chat is not configured.\n\n' + (status?.hint || '') +
      '\n\nFor example, add this to .env:\n' +
      'GEMINI_API_KEY=AIza…\n\nThen restart with `go run .`');
    $('#chat-input').disabled = true;
    $('#chat-send').disabled = true;
    return;
  }
  $('#chat-input').disabled = false;
  $('#chat-send').disabled = state.chat.pending;

  const history = currentHistory();
  if (history.length === 0) {
    appendMessage(container, 'assistant',
      "Hi! Ask me anything about this lesson — Go syntax, idioms, why something works, or how it compares to other languages. " +
      "I can see the lesson contents and your current editor code.");
  } else {
    for (const m of history) {
      appendMessage(container, m.role, m.content);
    }
  }
  if (state.chat.pending) appendThinking(container);

  container.scrollTop = container.scrollHeight;
}

// ERROR_SENTINEL lets us flow LLM/provider errors through the same history
// array as real messages while still rendering them with the error style.
// Stored in history (not just shown) so the user sees the failure after a
// reload, which makes "why didn't it work?" debuggable.
const ERROR_SENTINEL = '__ERROR__';

// appendMessage renders one message bubble into the container, choosing a
// safe rendering strategy based on the role:
//   user      → textContent (no HTML, ever)
//   assistant → markdown via renderMarkdown (limited safe subset)
//   error     → textContent (error text is not trusted markup)
function appendMessage(container, role, content) {
  const el = document.createElement('div');
  if (role === 'assistant' && typeof content === 'string' && content.startsWith(ERROR_SENTINEL)) {
    el.className = 'chat-message error';
    el.textContent = content.slice(ERROR_SENTINEL.length);
  } else if (role === 'assistant') {
    el.className = 'chat-message assistant';
    el.innerHTML = renderMarkdown(content);
  } else {
    el.className = 'chat-message ' + role;
    el.textContent = content;
  }
  container.appendChild(el);
}

// appendThinking shows the three-dot "thinking…" indicator while a request
// is in flight.
function appendThinking(container) {
  const el = document.createElement('div');
  el.className = 'chat-thinking';
  el.innerHTML = 'Thinking<span class="dot"></span><span class="dot"></span><span class="dot"></span>';
  container.appendChild(el);
}

function formatRetry(seconds) {
  if (!seconds || seconds <= 0) return 'a moment';
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.ceil(seconds / 60)}m`;
  return `${Math.ceil(seconds / 3600)}h`;
}

// updateMeta refreshes the small label under the chat input showing the
// active provider/model (or a "no key configured" hint).
function updateMeta() {
  const meta = $('#chat-meta');
  const s = state.chat.status;
  if (!s) { meta.textContent = ''; return; }
  if (s.available) {
    const limits = s.limits;
    const limitHint = limits
      ? ` · ${limits.perMinute}/min · ${limits.daily}/day`
      : '';
    meta.textContent = `${s.provider} · ${s.model}${limitHint}`;
    meta.classList.remove('error');
  } else {
    meta.textContent = 'no key configured';
    meta.classList.add('error');
  }
}

// send picks up the input field, optimistically appends the user turn,
// fires the request, and appends whatever the assistant returns. The
// pending flag prevents concurrent sends.
async function send() {
  if (state.chat.pending) return;
  const input = $('#chat-input');
  const text = input.value.trim();
  if (!text || !state.chat.status?.available) return;

  const history = currentHistory();
  history.push({ role: 'user', content: text });
  saveChatHistories(); // persist the question immediately
  input.value = '';
  state.chat.pending = true;
  $('#chat-send').disabled = true;
  renderMessages();

  try {
    const data = await sendChat(
      state.currentLesson?.id || '',
      getCurrentCode(),
      history,
    );
    if (data.error) {
      let msg = data.error;
      if (data.retryAfter) {
        msg += `\n\nTry again in ${formatRetry(data.retryAfter)}.`;
      }
      history.push({ role: 'assistant', content: ERROR_SENTINEL + msg });
    } else if (data.message?.content) {
      history.push({ role: 'assistant', content: data.message.content });
    } else {
      history.push({ role: 'assistant', content: ERROR_SENTINEL + 'empty response' });
    }
  } catch (err) {
    history.push({ role: 'assistant', content: ERROR_SENTINEL + (err?.message || 'request failed') });
  } finally {
    state.chat.pending = false;
    saveChatHistories(); // persist the reply (or the error)
    renderMessages();
  }
}
