import { createContext } from 'https://unpkg.com/preact@latest?module';

export const LoginContext = createContext({
  token: null,
  username: null,
  loginState: 0,
  onLoginRequired: () => console.debug('login required'),
  onLogout: () => console.debug('logout'),
});

export default LoginContext;
