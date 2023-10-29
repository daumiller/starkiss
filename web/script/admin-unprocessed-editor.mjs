import { h, render } from "/web/vendor/preact.mjs";
import { useState, useEffect, useMemo } from "/web/vendor/preact-hooks.mjs";
import htm from "/web/vendor/htm.mjs"; const html = htm.bind(h);

function baseName(path) {
  if(!path) { return ""; }
  const parts = path.split("/");
  return parts[parts.length - 1];
};

function UnprocessedEntry(props) {
  const [expanded, setExpanded] = useState(false);

  const durationString = (duration) => {
    duration = (duration || 0) | 0;
    const seconds = duration % 60; duration = (duration - seconds) / 60;
    const minutes = duration % 60; duration = (duration - minutes) / 60;
    const hours   = duration;
    let duration_string = `${seconds}s`;
    if(minutes > 0) { duration_string = `${minutes}m ${duration_string}`; }
    if(hours   > 0) { duration_string = `${hours}h ${duration_string}`; }
    return duration_string;
  };

  const dateTimeString = (unix_time) => {
    if(!unix_time) { return ""; }
    const date_time = new Date(unix_time * 1000);
    const year_string   = date_time.getFullYear();
    const month_string  = (date_time.getMonth() + 1).toString().padStart(2, "0");
    const day_string    = date_time.getDate().toString().padStart(2, "0");
    const hour_string   = date_time.getHours().toString().padStart(2, "0");
    const minute_string = date_time.getMinutes().toString().padStart(2, "0");
    return `${year_string}/${month_string}/${day_string} ${hour_string}:${minute_string}`;
  };

  const mapStreams = (event) => {
    this.props.showMapEditor(this.props.entry);
    event.preventDefault();
  };
  const toggleExpanded = (event) => {
    setExpanded(!expanded);
    event.preventDefault();
  };
  const selectedUpdater = (event) => {
    this.props.selectedUpdater(props.entry.id, !props.selected);
    event.preventDefault();
  };

  return html`<li key=${props.entry.id}>
    <input type="checkbox" checked=${props.selected} onChange=${selectedUpdater} />
    <label style="cursor:pointer;" onClick=${selectedUpdater}>${baseName(props.entry.source_location)}</label>
    <button onClick=${mapStreams}>Map Streams</button>
    <button onClick=${toggleExpanded}>Details</button>
    <div style=${{ display: expanded ? "block" : "none" }}>
      <table><tbody>
        <tr><td>Source Location</td><td>${props.entry.source_location || ''}</td></tr>
        <tr><td>Transcoded Location</td><td>${props.entry.transcoded_location || ''}</td></tr>
        <tr><td>Duration</td><td>${durationString(props.entry.duration)}</td></tr>
        <tr><td>Time Scanned</td><td>${dateTimeString(props.entry.created_at)}</td></tr>
      </tbody></table>
    </div>
  </li>`;
}

function MapEditor(props) {
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

  const updateMap = () => {
    if(!(props.entry && props.entry.id)) { return; }
    // console.log(`updating map from ${props.entry.transcoded_streams} to ${selectedStreams}`);
    fetch("/admin/unprocessed/map/" + props.entry.id, {
      method: "POST",
      headers:{ "Content-Type": "application/json" },
      body: JSON.stringify(selectedStreams),
    }).then((response) => {
      if(response.status != 200) { throw new Error(`HTTP ${response.status}`); }
      props.hide();
    }).catch((error) => {
      console.error(error);
      setErrorMessage(`Error updating record: ${error}`);
    });
  };

  const streamDescription = (stream) => {
    if(stream.type == "video") {
      return `#${stream.index} ${stream.type} ${stream.codec} ${stream.width}x${stream.height} ${stream.fps}fps`;
    }
    if(stream.type == "audio") {
      return `#${stream.index} ${stream.type} ${stream.codec} ${stream.channels}ch lang:${stream.language}`;
    }
    if(stream.type == "subtitle") {
      return `#${stream.index} ${stream.type} ${stream.codec} lang:${stream.language}`;
    }
    return `#${stream.index} (unknown) ${stream.type} ${stream.codec}`;
  };

  useEffect(() => {
    setSelectedStreams((props.entry && props.entry.transcoded_streams) || []);
  }, [props.entry]);

  const streamRows = useMemo(() => {
    if(!props.entry) { return []; }
    const source_streams = props.entry.source_streams || [];
    const rows = [];
    for(let idx=0; idx < source_streams.length; idx++) {
      const stream_curr = source_streams[idx];
      const stream_selected = selectedStreams.indexOf(stream_curr.index) >= 0
      rows.push(html`
        <li>
          <input type="checkbox" checked=${stream_selected} onChange=${() => { toggleSelectedStream(stream_curr.index); }}/>
          <label style="cursor:pointer;" onClick=${() => { toggleSelectedStream(stream_curr.index); }}>${streamDescription(stream_curr)}</label>
        </li>
      `);
    }
    return rows;
  }, [selectedStreams, props.entry]);

  return html`
    <div style="z-index:5; position:fixed; top:0px;bottom:0px;left:0px;right:0px; background-color:#00000088; display:${props.show ? "block" : "none" };" >
      <div style="display:inline-block; border:2px solid #000000; border-radius:12px; background-color:#FFFFFF; min-width:50vw; min-height:50vh;">
        <button onClick=${props.hide}>Close</button><br />
        <label>${(props.entry && props.entry.source_location) || ''}</label><br />
        <ul>${streamRows}</ul>
        <label>${errorMessage || ''}</label><br />
        <button onClick=${updateMap}>Update</button>
      </div>
    </div>
  `;
}

