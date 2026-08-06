import { state, activePreset, hasDraftChanges, isFavoritePreset, isHiddenPreset, visiblePresets } from './state.js';
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

  const s = state.draft ?? state.settings;
  const preset = activePreset(state.settings);
  const favoritePresetIds = s.favoritePresetIds ?? [];
  const favoriteIdSet = new Set(favoritePresetIds);
  const visiblePresetList = visiblePresets(s);
  const visiblePresetById = new Map(visiblePresetList.map(p => [p.id, p]));
  const favoritePresets = favoritePresetIds
    .map(id => visiblePresetById.get(id))
    .filter(Boolean);
  const otherPresets = visiblePresetList.filter(p => !favoriteIdSet.has(p.id));
  const hiddenPresets = s.presets.filter(p => isHiddenPreset(p.id, s));
  const canHideVisiblePreset = visiblePresetList.length > 1;

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

      <main class="shell" ${state.dialog !== null || state.saving ? 'inert' : ''} aria-busy="${state.saving}">
        <header class="page-header">
          <h1>Settings</h1>
          <p>Changes stay in this window until you save them.</p>
        </header>

        ${state.error ? `<div class="error-banner" role="alert">${escHtml(state.error)}</div>` : ''}
        ${s.firstRun ? renderFirstRun(state.settings, preset) : ''}

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

          <h3 class="subsection-title">Presets (pixels)</h3>
          <div class="card-group preset-list">
            ${favoritePresets.length > 0 ? `
              <div class="preset-group-label">Favorites</div>
              ${favoritePresets.map(p => renderPresetRow(p, s.activePresetId, canHideVisiblePreset)).join('')}
              <div class="preset-group-label">All presets</div>
            ` : ''}
            ${otherPresets.map(p => renderPresetRow(p, s.activePresetId, canHideVisiblePreset)).join('')}
          </div>
          ${state.presetNotice ? `<p class="preset-notice" role="status">${escHtml(state.presetNotice)}</p>` : ''}

          <div class="section-action">
            <button type="button" class="hyperlink-btn" data-action="add-preset">+ Add preset</button>
          </div>

          <section class="hidden-presets" aria-labelledby="hidden-presets-title">
            <div class="section-heading">
              <h3 id="hidden-presets-title">Hidden presets (${hiddenPresets.length})</h3>
              <p>Hidden presets stay saved but do not appear in resize selectors or quick menus.</p>
            </div>
            <div class="card-group preset-list">
              ${hiddenPresets.length > 0
                ? hiddenPresets.map(renderHiddenPresetRow).join('')
                : '<p class="empty-presets">No presets are hidden.</p>'}
            </div>
          </section>
        </section>

        <section class="settings-section" aria-labelledby="hotkey-title">
          <div class="section-heading">
            <h2 id="hotkey-title">Hotkey</h2>
            <p>Resize the active window or open a quick-pick flyout from anywhere in Windows.</p>
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

        ${renderSettingsActions()}
      </main>

      ${state.dialog !== null ? renderDialog() : ''}
    </div>
  `;

  if (scrollTop > 0) {
    const newShell = app.querySelector('.shell');
    if (newShell) newShell.scrollTop = scrollTop;
  }
}

function renderSettingsActions() {
  const hasChanges = hasDraftChanges();
  return `
    <section class="settings-actions" aria-label="Save settings">
      <span class="settings-actions-copy">Changes apply when you save.</span>
      <div class="settings-actions-buttons">
        <button type="button" class="standard-btn" data-action="revert-settings" ${hasChanges && !state.saving ? '' : 'disabled'}>Revert</button>
        <button type="button" class="accent-btn" data-action="save-settings" ${hasChanges && !state.saving ? '' : 'disabled'}>${state.saving ? 'Saving…' : 'Save'}</button>
      </div>
    </section>`;
}

function renderFirstRun(s, preset) {
  const hotkeyKeys = renderHotkeyKeys(s.hotkey);
  const presetDescription = preset
    ? `${escHtml(preset.name)} (${preset.width} × ${preset.height} px)`
    : 'your active preset';

  return `
    <aside class="first-run-card" aria-labelledby="first-run-title">
      <div class="first-run-title" id="first-run-title">Welcome to ResizeMe</div>
      <div class="first-run-desc">Resize the active window to exact dimensions with a keyboard shortcut.</div>
      <div class="first-run-steps">
        <div class="first-run-step">
          <span class="first-run-step-number" aria-hidden="true">1</span>
          <div>
            <div class="first-run-step-title">Use your global hotkey</div>
            <div class="first-run-step-desc">Press ${hotkeyKeys} to resize the active window to ${presetDescription}.</div>
          </div>
        </div>
        <div class="first-run-step">
          <span class="first-run-step-number" aria-hidden="true">2</span>
          <div>
            <div class="first-run-step-title">Find ResizeMe in the system tray</div>
            <div class="first-run-step-desc">Closing Settings keeps ResizeMe ready in the background. Use the tray icon to open Settings or quit.</div>
          </div>
        </div>
      </div>
      <div class="first-run-startup">
        <div>
          <div class="first-run-step-title">Launch at startup <span class="first-run-optional">Optional</span></div>
          <div class="first-run-step-desc">Start ResizeMe in the system tray when you sign in to Windows.</div>
        </div>
      </div>
      <div class="first-run-actions">
        <button type="button" class="accent-btn" data-action="first-run-yes">Launch at startup</button>
        <button type="button" class="standard-btn" data-action="first-run-no">Get started</button>
      </div>
    </aside>`;
}

function renderPresetRow(p, activeId, canHide) {
  const isActive = p.id === activeId;
  const isFavorite = isFavoritePreset(p.id);
  return `
    <div class="preset-row${isActive ? ' active' : ''}">
      <button type="button" class="preset-select" data-action="select-preset" data-id="${escAttr(p.id)}" data-focus-id="select-${escAttr(p.id)}" aria-pressed="${isActive}">
        <span class="radio-btn${isActive ? ' checked' : ''}" aria-hidden="true">
          ${isActive ? '<span class="radio-dot"></span>' : ''}
        </span>
        <span class="preset-name">${escHtml(p.name)}</span>
        <span class="preset-dims">${p.width} × ${p.height} px</span>
      </button>
      <button type="button" class="preset-favorite${isFavorite ? ' active' : ''}" data-action="toggle-favorite" data-id="${escAttr(p.id)}" aria-label="${isFavorite ? 'Remove ' + escAttr(p.name) + ' from favorites' : 'Add ' + escAttr(p.name) + ' to favorites'}">${isFavorite ? '&#xE735;' : '&#xE734;'}</button>
      <button type="button" class="preset-hide" data-action="hide-preset" data-id="${escAttr(p.id)}" data-focus-id="hide-${escAttr(p.id)}" aria-label="Hide ${escAttr(p.name)}" ${canHide ? '' : 'disabled'}>Hide</button>
      <button type="button" class="preset-edit" data-action="edit-preset" data-id="${escAttr(p.id)}" aria-label="Edit ${escAttr(p.name)}">&#xE70F;</button>
      <button type="button" class="preset-delete" data-action="delete-preset" data-id="${escAttr(p.id)}" aria-label="Delete ${escAttr(p.name)} permanently">&times;</button>
    </div>`;
}

function renderHiddenPresetRow(p) {
  const isFavorite = isFavoritePreset(p.id);
  return `
    <div class="preset-row preset-row-hidden">
      <div class="preset-hidden-copy">
        <span class="preset-name">${escHtml(p.name)}</span>
        <span class="preset-dims">${p.width} × ${p.height} px</span>
        <span class="preset-hidden-state">Hidden${isFavorite ? ' · Favorite' : ''}</span>
      </div>
      <button type="button" class="preset-favorite${isFavorite ? ' active' : ''}" data-action="toggle-favorite" data-id="${escAttr(p.id)}" aria-label="${isFavorite ? 'Remove ' + escAttr(p.name) + ' from favorites' : 'Add ' + escAttr(p.name) + ' to favorites'}">${isFavorite ? '&#xE735;' : '&#xE734;'}</button>
      <button type="button" class="preset-show" data-action="restore-preset" data-id="${escAttr(p.id)}" data-focus-id="restore-${escAttr(p.id)}" aria-label="Restore ${escAttr(p.name)}">Show</button>
      <button type="button" class="preset-edit" data-action="edit-preset" data-id="${escAttr(p.id)}" aria-label="Edit ${escAttr(p.name)}">&#xE70F;</button>
      <button type="button" class="preset-delete" data-action="delete-preset" data-id="${escAttr(p.id)}" aria-label="Delete ${escAttr(p.name)} permanently">&times;</button>
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
    const targetLabel = capture.target === 'quickPickHotkey' ? 'Pick a size hotkey' : 'Resize hotkey';
    return `
      <div class="card-group">
        <div class="hotkey-capture-card capturing">
          <div class="hotkey-header">
            <span class="setting-label">${targetLabel}</span>
            <div class="recording-indicator"><span class="recording-dot"></span> Recording</div>
          </div>
          <div class="capture-preview" id="capture-preview">${previewHtml}</div>
          <div class="capture-hint">Hold Ctrl, Alt, Shift, or Win — then press A-Z, 0-9, or F1-F24</div>
          <button type="button" class="standard-btn cancel-capture-btn" data-action="cancel-capture">Cancel</button>
        </div>
      </div>`;
  }

  const conflictError = s.hotkey && s.quickPickHotkey && s.hotkey === s.quickPickHotkey
      ? 'Resize and pick-a-size hotkeys must be different.'
      : '';

  return `
      <div class="card-group">
        ${renderHotkeyOption('hotkey', 'Resize hotkey', 'Run the active resize preset immediately.', s.hotkey, state.hotkeyError || conflictError)}
        ${renderHotkeyOption('quickPickHotkey', 'Pick a size hotkey', 'Open the preset flyout at the focused window.', s.quickPickHotkey, state.quickPickHotkeyError || conflictError)}
      </div>`;
}

