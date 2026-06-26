import {
  state,
  selectPreset,
  toggleFavoritePreset,
  deletePreset,
  openAddDialog,
  openEditDialog,
  closeDialog,
  confirmDialog,
  toggleCenter,
  toggleAutoStart,
  completeFirstRun,
} from './state.js';
import { startCapture, stopCapture } from './hotkey.js';

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
        case 'edit-preset':         openEditDialog(el.dataset.id, renderFn); break;
        case 'add-preset':          openAddDialog(renderFn); break;
        case 'start-capture':       startCapture(renderFn); break;
        case 'cancel-capture':      stopCapture(renderFn); break;
        case 'close-dialog-overlay': closeDialog(renderFn); break;
        case 'cancel-dialog':       closeDialog(renderFn); break;
        case 'confirm-dialog':      await confirmDialog(renderFn); break;
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
      else if (e.key === 'Enter') { await confirmDialog(renderFn); }
    };
    document.addEventListener('keydown', dialogKeyHandler);
  }
}
