// app.js — composition root for the frontend.
//
// Mirrors what main.go does on the backend: import each module, wire them
// together, kick off startup. No DOM logic of its own; that lives in the
// per-feature modules.
//
// The lessons module owns lesson state and notifies the playground + chat
// modules when the active lesson changes (via the onChange callback). That
// keeps lessons.js as the single owner of "what lesson is open" without
// either of the others having to know about each other.

import { initTheme } from './theme.js';
import { initLessons } from './lessons.js';
import { initPlayground, resetTo as resetPlayground } from './playground.js';
import { initChat, syncToLesson } from './chat.js';

// main bootstraps the page. Modules with no async setup go first; modules
// that need network (chat status) are awaited; finally the lessons module
// runs and fires its first onChange to populate the editor + chat panel.
async function main() {
  initTheme();
  initPlayground();
  await initChat();

  await initLessons({
    onChange(lesson) {
      resetPlayground(lesson.code || '');
      syncToLesson(lesson);
    },
  });
}

main();
