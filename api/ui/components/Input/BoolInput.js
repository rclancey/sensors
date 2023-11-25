import { h, Fragment } from 'https://unpkg.com/preact@latest?module';
import { useMemo, useCallback } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

export const BoolInput = ({
  type,
  label,
  checked,
  value,
  onChange,
  ...props
}) => {
  const id = useMemo(() => Math.random().toString(), []);
  const myOnChange = useCallback((evt) => onChange(evt.target.checked), [onChange]);
  if (!onChange) {
    return !!value ? '\u2713' : '\u00a0';
  }
  return html`<Fragment>
    <input
      id=${id}
      type="checkbox"
      value="true"
      checked=${!!value}
      onChange=${myOnChange}
      ${...props}
    />
    ${ label ? html`<label htmlFor=${id}>${label}</label>` : null }
  </Fragment>`;
};

export default BoolInput;
