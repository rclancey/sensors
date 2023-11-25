import { h } from 'https://unpkg.com/preact@latest?module';
import { useMemo } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import TextInput from './TextInput.js';

const html = htm.bind(h);

export const NumberInput = ({ onInput, ...props }) => {
  const myOnInput = useMemo(() => {
    if (!onInput) {
      return null;
    }
    return (val) => {
      const f = parseFloat(val);
      if (Number.isNaN(f)) {
        onInput(null);
      } else {
        onInput(f);
      }
    };
  }, [onInput]);
  return html`<${TextInput} type="number" onInput=${myOnInput} ${...props} />`;
};

export default NumberInput;
