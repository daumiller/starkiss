import { h, render } from "/web-admin/vendor/preact.mjs";
import htm from "/web-admin/vendor/htm.mjs"; const html = htm.bind(h);

import PropertyEditor from "/web-admin/script/properties.mjs";
import CategoryEditor from "/web-admin/script/categories.mjs";
import MetadataEditor from "/web-admin/script/metadata.mjs";
// import InputFileEditor from "/web-admin/script/inputfiles.mjs";

function Starkiss(props) {
  return html`
    <div id="admin-root">
      <section><h2>Properties </h2><${PropertyEditor} /></section>
      <section><h2>Categories </h2><${CategoryEditor} /></section>
      <section><h2>Metadata   </h2><${MetadataEditor} /></section>
    </div>
  `;
}

render(html`<${Starkiss} />`, document.body);
