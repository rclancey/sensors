import { h } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';

import { AsOf } from './AsOf.js';

const html = htm.bind(h);

export const Light = ({ name, light }) => {
  const { api } = useContext(AuthContext);
  const [state, setState] = useState(light);
  useEffect(() => setState(light), [light]);
  const turnOn = useCallback(() => {
    api.fetch(`/lights/${name}?value=100`, { method: 'PUT' })
      .then(() => setState(100));
  }, [api, name]);
  const turnOff = useCallback(() => {
    api.fetch(`/lights/${name}?value=0`, { method: 'PUT' })
      .then(() => setState(0));
  }, [api, name])
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

export const Lights = ({ lights }) => (
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

export default Lights;
