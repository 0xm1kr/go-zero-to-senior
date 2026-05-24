// Playground: code editor + Run button + output panel.
//
// The editor is a CodeMirror 5 instance mounted on the original
// <textarea id="code">. CodeMirror is loaded from /vendor/codemirror/ via
// plain <script> tags in index.html, so the global `window.CodeMirror` is
// guaranteed to exist by the time initPlayground() runs.
//
// Public API:
//   initPlayground()      — mount the editor and wire buttons (call once)
//   resetTo(code)         — load new starter code (called when lesson changes)
//   getCurrentCode()      — read the current editor buffer (used by chat.js
//                           so context-aware AI replies see live edits)
//
// We export getCurrentCode rather than letting callers poke at #code.value
// directly because CodeMirror doesn't keep the underlying textarea in sync
// on every keystroke — only on form submit / explicit save().

import { $ } from './dom.js';
import { runCode } from './api.js';

let editor = null;
let originalCode = '';

const DARK_THEME = 'material-darker';
const LIGHT_THEME = 'default';

// getCurrentCode returns whatever the user currently has in the editor,
// or the original textarea value if CodeMirror failed to load.
export function getCurrentCode() {
  if (editor) return editor.getValue();
  const ta = $('#code');
  return ta ? ta.value : '';
}

// initPlayground mounts CodeMirror on #code and wires up the Run/Reset
// buttons. Safe to call exactly once at startup.
export function initPlayground() {
  const textarea = $('#code');

  if (!window.CodeMirror) {
    // Fallback: keep the textarea wired up if CM failed to load. The app
    // is still usable; users just lose syntax highlighting.
    console.warn('CodeMirror not loaded; falling back to plain textarea.');
    wirePlainTextarea(textarea);
    return;
  }

  editor = window.CodeMirror.fromTextArea(textarea, {
    mode: 'go',
    theme: currentCmTheme(),
    lineNumbers: true,
    indentUnit: 4,
    tabSize: 4,
    indentWithTabs: true,
    smartIndent: true,
    matchBrackets: true,
    autoCloseBrackets: true,
    lineWrapping: false,
    viewportMargin: Infinity, // grow with content; size capped via CSS
    extraKeys: {
      // ⌘/Ctrl + Enter to run.
      'Cmd-Enter':  () => run(),
      'Ctrl-Enter': () => run(),
      // Tab inserts a tab (Go indentation), Shift-Tab dedents.
      Tab: (cm) => {
        if (cm.somethingSelected()) {
          cm.indentSelection('add');
        } else {
          cm.replaceSelection('\t', 'end', '+input');
        }
      },
      'Shift-Tab': (cm) => cm.indentSelection('subtract'),
    },
  });

  // Follow the app's light/dark toggle automatically.
  watchAppTheme((theme) => {
    editor.setOption('theme', theme === 'light' ? LIGHT_THEME : DARK_THEME);
  });

  $('#run-code').addEventListener('click', run);
  $('#reset-code').addEventListener('click', () => {
    editor.setValue(originalCode);
    editor.focus();
  });
}

// resetTo loads new starter code into the editor and remembers it as the
// "original" so the Reset button can restore it later. Called by app.js
// every time the user navigates to a new lesson.
export function resetTo(code) {
  const next = code || '';
  originalCode = next;
  if (editor) {
    editor.setValue(next);
    editor.clearHistory(); // don't let undo cross lesson boundaries
    editor.scrollTo(0, 0);
  } else {
    const ta = $('#code');
    if (ta) {
      ta.value = next;
      ta.dataset.original = next;
    }
  }
  $('#output').innerHTML = 'Click <b>Run</b> to execute the code above. <span class="meta">(⌘/Ctrl + Enter)</span>';
  $('#run-status').textContent = '';
  $('#run-status').className = '';
}

// run POSTs the current editor buffer to /api/run and renders stdout,
// stderr, errors, and duration into the output panel. Disables the Run
// button for the duration so impatient double-clicks don't queue requests.
async function run() {
  const btn = $('#run-code');
  const status = $('#run-status');
  const output = $('#output');

  btn.disabled = true;
  status.textContent = 'compiling…';
  status.className = 'running';
  output.textContent = '';

  try {
    const data = await runCode(getCurrentCode());
    output.innerHTML = '';

    if (data.stdout) appendSpan(output, 'stdout', data.stdout);
    if (data.stderr) appendSpan(output, 'stderr', data.stderr);
    if (data.error) {
      const prefix = (data.stdout || data.stderr) ? '\n' : '';
      appendSpan(output, 'error', prefix + 'error: ' + data.error);
    }
    if (!data.stdout && !data.stderr && !data.error) {
      output.innerHTML = '<span class="meta">(no output)</span>';
    }

    const meta = document.createElement('div');
    meta.className = 'meta';
    meta.style.marginTop = '8px';
    meta.textContent = `— exited in ${data.duration}`;
    output.appendChild(meta);

    status.textContent = data.error ? 'failed' : 'succeeded';
    status.className = data.error ? 'error' : 'success';
  } catch (err) {
    output.innerHTML = '';
    appendSpan(output, 'error', 'request failed: ' + err.message);
    status.textContent = 'failed';
    status.className = 'error';
  } finally {
    btn.disabled = false;
  }
}

// appendSpan adds a `<span class="…">text</span>` child to the output panel.
// Used to color-code stdout / stderr / error / meta segments without HTML
// injection risk (textContent escapes everything).
function appendSpan(parent, className, text) {
  const span = document.createElement('span');
  span.className = className;
  span.textContent = text;
  parent.appendChild(span);
}

// ── Theme integration ──────────────────────────────────────────────────────
//
// The app's light/dark toggle flips `data-theme` on <html>. We mirror that
// onto CodeMirror's theme so the editor follows the rest of the page.

// currentCmTheme reads the current app theme and maps it to a CM theme name.
function currentCmTheme() {
  const t = document.documentElement.getAttribute('data-theme');
  return t === 'light' ? LIGHT_THEME : DARK_THEME;
}

// watchAppTheme installs a MutationObserver on <html data-theme> and calls
// cb('light' | 'dark') whenever it changes.
function watchAppTheme(cb) {
  const obs = new MutationObserver(() => {
    cb(document.documentElement.getAttribute('data-theme') === 'light' ? 'light' : 'dark');
  });
  obs.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
}

// ── Plain-textarea fallback ────────────────────────────────────────────────
//
// wirePlainTextarea is used only when CodeMirror failed to load (e.g. blocked
// by a strict CSP). The app stays usable, just without syntax highlighting,
// line numbers, or bracket matching.
function wirePlainTextarea(ta) {
  $('#run-code').addEventListener('click', run);
  $('#reset-code').addEventListener('click', () => {
    ta.value = ta.dataset.original || '';
  });
  ta.addEventListener('keydown', (e) => {
    if (e.key === 'Tab') {
      e.preventDefault();
      const { selectionStart: s, selectionEnd: e2 } = ta;
      ta.value = ta.value.substring(0, s) + '\t' + ta.value.substring(e2);
      ta.selectionStart = ta.selectionEnd = s + 1;
      return;
    }
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      run();
    }
  });
}
