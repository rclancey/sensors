import { h, Fragment } from 'https://unpkg.com/preact@latest?module';
import { useState, useContext, useCallback, useEffect } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';
import _JSXStyle from 'https://unpkg.com/styled-jsx/style';
import zxcvbn from 'https://unpkg.com/zxcvbn';

import LoginContext from '../context/LoginContext.js';
import Grid from './Grid.js';
import Button from './Input/Button.js';
import TextInput from './Input/TextInput.js';
import Check from './Check.js';

const html = htm.bind(h);

const useEnterKey = (onEnter) => useCallback((evt) => {
  if (evt.key === 'Enter' && onEnter) {
    onEnter(evt);
  }
}, [onEnter]);

export const Username = ({ value = '', hidden = false, onChange, onEnter }) => {
  const onKeyPress = useEnterKey(onEnter);
  return html`<${TextInput}
      className="username"
      type="text"
      name="username"
      autocomplete="username"
      size=${15}
      value=${value}
      hidden=${hidden}
      onInput=${onChange}
      onKeyPress=${onKeyPress}
    />`;
};

export const Password = ({ value = '', name = 'password', onChange, onEnter }) => {
  const onKeyPress = useEnterKey(onEnter);
  return html`<${TextInput}
      className="password"
      type="password"
      name=${name}
      size=${15}
      autocomplete=${name === 'password' ? 'current-password' : 'new-password'}
      value=${value}
      onInput=${onChange}
      onKeyPress=${onKeyPress}
    />`;
};

export const TwoFactorCode = ({ value = '', onChange, onEnter }) => {
  const onInput = useCallback((evt) => onChange(evt.target.value.replace(/[^0-9]/g, '')), [onChange]);
  const onKeyPress = useEnterKey(onEnter);
  return html`<input
      className="twoFactorCode"
      type="text"
      name="2facode"
      inputmode="numeric"
      autofill="new-password"
      size=${6}
      maxlength=${6}
      value=${value}
      onInput=${onInput}
      onKeyPress=${onKeyPress}
    >
      <style jsx>${`
        .twoFactorCode {
          background-color: var(--gradient-end);
          color: var(--text);
          border: solid var(--border) 1px;
          border-radius: 4px;
          padding: 5px;
          padding-left: 15px;
          font-size: var(--font-size-huge);
          letter-spacing: 10px;
          width: calc(6ch + 65px);
          text-align: center;
        }
      `}</style>
    </input>`;
};

export const ResetCode = ({ value = '', onChange, onEnter }) => {
  //const onInput = useCallback((evt) => onChange(evt.target.value), [onChange]);
  const onKeyPress = useEnterKey(onEnter);
  return html`<${TextInput}
      className="resetCode"
      type="text"
      name="reset_code"
      size=${15}
      value=${value}
      onInput=${onChange}
      onKeyPress=${onKeyPress}
    />`;
}

export const PasswordStrength = ({ score }) => html`<div className="passwordStrength">
    <style jsx>${`
      .passwordStrength {
        width: 100%;
        display: flex;
        flex-direction: row;
        margin-top: 5px;
        opacity: 0.7;
      }
      .passwordStrength>div {
        flex: 1;
        height: 5px;
        border-radius: 5px;
        margin-left: 2px;
        background-color: var(--blur-background);
      }
      .passwordStrength>div:first-child {
        margin-left: 0px;
      }
      .passwordStrength>div.red.on {
        background-color: red;
      }
      .passwordStrength>div.orange.on {
        background-color: orange;
      }
      .passwordStrength>div.yellow.on {
        background-color: yellow;
      }
      .passwordStrength>div.green.on {
        background-color: green;
      }
    `}</style>
    <div className=${`red ${score >= 0 ? 'on' : 'off'}`} />
    <div className=${`orange ${score >= 1 ? 'on' : 'off'}`} />
    <div className=${`yellow ${score >= 2 ? 'on' : 'off'}`} />
    <div className=${`green ${score >= 3 ? 'on' : 'off'}`} />
    <div className=${`green ${score >= 4 ? 'on' : 'off'}`} />
  </div>`;

export const NewPassword = ({ value = '', minScore = 3, onChange, onEnter }) => {
  const [score, setScore] = useState(value ? zxcvbn(value).score : -1);
  const myOnChange = useCallback((val) => {
    const strength = val ? zxcvbn(val) : { score: -1 };
    setScore(strength.score);
    onChange({ password: val, score: strength.score, valid: strength.score >= minScore });
  }, [minScore, onChange]);
  return html`<div className="newPassword">
    <style jsx>${`
      .newPassword {
        display: flex;
        flex-direction: row;
        align-items: baseline;
      }
      .newPassword :global(svg) {
        width: 12px;
        height: 12px;
        margin-left: 5px;
      }
    `}</style>
    <div className="wrap">
      <${Password} value=${value} name="new_password" onChange=${myOnChange} onEnter=${onEnter} />
      <br />
      <${PasswordStrength} score=${score} />
    </div>
    <${Check} valid=${score >= minScore} />
  </div>`;
};

