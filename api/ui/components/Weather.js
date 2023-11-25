import { h } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';

import { AsOf } from './AsOf.js';

const html = htm.bind(h);

const temp = (val, unitSystem) => {
  switch (unitSystem) {
  case 'metric':
    return `${Intl.NumberFormat('en-US', { minimumFractionDigits: 1, maximumFractionDigits: 2 }).format(val - 273.15)} °C`;
  case 'imperial':
    return `${Math.round((val - 273.15) * 1.8) + 32} °F`;
  case 'scientific':
    let notation = 'standard';
    if (val > 0) {
      if (val < 0.1 || val >= 10000) {
        notation = 'scientific';
      }
    }
    return `${Intl.NumberFormat('en-US', { minimumSignificantDigits: 2, maximumSignificantDigits: 4, notation }).format(val)} K`;
  }
  return `${val}`;
};

const pres = (val, unitSystem) => {
  switch (unitSystem) {
  case 'metric':
    return `${Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(val * 0.7501)} mmHg`;
  case 'imperial':
    return `${Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(val * 0.7501 / 25.4)} in`;
  case 'scientific':
    return `${val} kPa`;
  }
  return `${val}`;
};

export const Weather = ({ weather }) => {
  if (!weather) {
    return null;
  }
  const current = weather.current;
  const forecast = weather.daily && weather.daily[0];
  if (!current || !forecast) {
    return null;
  }
  const unitSystem = 'imperial';
  return html`<div className="weather">
    <div className="header">Weather</div>
    <div className="grid">
      <div className="key">Temperature:</div>
      <div className="val">${temp(current.temp, unitSystem)}</div>
      <div className="key">Low / High:</div>
      <div className="val">${temp(forecast.temp.min, unitSystem)} / ${temp(forecast.temp.max, unitSystem)}</div>
      <div className="key">Humidity:</div>
      <div className="val">${current.humidity} %</div>
      <div className="key">Pressure:</div>
      <div className="val">${pres(current.pressure, unitSystem)}</div>
      <div className="key">Clouds:</div>
      <div className="val">${current.clouds} %</div>
    </div>
    <${AsOf} now=${weather.current.dt * 1000} />
  </div>`;
};

export default Weather;
