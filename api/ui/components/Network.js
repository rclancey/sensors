import { h } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';

import { AsOf } from './AsOf.js';

const html = htm.bind(h);

export const NetworkHost = ({ now, host }) => {
  const dt = new Date(host.last_seen);
  const mark = host.active ? '✓' : '✗';
  const className = host.active ? 'active' : 'inactive';
  return html`<span className="${className}">${mark}</span> ${host.ipv4} (${host.hostname})`;
}

export const Network = ({ info }) => {
  const now = new Date(info.now).getTime();
  return html`<div className="network">
    <div className="header">Hosts</div>
    <div className="grid">
      ${info.hosts.map((host) => (
        html`<div key="${host.mac}-val" className="val"><${NetworkHost} now=${now} host=${host} /></div>
        <div key="${host.mac}-val" className="key">${host.mac}</div>`
      ))}
    </div>
    <${AsOf} now=${info.now} />
  </div>`
};

export default Network;
