import { state, activePreset, isFavoritePreset } from './state.js';
import { capture } from './hotkey.js';

export function escHtml(v) {
  return String(v)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;');
}

function escAttr(v) { return escHtml(v); }

export function renderApp(app) {
  const shell = app.querySelector('.shell');
  const scrollTop = shell ? shell.scrollTop : 0;

  if (!state.settings) {
    app.innerHTML = `
      <div class="app-window">
        <div class="titlebar">
          <div class="app-icon">&#xE740;</div>
          <span class="titlebar-title">ResizeMe</span>
          <div class="titlebar-controls">
            <button class="titlebar-btn" data-waction="minimise" title="Minimise">&#xE921;</button>
            <button class="titlebar-btn titlebar-btn-close" data-waction="hide" title="Close">&#xE8BB;</button>
          </div>
        </div>
        <div class="shell" style="display:flex;align-items:center;justify-content:center;color:var(--text-secondary);font-size:13px">
          Loading…
        </div>
      </div>`;
    return;
  }

  const s = state.settings;
  const preset = activePreset(s);
  const favoritePresetIds = s.favoritePresetIds ?? [];
  const favoriteIdSet = new Set(favoritePresetIds);
  const favoritePresets = favoritePresetIds
    .map(id => s.presets.find(p => p.id === id))
    .filter(Boolean);
  const otherPresets = s.presets.filter(p => !favoriteIdSet.has(p.id));

  app.innerHTML = `
    <div class="app-window">
      <div class="titlebar">
        <div class="app-icon">&#xE740;</div>
        <span class="titlebar-title">ResizeMe</span>
        <div class="titlebar-controls">
          <button class="titlebar-btn" data-waction="minimise" title="Minimise">&#xE921;</button>
          <button class="titlebar-btn titlebar-btn-close" data-waction="hide" title="Close">&#xE8BB;</button>
        </div>
      </div>

      <div class="shell">
        ${s.firstRun ? renderFirstRun() : ''}
        ${state.error ? `<div class="error-banner">${escHtml(state.error)}</div>` : ''}
        ${renderHotkeyCard(s)}

        <div class="hero">
          <div class="hero-name">${escHtml(preset?.name ?? 'No preset')}</div>
          <div class="hero-meta">
            <span class="dims-badge">${preset?.width ?? 0} × ${preset?.height ?? 0}</span>
          </div>

        </div>

        <div class="section-label">Presets</div>
        <div class="card-group">
          ${favoritePresets.length > 0 ? `
            <div class="preset-group-label">Favorites</div>
            ${favoritePresets.map(p => renderPresetRow(p, s.activePresetId)).join('')}
            <div class="preset-group-label">All Presets</div>
          ` : ''}
          ${otherPresets.map(p => renderPresetRow(p, s.activePresetId)).join('')}
        </div>

        <div class="add-preset-row">
          <button class="hyperlink-btn" data-action="add-preset">+ Add preset</button>
        </div>

        <div class="section-label">Options</div>
        <div class="card-group">
          <div class="settings-card">
            <span class="setting-label">Center after resize</span>
            <input type="checkbox" class="toggle-switch" data-field="centerAfterResize" ${s.centerAfterResize ? 'checked' : ''} />
          </div>
          <div class="settings-card">
            <span class="setting-label">Launch at startup</span>
            <input type="checkbox" class="toggle-switch" data-field="autoStart" ${s.autoStart ? 'checked' : ''} />
          </div>
        </div>
      </div>

      ${state.dialog !== null ? renderDialog() : ''}
    </div>
  `;

  if (scrollTop > 0) {
    const newShell = app.querySelector('.shell');
    if (newShell) newShell.scrollTop = scrollTop;
  }
}

function renderFirstRun() {
  return `
    <div class="first-run-card">
      <div class="first-run-title">Start ResizeMe with Windows?</div>
      <div class="first-run-desc">Recommended — stays in the tray until you need it.</div>
      <div class="first-run-actions">
        <button class="accent-btn" data-action="first-run-yes">Yes</button>
        <button class="standard-btn" data-action="first-run-no">Not now</button>
      </div>
    </div>`;
}

