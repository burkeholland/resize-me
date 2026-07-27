import './style.css';
import { state, load, openAboutDialog } from './state.js';
import { renderApp } from './render.js';
import { bindEvents } from './events.js';
import { EventsOn, WindowHide, WindowMinimise } from '../wailsjs/runtime/runtime';

const app = document.querySelector('#app');

function render() {
  renderApp(app);
  bindEvents(app, render);
}

EventsOn('settings-updated', settings => {
  state.settings = settings;
  render();
});

EventsOn('show-about', () => openAboutDialog(render));

document.addEventListener('click', e => {
  const action = e.target.closest('[data-waction]')?.dataset.waction;
  if (action === 'hide') WindowHide();
  else if (action === 'minimise') WindowMinimise();
});

load(render);
