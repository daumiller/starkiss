import { h, render } from "/web-client/vendor/preact.mjs";
import htm from "/web-client/vendor/htm.mjs"; const html = htm.bind(h);

function Starkiss(props) {
  return html`
    <div id="client-root">
      <h1>Starkiss</h1>
    </div>
  `;
}

render(html`<${Starkiss} />`, document.body);
