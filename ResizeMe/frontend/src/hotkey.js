import { saveHotkey } from './state.js';

export const capture = {
  active: false,
  ctrl: false,
  alt: false,
  shift: false,
  win: false,
  key: null,
  hint: '',
};

function mapKey(key) {
  if (key.length === 1) {
    const upper = key.toUpperCase();
    if (/[A-Z0-9]/.test(upper)) return upper;
  }
  const fMatch = key.match(/^F(\d+)$/);
  if (fMatch) {
    const n = parseInt(fMatch[1]);
    if (n >= 1 && n <= 24) return key;
  }
  return null;
}

export function buildCombo() {
  const parts = [];
  if (capture.ctrl) parts.push('Ctrl');
  if (capture.alt) parts.push('Alt');
  if (capture.shift) parts.push('Shift');
  if (capture.win) parts.push('Win');
  if (capture.key) parts.push(capture.key);
  return parts.join('+');
}

function updateCaptureUI() {
  const preview = document.getElementById('capture-preview');
  if (!preview) return;
  const parts = [];
  if (capture.ctrl) parts.push('Ctrl');
  if (capture.alt) parts.push('Alt');
  if (capture.shift) parts.push('Shift');
  if (capture.win) parts.push('Win');
  if (capture.key) parts.push(capture.key);
  preview.innerHTML = parts.length > 0
    ? parts.map(p => `<kbd>${p}</kbd>`).join('<span class="key-sep">+</span>')
    : '<span class="capture-placeholder">Press a key combination…</span>';
  const hintEl = document.querySelector('.capture-hint');
  if (hintEl) hintEl.textContent = capture.hint || 'Hold Ctrl, Alt, Shift, or Win — then press A–Z, 0–9, or F1–F24';
}

let renderFnRef = null;

export function startCapture(renderFn) {
  renderFnRef = renderFn;
  capture.active = true;
  capture.ctrl = false;
  capture.alt = false;
  capture.shift = false;
  capture.win = false;
  capture.key = null;
  capture.hint = '';
  renderFn();
  document.addEventListener('keydown', onCaptureKeyDown, true);
  document.addEventListener('keyup', onCaptureKeyUp, true);
}

export function stopCapture(renderFn) {
  document.removeEventListener('keydown', onCaptureKeyDown, true);
  document.removeEventListener('keyup', onCaptureKeyUp, true);
  capture.active = false;
  const fn = renderFn || renderFnRef;
  renderFnRef = null;
  fn();
}

function onCaptureKeyDown(e) {
  e.preventDefault();
  e.stopPropagation();
  if (e.key === 'Escape') { stopCapture(); return; }
  // Track each modifier via its own key event so that ctrlKey/altKey
  // properties on subsequent keydown events (which can be false on Windows
  // due to AltGraph synthesis) don't silently reset the modifier state.
  if (e.key === 'Control') { capture.ctrl = true; }
  else if (e.key === 'Alt' || e.key === 'AltGraph') { capture.alt = true; }
  else if (e.key === 'Shift') { capture.shift = true; }
  else if (e.key === 'Meta') { capture.win = true; }
  else {
    const k = mapKey(e.key);
    if (k) { capture.key = k; capture.hint = ''; }
    else { capture.key = null; capture.hint = 'Unsupported key — use A–Z, 0–9, or F1–F24'; }
  }
  updateCaptureUI();
}

async function onCaptureKeyUp(e) {
  e.preventDefault();
  e.stopPropagation();
  if (e.key === 'Escape') { stopCapture(); return; }
  // Clear modifier flags on their own keyup so the display stays accurate.
  if (e.key === 'Control') { capture.ctrl = false; updateCaptureUI(); return; }
  if (e.key === 'Alt' || e.key === 'AltGraph') { capture.alt = false; updateCaptureUI(); return; }
  if (e.key === 'Shift') { capture.shift = false; updateCaptureUI(); return; }
  if (e.key === 'Meta') { capture.win = false; updateCaptureUI(); return; }
  if (!capture.key) return;
  const hasModifier = capture.ctrl || capture.alt || capture.shift || capture.win;
  if (!hasModifier) {
    capture.hint = 'Add a modifier: Ctrl, Alt, Shift, or Win';
    updateCaptureUI();
    return;
  }
  const combo = buildCombo();
  const renderFn = renderFnRef;
  stopCapture(renderFn);
  await saveHotkey(combo, renderFn);
}
