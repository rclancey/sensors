import { h } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';
import _JSXStyle from 'https://unpkg.com/styled-jsx/style';

const html = htm.bind(h);

export const Center = ({ orientation = 'horizontal', style, children }) => {
  const vert = orientation.substr(0, 1).toLowerCase() === 'v';
  return html`<div className="center" style=${style}>
    <div className="padding" />
    ${children}
    <div className="padding" />
    <style jsx>${`
      .center {
        display: flex;
        flex-direction: ${vert ? 'column' : 'row'};
        box-sizing: border-box;
      }
      .padding {
        flex: 10;
      }
    `}</style>
  </div>`;
};
