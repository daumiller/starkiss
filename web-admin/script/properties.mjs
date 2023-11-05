import { h } from "/web-admin/vendor/preact.mjs";
import { useState, useMemo, useEffect } from "/web-admin/vendor/preact-hooks.mjs";
import htm from "/web-admin/vendor/htm.mjs"; const html = htm.bind(h);
import api from "/web-admin/script/api.mjs";

// =============================================================================
// PropertyKeyEditor

function PropertyKeyEditor(props) {
  const [value, setValue] = useState(props.value);

  const updateKeyValue = async () => {
    const result = await api("properties", "POST", { [props.keyName]:value });
    if((result.status < 200) || (result.status > 299)) {
      props.onError(`Error ${(result.body && result.body.error) || result.status} updating ${props.keyName}`);
      return;
    }
    props.onUpdate(props.keyName, value);
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

  useEffect(async () => {
    const result = await api("properties", "GET");
    if((result.status < 200) || (result.status > 299)) {
      setError(`Error ${(result.body && result.body.error) || result.status} retrieving properties`);
      return;
    }
    setProperties(result.body);
    setError("");
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
