import { h } from 'https://unpkg.com/preact@latest?module';
import { useMemo } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import TextInput from './TextInput.js';

const html = htm.bind(h);

export const IntegerInput = ({ step = 1, onInput, ...props }) => {
  const myOnInput = useMemo(() => {
    if (!onInput) {
      return null;
    }
    return (val) => {
      const n = parseInt(val, 10);
      if (Number.isNaN(n)) {
        onInput(null);
      } else {
        onInput(n);
      }
    };
  }, [onInput]);
  return html`<${TextInput} type="number" step=${step} onInput=${myOnInput} ${...props} />`;
};

export default IntegerInput;