function renderPresetRow(p, activeId) {
  const isActive = p.id === activeId;
  const isFavorite = isFavoritePreset(p.id);
  return `
    <div class="preset-row${isActive ? ' active' : ''}" data-action="select-preset" data-id="${escAttr(p.id)}">
      <div class="radio-btn${isActive ? ' checked' : ''}">
        ${isActive ? '<div class="radio-dot"></div>' : ''}
      </div>
      <button class="preset-favorite${isFavorite ? ' active' : ''}" data-action="toggle-favorite" data-id="${escAttr(p.id)}" title="${isFavorite ? 'Remove from favorites' : 'Add to favorites'}">${isFavorite ? '&#xE735;' : '&#xE734;'}</button>
      <div class="preset-name">${escHtml(p.name)}</div>
      <div class="preset-dims">${p.width} × ${p.height}</div>
      <button class="preset-edit" data-action="edit-preset" data-id="${escAttr(p.id)}" title="Edit">&#xE70F;</button>
      <button class="preset-delete" data-action="delete-preset" data-id="${escAttr(p.id)}" title="Remove">&times;</button>
    </div>`;
}

function renderHotkeyCard(s) {
  if (capture.active) {
    const parts = [];
    if (capture.ctrl) parts.push('Ctrl');
    if (capture.alt) parts.push('Alt');
    if (capture.shift) parts.push('Shift');
    if (capture.win) parts.push('Win');
    if (capture.key) parts.push(capture.key);
    const previewHtml = parts.length > 0
      ? parts.map(p => `<kbd>${escHtml(p)}</kbd>`).join('<span class="key-sep">+</span>')
      : '<span class="capture-placeholder">Press a key combination…</span>';
    return `
      <div class="card-group" style="margin-bottom:14px">
        <div class="hotkey-capture-card capturing">
          <div class="hotkey-header">
            <span class="setting-label">Global hotkey</span>
            <div class="recording-indicator"><span class="recording-dot"></span> Recording</div>
          </div>
          <div class="capture-preview" id="capture-preview">${previewHtml}</div>
          <div class="capture-hint">Hold Ctrl, Alt, Shift, or Win — then press A–Z, 0–9, or F1–F24</div>
          <button class="standard-btn cancel-capture-btn" data-action="cancel-capture">Cancel</button>
        </div>
      </div>`;
  }

  const keysHtml = (s.hotkey || 'Ctrl+Alt+R').split('+')
    .map(p => `<kbd>${escHtml(p)}</kbd>`)
    .join('<span class="key-sep">+</span>');

  return `
    <div class="card-group" style="margin-bottom:14px">
      <div class="hotkey-capture-card" data-action="start-capture">
        <div class="hotkey-header">
          <span class="setting-label">Global hotkey</span>
          <span class="hotkey-edit-hint">click to change</span>
        </div>
        <div class="hotkey-key-display">${keysHtml}</div>
        ${state.hotkeyError ? `<div class="hotkey-error">${escHtml(state.hotkeyError)}</div>` : ''}
      </div>
    </div>`;
}

function renderDialog() {
  const d = state.dialog;
  const isEdit = d.mode === 'edit';
  const title = isEdit ? 'Edit preset' : 'Add preset';
  const confirmLabel = isEdit ? 'Save' : 'Add';
  return `
    <div class="dialog-overlay" data-action="close-dialog-overlay">
      <div class="dialog" data-stop-propagation>
        <div class="dialog-title">${title}</div>
        <div class="dialog-body">
          <div>
            <div class="field-label">Name</div>
            <input type="text" data-dialog-field="name" value="${escAttr(d.name)}" placeholder="My Preset" />
          </div>
          <div class="dialog-row">
            <div>
              <div class="field-label">Width</div>
              <input type="number" data-dialog-field="width" value="${d.width}" min="100" max="10000" />
            </div>
            <div>
              <div class="field-label">Height</div>
              <input type="number" data-dialog-field="height" value="${d.height}" min="100" max="10000" />
            </div>
          </div>
        </div>
        <div class="dialog-actions">
          <button class="standard-btn" data-action="cancel-dialog">Cancel</button>
          <button class="accent-btn" data-action="confirm-dialog">${confirmLabel}</button>
        </div>
      </div>
    </div>`;
}
