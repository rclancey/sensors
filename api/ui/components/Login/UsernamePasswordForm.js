import { h, Fragment } from 'https://unpkg.com/preact@latest?module';
import { useCallback, useEffect, useState } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import { Username, Password } from '../Password.js';
import Button from '../Input/Button.js';
import ResetPasswordForm from './ResetPasswordForm.js';
import SocialLoginForm from './SocialLoginForm.js';

const html = htm.bind(h);

export const UsernamePasswordForm = ({ username = '', token }) => {
  const [tmpUsername, setUsername] = useState(username || '');
  const [password, setPassword] = useState('');
  const [forgot, setForgot] = useState(false);
  const [error, setError] = useState(null);
  const onLogin = useCallback(() => token.login(tmpUsername, password)
    .then(() => setError(null))
    .catch((err) => setError(`${err}`)), [token, tmpUsername, password]);
  const onEnter = useCallback((evt) => {
    if (tmpUsername && password) {
      onLogin();
    } else if (tmpUsername) {
      evt.target.form.elements.password.focus();
    } else {
      evt.target.form.elements.username.focus();
    }
  }, [tmpUsername, password, onLogin]);
  const onChange = useCallback(() => {
    const u = new URL(document.location);
    const state = {};
    u.searchParams.delete('reset');
    history.pushState(state, 'Login', u.toString());
    setForgot(false);
  }, []);
  const onForgot = useCallback(() => {
    console.debug('onForgot');
    token.resetPassword(tmpUsername)
      .then((resp) => {
        console.debug('reset password response: %o', resp);
        const u = new URL(document.location);
        const state = {
          reset: true,
          username: tmpUsername,
        };
        u.searchParams.set('reset', true);
        u.searchParams.set('username', tmpUsername);
        history.pushState(state, 'Reset Password', u.toString());
        setForgot(true);
      });
  }, [token, tmpUsername]);
  useEffect(() => {
    const h = () => {
      const u = new URL(document.location);
      if (u.searchParams.get('reset') !== null) {
        setForgot(true);
      } else {
        setForgot(false);
      }
    };
    window.addEventListener('popstate', h);
    h();
    return () => {
      window.removeEventListener('popstate', h);
    };
  }, []);
  if (forgot) {
    return html`<${ResetPasswordForm} username=${tmpUsername} token=${token} onChange=${onChange} />`;
  }
  return html`<Fragment>
    <div className="header">Synos: Login Required</div>
    <div>Username:</div>
    <div>
      <${Username} value=${tmpUsername} onChange=${setUsername} onEnter=${onEnter} />
    </div>
    <div>Password:</div>
    <div>
      <${Password} value=${password} onChange=${setPassword} onEnter=${onEnter} />
    </div>
    ${ error !== null ? html`<Fragment>
      <div />
      <div className="error">${error}</div>
    </Fragment>` : null }
    <div />
    <div>
      <${Button} onClick=${onLogin}>Login</${Button}>
    </div>
    <div />
    <div>
      <${Button} type="text" disabled=${tmpUsername === ''} onClick=${onForgot}>I forgot my password</${Button}>
    </div>
    <${SocialLoginForm} />
  </Fragment>`;
};

export default UsernamePasswordForm;
