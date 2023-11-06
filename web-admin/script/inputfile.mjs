import { h } from "/web-admin/vendor/preact.mjs";
import { useState, useMemo, useEffect } from "/web-admin/vendor/preact-hooks.mjs";
import htm from "/web-admin/vendor/htm.mjs"; const html = htm.bind(h);
import api, { timeString, dateString } from "/web-admin/script/api.mjs";

function pathBase(path) {
  return path.split("/").pop();
}

const InputFileStatus = {
  NEEDS_STREAM_MAP:      "needs-stream-map",
  NEEDS_TRANSCODING:     "needs-transcoding",
  TRANSCODING_STARTED:   "transcoding-started",
  TRANSCODING_SUCCEEDED: "transcoding-succeeded",
  TRANSCODING_FAILED:    "transcoding-failed",
};

// =============================================================================
// InputFileMapEditor

function InputFileMapEditor(props) {
  const [selectedStreams, setSelectedStreams] = useState([]);
  const [errorMessage, setErrorMessage] = useState(null);

  const toggleSelectedStream = (stream_index) => {
    const newSelection = [ ...selectedStreams ];
    const idx = newSelection.indexOf(stream_index);
    if(idx >= 0) {
      newSelection.splice(idx, 1);
    } else {
      newSelection.push(stream_index);
    }
    newSelection.sort((a,b) => { return a - b; });
    setSelectedStreams(newSelection);
  };

  const updateMap = async () => {
    if(!(props.record && props.record.id)) { return; }
    // console.log(`updating map from ${props.record.stream_map} to ${selectedStreams}`);
    const result = await api(`input-file/${props.record.id}/map`, "POST", selectedStreams);
    if((result.status < 200) || (result.status > 299)) {
      setErrorMessage(`Error updating record: ${(result.body && result.body.error) || result.status}`);
      return;
    }
    props.hide();
    props.refresh();
  };

  const streamDescription = (stream) => {
    if(stream.stream_type == "video") {
      return `#${stream.index} ${stream.stream_type} ${stream.codec} ${stream.width}x${stream.height} ${stream.fps}fps`;
    }
    if(stream.stream_type == "audio") {
      return `#${stream.index} ${stream.stream_type} ${stream.codec} ${stream.channels}ch lang:${stream.language || "(unknown)"}`;
    }
    if(stream.stream_type == "subtitle") {
      return `#${stream.index} ${stream.stream_type} ${stream.codec} lang:${stream.language || "(unknown)"}`;
    }
    return `#${stream.index} (unknown) ${stream.stream_type} ${stream.codec}`;
  };

  useEffect(() => {
    setSelectedStreams((props.record && props.record.stream_map) || []);
  }, [props.record]);

  const streamRows = useMemo(() => {
    if(!props.record) { return []; }
    const source_streams = props.record.source_streams || [];
    const rows = [];
    for(let idx=0; idx < source_streams.length; idx++) {
      const stream_curr = source_streams[idx];
      const stream_selected = selectedStreams.indexOf(stream_curr.index) >= 0
      rows.push(html`
        <li key=${props.record.id}>
          <input type="checkbox" checked=${stream_selected} onChange=${() => { toggleSelectedStream(stream_curr.index); }}/>
          <label style="cursor:pointer;" onClick=${() => { toggleSelectedStream(stream_curr.index); }}>${streamDescription(stream_curr)}</label>
        </li>
      `);
    }
    return rows;
  }, [selectedStreams, props.record]);

  return html`
    <div style="z-index:5; position:fixed; top:0px;bottom:0px;left:0px;right:0px; background-color:#00000088; display:${props.show ? "block" : "none" };" >
      <div style="display:inline-block; border:2px solid #000000; border-radius:12px; background-color:#FFFFFF; min-width:50vw; min-height:50vh;">
        <button onClick=${props.hide}>Close</button><br />
        <label>${(props.record && props.record.source_location) || ''}</label><br />
        <ul>${streamRows}</ul>
        <label>${errorMessage || ''}</label><br />
        <button onClick=${updateMap}>Update</button>
      </div>
    </div>
  `;
}

// =============================================================================
// InputFileProperties

