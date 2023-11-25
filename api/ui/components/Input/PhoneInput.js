import { h } from 'https://unpkg.com/preact@latest?module';
import { useState, useCallback, useEffect, useMemo } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import TextInput from './TextInput.js';

const html = htm.bind(h);

const parsePhone = (phone) => {
  if (phone === null || phone === undefined) {
    return { raw: '' };
  }
  let rem = phone;
  const country = '+1';
  rem = rem.replace(/^\s*\+?1\s*/, '');
  let m = rem.match(/^\D*(\d{3})/);
  const area = m ? m[1] : '';
  rem = rem.replace(/^\D*\d{3}/, '');
  m = rem.match(/^\D*(\d{3})/);
  const prefix = m ? m[1] : '';
  rem = rem.replace(/^\D*\d{3}/, '');
  m = rem.match(/^\D*(\d{4})/);
  const num = m ? m[1] : '';
  return {
    raw: phone,
    country,
    area,
    prefix,
    num,
    valid: (country && area && prefix && num ? true : false),
  };
};

const formatPhone = (phone) => {
  if (!phone) {
    return null;
  }
  const {
    raw,
    country,
    area,
    prefix,
    num,
    valid,
  } = parsePhone(phone);
  if (!valid) {
    return raw;
  }
  return `${country} (${area}) ${prefix}-${num}`;
};

const simplifyPhone = (phone) => {
  const {
    raw,
    country,
    area,
    prefix,
    num,
    valid,
  } = parsePhone(phone);
  if (!valid) {
    console.debug('simplifyPhone(%o) => %o', phone, raw);
    return raw;
  }
  console.debug('simplifyPhone(%o) => %o', phone, `${country}${area}${prefix}${num}`);
  return `${country}${area}${prefix}${num}`;
};

const validate = (phone) => {
  if (!phone) {
    return true;
  }
  const { valid = false } = parsePhone(phone);
  return valid;
};

export const PhoneInput = ({ value, placeholder = '+1 (800) 777-3456', onInput, ...props }) => {
  const myOnInput = useMemo(() => {
    if (!onInput) {
      return null;
    }
    return (evt) => {
      const target = { evt };
      if (target.value === '') {
        onInput(null);
      } else {
        target.value = simplifyPhone(target.value);
        onInput(evt);
      }
    };
  }, [onInput]);
  const valid = useMemo(() => validate(value), [value]);
  if (!onInput) {
    return value ? formatPhone(value) : '\u00a0';
  }
  return html`<${TextInput}
    type="tel"
    value=${formatPhone(value)}
    valid=${valid}
    placeholder=${placeholder}
    onInput=${myOnInput}
    ${...props}
  />`;
};

export default PhoneInput;
