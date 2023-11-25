import { h } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';
import _JSXStyle from 'https://unpkg.com/styled-jsx/style';

const html = htm.bind(h);

export const FalseIcon = ({ color = 'currentColor' }) => html`
  <svg
    version="1.1"
    xmlns="http://www.w3.org/2000/svg"
    x="0px"
    y="0px"
    viewBox="0 0 512 512"
    style="enable-background:new 0 0 512 512;"
    fill=${color}
  >
    <g>
      <path d="M256,0C114.615,0,0,114.615,0,256s114.615,256,256,256s256-114.615,256-256S397.385,0,256,0z M327.115,365.904 L256,294.789l-71.115,71.115l-38.789-38.789L217.211,256l-71.115-71.115l38.789-38.789L256,217.211l71.115-71.115l38.789,38.789 L294.789,256l71.115,71.115L327.115,365.904z"/>
    </g>
  </svg>
`;

export const TrueIcon = ({ color = 'currentColor' }) => html`
  <svg
    version="1.1"
    xmlns="http://www.w3.org/2000/svg"
    x="0px"
    y="0px"
	  viewBox="0 0 32 32"
    style="enable-background:new 0 0 32 32;"
    fill=${color}
  >
    <g>
      <path d="M16,0C7.164,0,0,7.164,0,16s7.164,16,16,16s16-7.164,16-16S24.836,0,16,0z M13.52,23.383 L6.158,16.02l2.828-2.828l4.533,4.535l9.617-9.617l2.828,2.828L13.52,23.383z"/>
    </g>
  </svg>
`;

export const Check = ({ valid }) => {
  if (valid) {
    return html`<TrueIcon color="green" />`;
  }
  return html`<FalseIcon color="red" />`;
};

export default Check;
