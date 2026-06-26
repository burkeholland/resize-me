import {
  GetSettings,
  SaveSettings,
  SetActivePreset,
  SetAutoStart,
  SetCenterAfterResize,
  CompleteFirstRun,
  ResizeNow,
} from '../wailsjs/go/main/App';

export const state = {
  settings: null,
  dialog: null,
  error: '',
  hotkeyError: '',
};

let _requestSeq = 0;
function nextSeq() { return ++_requestSeq; }
function isStale(seq) { return seq !== _requestSeq; }

export function clearError() { state.error = ''; }

function setError(err, renderFn) {
  state.error = err?.message ?? String(err);
  renderFn();
}

export function clone(v) { return JSON.parse(JSON.stringify(v)); }

export function activePreset(settings) {
  return settings.presets.find(p => p.id === settings.activePresetId) ?? settings.presets[0];
}

export function isFavoritePreset(id) {
  const favorites = state.settings?.favoritePresetIds ?? [];
  return favorites.includes(id);
}

export async function load(renderFn) {
  try {
    const settings = await GetSettings();
    state.settings = settings;
    if (settings.loadError) {
      state.error = settings.loadError;
    }
    renderFn();
  } catch (err) {
    setError(err, renderFn);
  }
}

export async function selectPreset(id, renderFn) {
  clearError();
  const seq = nextSeq();
  try {
    const updated = await SetActivePreset(id);
    if (isStale(seq)) return;
    state.settings = updated;
    renderFn();
  } catch (err) {
    if (!isStale(seq)) setError(err, renderFn);
  }
}

export async function deletePreset(id, renderFn) {
  if (state.settings.presets.length <= 1) {
    state.error = 'At least one preset is required.';
    renderFn();
    return;
  }
  clearError();
  const presets = state.settings.presets.filter(p => p.id !== id);
  const activeId = presets.find(p => p.id === state.settings.activePresetId)
    ? state.settings.activePresetId
    : presets[0].id;
  const favoritePresetIds = (state.settings.favoritePresetIds ?? []).filter(favoriteId => favoriteId !== id);
  const updated = { ...clone(state.settings), presets, activePresetId: activeId, favoritePresetIds };
  const seq = nextSeq();
  try {
    const saved = await SaveSettings(updated);
    if (isStale(seq)) return;
    state.settings = saved;
    renderFn();
  } catch (err) {
    if (!isStale(seq)) setError(err, renderFn);
  }
}

export async function toggleFavoritePreset(id, renderFn) {
  clearError();
  const seq = nextSeq();
  const favoritePresetIds = [...(state.settings.favoritePresetIds ?? [])];
  const existing = favoritePresetIds.indexOf(id);
  if (existing >= 0) {
    favoritePresetIds.splice(existing, 1);
  } else {
    favoritePresetIds.push(id);
  }

  const updated = { ...clone(state.settings), favoritePresetIds };
  try {
    const saved = await SaveSettings(updated);
    if (isStale(seq)) return;
    state.settings = saved;
    renderFn();
  } catch (err) {
    if (!isStale(seq)) setError(err, renderFn);
  }
}

export async function saveHotkey(hotkey, renderFn) {
  if (!hotkey || hotkey === state.settings.hotkey) return;
  state.hotkeyError = '';
  const updated = { ...clone(state.settings), hotkey };
  const seq = nextSeq();
  try {
    const saved = await SaveSettings(updated);
    if (isStale(seq)) return;
    state.settings = saved;
    renderFn();
  } catch (err) {
    if (!isStale(seq)) {
      state.hotkeyError = friendlyHotkeyError(err?.message ?? String(err));
      renderFn();
    }
  }
}

function friendlyHotkeyError(msg) {
  if (/already registered/i.test(msg)) {
    return 'That combination is already in use by another app — try a different one.';
  }
  return msg;
}

export async function toggleCenter(checked, renderFn) {
  clearError();
  const seq = nextSeq();
  try {
    const updated = await SetCenterAfterResize(checked);
    if (isStale(seq)) return;
    state.settings = updated;
    renderFn();
  } catch (err) {
    if (!isStale(seq)) setError(err, renderFn);
  }
}

export async function toggleAutoStart(checked, renderFn) {
  clearError();
  const seq = nextSeq();
  try {
    const updated = await SetAutoStart(checked);
    if (isStale(seq)) return;
    state.settings = updated;
    renderFn();
  } catch (err) {
    if (!isStale(seq)) setError(err, renderFn);
  }
}

export async function resizeNow(renderFn) {
  clearError();
  try {
    await ResizeNow();
  } catch (err) {
    setError(err, renderFn);
  }
}

export async function completeFirstRun(enable, renderFn) {
  clearError();
  try {
    const updated = await CompleteFirstRun(enable);
    state.settings = updated;
    renderFn();
  } catch (err) {
    setError(err, renderFn);
  }
}

export async function confirmDialog(renderFn) {
  const d = state.dialog;
  if (!d) return;
  const nameEl = document.querySelector('[data-dialog-field="name"]');
  const widthEl = document.querySelector('[data-dialog-field="width"]');
  const heightEl = document.querySelector('[data-dialog-field="height"]');
  const name = (nameEl?.value.trim()) || 'Custom';
  const width = Math.max(100, Math.min(10000, Number(widthEl?.value) || 1920));
  const height = Math.max(100, Math.min(10000, Number(heightEl?.value) || 1080));

  let presets;
  if (d.mode === 'edit') {
    presets = clone(state.settings.presets).map(p =>
      p.id === d.id ? { ...p, name, width, height } : p
    );
  } else {
    presets = [...clone(state.settings.presets), { id: '', name, width, height }];
  }
  const updated = { ...clone(state.settings), presets };
  try {
    const saved = await SaveSettings(updated);
    state.settings = saved;
    state.dialog = null;
    renderFn();
  } catch (err) {
    setError(err, renderFn);
  }
}

export function openAddDialog(renderFn) {
  state.dialog = { mode: 'add', name: 'Custom', width: 1920, height: 1080 };
  renderFn();
  setTimeout(() => {
    const input = document.querySelector('[data-dialog-field="name"]');
    if (input) { input.focus(); input.select(); }
  }, 0);
}

export function openEditDialog(id, renderFn) {
  const preset = state.settings.presets.find(p => p.id === id);
  if (!preset) return;
  state.dialog = { mode: 'edit', id: preset.id, name: preset.name, width: preset.width, height: preset.height };
  renderFn();
  setTimeout(() => {
    const input = document.querySelector('[data-dialog-field="name"]');
    if (input) { input.focus(); input.select(); }
  }, 0);
}

export function closeDialog(renderFn) {
  state.dialog = null;
  renderFn();
}
