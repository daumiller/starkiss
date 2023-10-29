import { h, render } from "/web/vendor/preact.mjs";
import { useState, useEffect } from "/web/vendor/preact-hooks.mjs";
import htm from "/web/vendor/htm.mjs"; const html = htm.bind(h);

import PathEditor        from "/web/script/admin-path-editor.mjs";
import CategoryEditor    from "/web/script/admin-category-editor.mjs";
import UnprocessedEditor from "/web/script/admin-unprocessed-editor.mjs";
import TranscodingEditor from "/web/script/admin-transcoding-editor.mjs";

function Admin(props) {
  return html`
    <div id="admin-root">
      <section><h2>Paths             </h2><${PathEditor}        /></section>
      <section><h2>Categories        </h2><${CategoryEditor}    /></section>
      <section><h2>Unprocessed Queue </h2><${UnprocessedEditor} /></section>
      <section><h2>Transcoding Queue </h2><${TranscodingEditor} /></section>
    </div>
  `;
}

render(html`<${Admin} />`, document.body);