function InputFileProperties(props) {
  const [showMapEditor, setShowMapEditor] = useState(false);

  const selectedRecord = useMemo(() => {
    const selected_keys  = Object.keys(props.selectedRecords);
    const selected_count = selected_keys.length;
    if(selected_count == 0) { return null; }
    if(selected_count >  1) { return null; }
    return props.selectedRecords[selected_keys[0]];
  }, [props.selectedRecords]);

  const inputEditMap = () => { setShowMapEditor(true); };
  const inputEditMeta = () => {};

  if(!selectedRecord) { return html``; }

  return html`
    <div class="inputfile-properties" style="display:flex; flex-direction:column;">
      <span>      
        <button onClick=${inputEditMap}>Edit Stream Map</button>
        <button onClick=${inputEditMeta} disabled=${selectedRecord.status != InputFileStatus.TRANSCODING_SUCCEEDED}>Edit Metadata</button>
      </span>
      <table><tbody>
        <tr><td>Filename                 </td><td> ${selectedRecord.source_location}                      </td></tr>
        <tr><td>Status                   </td><td> ${selectedRecord.status}                               </td></tr>
        <tr><td>Time Scanned             </td><td> ${dateString(selectedRecord.time_scanned)}             </td></tr>
        <tr><td>Transcoding Command      </td><td> ${selectedRecord.transcoding_command}                  </td></tr>
        <tr><td>Transcoding Time Started </td><td> ${dateString(selectedRecord.transcoding_time_started)} </td></tr>
        <tr><td>Transcoding Time Elapsed </td><td> ${timeString(selectedRecord.transcoding_time_elapsed)} </td></tr>
        <tr><td>Transcoding Error        </td><td> ${selectedRecord.transcoding_error}                    </td></tr>
      </tbody></table>
      <${InputFileMapEditor} show=${showMapEditor} hide=${() => { setShowMapEditor(false); }} record=${selectedRecord} refresh=${props.refresh}/>
    </span>
  `;
}

// =============================================================================
// InputFileEditor

