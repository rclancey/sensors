import { h, Component, render } from 'https://unpkg.com/preact@latest?module';
import { useState, useEffect, useCallback, useMemo } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

// Initialize htm with Preact
const html = htm.bind(h);

const datefmt = Intl.DateTimeFormat('en-US', { timeStyle: 'medium', hour12: false });

const AsOf = ({ now }) => (
  html`<div className="asof">As of ${datefmt.format(new Date(now))}</div>`
);

const methods = [
  'GET',
  'POST',
  'PUT',
  'DELETE',
];

const directions = [
  'increasing',
  'decreasing',
];

const Webhook = ({ kind }) => {
  const [method, setMethod] = useState(methods[0]);
  const [url, setUrl] = useState('');
  const [direction, setDirection] = useState(directions[0]);
  const [threshold, setThreshold] = useState(0);
  const [resetThreshold, setResetThreshold] = useState(0);
  const [webhooks, setWebhooks] = useState([]);
  const onUpdate = useCallback(() => fetch(`/${kind}/webhook`, { method: 'GET' }).then((resp) => resp.json()).then(setWebhooks), [kind]);
  useEffect(() => onUpdate(), [onUpdate]);
  const onAdd = useCallback(() => {
    const hook = {
      callback_method: method,
      callback_url: url,
      direction,
      trigger_threshold: threshold,
      reset_threshold: resetThreshold,
    };
    fetch(`/${kind}/webhook`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(hook) })
      .then((resp) => resp.json())
      .then(console.debug);
  }, [kind, method, url, direction, threshold, resetThreshold]);
  const onDel = useCallback((evt) => {
    const hook = webhooks[parseInt(evt.target.dataset.idx)];
    fetch(`/${kind}/webhook`, { method: 'DELETE', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(hook) })
      .then((resp) => resp.json())
      .then(onUpdate);
  }, [kind, webhooks, onUpdate]);
  const onSetMethod = useCallback((evt) => setMethod(evt.target.value), []);
  const onSetUrl = useCallback((evt) => setUrl(evt.target.value), []);
  const onSetDirection = useCallback((evt) => setDirection(evt.target.value), []);
  const onSetThreshold = useCallback((evt) => setThreshold(parseFloat(evt.target.value)), []);
  const onSetResetThreshold = useCallback((evt) => setResetThreshold(parseFloat(evt.target.value)), []);
  return html`<div className="webhook">
    <div className="grid">
      <div className="key">URL:</div>
      <div className="val">
        <select value=${method} onChange=${onSetMethod}>
          ${methods.map((meth) => html`<option key=${meth} value=${meth}>${meth}</option>`)}
        </select>
        <input type="url" value=${url} onInput=${onSetUrl} />
      </div>
      <div className="key">Dir:</div>
      <div className="val">
        <select value=${direction} onChange=${onSetDirection}>
          ${directions.map((dir) => html`<option key=${dir} value=${dir}>${dir}</option>`)}
        </select>
      </div>
      <div className="key">Thresh:</div>
      <div className="val">
        <input type="number" value=${threshold} onInput=${onSetThreshold} />
      </div>
      <div className="key">Reset:</div>
      <div className="val">
        <input type="number" value=${resetThreshold} onInput=${onSetResetThreshold} />
      </div>
    </div>
    <button onClick=${onAdd}>Add Webhook</button>
    ${webhooks.map((hook, i) => (
      html`<div data-idx=${i} onClick=${onDel}>${hook.callback_method} ${hook.callback_url}</div>`
    ))}
  </div>`
};

const Brightness = ({ light }) => (
  html`<div className="brightness">
    <div className="header">Brightness</div>
    <div className="grid">
      <div className="key">Visible:</div>
      <div className="nval">${light.visible}</div>
      <div className="key">Infrared:</div>
      <div className="vnal">${light.infrared}</div>
      <div className="key">Full Spectrum:</div>
      <div className="vnal">${light.full_spectrum}</div>
      <div className="key">Lux:</div>
      <div className="vnal">${light.lux}</div>
    </div>
    <${AsOf} now=${light.now} />
  </div>`
);

const tfmt = Intl.NumberFormat('en-US', { minimumIntegerDigits: 2 });

const ElapsedTime = ({ since }) => {
  const [now, setNow] = useState(Date.now());
  useEffect(() => {
    const interval = setInterval(() => setNow(Date.now()), 1000);
    return () => {
      clearInterval(interval);
    };
  }, []);
  const elapsed = now - new Date(since).getTime();
  if (elapsed < 0 || elapsed > 86400000) {
    return '--:--:--';
  }
  const hr = Math.floor(elapsed / 3600000);
  const mn = Math.floor(elapsed / 60000) % 60;
  const sc = Math.floor(elapsed / 1000) % 60;
  return `${tfmt.format(hr)}:${tfmt.format(mn)}:${tfmt.format(sc)}`;
};

