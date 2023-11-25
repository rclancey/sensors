import { h } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';

import { AsOf } from './AsOf.js';

const html = htm.bind(h);

const tfmt = Intl.NumberFormat('en-US', { minimumIntegerDigits: 2 });

export const ElapsedTime = ({ since }) => {
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

export const Motion = ({ motion }) => {
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

export default Motion;