function UnprocessedEditor(props) {
  const [unprocessed, setUnprocessed] = useState([]);
  const [filterValues, setFilterValues] = useState({ chkMap:false, chkTrans:false, chkMeta:false });
  const [selectedIds, setSelectedIds] = useState({});
  const [mapEditorProperties, setMapEditorProperties] = useState({ show:false, entry:null });
  const [errorMessage, setErrorMessage] = useState(null);

  const queueForTrans = () => {
    const ids_to_queue = [];
    for(let id in selectedIds) {
      if(!selectedIds.hasOwnProperty(id)) { continue; }
      ids_to_queue.push(id);
    }
    if(ids_to_queue.length == 0) { return; }

    fetch("/admin/unprocessed/queue", {
      method: "POST",
      headers:{ "Content-Type": "application/json" },
      body: JSON.stringify(ids_to_queue),
    }).then((response) => {
      if(response.status == 200) { setErrorMessage(null); return; }
      if(response.status != 400) { throw new Error(`HTTP ${response.status}`); }
      response.json().then((data) => {
        setErrorMessage(`Error queueing records: ${JSON.stringify(data)}`);
      });
    }).catch((error) => {
      console.error(error);
      setErrorMessage(`Error queueing records: ${error}`);
    });
  };

  const createFilterHandler = (fieldName) => {
    return (event) => {
      setFilterValues({ ...filterValues, [fieldName]:event.target.checked });
    };
  }

  const loadUnprocessed = () => {
    fetch("/admin/unprocessed", { method: "GET" }).then((response) => {
      return response.json();
    }).then((data) => {
      const unprocessedMap = {};
      for(let idx=0; idx < data.length; idx++) {
        const unprocessed = data[idx];
        unprocessedMap[unprocessed.id] = unprocessed;
      }
      setUnprocessed(unprocessedMap);
    }).catch((error) => {
      console.error(error);
    });
  }

  const selectedUpdater = (id, selected) => {
    const newSelection = { ...selectedIds };
    if(selected) {
      newSelection[id] = true;
    } else {
      delete newSelection[id];
    }
    setSelectedIds(newSelection);
  };

  const showMapEditor = (entry) => {
    setMapEditorProperties({ show:true, entry:entry });
  };
  const hideMapEditor = () => {
    setMapEditorProperties({ show:false, entry:null });
  };

  useEffect(() => {
    setSelectedIds({});
  }, [filterValues]);

  const unprocessedRows = useMemo(() => {
    const rows = [];
    if(!unprocessed) { return rows; }
    for(let id in unprocessed) {
      if(!unprocessed.hasOwnProperty(id)) { continue; }
      if(filterValues.chkMap   && !unprocessed[id].needs_stream_map ) { continue; }
      if(filterValues.chkTrans && !unprocessed[id].needs_transcoding) { continue; }
      if(filterValues.chkMeta  && !unprocessed[id].needs_metadata   ) { continue; }
      const entry = unprocessed[id];
      if(!entry) { continue; }
      rows.push(html`
        <${UnprocessedEntry}
          entry=${entry}
          selected=${!!selectedIds[id]}
          selectedUpdater=${selectedUpdater}
          showMapEditor=${showMapEditor}
        />
      `);
    }
    return rows;
  }, [unprocessed, filterValues, selectedIds]);

  return html`
    <div id="unprocessed-editor-root">
      <label for="chkMap">Needs Map</label><input type="checkbox" checked=${filterValues.chkMap} onChange=${createFilterHandler("chkMap")} />
      <label for="chkTrans">Needs Trans</label><input type="checkbox" checked=${filterValues.chkTrans} onChange=${createFilterHandler("chkTrans")} />
      <label for="chkMeta">Needs Meta</label><input type="checkbox" checked=${filterValues.chkMeta} onChange=${createFilterHandler("chkMeta")}/>
      <button onClick=${loadUnprocessed}>Load</button><br />
      <ul>${unprocessedRows}</ul>
      <label>${errorMessage || ''}</label><br />
      <button onClick=${queueForTrans}>Queue for Transcoding</button>
      <${MapEditor} show=${mapEditorProperties.show} entry=${mapEditorProperties.entry} hide=${hideMapEditor} />
    </div>
  `;
}

export default UnprocessedEditor;
