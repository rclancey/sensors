import base64 from 'base-64';

import Emitter from '../../lib/emitter.js';
import { API } from '../../lib/api.js';

const getCookie = (name) => {
  if (typeof document === 'undefined') {
    return null;
  }
  return document.cookie.split(/;\s*/)
    .map(cookie => {
      const parts = cookie.split(/=/);
      const name = unescape(parts.shift());
      const value = unescape(parts.join('='));
      return { name, value };
    })
    .find(cookie => cookie.name === name);
};

export const LOGIN_STATE = {
  LOGGED_OUT:    0,
  TOKEN_EXPIRED: 1,
  NEEDS_2FA:     2,
  LOGGED_IN:     3,
};

export class Token extends Emitter {
  #loginState;
  #claims;
  #watchInterval;
  #api;
  #userinfo;

  constructor(claims) {
    super();
    this.#loginState = LOGIN_STATE.LOGGED_OUT;
    this.on('login', () => this.#loadUserInfo());
    if (claims === undefined) {
      this.updateFromCookie();
    } else {
      this.#setClaims(claims);
    }
    this.#userinfo = null;
  }

  login(username, password) {
    console.debug('login(%o, %o)', username, password);
    const api = new API();
    return api.post('/api/login', { username, password })
      .then((resp) => {
        this.updateFromCookie();
        return resp;
      })
      .catch(() => {
        this.updateFromCookie();
        return null;
      });
  }

  twoFactor(code) {
    const api = new API();
    return api.post('/api/login/twofactor', { two_factor_code: code })
      .then((resp) => {
        this.updateFromCookie();
        return resp;
      })
      .catch((err) => {
        this.updateFromCookie();
        throw err;
      });
  }

  logout() {
    if (typeof document !== 'undefined') {
      document.cookie = 'auth=;expires=Thu, 01 Jan 1970 00:00:01 GMT';
    }
    this.#setClaims({});
  }

  resetPassword(username) {
    const api = new API();
    const body = { username };
    return api.post('/api/password/reset', body);
  }

  changePassword(body) {
    const api = new API();
    return api.post('/api/password', body)
      .then((resp) => {
        this.logout();
        return resp;
      });
  }

  #loadUserInfo() {
    const api = new API();
    api.get('/api/admin/user/__myself__')
      .then((obj) => {
        this.#userinfo = obj;
        this.emit('info', obj);
      })
      .catch(() => this.updateFromCookie());
  }

  get userinfo() {
    return this.#userinfo;
  }

  checkState() {
    if (this.expired()) {
      if (!this.username) {
        if (this.#loginState != LOGIN_STATE.LOGGED_OUT) {
          this.#loginState = LOGIN_STATE.LOGGED_OUT;
          this.emit('logout');
        }
      } else if (this.#loginState != LOGIN_STATE.TOKEN_EXPIRED) {
        this.#loginState = LOGIN_STATE.TOKEN_EXPIRED;
        this.emit('expire');
      }
    } else if (this.needs2fa()) {
      if (this.#loginState != LOGIN_STATE.NEEDS_2FA) {
        this.#loginState = LOGIN_STATE.NEEDS_2FA;
        this.emit('2fa');
      }
    } else if (this.#loginState != LOGIN_STATE.LOGGED_IN) {
      this.#loginState = LOGIN_STATE.LOGGED_IN;
      console.debug('emit login');
      this.emit('login');
    }
    return this.loginState;
  }

  #setClaims(claims) {
    this.#claims = claims;
    this.checkState();
  }

  updateFromCookie() {
    if (typeof document === 'undefined') {
      this.#setClaims({});
      return this;
    }
    const cookie = getCookie('auth');
    if (!cookie || !cookie.value) {
      this.#setClaims({});
      return this;
    }
    try {
      const parts = cookie.value.split('.');
      const claims = JSON.parse(base64.decode(parts[1].replace(/_/g, '/')));
      this.#setClaims(claims);
    } catch (err) {
      this.#setClaims({});
    }
    return this;
  }

  get state() {
    return this.#loginState;
  }

  get id() {
    return this.#claims['x-userid'];
  }

  get username() {
    return this.#claims.preferred_username;
  }

  get email() {
    return this.#claims.email;
  }

  get phoneNumber() {
    return this.#claims.phone_number;
  }

  get familyName() {
    return this.#claims.family_name;
  }

  get givenName() {
    return this.#claims.given_name;
  }

  get expires() {
    if (!this.#claims.exp) {
      return 0;
    }
    return this.#claims.exp * 1000;
  }

  expired() {
    return this.expires < Date.now();
  }

  get authTime() {
    if (!this.#claims.auth_time) {
      return 0;
    }
    return this.#claims.auth_time * 1000;
  }

  get issueTime() {
    if (!this.#claims.iat) {
      return 0;
    }
    return this.#claims.iat * 1000;
  }

  get isser() {
    return this.#claims.iss;
  }

  get notBeforeTime() {
    if (!this.#claims.nbf) {
      return 0;
    }
    return this.#claims.nbf * 1000;
  }

  get ttl() {
    if (!this.#claims['x-ttl']) {
      return 0;
    }
    return this.#claims['x-ttl'];
  }

  get initials() {
    const names = [
      this.#claims.given_name,
      this.#claims.family_name,
    ];
    const inits = names.filter((name) => name !== null && name !== undefined && name !== '')
      .map((name) => name.substr(0, 1).toUpperCase());
    return inits.join('');
  }

  get avatar() {
    return this.#claims.picture;
  }

  needs2fa() {
    if (!this.#claims['x-2fa']) {
      return true;
    }
    return false;
  }

  watch() {
    if (this.#watchInterval) {
      clearInterval(this.#watchInterval);
      this.#watchInterval = null;
    }
    this.#watchInterval = setInterval(() => this.checkState(), 60000);
  }

  dispose() {
    if (this.#watchInterval) {
      clearInterval(this.#watchInterval);
      this.#watchInterval = null;
    }
  }
}

export default Token;
