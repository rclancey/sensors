import { h, render } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';

import { WithLogin } from './components/Login/WithLogin.js';
import { Status } from './components/Status.js';

// Initialize htm with Preact
const html = htm.bind(h);

function App(props) {
  return html`<${WithLogin}><${Status} /></${WithLogin}>`;
}

render(html`<${App} name="World" />`, document.body);
