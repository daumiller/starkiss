import { h } from "/web-admin/vendor/preact.mjs";
import { useState, useMemo, useEffect } from "/web-admin/vendor/preact-hooks.mjs";
import htm from "/web-admin/vendor/htm.mjs"; const html = htm.bind(h);
import api from "/web-admin/script/api.mjs";

// =============================================================================
// PropertyKeyEditor

function PropertyKeyEditor(props) {
  const [value, setValue] = useState(props.value);

  const updateKeyValue = () => {
    api("properties", "POST", { [props.keyName]:value }).then(() => {
      props.onUpdate(props.keyName, value);
    }).catch((error) => {
      props.onError(`Error \"${error}\" updating ${props.keyName}`);
    });
  }

  const cancelUpdate = () => { setValue(props.value); }

  return html`
    <tr>
      <td><label for=${props.keyName}>${props.keyName}</label></td>
      <td><input type="text" name="${props.keyName}" value="${value}" onInput=${(event) => { setValue(event.target.value); }} /></td>
      <td>
        <button onClick=${updateKeyValue} disabled=${value == props.value}>Update</button>
        <button onClick=${cancelUpdate}   disabled=${value == props.value}>Cancel</button>
      </td>
    </tr>
  `;
}

// =============================================================================
// PropertyEditor

function PropertyEditor(props) {
  const [properties, setProperties] = useState({});
  const [error, setError] = useState("");

  const onKeyValueUpdated = (key, value) => {
    setError("");
    setProperties({ ...properties, [key]:value });
  };
  const onKeyValueError = (message) => {
    setError(message);
  };

  const keyValueEditors = useMemo(() => {
    return Object.entries(properties).map(([key, value]) => {
      return html`<${PropertyKeyEditor} key=${key} keyName=${key} value=${value} onUpdate=${onKeyValueUpdated} onError=${onKeyValueError} />`;
    });
  }, [properties]);

  useEffect(() => {
    api("properties", "GET", false, true).then((data) => {
      setProperties(data);
      setError("");
    }).catch((error) => {
      setError(`Error \"${error}\" retrieving properties`);
    });
  }, []);

  return html`
    <div id="property-editor-root">
      <span class="error">${error}</span>
      <table>
        <thead><tr>
          <th>Key</th>
          <th>Value</th>
          <th></th>
        </tr></thead>
        <tbody>${keyValueEditors}</tbody>
      </table>
    </div>
  `;
}

export default PropertyEditor;
