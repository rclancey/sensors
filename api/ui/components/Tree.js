import { h } from 'https://unpkg.com/preact@latest?module';
import htm from 'https://unpkg.com/htm?module';

const html = htm.bind(h);

const Comment = ({ children }) => (
  html`<span className="comment">// ${children}</span>`
);

const Toggle = ({ open, onToggle }) => (
  html`<div className="toggleWrapper">
    <div className="toggle ${open ? 'open' : 'closed'}" onClick=${onToggle} />
  </diiv>`
);

const Comma = ({ last }) => last ? null : html`<span className="comma">,</span>`;

const PairVal = ({ val }) => {
  if (val === null) {
    return html`<${NullVal} />`;
  }
  switch(typeof val) {
  case 'string':
    return html`<${Str} str=${val} />`;
  case 'number':
    return html`<${Num} num=${val} />`;
  case 'boolean':
    return html`<${Bool} bool=${val} />`;
  }
  return null;
};

const Pair = ({ name, val, last, depth=1 }) => {
  const [open, setOpen] = useState(depth < 4);
  const onToggle = useCallback(() => setOpen((orig) => !orig), []);
  if (Array.isArray(val)) {
    if (open) {
      if (val.length === 0) {
        return html`<div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <${Key} name=${name} />
          <span>: [ ]</span>
          <${Comma} last=${last} />
        </div>`;
      }
      return html`<Fragment>
        <div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <${Key} name=${name} />
          <span>: [</span>
        </div>
        <div className="child">
          <${Arr} arr=${val} depth=${depth+1} />
        </div>
        <div className="closing">
          <span>]</span>
          <${Comma} last=${last} />
        </div>
      </Fragment>`;
    }
    return html`<div className="pair closed">
      <${Toggle} open=${open} onToggle=${onToggle} />
      <${Key} name=${name} />
      <span>: [...]</span>
      <${Comma} last=${last} />
      <${Comment}>${val.length} ${val.length === 1 ? 'item' : 'items'}</${Comment}>
    </div>`;
  }
  if (val !== null && val !== undefined && typeof val === 'object') {
    if (open) {
      if (Object.keys(val).length === 0) {
        return html`<div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <${Key} name=${name} />
          <span>: { }</span>
          <${Comma} last=${last} />
        </div>`;
      }
      return html`<Fragment>
        <div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <${Key} name=${name} />
          <span>: {</span>
        </div>
        <div className="child">
          <${Obj} obj=${val} depth=${depth+1} />
        </div>
        <div className="closing">
          <span>}</span>
          <${Comma} last=${last} />
        </div>
      </Fragment>`;
    }
    return html`<div className="pair closed">
      <${Toggle} open=${open} onToggle=${onToggle} />
      <${Key} name=${name} />
      <span>: {...}</span>
      <${Comma} last=${last} />
      <${Comment}>${Object.keys(val).length} ${Object.keys(val).length === 1 ? 'item' : 'items'}</${Comment}>
    </div>`
  }
  return html`<div className="pair">
    <${Key} name=${name} />
    <span>: </span>
    <${PairVal} val=${val} />
    <${Comma} last=${last} />
  </div>`;
};

const Obj = ({ obj, depth }) => {
  const keys = Object.keys(obj).sort((a, b) => a < b ? -1 : 1);
  const last = keys.length - 1;
  return keys.map((k, i) => html`<${Pair} key=${k} name=${k} val=${obj[k]} last=${i === last} depth=${depth} />`);
};

const ArrVal = ({ val }) => {
  if (val === null) {
    return html`<${NullVal} />`;
  }
  switch (typeof val) {
  case 'string':
    return html`<${Str} str=${val} />`;
  case 'number':
    return html`<${Num} num=${val} />`;
  case 'boolean':
    return html`<${Bool} bool=${val} />`;
  }
  return null;
};

const ArrItem = ({ item, last, depth=1 }) => {
  const [open, setOpen] = useState(depth < 4);
  const onToggle = useCallback(() => setOpen((orig) => !orig), []);
  if (Array.isArray(item)) {
    if (open) {
      if (item.length === 0) {
        return html`<div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <span>[ ]</span>
          <${Comma} last=${last} />
        </div>`;
      }
      return html`<Fragment>
        <div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <span>[</span>
        </div>
        <div className="child">
          <${Arr} arr=${item} depth=${depth+1} />
        </div>
        <div className="closing">
          <span>]</span>
          <${Comma} last=${last} />
        </div>
      </Fragment>`;
    }
    return html`<div className="pair closed">
      <${Toggle} open=${open} onToggle=${onToggle} />
      <span>[...]</span>
      <${Comma} last=${last} />
      <${Comment}>${item.length} ${item.length === 1 ? 'item' : 'items'}</${Comment}>
    </div>`;
  }
  if (typeof item === 'object') {
    if (open) {
      if (Object.keys(item).length === 0) {
        return html`<div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <span>{ }</span>
          <${Comma} last=${last} />
        </div>`;
      }
      return html`<Fragment>
        <div className="pair open">
          <${Toggle} open=${open} onToggle=${onToggle} />
          <span>{</span>
        </div>
        <div className="child">
          <${Obj} obj=${item} depth=${depth+1} />
        </div>
        <div className="closing">
          <span>}</span>
          <${Comma} last=${last} />
        </div>
      </Fragment>`;
    }
    return html`<div className="pair closed">
      <${Toggle} open=${open} onToggle=${onToggle} />
      <span>{...}</span>
      <${Comma} last=${last} />
      <${Comment}>${Object.keys(item).length} ${Object.keys(item).length === 1 ? 'item' : 'items'}</${Comment}>
    </div>`
  }
  return html`<div className="pair">
    <${ArrVal} val=${item} />
    <${Comma} last=${last} />
  </div>`;
};

const Arr = ({ arr, depth }) => {
  const last = arr.length - 1;
  return arr.map((item, i) => html`<${ArrItem} key=${i} item=${item} last=${i === last} depth=${depth} />`);
};

const Key = ({ name }) => html`<div className="key">"${name}"</div>`;

const Str = ({ str }) => html`<div className="string">"${str}"</div>`;

const Num = ({ num }) => html`<div className="number">${num}</div>`;

const Bool = ({ bool }) => html`<div className="boolean">${bool.toString()}</div>`;

const NullVal = () => html`<div className="null">null</div>`;

export const Tree = ({ tree }) => html`<div className="tree"><${ArrItem} item=${tree} last /></div>`;

export default Tree;
