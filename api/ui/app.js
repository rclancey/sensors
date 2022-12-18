import { Fragment, h, Component, render } from 'https://unpkg.com/preact@latest?module';
import { useState, useEffect, useCallback, useMemo } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

// Initialize htm with Preact
const html = htm.bind(h);

const datefmt = Intl.DateTimeFormat('en-US', { timeStyle: 'medium', hour12: false });

const AsOf = ({ now }) => (
  html`<div className="asof">As of ${datefmt.format(new Date(now))}</div>`
);

const Comment = ({ children }) => (
  html`<span className="comment">// ${children}</span>`
);

const Toggle = ({ open, onToggle }) => (
  html`<div className="toggleWrapper">
    <div className="toggle ${open ? 'open' : 'closed'}" onClick=${onToggle} />
  </diiv>`
);

const Comma = ({ last }) => last ? null : html`<span className="comma">,</span>`;

const PairVal = ({ val }) => {
  if (val === null) {
    return html`<${NullVal} />`;
  }
  switch(typeof val) {
  case 'string':
    return html`<${Str} str=${val} />`;
  case 'number':
    return html`<${Num} num=${val} />`;
  case 'boolean':
    return html`<${Bool} bool=${val} />`;
  }
  return null;
};

const Pair = ({ name, val, last, depth=1 }) => {
  const [open, setOpen] = useState(depth < 4);
  const onToggle = useCallback(() => setOpen((orig) => !orig), []);
  if (Array.isArray(val)) {
    if (open) {
      if (val.length === 0) {
        return html`<div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <${Key} name=${name} />
          <span>: [ ]</span>
          <${Comma} last=${last} />
        </div>`;
      }
      return html`<Fragment>
        <div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <${Key} name=${name} />
          <span>: [</span>
        </div>
        <div className="child">
          <${Arr} arr=${val} depth=${depth+1} />
        </div>
        <div className="closing">
          <span>]</span>
          <${Comma} last=${last} />
        </div>
      </Fragment>`;
    }
    return html`<div className="pair closed">
      <${Toggle} open=${open} onToggle=${onToggle} />
      <${Key} name=${name} />
      <span>: [...]</span>
      <${Comma} last=${last} />
      <${Comment}>${val.length} ${val.length === 1 ? 'item' : 'items'}</${Comment}>
    </div>`;
  }
  if (val !== null && val !== undefined && typeof val === 'object') {
    if (open) {
      if (Object.keys(val).length === 0) {
        return html`<div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <${Key} name=${name} />
          <span>: { }</span>
          <${Comma} last=${last} />
        </div>`;
      }
      return html`<Fragment>
        <div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <${Key} name=${name} />
          <span>: {</span>
        </div>
        <div className="child">
          <${Obj} obj=${val} depth=${depth+1} />
        </div>
        <div className="closing">
          <span>}</span>
          <${Comma} last=${last} />
        </div>
      </Fragment>`;
    }
    return html`<div className="pair closed">
      <${Toggle} open=${open} onToggle=${onToggle} />
      <${Key} name=${name} />
      <span>: {...}</span>
      <${Comma} last=${last} />
      <${Comment}>${Object.keys(val).length} ${Object.keys(val).length === 1 ? 'item' : 'items'}</${Comment}>
    </div>`
  }
  return html`<div className="pair">
    <${Key} name=${name} />
    <span>: </span>
    <${PairVal} val=${val} />
    <${Comma} last=${last} />
  </div>`;
};

const Obj = ({ obj, depth }) => {
  const keys = Object.keys(obj).sort((a, b) => a < b ? -1 : 1);
  const last = keys.length - 1;
  return keys.map((k, i) => html`<${Pair} key=${k} name=${k} val=${obj[k]} last=${i === last} depth=${depth} />`);
};

const ArrVal = ({ val }) => {
  if (val === null) {
    return html`<${NullVal} />`;
  }
  switch (typeof val) {
  case 'string':
    return html`<${Str} str=${val} />`;
  case 'number':
    return html`<${Num} num=${val} />`;
  case 'boolean':
    return html`<${Bool} bool=${val} />`;
  }
  return null;
};