export const ConfirmPassword = ({ password = '', onChange, onEnter }) => {
  const [confirmPassword, setConfirmPassword] = useState('');
  useEffect(() => {
    onChange(confirmPassword === password);
  }, [password, confirmPassword, onChange]);
  return html`<div className="confirmPassword">
    <style jsx>${`
      .confirmPassword :global(svg) {
        width: 12px;
        height: 12px;
        margin-left: 5px;
      }
    `}</style>
    <${Password} value=${confirmPassword} name="confirm_password" onChange=${setConfirmPassword} onEnter=${onEnter} />
    <${Check} valid=${password && confirmPassword === password} />
  </div>`;
};

export const ChangePassword = ({ onComplete, onCancel }) => {
  const { username, token } = useContext(LoginContext);
  const [password, setPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [valid, setValid] = useState(false);
  const [match, setMatch] = useState(true);
  const [error, setError] = useState(null);
  const onInput = useCallback((obj) => {
    setNewPassword(obj.password);
    setValid(obj.valid);
  }, []);
  const onChange = useCallback(() => {
    if (password === '') {
      setError('Missing current password');
      return;
    }
    if (newPassword === '') {
      setError('Missing new password');
      return;
    }
    if (!valid) {
      setError('New password is too weak');
      return;
    }
    if (!match) {
      setError("New passwords don't match");
      return;
    }
    setError(null);
    token.changePassword({ username, password, new_password: newPassword })
      .then((resp) => {
        onComplete(resp);
      })
      .catch((err) => setError(`${err}`));
  }, [username, password, newPassword, token]);
  const preventSubmit = useCallback((evt) => {
    evt.preventDefault();
    evt.stopPropagation();
    return false;
  }, []);
  const disabled = (password === '' || newPassword === '' || !valid || !match);
  return html`<form name="changePassword" onSubmit=${preventSubmit}>
    <style jsx>${`
      .error {
        color: var(--highlight);
        font-weight: bold;
      }
    `}</style>
    <${Username} value=${username} hidden />
    <${Grid}>
      <div>Current Password:</div>
      <div><Password value=${password} onChange=${setPassword} onEnter=${onChange} /></div>
      <div>New Password:</div>
      <${NewPassword} value=${newPassword} onChange=${onInput} onEnter=${onChange} />
      <div>Confirm Password:</div>
      <${ConfirmPassword} password=${newPassword} onChange=${setMatch} onEnter=${onChange} />
      ${ error !== null ? html`<div colspan=${2} className="error">${error}</div>` : null }
      <div />
      <div>
        <${Button} onClick=${onChange} disabled=${disabled}>Update Password</Button>
      </div>
    </${Grid}>
  </form>`;
};

export const ResetPassword = ({ onComplete, onCancel }) => {
  const { token } = useContext(LoginContext);
  const [username, setUsername] = useState('');
  const [code, setCode] = useState('');
  const [codeFromUrl, setCodeFromUrl] = useState(false);
  const [newPassword, setNewPassword] = useState('');
  const [valid, setValid] = useState(false);
  const [match, setMatch] = useState(true);
  const [error, setError] = useState(null);
  const onInput = useCallback((obj) => {
    setNewPassword(obj.password);
    setValid(obj.valid);
  }, []);
  const onChange = useCallback(() => {
    if (code === '') {
      setError('Missing reset code');
      return;
    }
    if (newPassword === '') {
      setError('Missing new password');
      return;
    }
    if (!valid) {
      setError('New password is too weak');
      return;
    }
    if (!match) {
      setError("New passwords don't match");
      return;
    }
    setError(null);
    token.changePassword({ username, reset_code: code, new_password: newPassword })
      .then((resp) => {
        onComplete(resp);
      })
      .catch((err) => setError(`${err}`));
  }, [username, code, newPassword, token]);
  useEffect(() => {
    const u = new URL(document.location);
    setUsername(u.searchParams.get('username'));
    if (u.searchParams.has('code')) {
      setCode(u.searchParams.get('code'));
      setCodeFromUrl(true);
    }
  }, []);
  const preventSubmit = useCallback((evt) => {
    evt.preventDefault();
    evt.stopPropagation();
    return false;
  }, []);
  const disabled = (code === '' || newPassword === '' || !valid || !match);
  return html`<form name="resetPassword" onSubmit=${preventSubmit}>
    <style jsx>${`
      .error {
        color: var(--highlight);
        font-weight: bold;
      }
    `}</style>
    <${Grid}>
      <div>Username:</div>
      <div>
        ${username}
        <input type="hidden" name="username" value=${username} />
      </div>
      ${ codeFromUrl ? null : html`<Fragment>
        <div colspan=${2}>
          <p>Check your email for a code to enter below:</p>
        </div>
        <div>Reset Code:</div>
        <${ResetCode} value=${code} onChange=${setCode} onEnter=${onChange} />
      </Fragment>` }
      <div>New Password:</div>
      <${NewPassword} value=${newPassword} onChange=${onInput} onEnter=${onChange} />
      <div>Confirm Password:</div>
      <${ConfirmPassword} password=${newPassword} onChange=${setMatch} onEnter=${onChange} />
      ${ error !== null ? html`<div colspan=${2} className="error">${error}</div>` : null }
      <div />
      <div>
        <${Button} onClick=${onChange} disabled=${disabled}>Reset Password</Button>
      </div>
    </${Grid}>
  </form>`;
};
