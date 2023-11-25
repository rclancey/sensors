import { h } from 'https://unpkg.com/preact@latest?module';
import { useCallback } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';
import _JSXStyle from 'https://unpkg.com/styled-jsx/style';

const html = htm.bind(h);

export const valueLabel = (obj) => {
  if (typeof obj === 'string' || typeof obj === 'number') {
    return {
      value: obj,
      label: obj,
    };
  }
  if (typeof obj === 'boolean') {
    return {
      value: obj,
      label: obj ? 'Yes' : 'No',
    };
  }
  if (Array.isArray(obj)) {
    return {
      value: obj[0],
      label: obj[1],
    };
  }
  let value;
  let label;
  if (Object.hasOwnProperty.call(obj, 'id')) {
    value = obj.id;
  } else if (Object.hasOwnProperty.call(obj, 'value')) {
    value = obj.value;
  } else {
    throw new Error(`can't find a suitible value for option ${obj}`);
  }
  if (Object.hasOwnProperty.call(obj, 'name')) {
    label = obj.name;
  } else if (Object.hasOwnProperty.call(obj, 'label')) {
    label = obj.label;
  } else {
    throw new Error(`can't find a suitible label for option ${obj}`);
  }
  return { value, label };
};

export const valueKey = (obj) => valueLabel(obj).value;

export const Option = ({ option }) => {
  if (option === null || option === undefined) {
    return html`<option value="" disabled>${' '}</option>`;
  }
  const { value, label } = valueLabel(option);
  return html`<option value=${value} disabled=${option.disabled ? true : false}>${`${label}`}</option>`;
};

export const MenuInput = ({
  groups,
  options,
  value,
  empty,
  onChange,
  ...props
}) => {
  const callback = useCallback((evt) => {
    if (empty && evt.target.selectedIndex === 0) {
      return onChange(null);
    }
    let n = evt.target.selectedIndex - (empty ? 1 : 0);
    if (groups) {
      const group = groups.find((grp) => {
        if (grp.options.length < n) {
          n -= grp.options.length;
          return false;
        }
        return true;
      });
      if (group) {
        return onChange(group.options[n]);
      }
    }
    if (options) {
      return onChange(options[n]);
    }
    return onChange(null);
  }, [onChange, empty, options, groups]);
  if (!onChange) {
    return value || '\u00a0';
  }
  return html`<select
      value=${value || ''}
      onChange=${callback}
      ${...props}
    >
      <style jsx>${`
        background: var(--gradient-end);
        color: var(--text);
        border: solid var(--border) 1px;
        border-radius: 4px;
        padding: 3px 5px;
        font-size: var(--font-size-normal);
      `}</style>
      ${ empty ? <${Option} /> : null }
      ${ groups ? groups.map((group, i) => (
        group.label ? html`<optgroup key=${i} label=${group.label}>
          ${ group.options.map((opt, j) => html`<${Option} key=${`${i}.${j}`} option=${opt} />`) }
        </optgroup>` : 
          group.options.map((opt, j) => html`<${Option} key=${`${i}.${j}`} option=${opt} />`)
        )
      ) : null }
      ${ options ? options.map((opt, i) => html`<${Option} key=${i} option=${opt} />`) : null }
    </select>`;
  );
};

export default MenuInput;
