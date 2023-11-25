import { h, Fragment } from 'https://unpkg.com/preact@latest?module';
import { forwardRef } from 'https://unpkg.com/preact@10.16.0/compat/dist/compat.module.js?module';
import { useEffect, useMemo, useState } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import { checkLoginCookie, doLogin, LoginContext } from '../../lib/login.js';
import { Center } from '../Center.js';

const html = htm.bind(h);

export const CheckLogin = ({ theme, dark, children }) => {
  const [login, setLogin] = useState({ loggedIn: null, username: null });
  const [userinfo, setUserinfo] = useState(null);
  useEffect(() => {
    setLogin(checkLoginCookie());
  }, []);
  const onLoginRequired = useMemo(() => {
    return () => setLogin(orig => Object.assign({}, orig, { loggedIn: false }));
  }, [setLogin]);
  const onLogin = useMemo(() => {
    return (username, password) => doLogin(username, password).then(setLogin);
  }, [setLogin])
  useEffect(() => {
    if (!login.loggedIn) {
      setUserinfo(null);
    } else {
      fetch('/api/admin/user/__myself__', { method: 'GET' })
        .then((resp) => resp.json())
        .then(setUserinfo)
        .catch((err) => console.error("bad userinfo: %o", err));
    }
  }, [login]);
  const ctx = useMemo(() => {
    return { ...login, userinfo, onLoginRequired };
  }, [login, userinfo, onLoginRequired]);
  if (login.loggedIn === null) {
    return null;
  }
  if (login.loggedIn) {
    return html`<${LoginContext.Provider} value=${ctx}>
      ${children}
    </${LoginContext.Provider}>`;
  }
  return html`<${Login} theme=${theme} dark=${dark} username=${login.username} onLogin=${onLogin} />`;
};

const GithubButton = () => html`<a href="/api/login/github" className="github">
  <span className="logo">${'\u00a0'}</span>
  Login with GitHub
</a>`;

const GoogleButton = () => html`<a href="/api/login/google" className="google">
  <span className="logo">${'\u00a0'}</span>
  Sign in with Google
</a>`;

const AmazonButton = () => html`<a href="/api/login/amazon" className="amazon">
  <span className="fab fa-amazon"/>
  Login with Amazon
</a>`;

const FacebookButton = () => html`<a href="/auth/facebook" className="facebook">
  <span className="logo">${'\u00a0'}</span>
  Login with Facebook
</a>`;

const AppleButton = () => html`<a href="/auth/apple" className="apple">
  <span className="fab fa-apple"/>
  Sign in with Apple
</a>`;

export const Login = ({ theme, dark, username, onLogin }) => {
  const api = useAPI(API);
  const [tmpUsername, setUsername] = useState(username || '');
  const [password, setPassword] = useState('');
  const [error, setError] = useState(null);
  const sendLogin = useCallback(() => {
  }, [api]);
  return html`<div id="app" className=${`${theme} ${dark ? 'dark' : 'light'}`}>
    <${Center} orientation="horizontal" style=${{ height: '100vh' }}>
      <${Center} orientation="vertical">
        <div className="login">
          <div className="header">Synos: Login Required</div>
          <div>Username:</div>
          <div>
            <input
              type="text"
              value=${tmpUsername}
              onInput=${evt => setUsername(evt.target.value)}
            />
          </div>
          <div>Password:</div>
          <div>
            <input
              type="password"
              value=${password}
              onInput=${evt => setPassword(evt.target.value)}
            />
          </div>
          ${ error !== null ? (<>
            <div />
            <div className="error">${error}</div>
          </>) : null }
          <div />
          <div>
            <input
              type="button"
              value="Login"
              onClick=${() => onLogin(tmpUsername, password).catch(err => setError(err.message))}
            />
          </div>
          <div className="social">
            <${GithubButton} />
            <${GoogleButton} />
            <${AmazonButton} />
          </div>
        </div>
      </${Center}>
    </${Center}>
  </div>`;
};
