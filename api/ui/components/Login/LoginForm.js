import { h } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';
import _JSXStyle from 'https://unpkg.com/styled-jsx/style';

import { Center } from '../Center.js';
import Grid from '../Grid.js';

const html = htm.bind(h);

export const LoginForm = ({ children }) => html`<div id="app">
  <style jsx>${`
    :global(.grid.login) {
      grid-template-columns: min-content min-content !important;
    }
  `}</style>
  <${Center} orientation="horizontal" style=${{ width: '100vw', height: '100vh' }}>
    <${Center} orientation="vertical">
      <${Grid} className="login">
        ${children}
      </${Grid}>
    </${Center}>
  </${Center}>
</div>`;

export default LoginForm;
