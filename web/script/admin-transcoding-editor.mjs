import { h, render } from "/web/vendor/preact.mjs";
import { useState, useEffect } from "/web/vendor/preact-hooks.mjs";
import htm from "/web/vendor/htm.mjs"; const html = htm.bind(h);

function TranscodingEditor(props) {
  return html`
    <div id="transcoding-editor-root">
      Transcoding Editor
    </div>
  `;
}

export default TranscodingEditor;
