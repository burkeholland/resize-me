import {
  GetSettings,
  SaveSettingsIfUnchanged,
  CompleteFirstRun,
  ResizeNow,
  GetVersion,
  CheckForUpdates,
} from '../wailsjs/go/main/App';
import { ClipboardSetText } from '../wailsjs/runtime/runtime';

export const state = {
  settings: null,
  draft: null,
  draftBase: null,
  version: '',
  dialog: null,
  error: '',
  hotkeyError: '',
  quickPickHotkeyError: '',
  saving: false,
  update: null,
  updateError: '',
  updateActionError: '',
  updateNotice: '',
  checkingForUpdates: false,
  presetNotice: '',
};

export function clearError() { state.error = ''; }

function setError(err, renderFn) {
  state.error = err?.message ?? String(err);
  renderFn();
}

export function clone(v) { return JSON.parse(JSON.stringify(v)); }

function persistedSettings(settings) {
  if (!settings) return null;
  const { loadError, ...persisted } = settings;
  return persisted;
}

function sameSettings(left, right) {
  return JSON.stringify(persistedSettings(left)) === JSON.stringify(persistedSettings(right));
}

export function activePreset(settings) {
  const presets = visiblePresets(settings);
  return presets.find(p => p.id === settings.activePresetId) ?? presets[0];
}

export function isHiddenPreset(id, settings = state.draft ?? state.settings) {
  return (settings?.hiddenPresetIds ?? []).includes(id);
}

export function visiblePresets(settings) {
  if (!settings) return [];
  const hiddenPresetIds = new Set(settings.hiddenPresetIds ?? []);
  return settings.presets.filter(preset => !hiddenPresetIds.has(preset.id));
}

export function isFavoritePreset(id) {
  const favorites = state.draft?.favoritePresetIds ?? [];
  return favorites.includes(id);
}

export function hasDraftChanges() {
  return state.draft !== null && !sameSettings(state.draft, state.settings);
}

function resetDraft() {
  state.draft = clone(state.settings);
  state.draftBase = clone(state.settings);
  state.hotkeyError = '';
  state.quickPickHotkeyError = '';
  state.presetNotice = '';
}

export function receiveSettingsUpdate(settings, renderFn) {
  const hadChanges = hasDraftChanges();
  state.settings = settings;
  if (!hadChanges) {
    resetDraft();
  }
  renderFn();
}

export async function load(renderFn) {
  try {
    const [settings, version] = await Promise.all([GetSettings(), GetVersion()]);
    state.settings = settings;
    state.version = version;
    resetDraft();
    if (settings.loadError) {
      state.error = settings.loadError;
    }
    renderFn();
  } catch (err) {
    setError(err, renderFn);
  }
}

export function selectPreset(id, renderFn) {
  if (state.saving) return;
  clearError();
  state.draft.activePresetId = id;
  renderFn();
}

export function deletePreset(id, renderFn) {
  if (state.saving) return;
  const preset = state.draft.presets.find(candidate => candidate.id === id);
  if (!preset) {
    state.error = 'Preset no longer exists.';
    renderFn();
    return;
  }
  const hidden = isHiddenPreset(id);
  if (!hidden && visiblePresets(state.draft).length <= 1) {
    state.error = 'At least one visible preset is required.';
    renderFn();
    return;
  }
  clearError();
  const presets = state.draft.presets.filter(p => p.id !== id);
  const hiddenPresetIds = (state.draft.hiddenPresetIds ?? []).filter(hiddenId => hiddenId !== id);
  const visible = visiblePresets({ ...state.draft, presets, hiddenPresetIds });
  const activeId = visible.some(p => p.id === state.draft.activePresetId)
    ? state.draft.activePresetId
    : visible[0].id;
  const favoritePresetIds = (state.draft.favoritePresetIds ?? []).filter(favoriteId => favoriteId !== id);
  state.draft = {
    ...clone(state.draft),
    presets,
    activePresetId: activeId,
    favoritePresetIds,
    hiddenPresetIds,
  };
  state.presetNotice = `Deleted ${preset.name} permanently.`;
  renderAndFocus(`select-${activeId}`, renderFn);
}

