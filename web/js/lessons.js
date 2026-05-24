// lessons.js — sidebar list, current-lesson rendering, progress, navigation.
//
// Owns the "what lesson is open" state and emits an event (the onChange
// callback) when it changes so the playground and chat modules can react
// without depending on each other. Lesson selection itself is driven by
// location.hash so deep links and the browser Back button just work.

import { $ } from './dom.js';
import { state, saveProgress } from './state.js';
import { fetchLessons, fetchLesson } from './api.js';

// onChange is the callback supplied by app.js. Fires every time the active
// lesson changes — used to repopulate the editor + chat panel.
let onChange = () => {};

// initLessons fetches the catalog, wires up sidebar interactions and
// navigation buttons, and renders whichever lesson is referenced by the URL
// hash (defaulting to the first lesson on a bare load).
export async function initLessons(opts = {}) {
  onChange = opts.onChange || (() => {});

  state.lessons = await fetchLessons();
  renderSidebar();
  updateProgressBar();

  $('#search').addEventListener('input', renderSidebar);
  $('#prev-lesson').addEventListener('click', () => navigate(-1));
  $('#next-lesson').addEventListener('click', () => navigate(+1));
  $('#mark-complete').addEventListener('click', toggleComplete);
  $('#reset-progress').addEventListener('click', () => {
    if (confirm('Clear all completed lessons?')) {
      state.progress.clear();
      saveProgress();
      renderSidebar();
      updateProgressBar();
      renderCurrentLesson();
    }
  });

  window.addEventListener('hashchange', renderCurrentLesson);
  await renderCurrentLesson();
}

// renderSidebar rebuilds the lesson list from state.lessons, applying the
// search filter and inserting a category header before each first lesson
// of a new category. Cheap to call on every search keystroke — the whole
// catalog is well under a thousand DOM nodes.
function renderSidebar() {
  const query = ($('#search').value || '').toLowerCase().trim();
  const filtered = state.lessons.filter(l =>
    !query ||
    l.title.toLowerCase().includes(query) ||
    l.category.toLowerCase().includes(query)
  );

  const list = $('#lesson-list');
  list.innerHTML = '';

  let lastCategory = null;
  filtered.forEach((l) => {
    if (l.category !== lastCategory) {
      const h = document.createElement('div');
      h.className = 'category-header';
      h.textContent = l.category;
      list.appendChild(h);
      lastCategory = l.category;
    }
    const item = document.createElement('div');
    item.className = 'lesson-item';
    if (state.progress.has(l.id)) item.classList.add('done');
    if (state.currentLesson && state.currentLesson.id === l.id) item.classList.add('active');

    const check = document.createElement('span');
    check.className = 'check';
    item.appendChild(check);

    const text = document.createElement('span');
    text.textContent = l.title;
    item.appendChild(text);

    item.addEventListener('click', () => { location.hash = l.id; });
    list.appendChild(item);
  });
}

// updateProgressBar refreshes the fill width and "N / M complete" label.
// "done" counts only lessons that still exist in the current catalog so
// stale ids in localStorage (e.g. lessons renamed in a future version)
// don't inflate the count.
function updateProgressBar() {
  const total = state.lessons.length;
  const done = [...state.progress].filter(id => state.lessons.some(l => l.id === id)).length;
  const pct = total ? (done / total) * 100 : 0;
  $('#progress-fill').style.width = pct + '%';
  $('#progress-label').textContent = `${done} / ${total} complete`;
}

// renderCurrentLesson reads location.hash, fetches the full lesson body,
// and renders title/description/notes. On a missing id it shows a "not
// found" title and bails. Fires onChange last so consumers see fully
// populated state.currentLesson.
async function renderCurrentLesson() {
  let id = location.hash.replace(/^#/, '');
  if (!id && state.lessons.length) id = state.lessons[0].id;
  if (!id) return;

  let lesson;
  try {
    lesson = await fetchLesson(id);
  } catch {
    $('#lesson-title').textContent = 'Lesson not found';
    return;
  }
  state.currentLesson = lesson;

  $('#lesson-category').textContent = lesson.category;
  $('#lesson-title').textContent = lesson.title;
  $('#lesson-description').innerHTML = lesson.description || '';

  const notesEl = $('#lesson-notes');
  notesEl.innerHTML = '';
  (lesson.notes || []).forEach(n => {
    const li = document.createElement('li');
    li.textContent = n;
    notesEl.appendChild(li);
  });
  $('#key-takeaways').hidden = !(lesson.notes && lesson.notes.length);

  updateMarkCompleteButton();
  updateNavButtons();
  renderSidebar();
  window.scrollTo({ top: 0, behavior: 'instant' });

  onChange(lesson);
}

// updateMarkCompleteButton toggles the "Mark complete" / "✓ Completed"
// button text and its .done style based on the active lesson's status.
function updateMarkCompleteButton() {
  const btn = $('#mark-complete');
  const done = state.currentLesson && state.progress.has(state.currentLesson.id);
  btn.textContent = done ? '✓ Completed' : 'Mark complete';
  btn.classList.toggle('done', !!done);
}

// updateNavButtons enables / disables Prev and Next so users hit a wall at
// either end of the catalog rather than wrapping around.
function updateNavButtons() {
  const idx = state.lessons.findIndex(l => l.id === state.currentLesson?.id);
  $('#prev-lesson').disabled = idx <= 0;
  $('#next-lesson').disabled = idx === -1 || idx >= state.lessons.length - 1;
}

// navigate jumps `delta` lessons forward (+1) or backward (-1) by setting
// location.hash, which the hashchange listener picks up to re-render.
function navigate(delta) {
  const idx = state.lessons.findIndex(l => l.id === state.currentLesson?.id);
  const next = state.lessons[idx + delta];
  if (next) location.hash = next.id;
}

// toggleComplete flips the active lesson's completion state, persists the
// change, and refreshes every UI surface that depends on progress.
function toggleComplete() {
  if (!state.currentLesson) return;
  const id = state.currentLesson.id;
  if (state.progress.has(id)) state.progress.delete(id);
  else state.progress.add(id);
  saveProgress();
  updateMarkCompleteButton();
  updateProgressBar();
  renderSidebar();
}
