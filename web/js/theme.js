// theme.js — light/dark mode toggle.
//
// The initial theme is applied by an inline <script> in index.html before
// the stylesheet paints to avoid a flash of unstyled content. This module
// only handles user-driven toggling after page load and persists the
// choice to localStorage so future visits start on the same theme.

import { $ } from './dom.js';

// initTheme wires the header toggle button. Call once at startup.
export function initTheme() {
  $('#theme-toggle').addEventListener('click', toggle);
  syncTitle();
}

// current reads the active theme from the <html data-theme> attribute,
// normalising anything unexpected to 'dark'.
function current() {
  return document.documentElement.getAttribute('data-theme') === 'light' ? 'light' : 'dark';
}

// toggle flips the theme, mirrors it onto <html data-theme> (which CSS
// reads), and persists the choice. The playground module observes the
// attribute change and updates its CodeMirror theme automatically.
function toggle() {
  const next = current() === 'light' ? 'dark' : 'light';
  document.documentElement.setAttribute('data-theme', next);
  try { localStorage.setItem('go-tut-theme', next); } catch {}
  syncTitle();
}

// syncTitle keeps the toggle button's tooltip + aria-label in sync with
// the action it would perform on the next click.
function syncTitle() {
  const btn = $('#theme-toggle');
  if (!btn) return;
  const next = current() === 'light' ? 'dark' : 'light';
  btn.title = `Switch to ${next} mode`;
  btn.setAttribute('aria-label', `Switch to ${next} mode`);
}
