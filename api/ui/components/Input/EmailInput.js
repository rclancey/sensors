import { h } from 'https://unpkg.com/preact@latest?module';
import { useState, useEffect, useMemo } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import TextInput from './TextInput.js';

const html = htm.bind(h);

const validate = (value) => {
  if (!value) {
    return true;
  }
  const parts = value.split(/@/);
  if (parts.length != 2) {
    return false;
  }
  if (parts[0].match(/[^0-9A-Za-z\+\._\-]/)) {
    return false;
  }
  const domain = parts[1].split('.');
  if (domain.length < 2) {
    return false;
  }
  if (parts[1].match(/[^0-9A-Za-z\-\.]/)) {
    return false;
  }
  if (parts[1].match(/--/)) {
    return false;
  }
  if (domain.some((name) => name === '')) {
    return false;
  }
  if (domain.some((name) => (name.startsWith('-') || name.endsWith('-')))) {
    return false;
  }
  return true;
};

export const EmailInput = ({ value, placeholder = 'username@domain.com', ...props }) => {
  const valid = useMemo(() => validate(value), [value]);
  return html`<${TextInput}
    type="email"
    value=${value}
    valid=${valid}
    placeholder=${placeholder}
    ${...props}
  />`;
};

export default EmailInput;
