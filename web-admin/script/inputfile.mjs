import { h } from "/web-admin/vendor/preact.mjs";
import { useState, useMemo, useEffect } from "/web-admin/vendor/preact-hooks.mjs";
import htm from "/web-admin/vendor/htm.mjs"; const html = htm.bind(h);
import api from "/web-admin/script/api.mjs";

// =============================================================================
// InputFileProperties

function InputFileProperties(props) {
  return html`
    <span>properties</span>
  `;
}

// =============================================================================
// InputFileEditor

function InputFileEditor(props) {
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [records, setRecords] = useState([]);
  const [selectedRecords, setSelectedRecords] = useState({});

  const refresh = () => {
    setLoading(true);
    setRecords([]);
    setError("");
    api("input-files", "GET", false, true).then((data) => {
      setRecords(data);
      const new_selected_records = {};
      data.forEach((record) => {
        if(selectedRecords[record.id]) {
          new_selected_records[record.id] = record;
        }
      });
      setSelectedRecords(new_selected_records);
      setLoading(false);
    }).catch((error) => {
      setError(`Error \"${error}\" retrieving input files`);
    });
  };
  useEffect(() => { refresh(); }, []);

  const selectNone = () => { setSelectedRecords({}); };
  const selectAll = () => {
    const new_selection = {};
    records.forEach((record) => { new_selection[record.id] = record; });
    setSelectedRecords(new_selection);
  };
  const selectRecord = (recordId) => {
    const new_selection = { ...selectedRecords };
    if(new_selection[recordId]) {
      delete(new_selection[recordId]);
    } else {
      new_selection[recordId] = records.find((record) => { return record.id == recordId; });
    }
    setSelectedRecords(new_selection);
  };

  const recordElements = useMemo(() => {
    return records.map((record) => { return html`
      <span class="inputfile-row" style="cursor:pointer; background-color:${!!selectedRecords[record.id] ? 'blue' : 'none'};" onClick=${() => { selectRecord(record.id); }}>
        <span class="inputfile-row-source">${record.source_location}</span>
      </span>
    `; })
  }, [records, selectedRecords]);

  return html`
    <div class="inputfile-editor" style="display:flex; flex-direction:column;">
      <span class="inputfile-header">
        <button onClick=${refresh}>Refresh</button>
        <button onClick=${selectAll}>Select All</button>
        <button onClick=${selectNone}>Select None</button>
        <span>(${Object.keys(selectedRecords).length} selected)</span>
      </span>
      <div class="inputfile-body" style="display:flex; flex-direction:row;">
        <div class="inputfile-listing" style="display:flex; flex-direction:column;">
          <div>${recordElements}</div>
        </div>
        <${InputFileProperties} records=${records} selectedRecords=${selectedRecords} loading=${loading} refresh=${refresh}/>
      </div>
    </div>
  `;
}

export default InputFileEditor;
