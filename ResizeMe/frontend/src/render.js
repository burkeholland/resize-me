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
            <button type="button" class="titlebar-btn" data-waction="minimise" aria-label="Minimise">&#xE921;</button>
            <button type="button" class="titlebar-btn titlebar-btn-close" data-waction="hide" aria-label="Close">&#xE8BB;</button>
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
      <div class="titlebar" ${state.dialog !== null ? 'inert' : ''}>
        <div class="app-icon">&#xE740;</div>
        <span class="titlebar-title">ResizeMe</span>
        <div class="titlebar-controls">
          <button type="button" class="titlebar-btn" data-waction="minimise" aria-label="Minimise">&#xE921;</button>
          <button type="button" class="titlebar-btn titlebar-btn-close" data-waction="hide" aria-label="Close">&#xE8BB;</button>
        </div>
      </div>

      <main class="shell" ${state.dialog !== null ? 'inert' : ''}>
        <header class="page-header">
          <h1>Settings</h1>
          <p>Choose how ResizeMe sizes windows and works in the background.</p>
        </header>

        ${state.error ? `<div class="error-banner" role="alert">${escHtml(state.error)}</div>` : ''}
        ${s.firstRun ? renderFirstRun() : ''}

        <section class="current-resize" aria-labelledby="current-resize-title">
          <div>
            <div class="current-resize-label">Current resize preset</div>
            <h2 id="current-resize-title">${escHtml(preset?.name ?? 'No preset')}</h2>
            <p>${preset?.width ?? 0} × ${preset?.height ?? 0} pixels</p>
          </div>
          <button type="button" class="accent-btn" data-action="resize-now">Resize now</button>
        </section>

        <section class="settings-section" aria-labelledby="resize-behavior-title">
          <div class="section-heading">
            <h2 id="resize-behavior-title">Resize behavior</h2>
            <p>Choose the size ResizeMe applies and what happens afterward.</p>
          </div>

          <div class="card-group">
            <label class="settings-card">
              <span class="setting-copy">
                <span class="setting-label">Center after resizing</span>
                <span class="setting-description">Move the resized window to the center of its current screen.</span>
              </span>
              <input type="checkbox" class="toggle-switch" data-field="centerAfterResize" ${s.centerAfterResize ? 'checked' : ''} />
            </label>
          </div>

          <h3 class="subsection-title">Presets</h3>
          <div class="card-group preset-list">
            ${favoritePresets.length > 0 ? `
              <div class="preset-group-label">Favorites</div>
              ${favoritePresets.map(p => renderPresetRow(p, s.activePresetId)).join('')}
              <div class="preset-group-label">All presets</div>
            ` : ''}
            ${otherPresets.map(p => renderPresetRow(p, s.activePresetId)).join('')}
          </div>

          <div class="section-action">
            <button type="button" class="hyperlink-btn" data-action="add-preset">+ Add preset</button>
          </div>
        </section>

        <section class="settings-section" aria-labelledby="hotkey-title">
          <div class="section-heading">
            <h2 id="hotkey-title">Hotkey</h2>
            <p>Run the current resize preset from anywhere in Windows.</p>
          </div>
          ${renderHotkeyCard(s)}
        </section>

        <section class="settings-section" aria-labelledby="general-title">
          <div class="section-heading">
            <h2 id="general-title">General</h2>
            <p>Control when ResizeMe is available in the system tray.</p>
          </div>
          <div class="card-group">
            <label class="settings-card">
              <span class="setting-copy">
                <span class="setting-label">Launch ResizeMe at startup</span>
                <span class="setting-description">Start in the system tray when you sign in to Windows.</span>
              </span>
              <input type="checkbox" class="toggle-switch" data-field="autoStart" ${s.autoStart ? 'checked' : ''} />
            </label>
            <button type="button" class="settings-card about-row" data-action="open-about">
              <span class="setting-copy">
                <span class="setting-label">About ResizeMe</span>
                <span class="setting-description">Version information, project details, and updates.</span>
              </span>
              <span class="row-chevron" aria-hidden="true">&#xE76C;</span>
            </button>
          </div>
        </section>
      </main>

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
    <aside class="first-run-card" aria-labelledby="first-run-title">
      <div class="first-run-title" id="first-run-title">Start ResizeMe with Windows?</div>
      <div class="first-run-desc">Keep ResizeMe ready in the system tray after you sign in.</div>
      <div class="first-run-actions">
        <button type="button" class="accent-btn" data-action="first-run-yes">Launch at startup</button>
        <button type="button" class="standard-btn" data-action="first-run-no">Not now</button>
      </div>
    </aside>`;
}

function renderPresetRow(p, activeId) {
  const isActive = p.id === activeId;
  const isFavorite = isFavoritePreset(p.id);
  return `
    <div class="preset-row${isActive ? ' active' : ''}">
      <button type="button" class="preset-select" data-action="select-preset" data-id="${escAttr(p.id)}" aria-pressed="${isActive}">
        <span class="radio-btn${isActive ? ' checked' : ''}" aria-hidden="true">
          ${isActive ? '<span class="radio-dot"></span>' : ''}
        </span>
        <span class="preset-name">${escHtml(p.name)}</span>
        <span class="preset-dims">${p.width} × ${p.height}</span>
      </button>
      <button type="button" class="preset-favorite${isFavorite ? ' active' : ''}" data-action="toggle-favorite" data-id="${escAttr(p.id)}" aria-label="${isFavorite ? 'Remove ' + escAttr(p.name) + ' from favorites' : 'Add ' + escAttr(p.name) + ' to favorites'}">${isFavorite ? '&#xE735;' : '&#xE734;'}</button>
      <button type="button" class="preset-edit" data-action="edit-preset" data-id="${escAttr(p.id)}" aria-label="Edit ${escAttr(p.name)}">&#xE70F;</button>
      <button type="button" class="preset-delete" data-action="delete-preset" data-id="${escAttr(p.id)}" aria-label="Remove ${escAttr(p.name)}">&times;</button>
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
      <div class="card-group">
        <div class="hotkey-capture-card capturing">
          <div class="hotkey-header">
            <span class="setting-label">Global hotkey</span>
            <div class="recording-indicator"><span class="recording-dot"></span> Recording</div>
          </div>
          <div class="capture-preview" id="capture-preview">${previewHtml}</div>
          <div class="capture-hint">Hold Ctrl, Alt, Shift, or Win — then press A–Z, 0–9, or F1–F24</div>
          <button type="button" class="standard-btn cancel-capture-btn" data-action="cancel-capture">Cancel</button>
        </div>
      </div>`;
  }

  const keysHtml = (s.hotkey || 'Ctrl+Alt+R').split('+')
    .map(p => `<kbd>${escHtml(p)}</kbd>`)
    .join('<span class="key-sep">+</span>');

  return `
    <div class="card-group">
      <button type="button" class="hotkey-capture-card" data-action="start-capture">
        <div class="hotkey-header">
          <span class="setting-copy">
            <span class="setting-label">Global hotkey</span>
            <span class="setting-description">Select this row, then press a new key combination.</span>
          </span>
          <span class="hotkey-edit-hint">Change</span>
        </div>
        <div class="hotkey-key-display">${keysHtml}</div>
        ${state.hotkeyError ? `<div class="hotkey-error" role="alert">${escHtml(state.hotkeyError)}</div>` : ''}
      </button>
    </div>`;
}

function renderDialog() {
  const d = state.dialog;
  if (d.mode === 'about') {
    return renderAboutDialog();
  }
  const isEdit = d.mode === 'edit';
  const title = isEdit ? 'Edit preset' : 'Add preset';
  const confirmLabel = isEdit ? 'Save' : 'Add';
  return `
    <div class="dialog-overlay" data-action="close-dialog-overlay">
      <div class="dialog" role="dialog" aria-modal="true" aria-labelledby="dialog-title" data-stop-propagation>
        <div class="dialog-title" id="dialog-title">${title}</div>
        <div class="dialog-body">
          <label class="field-label">
            Name
            <input type="text" data-dialog-field="name" value="${escAttr(d.name)}" placeholder="My Preset" />
          </label>
          <div class="dialog-row">
            <label class="field-label">
              Width
              <input type="number" data-dialog-field="width" value="${d.width}" min="100" max="10000" />
            </label>
            <label class="field-label">
              Height
              <input type="number" data-dialog-field="height" value="${d.height}" min="100" max="10000" />
            </label>
          </div>
        </div>
        <div class="dialog-actions">
          <button type="button" class="standard-btn" data-action="cancel-dialog">Cancel</button>
          <button type="button" class="accent-btn" data-action="confirm-dialog">${confirmLabel}</button>
        </div>
      </div>
    </div>`;
}

function renderAboutDialog() {
  return `
    <div class="dialog-overlay">
      <div class="dialog about-dialog" role="dialog" aria-modal="true" aria-labelledby="about-title" data-stop-propagation>
        <div class="about-icon" aria-hidden="true">&#xE740;</div>
        <div class="dialog-title" id="about-title">ResizeMe</div>
        <div class="about-version">Version ${escHtml(state.version)}</div>
        <div class="dialog-body about-body">
          <p>Quickly resize the active window with custom presets and a global hotkey.</p>
          <p>Updates are available through GitHub Releases or <code>winget upgrade BurkeHolland.ResizeMe</code>.</p>
          <button type="button" class="hyperlink-btn about-link" data-action="open-project">View project on GitHub</button>
        </div>
        <div class="dialog-actions">
          <button type="button" class="accent-btn" data-action="close-about">Close</button>
        </div>
      </div>
    </div>`;
}
