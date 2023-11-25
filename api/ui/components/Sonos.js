import { h } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';

import { AsOf } from './AsOf.js';

const html = htm.bind(h);

const tfmt = Intl.NumberFormat('en-US', { minimumIntegerDigits: 2 });

const msTime = (ms) => {
  const hr = Math.floor(ms / 3600000);
  const mn = Math.floor((ms % 3600000) / 60000);
  const sc = Math.floor((ms % 60000) / 1000);
  return `${tfmt.format(hr)}:${tfmt.format(mn)}:${tfmt.format(sc)}`;
};

export const Sonos = ({ info }) => (
  html`<div className="sonos">
    <div className="header">Sonos</div>
    <div className="grid">
      <div className="key">Status:</div>
      <div className="val">${info.state}</div>
      <div className="key">Volume:</div>
      <div className="val">${info.volume}</div>
      <div className="key">Playlist:</div>
      <div className="val">${info.index + 1} / ${info.tracks.length}</div>
      <div className="key">Track:</div>
      <div className="val">${info.tracks[info.index].artist} - ${info.tracks[info.index].title}</div>
      <div className="key">Time:</div>
      <div className="val">${msTime(info.time)} / ${msTime(info.duration)}</div>
    </div>
    <${AsOf} now=${info.now} />
  </div>`
);

export default Sonos;