function InputFileEditor(props) {
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [records, setRecords] = useState([]);
  const [sortMethod, setSortMethod] = useState("base_name");
  const [sortOrder, setSortOrder] = useState("ascending");
  const [selectedRecords, setSelectedRecords] = useState({});
  const [selectedCount, setSelectedCount] = useState(0);

  const refresh = async () => {
    setLoading(true);
    setRecords([]);
    setError("");
    const result = await api("input-files", "GET");
    if((result.status < 200) || (result.status > 299)) {
      setError(`Error ${(result.body && result.body.error) || result.status} retrieving input files`);
      return;
    }
    const data = result.body;
    for(let idx=0; idx<data.length; ++idx) {
      data[idx].status = InputFileStatus.NEEDS_STREAM_MAP;
      if(data[idx].stream_map.length        > 0) { data[idx].status = InputFileStatus.NEEDS_TRANSCODING;     }
      if(data[idx].transcoding_time_started > 0) { data[idx].status = InputFileStatus.TRANSCODING_STARTED;   }
      if(data[idx].transcoding_time_elapsed > 0) { data[idx].status = InputFileStatus.TRANSCODING_SUCCEEDED; }
      if(data[idx].transcoding_error.length > 0) { data[idx].status = InputFileStatus.TRANSCODING_FAILED;    }
      data[idx].base_name = pathBase(data[idx].source_location);
    }
    setRecords(data);
    const new_selected_records = {};
    data.forEach((record) => {
      if(selectedRecords[record.id]) {
        new_selected_records[record.id] = record;
      }
    });
    setSelectedRecords(new_selected_records);
    setLoading(false);
  };
  useEffect(() => { refresh(); }, []);

  const resetStatus = async () => {
    const ids = Object.keys(selectedRecords);
    if(ids.length == 0) { return; }

    setLoading(true);
    setError("");
    const promises = [];
    for(let index=0; index<ids.length; ++index) {
      const id = ids[index];
      promises.push(api(`input-file/${id}/reset`, "POST"));
    }
    const results = await Promise.all(promises);
    const errors = [];
    for(let index=0; index<results.length; ++index) {
      const result = results[index];
      if((result.status < 200) || (result.status > 299)) {
        errors.push(`Error ${(result.body && result.body.error) || result.status} resetting input file status`);
      }
    }
    setLoading(false);
    if(errors.length > 0) {
      setError(errors.join(" \n"));
    } else {
      window.setTimeout(refresh, 0);
    }
  };
  const deleteRecords = async() => {
    const ids = Object.keys(selectedRecords);
    if(ids.length == 0) { return; }
    if(confirm(`Delete ${ids.length} record(s)?`) == false) { return; }

    setLoading(true);
    setError("");
    const errors = [];
    for(let index=0; index<ids.length; ++index) {
      // TODO: we 500 here if multiple deletes run simultaneously...
      // this really shouldn't happen, the server should block if it needs time to update...
      const id = ids[index];
      const result = await api(`input-file/${id}`, "DELETE");
      if((result.status < 200) || (result.status > 299)) {
        errors.push(`Error ${(result.body && result.body.error) || result.status} deleting input file`);
      }
    }
    setLoading(false);
    if(errors.length > 0) {
      setError(errors.join(" \n"));
    } else {
      window.setTimeout(refresh, 0);
    }
  };

  const selectNone = () => {
    setSelectedRecords({});
    setSelectedCount(0);
  };
  const selectAll = () => {
    const new_selection = {};
    let new_selected_count = 0;
    records.forEach((record) => { ++new_selected_count; new_selection[record.id] = record; });
    setSelectedRecords(new_selection);
    setSelectedCount(new_selected_count);
  };
  const selectRecord = (event, recordId) => {
    if(event.shiftKey == false) {
      const new_selection = {};
      new_selection[recordId] = records.find((record) => { return record.id == recordId; });
      setSelectedRecords(new_selection);
      setSelectedCount(1);
      return;
    }

    const new_selection = { ...selectedRecords };
    let new_selection_count = selectedCount;
    if(new_selection[recordId]) {
      delete(new_selection[recordId]);
      --new_selection_count
    } else {
      new_selection[recordId] = records.find((record) => { return record.id == recordId; });
      ++new_selection_count;
    }
    setSelectedRecords(new_selection);
    setSelectedCount(new_selection_count);
  };

  const setSort = (method) => {
    if(method == sortMethod) {
      setSortOrder(sortOrder == "ascending" ? "descending" : "ascending");
    } else {
      setSortMethod(method);
      setSortOrder("ascending");
    }
  };

  const recordElements = useMemo(() => {
    var sort_funct = null;
    switch(sortMethod) {
      case "base_name":
        if(sortOrder == "ascending") {
          sort_funct = (a, b) => { return a.base_name.localeCompare(b.base_name); }
        } else {
          sort_funct = (a, b) => { return b.base_name.localeCompare(a.base_name); }
        }
        break;
      case "duration":
        if(sortOrder == "ascending") {
          sort_funct = (a, b) => { return a.source_duration - b.source_duration; }
        } else {
          sort_funct = (a, b) => { return b.source_duration - a.source_duration; }
        }
        break;
      case "status":
        if(sortOrder == "ascending") {
          sort_funct = (a, b) => { return a.status.localeCompare(b.status); }
        } else {
          sort_funct = (a, b) => { return b.status.localeCompare(a.status); }
        }
        break;
    }
    records.sort(sort_funct);

    return records.map((record) => {
      return html`
        <tr class="inputfile-row" style="cursor:pointer; background-color:${!!selectedRecords[record.id] ? 'blue' : 'none'};" onClick=${(event) => { selectRecord(event, record.id); }}>
          <td>${pathBase(record.source_location)}</td>
          <td>${timeString(record.source_duration)}</td>
          <td>${record.status}</td>
        </tr>
      `;
    });
  }, [records, selectedRecords, sortOrder, sortMethod]);

  return html`
    <div class="inputfile-editor" style="display:flex; flex-direction:column;">
      <span class="inputfile-error">${error}</span>
      <span class="inputfile-loading">${loading && "Loading..."}</span>
      <span class="inputfile-header">
        <button onClick=${refresh}>Refresh</button>
        <button onClick=${selectAll}>Select All</button>
        <button onClick=${selectNone}>Select None</button>
        <button onClick=${resetStatus}>Reset Status</button>
        <button onClick=${deleteRecords}>Delete Record(s)</button>
      </span>
      <div class="inputfile-body" style="display:flex; flex-direction:row;">
        <table>
          <thead>
            <th style="cursor:pointer;" onClick=${() => { setSort("base_name"); }}>Filename</th>
            <th style="cursor:pointer;" onClick=${() => { setSort("duration" ); }}>Duration</th>
            <th style="cursor:pointer;" onClick=${() => { setSort("status"   ); }}>Status</th>
          </thead>
          <tbody>${recordElements}</tbody>
        </table>
        ${(selectedCount == 1) && html`<${InputFileProperties} selectedRecords=${selectedRecords} refresh=${refresh}/>`}
      </div>
    </div>
  `;
}

export default InputFileEditor;
