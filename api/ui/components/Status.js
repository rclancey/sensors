import { h } from 'https://unpkg.com/preact@latest?module';
import { useCallback, useContext, useEffect, useMemo, useState } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import { LoginContext } from '../context/LoginContext.js';
import { Lights } from './Lights.js';
import { Weather } from './Weather.js';
import { Brightness } from './Brightness.js';
import { Motion } from './Motion.js';
import { AppleTV } from './AppleTV.js';
import { Sonos } from './Sonos.js';
import { Network } from './Network.js';
import { Tree } from './Tree.js';

const html = htm.bind(h);

class API {
  constructor(token, logout) {
    this.token = token;
    this.logout = logout;
  }

  async fetch(uri, opts) {
    if (!opts.headers) {
      opts.headers = {};
    }
    opts.credentials = 'same-origin';
    opts.headers['Authorization'] = `Bearer ${this.token}`;
    return fetch(uri, opts).then(resp => {
      if (resp.status === 401) {
        this.logout();
      }
      return resp;
    });
  }
}

export const Status = () => {
  const { token, onLoginRequired } = useContext(LoginContext);
  const api = useMemo(() => new API(token, logout), [token, logout]);
  const [state, setState] = useState(null);
  const onUpdate = useCallback(() => {
    api.fetch('/status', { method: 'GET' })
      .then((resp) => resp.json())
      .then(setState);
  }, [api]);
  useEffect(() => {
    onUpdate();
    const interval = setInterval(() => onUpdate(), [60000]);
    return () => clearInterval(interval);
  }, [onUpdate]);
  if (state === null) {
    return null;
  }
  return html`<div>
    <div className="status">
      <${Lights} lights=${state.lights} />
      <${Weather} weather=${state.weather} />
      <${Brightness} light=${state.brightness} />
      <${Motion} motion=${state.motion} />
      <${AppleTV} info=${state.appletv} />
      <${Sonos} info=${state.sonos} />
      <${Network} info=${state.network} />
      <${Tree} tree=${state} />
    </div>
    <button onClick=${onUpdate}>Refresh</button>
  </div>`;
};

export default Status;
