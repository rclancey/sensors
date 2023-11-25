import { h } from 'https://unpkg.com/preact@latest?module';
import { useCallback } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import TextInput from './TextInput.js';

const html = htm.bind(h);

const validate = (value) => {
  if (!value) {
    return true;
  }
  try {
    const u = new URL(value);
  } catch (err) {
    return false;
  }
  return true;
};

export const URLInput = ({ value, placeholder = 'https://www.example.com/', onInput, ...props }) => {
  const valid = useMemo(() => validate(value), [value]);
  if (!onInput) {
    return (<a href={value}>{value}</a>);
  }
  return html`<${TextInput}
    type="url"
    value=${value}
    valid=${valid}
    placeholder=${placeholder}
    ${...props}
  />`;
};

export default URLInput;
