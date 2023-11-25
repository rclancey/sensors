import { h } from 'https://unpkg.com/preact@latest?module';
import { useCallback, useEffect, useMemo, useState } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import LoginContext from '../../context/LoginContext.js';
import Token, { LOGIN_STATE } from './token.js';
import LoginForm from './LoginForm.js';
import UsernamePasswordForm from './UsernamePasswordForm.js';
import TwoFactor from './TwoFactor.js';

const html = htm.bind(h);

const getToken = () => {
  const t = Token();
  if (t.expired()) {
    return null;
  }
  return t;
};

export const WithLogin = ({ children }) => {
  const token = useMemo(() => new Token(), []);
  const [loginState, setLoginState] = useState(token.state);
  const [username, setUsername] = useState(token.username);
  const [userinfo, setUserinfo] = useState(null);
  useEffect(() => {
    const h = () => {
      setLoginState(token.state);
      setUsername(token.username);
      setUserinfo(token.userinfo);
    };
    token.on('login', h);
    token.on('logout', h);
    token.on('expire', h);
    token.on('2fa', h);
    token.on('info', h);
    return () => {
      token.dispose();
    };
  }, [token]);
  const onLogout = useCallback(() => token.logout(), [token]);
  /*
  const onLoginRequired = useCallback(() => token.updateFromCookie(), [token]);
  */
  const onLoginRequired = useMemo(() => {
    console.debug('onLoginRequired changing for updated token %o', token);
    return () => token.updateFromCookie();
  }, [token]);
  const ctx = useMemo(() => ({
    token,
    username,
    userinfo,
    loginState,
    onLoginRequired,
    onLogout,
  }), [token, username, userinfo, loginState, onLoginRequired, onLogout]);
  switch (loginState) {
    case LOGIN_STATE.LOGGED_OUT:
    case LOGIN_STATE.EXPIRED:
      //console.debug('WithLogin rendering login form');
      return html`<${LoginForm}>
        <${UsernamePasswordForm} username=${username} token=${token} />
      </${LoginForm}>`;
    case LOGIN_STATE.NEEDS_2FA:
      //console.debug('WithLogin rendering 2fa form');
      return html`<${LoginForm}>
        <${TwoFactor} token=${token} />
      </${LoginForm}>`;
    case LOGIN_STATE.LOGGED_IN:
      //console.debug('WithLogin rendering children');
      return html`<${LoginContext.Provider} value=${ctx}>
        ${children}
      </${LoginContext.Provider}>`;
    default:
      //console.debug('WithLogin rendering default (%o)', loginState);
      return html`<${LoginForm} token=${token} />`;
  }
};

export default WithLogin;
