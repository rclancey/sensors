import { h } from 'https://unpkg.com/preact@latest?module';
import { useCallback, useMemo } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';
import _JSXStyle from 'https://unpkg.com/styled-jsx/style';

const html = htm.bind(h);

const Star = ({ value, children, onChange }) => {
  const onClick = useCallback((evt) => {
    console.debug({ layerX: evt.layerX, offsetX: evt.offsetLeft, offset: evt.target.offsetLeft, width: evt.target.offsetWidth, evt });
    if (evt.layerX >= evt.target.offsetLeft + (evt.target.offsetWidth / 2)) {
      console.debug('onChange(%o)', value);
      onChange(value);
    } else {
      console.debug('onChange(%o)', value - 1);
      onChange(value - 1);
    }
  }, [value, onChange]);
  return html`<span data=${value} className="star" onClick=${onClick}>${children}</span>`;
};

export const StarInput = ({
  value,
  min = 0,
  max = 5,
  filled = '\u2605',
  empty = '\u2606',
  onInput,
}) => {
  const stars = useMemo(() => new Array(max).fill(0).map((x, i) => i + 1), [max]);
  const xval = value < min ? min : value;
  console.debug('stars = %o', stars);
  return html`<div className="stars">
    <style jsx>${`
      .stars {
        color: var(--highlight);
        display: inline-block;
        font-size: 16px;
      }
      .stars :global(.star) {
        cursor: pointer;
      }
    `}</style>
    ${stars.map((star) => html`<${Star} key=${star} value=${star} onChange=${onInput}>
      ${star <= xval ? filled : empty}
    </${Star}>`)}
  </div>`;
};

export default StarInput;
