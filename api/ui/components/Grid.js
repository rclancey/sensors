import { h } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';
import _JSXStyle from 'https://unpkg.com/styled-jsx/style';

const html = htm.bind(h);

export const Grid = ({ cols = 2, className = '', children, ...props }) => (
  html`<div className=${`grid ${className}`} ${...props}>
    <style jsx>${`
      .grid {
        display: grid;
        grid-template-columns: min-content ${' auto'.repeat(cols - 1)};
        column-gap: 4px;
        row-gap: 10px;
        align-items: baseline;
      }
      .grid>:global(div) {
        white-space: nowrap;
      }
      .grid>:global(div[colspan]) {
        grid-column: span 2;
      }
    `}</style>
    {children}
  </div>`
);
