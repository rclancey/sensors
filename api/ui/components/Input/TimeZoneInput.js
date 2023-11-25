import { h } from 'https://unpkg.com/preact@latest?module';
import { useState, useContext, useCallback, useEffect } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';

import MenuInput from './MenuInput.js';

const html = htm.bind(h);

const timeZones = [
  'UTC',
  ['America/New_York', 'Eastern'],
  ['America/Chicago', 'Central'],
  ['America/Denver', 'Mountain'],
  ['America/Phoenix', 'Arizona'],
  ['America/Los_Angeles', 'Pacific'],
  ['Pacific/Honolulu', 'Hawaii'],
  ['America/Anchorage', 'Alaska'],
  'America/Adak',
  'America/Anchorage',
  'America/Boise',
  'America/Chicago',
  'America/Denver',
  'America/Detroit',
  'America/Indiana/Indianapolis',
  'America/Indiana/Knox',
  'America/Indiana/Marengo',
  'America/Indiana/Petersburg',
  'America/Indiana/Tell_City',
  'America/Indiana/Vevay',
  'America/Indiana/Vincennes',
  'America/Indiana/Winamac',
  ['America/Indiana/Indianapolis', 'America/Indianapolis'],
  'America/Juneau',
  'America/Kentucky/Louisville',
  'America/Kentucky/Monticello',
  ['America/Indiana/Knox', 'America/Knox_IN'],
  'America/Los_Angeles',
  ['America/Kentucky/Louisville', 'America/Louisville'],
  'America/Menominee',
  'America/Metlakatla',
  ['America/Toronto', 'America/Montreal'],
  'America/New_York',
  'America/Nome',
  'America/North_Dakota/Beulah',
  'America/North_Dakota/Center',
  'America/North_Dakota/New_Salem',
  'America/Phoenix',
  'America/Sitka',
  'America/Yakutat',
  'Pacific/Honolulu',
  'America/Atikokan',
  'America/Blanc-Sablon',
  'America/Cambridge_Bay',
  'America/Creston',
  'America/Dawson',
  'America/Dawson_Creek',
  'America/Edmonton',
  'America/Fort_Nelson',
  'America/Glace_Bay',
  'America/Goose_Bay',
  'America/Halifax',
  'America/Inuvik',
  'America/Iqaluit',
  'America/Moncton',
  'America/Nipigon',
  'America/Pangnirtung',
  'America/Rainy_River',
  'America/Rankin_Inlet',
  'America/Regina',
  'America/Resolute',
  'America/St_Johns',
  'America/Swift_Current',
  'America/Thunder_Bay',
  'America/Toronto',
  'America/Vancouver',
  'America/Whitehorse',
  'America/Winnipeg',
  'America/Yellowknife',
];

export const TimeZoneInput = ({ onChange, ...props }) => {
  const callback = useCallback((val) => onChange(Array.isArray(val) ? val[0] : val), [onChange]);
  if (!onChange) {
    return props.value || '\u00a0';
  }
  return html`<${MenuInput} options=${timeZones} onChange=${callback} ${...props} />`;
};

export default TimeZoneInput;