const Motion = ({ motion }) => {
  if (new Date(motion.last_motion).getTime() < 0) {
    return null;
  }
  const elapsed = Date.now() - new Date(motion.last_motion).getTime();
  const lastMotion = elapsed > 0 && elapsed < 86400000 ? datefmt.format(new Date(motion.last_motion)) : '--:--:--';
  return html`<div className="motion">
    <div className="header">Motion</div>
    <div className="grid">
      <div className="key">Updated:</div>
      <div className="val">${datefmt.format(new Date(motion.now))}</div>
      <div className="key">Last motion:</div>
      <div className="val">${lastMotion}</div>
      <div className="key">Elapsed time:</div>
      <div className="val"><${ElapsedTime} since=${motion.last_motion} /></div>
    </div>
    <${AsOf} now=${motion.now} />
    <${Webhook} kind="motion" />
  </div>`
};

/*
const isOn = (device) => {
  const info = device.info.system.get_sysinfo;
  if (info.light_state) {
    return info.light_state.brightness || 0;
  }
  return (info.relay_state || 0) * 100;
}
*/

const Light = ({ name, light }) => {
  const [state, setState] = useState(light);
  useEffect(() => setState(light), [light]);
  const turnOn = useCallback(() => {
    fetch(`/lights/${name}?value=100`, { method: 'PUT' })
      .then(() => setState(100));
  }, [name]);
  const turnOff = useCallback(() => {
    fetch(`/lights/${name}?value=0`, { method: 'PUT' })
      .then(() => setState(0));
  }, [name])
  const onToggle = useCallback(() => {
    if (state > 0) {
      turnOff();
    } else {
      turnOn();
    }
  }, [state, turnOn, turnOff]);
  return html`<div className="key">${name}</div>
    <div className="val">
      <input type="checkbox" checked=${state > 0} onClick=${onToggle} />
      ${state < 0 ? '?' : state}
    </div>`
};

const Lights = ({ lights }) => (
  html`<div className="lights">
    <div className="header">Lights</div>
    <div className="grid">
      ${Object.keys(lights.devices).sort().map((k) => (
        html`<${Light} key=${k} name=${k} light=${lights.devices[k]} />`
      ))}
    </div>
    <${AsOf} now=${lights.now} />
  </div>`
);

const AppleTV = ({ info }) => (
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

const msTime = (ms) => {
  const hr = Math.floor(ms / 3600000);
  const mn = Math.floor((ms % 3600000) / 60000);
  const sc = Math.floor((ms % 60000) / 1000);
  return `${tfmt.format(hr)}:${tfmt.format(mn)}:${tfmt.format(sc)}`;
};

const Sonos = ({ info }) => (
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

const NetworkHost = ({ now, host }) => {
  const dt = new Date(host.last_seen);
  const mark = host.active ? '✓' : '✗';
  const className = host.active ? 'active' : 'inactive';
  return html`<span className="${className}">${mark}</span> ${host.ipv4} (${host.hostname})`;
}

const Network = ({ info }) => {
  const now = new Date(info.now).getTime();
  return html`<div className="network">
    <div className="header">Hosts</div>
    <div className="grid">
      ${info.hosts.map((host) => (
        html`<div key="${host.mac}-key" className="key">${host.mac}</div>
        <div key="${host.mac}-val" className="val"><${NetworkHost} now=${now} host=${host} /></div>`
      ))}
    </div>
    <${AsOf} now=${info.now} />
  </div>`
};

const Status = () => {
  const [state, setState] = useState(null);
  const onUpdate = useCallback(() => {
    fetch('/status', { method: 'GET' })
      .then((resp) => resp.json())
      .then(setState);
  }, []);
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
      <${Brightness} light=${state.brightness} />
      <${Motion} motion=${state.motion} />
      <${Lights} lights=${state.lights} />
      <${AppleTV} info=${state.appletv} />
      <${Sonos} info=${state.sonos} />
      <${Network} info=${state.network} />
    </div>
    <button onClick=${onUpdate}>Refresh</button>
  </div>`;
};

function App (props) {
  return html`<${Status} />`;
  //return html`<h1>Hello ${props.name}!</h1>`;
}

render(html`<${App} name="World" />`, document.body);
