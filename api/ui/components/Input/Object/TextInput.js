import { h } from 'https://unpkg.com/preact@latest?module';
import { useCallback } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import TextInput from '../TextInput.js';

const html = htm.bind(h);

export const ObjectTextInput = ({ obj, field, onInput, ...props }) => {
  const myOnInput = useCallback((val) => onInput({ [field]: val }), [field, onInput]);
  return html`<${TextInput}
    value=${obj ? obj[field] : null}
    onInput=${myOnInput}
    ${...props}
  />`;
};

export default ObjectTextInput;
