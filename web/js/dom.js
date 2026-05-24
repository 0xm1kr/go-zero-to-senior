// dom.js — tiny DOM query helpers.
//
// Saves a hundred document.querySelector calls. The whole module is two
// lines because the alternative is a small build step or a runtime
// dependency, and neither is worth the cost for an app this size.

// $ returns the first element matching `sel` (or null).
export const $ = (sel) => document.querySelector(sel);

// $$ returns ALL elements matching `sel` as a real Array (not a
// NodeList), so .map/.filter/.find work without conversion.
export const $$ = (sel) => Array.from(document.querySelectorAll(sel));
