import { h } from 'https://unpkg.com/preact@latest?module';
import { useMemo } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import XTextInput from '../TextInput.js';
import XEmailInput from '../EmailInput.js';
import XPhoneInput from '../PhoneInput.js';
import XNumberInput from '../NumberInput.js';
import XIntegerInput from '../IntegerInput.js';
import XDateInput from '../DateInput.js';

const html = htm.bind(h);

const ObjectInput = (Comp) => (({ obj, field, onInput, onChange, ...props }) => {
  const myOnInput = useMemo(() => {
    if (!onInput) {
      return null;
    }
    return (val) => onInput({ [field]: val });
  }, [field, onInput]);
  const myOnChange = useMemo(() => {
    if (!onChange) {
      return null;
    }
    return (val) => onChange({ [field]: val });
  }, [field, onChange]);
  return html`<${Comp}
    value=${obj ? obj[field] : null}
    onInput=${myOnInput}
    onChange=${myOnChange}
    ${...props}
  />`;
});

export const TextInput = ObjectInput(XTextInput);
export const DateInput = ObjectInput(XDateInput);
export const IntegerInput = ObjectInput(XIntegerInput);
export const NumberInput = ObjectInput(XNumberInput);
export const BoolInput = ObjectInput(XBoolInput);
export const PhoneInput = ObjectInput(XPhoneInput);
export const EmailInput = ObjectInput(XEmailInput);
export const RadioInput = ObjectInput(XRadioInput);
export const MenuInput = ObjectInput(XMenuInput);