export function hidePreset(id, renderFn) {
  if (state.saving) return;
  const preset = state.draft.presets.find(candidate => candidate.id === id);
  const visible = visiblePresets(state.draft);
  if (!preset || !visible.some(candidate => candidate.id === id)) {
    state.error = 'Preset is no longer available.';
    renderFn();
    return;
  }
  if (visible.length <= 1) {
    state.error = 'At least one visible preset is required.';
    renderFn();
    return;
  }

  clearError();
  const replacement = visible.find(candidate => candidate.id !== id);
  const hiddenPresetIds = [...new Set([...(state.draft.hiddenPresetIds ?? []), id])];
  const hidActivePreset = state.draft.activePresetId === id;
  const activePresetId = hidActivePreset
    ? replacement.id
    : state.draft.activePresetId;
  state.draft = { ...clone(state.draft), hiddenPresetIds, activePresetId };
  state.presetNotice = hidActivePreset
    ? `Hidden ${preset.name}. ${replacement.name} is now the active preset.`
    : `Hidden ${preset.name}.`;
  renderAndFocus(`restore-${id}`, renderFn);
}

export function restorePreset(id, renderFn) {
  if (state.saving) return;
  const preset = state.draft.presets.find(candidate => candidate.id === id);
  if (!preset || !isHiddenPreset(id)) {
    state.error = 'Preset is no longer hidden.';
    renderFn();
    return;
  }

  clearError();
  const hiddenPresetIds = (state.draft.hiddenPresetIds ?? []).filter(hiddenId => hiddenId !== id);
  state.draft = { ...clone(state.draft), hiddenPresetIds };
  state.presetNotice = `Restored ${preset.name}.`;
  renderAndFocus(`hide-${id}`, renderFn);
}

function renderAndFocus(focusId, renderFn) {
  renderFn();
  requestAnimationFrame(() => {
    const target = [...document.querySelectorAll('[data-focus-id]')]
      .find(element => element.dataset.focusId === focusId);
    target?.focus();
  });
}

export function toggleFavoritePreset(id, renderFn) {
  if (state.saving) return;
  clearError();
  const favoritePresetIds = [...(state.draft.favoritePresetIds ?? [])];
  const existing = favoritePresetIds.indexOf(id);
  if (existing >= 0) {
    favoritePresetIds.splice(existing, 1);
  } else {
    favoritePresetIds.push(id);
  }

  state.draft = { ...clone(state.draft), favoritePresetIds };
  renderFn();
}

export function saveHotkey(hotkey, target, renderFn) {
  const field = target === 'quickPickHotkey' ? 'quickPickHotkey' : 'hotkey';
  const errorField = field === 'quickPickHotkey' ? 'quickPickHotkeyError' : 'hotkeyError';
  if (!hotkey || hotkey === state.draft[field] || state.saving) return;
  state[errorField] = '';
  state.draft = { ...clone(state.draft), [field]: hotkey };
  renderFn();
}

function friendlyHotkeyError(msg) {
  if (/already registered/i.test(msg)) {
    return 'That combination is already in use by another app — try a different one.';
  }
  return msg;
}

export function toggleCenter(checked, renderFn) {
  if (state.saving) return;
  clearError();
  state.draft.centerAfterResize = checked;
  renderFn();
}

export function toggleAutoStart(checked, renderFn) {
  if (state.saving) return;
  clearError();
  state.draft.autoStart = checked;
  renderFn();
}

export function revertDraft(renderFn) {
  if (state.saving) return;
  clearError();
  resetDraft();
  renderFn();
}

