import { h, render } from "/web/vendor/preact.mjs";
import { useState, useEffect } from "/web/vendor/preact-hooks.mjs";
import htm from "/web/vendor/htm.mjs"; const html = htm.bind(h);

function PathEditor(props) {
  const [formFields, setFormFields] = useState({ media:"", poster:"" });
  const createFormHandler = (fieldName) => {
    return (event) => { setFormFields({ ...formFields, [fieldName]:event.target.value }); };
  };

  const readPaths = () => {
    fetch("/admin/paths", { method: "GET" }).then((response) => {
      return response.json();
    }).then((data) => {
      setFormFields(data);
    }).catch((error) => {
      console.error(error);
    });
  }

  const writePaths = () => {
    fetch("/admin/paths", { method: "POST", headers:{ "Content-Type": "application/json"}, body: JSON.stringify(formFields) }).then((response) => {
      return response.json();
    }).then((data) => {
      setFormFields(data);
    }).catch((error) => {
      console.error(error);
    });
  }

  return html`
    <div id="path-editor-root">
      <label for="mediapath">Media Path</label><input type="text" value=${formFields.media} onChange=${createFormHandler("media")} /><br />
      <label for="posterpath">Poster Path</label><input type="text" value=${formFields.poster} onChange=${createFormHandler("poster")} /><br />
      <button onClick=${readPaths}>Read</button><button onClick=${writePaths}>Write</button>
    </div>
  `;
}

export default PathEditor;
