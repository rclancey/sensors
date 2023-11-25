import { h } from 'https://unpkg.com/preact@latest?module';
import { useState, useMemo, useCallback } from 'https://unpkg.com/preact@latest/hooks/dist/hooks.module.js?module';
import htm from 'https://unpkg.com/htm?module';
import _JSXStyle from 'https://unpkg.com/styled-jsx/style';

import DateInput from './DateInput.js';
import TimeInput from './TimeInput.js';
import TimeZoneInput from './TimeZoneInput.js';

const html = htm.bind(h);

const dateParts = (t, timeZone = 'UTC') => {
  const fmt = {
    hour12: false,
    year: 'numeric',
    month: 'numeric',
    day: 'numeric',
    hour: 'numeric',
    minute: 'numeric',
  };
  const [
    month,
    day,
    year,
    hour,
    minute,
  ] = Intl.DateTimeFormat('en-US', { ...fmt, timeZone })
    .format(new Date(t))
    .split(/\D+/)
    .map((v) => parseInt(v, 10));
  const dt = (year * 10000) + (month * 100) + day;
  const tm = (hour * 60) + minute;
  return [dt, tm];
};

const tzOffset = (t, tz) => {
  const [utcdt, utctm] = dateParts(t, 'UTC');
  let [tzdt, tztm] = dateParts(t, tz);
  if (tzdt > utcdt) {
    tztm += (24 * 60);
  }
  return tztm - utctm;
};

const withTimeZone = ({ year, month, day, hour = 0, minute = 0, second = 0, millisecond = 0, timeZone = 'local' }) => {
  const args = [year, month, day, hour, minute, second, millisecond];
  if (timeZone === 'local') {
    return new Date(...args);
  }
  const utc = new Date(Date.UTC(...args));
  if (timeZone === 'UTC') {
    return utc;
  }
  const offset = tzOffset(utc.getTime(), timeZone);
  args[3] -= Math.trunc(offset / 60);
  args[4] -= (offset % 60);
  const dt = new Date(Date.UTC(...args));
  const fmt = { hour12: false, hour: 'numeric', timeZone };
  let hr = parseInt(Intl.DateTimeFormat('en-US', fmt).format(dt), 10);
  if (hr === hour) {
    return dt;
  }
  if (((hr + 23) % 24) === hour) {
    dt.setUTCHours(dt.getUTCHours() - 1);
  } else if (((hr + 1) % 24) === hour) {
    dt.setUTCHours(dt.getUTCHours() + 1);
  }
  return dt;
};

const parse = (date, time, timeZone) = {
  const now = new Date();
  let xdate = date;
  if (!xdate) {
    xdate = Intl.DateTimeFormat('fr-CA', { year: 'numeric', month: '2-digit', day: '2-digit' }).format(now);
  }
  let [year, month, day] = xdate.split('-').map((v) => parseInt(v, 10));
  month -= 1;
  let [hour, min, sec] = time ? time.split(':').map((v) => parseInt(v, 10)) : [0, 0, 0];
  if (!sec) {
    sec = 0;
  }
  return withTimeZone({ year, month, day, hour, min, sec, timeZone });
};

const dateInfo = (dt, timeZone) => {
  const d = Intl.DateTimeFormat('fr-CA', { year: 'numeric', month: '2-digit', day: '2-digit', timeZone }).format(dt);
  const t = Intl.DateTimeFormat('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit', fractionalSecondDigits: 3, timeZone }).format(dt);
  return {
    epoch: dt.getTime(),
    datetime: dt,
    timeZone,
    iso: dt.toISOString(),
    date: d,
    time: t,
    string: `${d}T${t}`,
  };
};

export const DateTimeInput = ({ value, timeZone = 'UTC', onInput, ...props }) => {
  const [date, time] = useMemo(() => {
    if (value === null || value === undefined) {
      return [null, null];
    }
    try {
      const dt = new Date(value);
      const dateFmt = Intl.DateTimeFormat('fr-CA', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        timeZone,
      });
      const timeFmt = Intl.DateTimeFormat('en-US', {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        timeZone,
      });
      return [dateFmt.format(dt), timeFmt.format(dt)];
    } catch (err) {
      return [null, null];
    }
  }, [value, timeZone]);
  const onDateInput = useCallback((val) => {
    if (!val) {
      onInput(null);
    } else {
      const dt = parse(val, time, timeZone);
      onInput(dateInfo(dt, timeZone));
    }
  }, [time, timeZone, onInput]);
  const onTimeInput = useCallback((val) => {
    if (!val) {
      if (!date) {
        onInput(null);
      } else {
        const dt = parse(date, '00:00:00', timeZone);
        onInput(dateInfo(dt, timeZone);
      }
    } else {
      const dt = parse(date, val, timeZone);
      onInput(dateInfo(dt, timeZone);
    }
  }, [date, timeZone, onInput]);
  const onTimeZoneChange = useCallback((opt) => {
    if (!date && !time) {
      onInput(null);
    } else {
      const dt = parse(date, time, val);
      onInput(dateInfo(dt, val));
    }
  }, [date, time, onInput]);
  return html`<div className="dateTimeInput">
    <style jsx>${`
      .dateTimeInput {
        display: flex;
        flex-direction: row;
        align-items: baseline;
      }
    `}</style>
    <${TextInput}
      type="date"
      value=${date}
      onInput=${onDateInput}
      ${...props}
    />
    <${TextInput}
      type="time"
      value=${time}
      onInput=${onTimeInput}
      ${...props}
    />
    <${TimeZoneInput} value=${timeZone} onChange=${onTimeZoneChange} />
  </div>`;
};

export default DateTimeInput;
