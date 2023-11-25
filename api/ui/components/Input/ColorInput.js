import { h } from 'https://unpkg.com/preact@latest?module';
import { useCallback } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';
import _JSXStyle from 'https://unpkg.com/styled-jsx/style';

const html = htm.bind(h);

export const ColorInput = ({ value, placeholder = '#000000', onInput }) => {
  const myOnInput = useCallback((evt) => onInput(evt.target.value), [onInput]);
  if (!onInput) {
    return html`<div className="color">
      <style jsx>${`
        .color {
          display: inline-block;
          width: 2em;
          height: 1em;
          border: solid var(--border) 1px;
          border-radius: 4px;
          background-color: ${value || '#000'};
        }
      `}</style>
    </div>`;
  }
  return html`<${TextInput} value=${value} placeholder=${placeholder} onInput=${myOnInput} ${...props} />`;
};

export default ColorInput;