const ArrItem = ({ item, last, depth=1 }) => {
  const [open, setOpen] = useState(depth < 4);
  const onToggle = useCallback(() => setOpen((orig) => !orig), []);
  if (Array.isArray(item)) {
    if (open) {
      if (item.length === 0) {
        return html`<div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <span>[ ]</span>
          <${Comma} last=${last} />
        </div>`;
      }
      return html`<Fragment>
        <div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <span>[</span>
        </div>
        <div className="child">
          <${Arr} arr=${item} depth=${depth+1} />
        </div>
        <div className="closing">
          <span>]</span>
          <${Comma} last=${last} />
        </div>
      </Fragment>`;
    }
    return html`<div className="pair closed">
      <${Toggle} open=${open} onToggle=${onToggle} />
      <span>[...]</span>
      <${Comma} last=${last} />
      <${Comment}>${item.length} ${item.length === 1 ? 'item' : 'items'}</${Comment}>
    </div>`;
  }
  if (typeof item === 'object') {
    if (open) {
      if (Object.keys(item).length === 0) {
        return html`<div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <span>{ }</span>
          <${Comma} last=${last} />
        </div>`;
      }
      return html`<Fragment>
        <div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <span>{</span>
        </div>
        <div className="child">
          <${Obj} obj=${item} depth=${depth+1} />
        </div>
        <div className="closing">
          <span>}</span>
          <${Comma} last=${last} />
        </div>
      </Fragment>`;
    }
    return html`<div className="pair closed">
      <${Toggle} open=${open} onToggle=${onToggle} />
      <span>{...}</span>
      <${Comma} last=${last} />
      <${Comment}>${Object.keys(item).length} ${Object.keys(item).length === 1 ? 'item' : 'items'}</${Comment}>
    </div>`
  }
  return html`<div className="pair">
    <${ArrVal} val=${item} />
    <${Comma} last=${last} />
  </div>`;
};

const Arr = ({ arr, depth }) => {
  const last = arr.length - 1;
  return arr.map((item, i) => html`<${ArrItem} key=${i} item=${item} last=${i === last} depth=${depth} />`);
};

const Key = ({ name }) => html`<div className="key">"${name}"</div>`;

const Str = ({ str }) => html`<div className="string">"${str}"</div>`;

const Num = ({ num }) => html`<div className="number">${num}</div>`;

const Bool = ({ bool }) => html`<div className="boolean">${bool.toString()}</div>`;

const NullVal = () => html`<div className="null">null</div>`;

const Tree = ({ tree }) => html`<div className="tree"><${ArrItem} item=${tree} last /></div>`;

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

/*
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
*/

const Brightness = ({ light }) => (
  html`<div className="brightness">
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
        html`<div key="${host.mac}-val" className="val"><${NetworkHost} now=${now} host=${host} /></div>
        <div key="${host.mac}-val" className="key">${host.mac}</div>`
      ))}
    </div>
    <${AsOf} now=${info.now} />
  </div>`
};

const temp = (val, unitSystem) => {
  switch (unitSystem) {
  case 'metric':
    return `${Intl.NumberFormat('en-US', { minimumFractionDigits: 1, maximumFractionDigits: 2 }).format(val - 273.15)} °C`;
  case 'imperial':
    return `${Math.round((val - 273.15) * 1.8) + 32} °F`;
  case 'scientific':
    let notation = 'standard';
    if (val > 0) {
      if (val < 0.1 || val >= 10000) {
        notation = 'scientific';
      }
    }
    return `${Intl.NumberFormat('en-US', { minimumSignificantDigits: 2, maximumSignificantDigits: 4, notation }).format(val)} K`;
  }
  return `${val}`;
};

const pres = (val, unitSystem) => {
  switch (unitSystem) {
  case 'metric':
    return `${Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(val * 0.7501)} mmHg`;
  case 'imperial':
    return `${Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(val * 0.7501 / 25.4)} in`;
  case 'scientific':
    return `${val} kPa`;
  }
  return `${val}`;
};

const Weather = ({ weather }) => {
  if (!weather) {
    return null;
  }
  const current = weather.current;
  const forecast = weather.daily && weather.daily[0];
  if (!current || !forecast) {
    return null;
  }
  const unitSystem = 'imperial';
  return html`<div className="weather">
    <div className="header">Weather</div>
    <div className="grid">
      <div className="key">Temperature:</div>
      <div className="val">${temp(current.temp, unitSystem)}</div>
      <div className="key">Low / High:</div>
      <div className="val">${temp(forecast.temp.min, unitSystem)} / ${temp(forecast.temp.max, unitSystem)}</div>
      <div className="key">Humidity:</div>
      <div className="val">${current.humidity} %</div>
      <div className="key">Pressure:</div>
      <div className="val">${pres(current.pressure, unitSystem)}</div>
      <div className="key">Clouds:</div>
      <div className="val">${current.clouds} %</div>
    </div>
    <${AsOf} now=${weather.current.dt * 1000} />
  </div>`;
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
      <${Weather} weather=${state.weather} />
      <${Brightness} light=${state.brightness} />
      <${Motion} motion=${state.motion} />
      <${Lights} lights=${state.lights} />
      <${AppleTV} info=${state.appletv} />
      <${Sonos} info=${state.sonos} />
      <${Network} info=${state.network} />
      <${Tree} tree=${state} />
    </div>
    <button onClick=${onUpdate}>Refresh</button>
  </div>`;
};

function App (props) {
  return html`<${Status} />`;
  //return html`<h1>Hello ${props.name}!</h1>`;
}

render(html`<${App} name="World" />`, document.body);
