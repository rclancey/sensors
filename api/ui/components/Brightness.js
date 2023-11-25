import { h } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';

import { AsOf } from './AsOf.js';

const html = htm.bind(h);

export const Brightness = ({ light }) => html`<div className="brightness">
  <div className="header">Brightness</div>
  <div className="grid">
    <div className="key">Visible:</div>
    <div className="nval">${light.visible}</div>
    <div className="key">Infrared:</div>
    <div className="nval">${light.infrared}</div>
    <div className="key">Full Spectrum:</div>
    <div className="nval">${light.full_spectrum}</div>
    <div className="key">Lux:</div>
    <div className="nval">${light.lux}</div>
  </div>
  <${AsOf} now=${light.now} />
</div>`;

export default Brightness;
