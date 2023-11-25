import { h } from 'https://unpkg.com/preact@latest?module';
import { useMemo } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import TextInput from './TextInput.js';

const html = htm.bind(h);

const fmt = Intl.DateTimeFormat('fr-CA', { year: 'numeric', month: 'numeric', day: 'numeric' });

export const DateInput = ({ value, onInput, ...props }) => {
  const myOnInput = useMemo(() => {
    if (!onInput) {
      return null;
    }
    return (val) => onInput(new Date(`${val}T00:00:00`).getTime());
  }, [onInput]);
  const date = useMemo(() => {
    if (value === null || value === undefined) {
      return '';
    }
    return fmt.format(new Date(value));
  }, [value]);
  return html`<${TextInput} type="date" value=${date} onInput=${myOnInput} ${...props} />`;
};

export default DateInput;
