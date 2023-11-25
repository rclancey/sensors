import { h } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';

import { AsOf } from './AsOf.js';

const html = htm.bind(h);

export const AppleTV = ({ info }) => (
  html`<div className="appletv">
    <div className="header">AppleTV</div>
    <div className="grid">
      <div className="key">Power:</div>
      <div className="val">${info.power_state}</div>
      <div className="key">Status:</div>
      <div className="val">${info.device_state}</div>
      ${info.device_state !== 'idle' && (
        html`<div className="key">App:</div>
          <div className="val">${info.app}</div>
          <div className="key">Title:</div>
          <div className="val">${info.title}</div>`
      )}
    </div>
    <${AsOf} now=${info.now} />
  </div>`
);

export default AppleTV;