export async function saveDraft(renderFn) {
  if (!hasDraftChanges() || state.saving) return;
  clearError();
  state.hotkeyError = '';
  state.quickPickHotkeyError = '';
  if (state.draft.hotkey && state.draft.quickPickHotkey && state.draft.hotkey === state.draft.quickPickHotkey) {
    const message = 'Resize and pick-a-size hotkeys must be different.';
    state.hotkeyError = message;
    state.quickPickHotkeyError = message;
    renderFn();
    return;
  }
  state.saving = true;
  renderFn();

  try {
    const saved = await SaveSettingsIfUnchanged(clone(state.draft), clone(state.draftBase));
    state.settings = saved;
    resetDraft();
  } catch (err) {
    const message = err?.message ?? String(err);
    if (/settings changed elsewhere/i.test(message)) {
      try {
        state.settings = await GetSettings();
        resetDraft();
      } catch (loadError) {
        state.error = `${message} Unable to load the latest settings: ${loadError?.message ?? String(loadError)}`;
        return;
      }
    }
    if (/quick-pick hotkey/i.test(message) || (state.draft.quickPickHotkey && message.includes(state.draft.quickPickHotkey))) {
      state.quickPickHotkeyError = friendlyHotkeyError(message);
    } else if (/register hotkey|already registered/i.test(message)) {
      state.hotkeyError = friendlyHotkeyError(message);
    } else {
      state.error = message;
    }
  } finally {
    state.saving = false;
    renderFn();
  }
}

export async function resizeNow(renderFn) {
  clearError();
  try {
    await ResizeNow();
    renderFn();
  } catch (err) {
    setError(err, renderFn);
  }
}

export async function completeFirstRun(enable, renderFn) {
  clearError();
  try {
    const updated = await CompleteFirstRun(enable);
    receiveSettingsUpdate(updated, renderFn);
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

  if (state.saving) return;
  let presets;
  if (d.mode === 'edit') {
    presets = clone(state.draft.presets).map(p =>
      p.id === d.id ? { ...p, name, width, height } : p
    );
  } else {
    presets = [...clone(state.draft.presets), { id: '', name, width, height }];
  }
  state.draft = { ...clone(state.draft), presets };
  state.dialog = null;
  renderFn();
}

export function openAddDialog(renderFn) {
  if (state.saving) return;
  state.dialog = { mode: 'add', name: 'Custom', width: 1920, height: 1080 };
  renderFn();
  setTimeout(() => {
    const input = document.querySelector('[data-dialog-field="name"]');
    if (input) { input.focus(); input.select(); }
  }, 0);
}

export function openEditDialog(id, renderFn) {
  if (state.saving) return;
  const preset = state.draft.presets.find(p => p.id === id);
  if (!preset) return;
  state.dialog = { mode: 'edit', id: preset.id, name: preset.name, width: preset.width, height: preset.height };
  renderFn();
  setTimeout(() => {
    const input = document.querySelector('[data-dialog-field="name"]');
    if (input) { input.focus(); input.select(); }
  }, 0);
}

export function openAboutDialog(renderFn) {
  state.updateError = '';
  state.updateActionError = '';
  state.updateNotice = '';
  state.dialog = { mode: 'about' };
  renderFn();
  setTimeout(() => document.querySelector('[data-action="close-about"]')?.focus(), 0);
}

export async function checkForUpdates(renderFn) {
  if (state.checkingForUpdates) return;
  state.update = null;
  state.updateError = '';
  state.updateActionError = '';
  state.updateNotice = '';
  state.checkingForUpdates = true;
  renderFn();

  try {
    state.update = await CheckForUpdates();
  } catch (err) {
    state.updateError = `Unable to check for updates: ${err?.message ?? String(err)}`;
  } finally {
    state.checkingForUpdates = false;
    renderFn();
  }
}

export async function copyUpdateCommand(renderFn) {
  const command = state.update?.updateCommand;
  if (!command) return;
  state.updateActionError = '';
  state.updateNotice = '';

  try {
    const copied = await ClipboardSetText(command);
    if (!copied) {
      throw new Error('Windows did not confirm that the command was copied.');
    }
    state.updateNotice = 'Update command copied to the clipboard.';
  } catch (err) {
    state.updateActionError = `Unable to copy the update command: ${err?.message ?? String(err)}`;
  }
  renderFn();
}

export function closeDialog(renderFn) {
  state.dialog = null;
  renderFn();
}