function renderHotkeyOption(target, label, description, hotkey, error) {
  return `
        <button type="button" class="hotkey-capture-card" data-action="start-capture" data-target="${target}">
          <div class="hotkey-header">
            <span class="setting-copy">
              <span class="setting-label">${label}</span>
              <span class="setting-description">${description}</span>
            </span>
            <span class="hotkey-edit-hint">Change</span>
          </div>
          <div class="hotkey-key-display">${renderHotkeyKeys(hotkey, target === 'quickPickHotkey' ? 'Ctrl+Alt+Shift+R' : 'Ctrl+Alt+R')}</div>
          ${error ? `<div class="hotkey-error" role="alert">${escHtml(error)}</div>` : ''}
        </button>`;
}

function renderHotkeyKeys(hotkey, fallback) {
  return (hotkey || fallback || 'Ctrl+Alt+R').split('+')
      .map(part => `<kbd>${escHtml(part)}</kbd>`)
      .join('<span class="key-sep">+</span>');
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
              Width (px)
              <input type="number" data-dialog-field="width" value="${d.width}" min="100" max="10000" />
            </label>
            <label class="field-label">
              Height (px)
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
          ${renderUpdateStatus()}
          <button type="button" class="hyperlink-btn about-link" data-action="open-project">View project on GitHub</button>
        </div>
        <div class="dialog-actions">
          <button type="button" class="accent-btn" data-action="close-about">Close</button>
        </div>
      </div>
    </div>`;
}

function renderUpdateStatus() {
  if (state.checkingForUpdates) {
    return `
      <section class="about-update" aria-labelledby="update-title" aria-live="polite">
        <div class="about-update-heading">
          <span class="setting-label" id="update-title">Updates</span>
          <button type="button" class="standard-btn" disabled>Checking…</button>
        </div>
        <p>Checking compatible published Windows releases…</p>
      </section>`;
  }

  if (state.updateError) {
    return `
      <section class="about-update" aria-labelledby="update-title" aria-live="polite">
        <div class="about-update-heading">
          <span class="setting-label" id="update-title">Updates</span>
          <button type="button" class="standard-btn" data-action="check-for-updates">Try again</button>
        </div>
        <p class="about-update-error" role="alert">${escHtml(state.updateError)}</p>
      </section>`;
  }

  if (state.update?.available) {
    return `
      <section class="about-update" aria-labelledby="update-title" aria-live="polite">
        <div class="about-update-heading">
          <span class="setting-label" id="update-title">Updates</span>
          <button type="button" class="standard-btn" data-action="check-for-updates">Check again</button>
        </div>
        <p>Version ${escHtml(state.update.latestVersion)} is published for this Windows architecture.</p>
        <p>ResizeMe updates safely through winget. Close ResizeMe, then run:</p>
        <code class="update-command">${escHtml(state.update.updateCommand)}</code>
        <div class="about-update-actions">
          <button type="button" class="standard-btn" data-action="copy-update-command">Copy command</button>
          <button type="button" class="hyperlink-btn" data-action="open-release">View release notes</button>
        </div>
        ${state.updateActionError ? `<p class="about-update-error" role="alert">${escHtml(state.updateActionError)}</p>` : ''}
        ${state.updateNotice ? `<p class="about-update-notice" role="status">${escHtml(state.updateNotice)}</p>` : ''}
        <p class="about-update-note">Winget publication can take a little time after a GitHub release is published.</p>
      </section>`;
  }

  if (state.update) {
    return `
      <section class="about-update" aria-labelledby="update-title" aria-live="polite">
        <div class="about-update-heading">
          <span class="setting-label" id="update-title">Updates</span>
          <button type="button" class="standard-btn" data-action="check-for-updates">Check again</button>
        </div>
        <p>You are up to date. Version ${escHtml(state.update.latestVersion)} is the latest compatible published Windows release.</p>
      </section>`;
  }

  return `
    <section class="about-update" aria-labelledby="update-title">
      <div class="about-update-heading">
        <span class="setting-label" id="update-title">Updates</span>
        <button type="button" class="standard-btn" data-action="check-for-updates">Check for updates</button>
      </div>
      <p>Check GitHub Releases for a compatible Windows update. ResizeMe installs updates only through winget.</p>
    </section>`;
}
