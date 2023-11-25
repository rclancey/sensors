import { h, Fragment } from 'https://unpkg.com/preact@latest?module';
import { useCallback, useMemo, useState } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import { TwoFactorCode } from '../Password.js';
import Button from '../Input/Button.js';

const html = htm.bind(h);

export const TwoFactor = ({ token }) => {
  const [code, setCode] = useState('');
  const [error, setError] = useState(null);
  const onAuth = useCallback(() => {
    token.twoFactor(code)
      .catch((err) => setError(`${err}`));
  }, [token, code]);
  return html`<Fragment>
    <div colspan={2}>
      <p>Enter the 6-digit code from your authenticator app</p>
    </div>
    <div>Code:</div>
    <div>
      <${TwoFactorCode} value=${code} onChange=${setCode} />
    </div>
    <div />
    <div>
      <${Button} disabled=${code.length != 6} onClick=${onAuth}>Login</Button>
    </div>
  </Fragment>`;
};

export default TwoFactor;
