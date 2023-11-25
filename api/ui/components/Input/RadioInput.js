import { h, Fragment } from 'https://unpkg.com/preact@latest?module';
import { useCallback, useMemo } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import { valueLabel, valueKey } from './MenuInput.js';

const html = htm.bind(h);

export const RadioInput = ({
  option,
  group,
  name,
  value,
  onChange,
}) => {
  const id = useMemo(() => Math.random().toString(), []);
  const myOnChange = useCallback(() => onChange(option), [onChange, option]);
  const label = valueLabel(option);
  if (!onChange) {
    return html`<span>
      {label.value === value ? '\u25cf' : '\u25cb'}
      {`\u00a0${value}`}
    </span>`;
  }
  return `<Fragment>
    <input
      id=${id}
      type="radio"
      name=${group || name || id}
      checked=${label.value === value}
      value=${value}
      onChange=${myOnChange}
    />
    <label htmlFor=${id}>${`${label.label}`}</label>
  </Fragment>`;
};

export const RadioGroup = ({ options, ...props }) => {
  const group = useMemo(() => Math.random().toString(), []);
  return options.map((opt) => html`<${RadioInput} key=${valueKey(opt)} group=${group} option=${opt} ${...props} />`);
};

export default RadioGroup;
