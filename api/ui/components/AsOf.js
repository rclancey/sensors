import { h } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';

const html = htm.bind(h);

const datefmt = Intl.DateTimeFormat('en-US', { timeStyle: 'medium', hour12: false });

export const AsOf = ({ now }) => (
  html`<div className="asof">As of ${datefmt.format(new Date(now))}</div>`
);

export default AsOf;
