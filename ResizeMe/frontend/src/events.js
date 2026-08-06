import {
  state,
  selectPreset,
  toggleFavoritePreset,
  deletePreset,
  hidePreset,
  restorePreset,
  openAddDialog,
  openEditDialog,
  openAboutDialog,
  closeDialog,
  checkForUpdates,
  copyUpdateCommand,
  confirmDialog,
  toggleCenter,
  toggleAutoStart,
  saveDraft,
  revertDraft,
  completeFirstRun,
  resizeNow,
} from './state.js';
import { startCapture, stopCapture } from './hotkey.js';
import { BrowserOpenURL } from '../wailsjs/runtime/runtime';

const projectUrl = 'https://github.com/burkeholland/resize-me';

let dialogKeyHandler = null;

export function bindEvents(app, renderFn) {
  if (dialogKeyHandler) {
    document.removeEventListener('keydown', dialogKeyHandler);
    dialogKeyHandler = null;
  }

  app.querySelectorAll('[data-action]').forEach(el => {
    el.addEventListener('click', async e => {
      e.stopPropagation();
      const action = el.dataset.action;
      switch (action) {
        case 'first-run-yes':       await completeFirstRun(true, renderFn); break;
        case 'first-run-no':        await completeFirstRun(false, renderFn); break;
        case 'select-preset':await selectPreset(el.dataset.id, renderFn); break;
        case 'toggle-favorite':     await toggleFavoritePreset(el.dataset.id, renderFn); break;
        case 'delete-preset':       await deletePreset(el.dataset.id, renderFn); break;
        case 'hide-preset':         await hidePreset(el.dataset.id, renderFn); break;
        case 'restore-preset':      await restorePreset(el.dataset.id, renderFn); break;
        case 'edit-preset':         openEditDialog(el.dataset.id, renderFn); break;
        case 'add-preset':          openAddDialog(renderFn); break;
        case 'open-about':          openAboutDialog(renderFn); break;
        case 'open-project':        BrowserOpenURL(projectUrl); break;
        case 'open-release':        BrowserOpenURL(state.update?.releaseURL ?? `${projectUrl}/releases`); break;
        case 'check-for-updates':   await checkForUpdates(renderFn); break;
        case 'copy-update-command': await copyUpdateCommand(renderFn); break;
        case 'close-about':         closeDialog(renderFn); break;
        case 'resize-now':          await resizeNow(renderFn); break;
        case 'start-capture':       startCapture(renderFn, el.dataset.target); break;
        case 'cancel-capture':      stopCapture(renderFn); break;
        case 'close-dialog-overlay': closeDialog(renderFn); break;
        case 'cancel-dialog':       closeDialog(renderFn); break;
        case 'confirm-dialog':      await confirmDialog(renderFn); break;
        case 'save-settings':       await saveDraft(renderFn); break;
        case 'revert-settings':     revertDraft(renderFn); break;
      }
    });
  });

  const dialogEl = app.querySelector('[data-stop-propagation]');
  if (dialogEl) {
    dialogEl.addEventListener('click', e => e.stopPropagation());
  }

  const centerToggle = app.querySelector('[data-field="centerAfterResize"]');
  if (centerToggle) {
    centerToggle.addEventListener('change', () => toggleCenter(centerToggle.checked, renderFn));
  }

  const autoStartToggle = app.querySelector('[data-field="autoStart"]');
  if (autoStartToggle) {
    autoStartToggle.addEventListener('change', () => toggleAutoStart(autoStartToggle.checked, renderFn));
  }

  if (state.dialog !== null) {
    dialogKeyHandler = async e => {
      if (e.key === 'Escape') { closeDialog(renderFn); }
      else if (e.key === 'Enter') {
        if (state.dialog.mode === 'about') {
          if (document.activeElement?.dataset.action === 'close-about') {
            closeDialog(renderFn);
          }
          return;
        }
        await confirmDialog(renderFn);
      }
      else if (e.key === 'Tab') {
        const focusable = [...app.querySelectorAll('.dialog input, .dialog button:not(:disabled)')];
        if (focusable.length === 0) return;
        const first = focusable[0];
        const last = focusable[focusable.length - 1];
        if (e.shiftKey && document.activeElement === first) {
          e.preventDefault();
          last.focus();
        } else if (!e.shiftKey && document.activeElement === last) {
          e.preventDefault();
          first.focus();
        }
      }
    };
    document.addEventListener('keydown', dialogKeyHandler);
  }
}
