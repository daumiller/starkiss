import { h } from "/web-admin/vendor/preact.mjs";
import { useState, useMemo, useEffect } from "/web-admin/vendor/preact-hooks.mjs";
import htm from "/web-admin/vendor/htm.mjs"; const html = htm.bind(h);
import api from "/web-admin/script/api.mjs";

function pathBase(path) {
  return path.split("/").pop();
}

function timeString(duration) {
  const seconds = duration % 60; duration = (duration - seconds) / 60;
  const minutes = duration % 60; duration = (duration - minutes) / 60;
  const hours   = duration;
  var time_string = `${seconds}s`;
  if(minutes > 0) { time_string = `${minutes}m ${time_string}`; }
  if(hours   > 0) { time_string = `${hours}h ${time_string}`; }
  return time_string;
}

function dateString(unix_seconds) {
  if(unix_seconds == 0) { return ""; }

  const date_time = new Date(unix_seconds * 1000);
  const year   = date_time.getFullYear();
  const month  = (date_time.getMonth() + 1).toString().padStart(2, "0");
  const date   = date_time.getDate().toString().padStart(2, "0");
  const hour   = date_time.getHours().toString().padStart(2, "0");
  const minute = date_time.getMinutes().toString().padStart(2, "0");
  return `${year}-${month}-${date} ${hour}:${minute}`;
}

// =============================================================================
// InputFileProperties

function InputFileProperties(props) {
  const selectedRecord = useMemo(() => {
    const selected_keys  = Object.keys(props.selectedRecords);
    const selected_count = selected_keys.length;
    if(selected_count == 0) { return null; }
    if(selected_count >  1) { return null; }
    return props.selectedRecords[selected_keys[0]];
  }, [props.selectedRecords]);

  const inputEditMap = () => {};
  const inputEditMeta = () => {};

  return html`
    <div class="inputfile-properties" style="display:flex; flex-direction:column;">
      <span>      
        <button onClick=${inputEditMap}>Edit Stream Map</button>
        <button onClick=${inputEditMeta}>Edit Metadata</button>
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

  const refresh = () => {
    setLoading(true);
    setRecords([]);
    setError("");
    api("input-files", "GET", false, true).then((data) => {
      for(let idx=0; idx<data.length; ++idx) {
        data[idx].status = "needs_stream_map";
        if(data[idx].stream_map.length        > 0) { data[idx].status = "needs_transcoding";     }
        if(data[idx].transcoding_time_started > 0) { data[idx].status = "transcoding_started";   }
        if(data[idx].transcoding_time_elapsed > 0) { data[idx].status = "transcoding_succeeded"; }
        if(data[idx].transcoding_error.length > 0) { data[idx].status = "transcoding_failed";    }
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
    }).catch((error) => {
      setError(`Error \"${error}\" retrieving input files`);
    });
  };
  useEffect(() => { refresh(); }, []);

  const resetStatus = () => {};
  const deleteRecords = () => {};

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
  const selectRecord = (recordId) => {
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
          sort_funct = () => { return a.base_name.localeCompare(b.base_name); }
        } else {
          sort_funct = () => { return b.base_name.localeCompare(a.base_name); }
        }
        break;
      case "duration":
        if(sortOrder == "ascending") {
          sort_funct = () => { return a.source_duration - b.source_duration; }
        } else {
          sort_funct = () => { return b.source_duration - a.source_duration; }
        }
        break;
      case "status":
        if(sortOrder == "ascending") {
          sort_funct = () => { return a.status.localeCompare(b.status); }
        } else {
          sort_funct = () => { return b.status.localeCompare(a.status); }
        }
        break;
    }
    records.sort(sort_funct);

    return records.map((record) => {
      return html`
        <tr class="inputfile-row" style="cursor:pointer; background-color:${!!selectedRecords[record.id] ? 'blue' : 'none'};" onClick=${() => { selectRecord(record.id); }}>
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
