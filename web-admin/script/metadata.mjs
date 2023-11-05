import { h } from "/web-admin/vendor/preact.mjs";
import { useState, useMemo, useEffect } from "/web-admin/vendor/preact-hooks.mjs";
import htm from "/web-admin/vendor/htm.mjs"; const html = htm.bind(h);
import api, { apiFile, timeString, sizeString } from "/web-admin/script/api.mjs";

const MetadataMediaType = {
  FILE_VIDEO: "file-video",
  FILE_AUDIO: "file-audio",
  SEASON:     "season",
  SERIES:     "series",
  ALBUM:      "album",
  ARTIST:     "artist",
  MOVIE:      "movie",
  MUSIC:      "music",
};

// =============================================================================
// MetadataGroupCreator

function MetadataGroupCreator(props) {
  const [displayName, setDisplayName] = useState("");
  const [mediaType, setMediaType] = useState(props.allowedTypes[0]);

  const createGroup = async function() {
    const md_record = {};
    md_record.id           = "";
    md_record.parent_id    = props.parentId;
    md_record.media_type   = mediaType;
    md_record.name_display = displayName;
    md_record.name_sort    = "";
    md_record.streams      = [];
    md_record.duration     = 0;
    md_record.size         = 0;
    const result = await api("metadata", "POST", md_record);
    if((result.status < 200) || (result.status > 299)) {
      props.setError(`Error creating metadata: ${(result.body && result.body.error) || result.status}`);
      return;
    }
    props.setError("");
    props.hide();
    props.refresh();
  };

  return html`
  <div style="z-index:5; position:fixed; top:0px;bottom:0px;left:0px;right:0px; background-color:#00000088; display:${props.visible ? "block" : "none" };" >
    <div style="display:inline-block; border:2px solid #000000; border-radius:12px; background-color:#FFFFFF; min-width:50vw; min-height:50vh;">
      <button onClick=${props.hide}>Close</button><br />
      <span><label>Display name: </label><input type="text" value=${displayName} onInput=${(event) => { setDisplayName(event.target.value); }} /></span><br />
      <span><label>Media type: </label><select name="media_type" value=${mediaType} onInput=${(event) => { setMediaType(event.target.value); }}>${props.allowedTypes.map((type) => { return html`<option value=${type}>${type}</option>`; })}</select></span><br />
      <button onClick=${createGroup}>Create</button>
    </div>
  </div>
  `;
}

// =============================================================================
// MetadataPathSelector

function treePathForId(tree, id, path=[]) {
  const listing = (path.length == 0) ? tree : tree.children;
  for(let idx=0; idx<listing.length; ++idx) {
    const entry = listing[idx];
    if(entry.id == id) {
      return [...path, entry.id];
    } else if(entry.children) {
      const result = treePathForId(entry, id, [...path, entry.id]);
      if(result) { return result; }
    }
  }
  return null;
}

function treeRecordForId(listing, id) {
  for(let idx=0; idx<listing.length; ++idx) {
    const entry = listing[idx];
    if(entry.id == id) { return entry; }
    if(entry.children) {
      const result = treeRecordForId(entry.children, id);
      if(result) { return result; }
    }
  }
  return null;
}

function MetadataPathElementSelector(props) {
  const includeEmptyOptions = useMemo(() => { return props.parentId != ""; }, [props.parentId]);

  const onInput = (event) => {
    if(event.target.value == "---") {
      props.onInput({ target: { value: props.parentId } });
    } else {
      props.onInput(event);
    }
  };

  return html`
    <select key=${props.value} value=${props.value} onInput=${onInput}>
      ${includeEmptyOptions && html`<option value="---">--</option>`}
      ${props.children.map((child) => { return html`<option key=${child.id} value=${child.id}>${child.name}</option>`; })}
    </select>
  `;
}

function MetadataPathSelector(props) {
  const [pathIds, setPathIds] = useState([ props.value ]);
  const [record, setRecord] = useState(null);
  const [groupCreatorVisible, setGroupCreatorVisible] = useState(false);

  useEffect(() => {
    const new_path_ids = treePathForId(props.tree, props.value);
    if(new_path_ids) {
      props.setError("");
      setPathIds(new_path_ids);
    } else {
      props.setError(`Error: could not find path (1) for ${props.value}`);
      setRecord(null);
      return;
    }
    const new_record = treeRecordForId(props.tree, props.value);
    setRecord(new_record);
  }, [props.value]);

  const pathSelectors = useMemo(() => {
    const selectors = [];

    let children = props.tree;
    for(let idx=0; idx<pathIds.length; ++idx) {
      const currentId = pathIds[idx];
      const parentId = (idx > 0) ? pathIds[idx-1] : '';
      selectors.push(html`<${MetadataPathElementSelector} key=${parentId} parentId=${parentId} value=${currentId} onInput=${props.onInput} children=${children}/>`);

      let next_children = null;
      for(let child_idx=0; child_idx<children.length; ++child_idx) {
        const child = children[child_idx];
        if(child.id == currentId) { next_children = child.children; break; }
      }
      if(!next_children) { props.setError(`Error: could not find path (2) for ${pathIds}`); return null; } else { props.setError(""); }
      children = next_children;
    }
    if(children.length > 0) {
      const parentId = children[0].parentId;
      selectors.push(html`<${MetadataPathElementSelector} key=${parentId} parentId=${parentId} value=${"---"} onInput=${props.onInput} children=${children}/>`);
    }

    return selectors;
  }, [pathIds]);

  const allowedTypes = useMemo(() => {
    if(!props.canCreateGroup) { return []; }
    if(!record) { return []; }
    if(record.id == "lost") { return []; }
    if(record.media_type == MetadataMediaType.FILE_VIDEO) { return []; }
    if(record.media_type == MetadataMediaType.FILE_AUDIO) { return []; }
    if(record.media_type == MetadataMediaType.MOVIE     ) { return []; }
    if(record.media_type == MetadataMediaType.ALBUM     ) { return []; }
    if(record.media_type == MetadataMediaType.SERIES    ) { return [MetadataMediaType.SERIES, MetadataMediaType.SEASON]; }
    if(record.media_type == MetadataMediaType.SEASON    ) { return []; }
    if(record.media_type == MetadataMediaType.MUSIC     ) { return [MetadataMediaType.ALBUM, MetadataMediaType.ARTIST]; }
    if(record.media_type == MetadataMediaType.ARTIST    ) { return [MetadataMediaType.ALBUM]; }
    if(record.media_type == MetadataMediaType.ALBUM     ) { return []; }
  }, [record, props.canCreateGroup]);

  const createGroup = () => { setGroupCreatorVisible(true); };
  const hideGroupCreator = () => { setGroupCreatorVisible(false); }

  return html`
    <div class="metadata-path-selector" style="display:flex; flex-direction:row;">
      ${props.tree && pathSelectors}
      ${props.canCreateGroup && (allowedTypes.length > 0) && html`<button onClick=${createGroup}>Create Group</button>`}
      ${props.canCreateGroup && (allowedTypes.length > 0) && html`
        <${MetadataGroupCreator}
          visible=${groupCreatorVisible}
          hide=${hideGroupCreator}
          refresh=${props.refresh}
          parentId=${props.value}
          allowedTypes=${allowedTypes}
          setError=${props.setError}
        />
      `}
    </div>
  `;
}

// =============================================================================
// MetadataParentListing

function MetadataParentListing(props) {
  const selectedLookup = useMemo(() => {
    const lookup = {};
    for(let idx=0; idx<props.selectedRecords.length; ++idx) {
      lookup[props.selectedRecords[idx].id] = true;
    }
    return lookup;
  }, [props.selectedRecords]);

  const selectAll  = () => {
    const new_selection = [];
    for(let idx=0; idx<props.records.length; ++idx) { new_selection.push(props.records[idx]); }
    props.setSelectedRecords(new_selection);
  };
  const selectNone = () => { props.setSelectedRecords([]); };
  const selectRecord = (event, recordId) => {
    const record = props.records.find((record) => { return record.id == recordId; });

    if(event.shiftKey == false) {
      props.setSelectedRecords([ record ]);
      return;
    }

    if(selectedLookup[recordId]) {
      props.setSelectedRecords(props.selectedRecords.filter((record) => { return record.id != recordId; }));
    } else {
      const new_selection = [];
      for(let idx=0; idx<props.selectedRecords.length; ++idx) { new_selection.push(props.selectedRecords[idx]); }
      new_selection.push(record);
      props.setSelectedRecords(new_selection);
    }
  };

  const recordElements = useMemo(() => {
    return props.records.map((record) => { return html`
      <span class="metadata-row" style="cursor:pointer; background-color:${!!selectedLookup[record.id] ? 'blue' : 'none'};" onClick=${(event) => { selectRecord(event, record.id); }}>
        <span class="metadata-row-name">${record.name_display}</span>
      </span>
    `; })
  }, [props.records, props.selectedRecords]);

  return html`
    <div class="metadata-parent-listing" style="display:flex; flex-direction:column;">
      <span>
        <button onClick=${selectAll}>Select All</button>
        <button onClick=${selectNone}>Select None</button>
        <span>(${props.selectedRecords.length} selected)</span>
      </span>
      ${props.records && recordElements}
    </div>
  `;
}

// =============================================================================
// MetadataProperties

function MetadataProperties(props) {
  const [parentId, setParentId] = useState("lost");
  const [mediaType, setMediaType] = useState("");
  const [nameDisplay, setNameDisplay] = useState("");
  const [nameSort, setNameSort] = useState("");
  const [duration, setDuration] = useState(0);
  const [size, setSize] = useState(0);

  const onParentChanged = (event) => { setParentId(event.target.value); };

  useEffect(() => {
    if(props.records.length < 1) { return; }

    const single = (props.records.length == 1);
    if(!single) {
      let current_parrent_id = props.records[0].parent_id;
      for(let idx=1; idx<props.records.length; ++idx) {
        if(props.records[idx].parent_id != current_parrent_id) {
          current_parrent_id = "lost";
          break;
        }
      }
      setParentId(current_parrent_id || "lost");
      setMediaType("");
      setNameDisplay("");
      setNameSort("");
      setDuration(0);
      setSize(0);
      return;
    }

    const record = props.records[0];
    setParentId(record.parent_id || "lost");
    setMediaType(record.media_type);
    setNameDisplay(record.name_display);
    setNameSort(record.name_sort);
    setDuration(record.duration);
    setSize(record.size);
  }, [props.records]);

  const doSave = async () => {
    if(props.records.length < 1) { return; }

    const api_parent_id = (parentId == "lost") ? "" : parentId;
    if(props.records.length == 1) {
      const result = await api(`metadata/${props.records[0].id}`, "POST", { parent_id:api_parent_id, name_display:nameDisplay, name_sort:nameSort });
      if((result.status < 200) || (result.status > 299)) {
        props.setError(`Error saving metadata: ${(result.body && result.body.error) || result.status}`);
        return;
      }
      props.setError("");
      props.refresh();
      return;
    }
    for(let idx=0; idx<props.records.length; ++idx) {
      const result = await api(`metadata/${props.records[idx].id}`, "POST", { parent_id:api_parent_id });
      if((result.status < 200) || (result.status > 299)) {
        props.setError(`Error saving metadata: ${(result.body && result.body.error) || result.status}`);
        return;
      }
    }
    props.setError("");
    props.refresh();
  };

  const doDelete = async () => {
    if(props.records.length < 1) { return; }
    if(confirm(`Delete ${props.records.length} record(s)?`) == false) { return; }

    for(let idx=0; idx<props.records.length; ++idx) {
      const result = await api(`metadata/${props.records[idx].id}`, "DELETE", { "delete_children":false });
      if((result.status < 200) || (result.status > 299)) {
        props.setError(`Error deleting metadata: ${(result.body && result.body.error) || result.status}`);
        return;
      }
    }
    props.setError("");
    props.refresh();
  };

  const onPosterDrag = (event) => {
    event.preventDefault();
  };
  const onPosterDrop = async (event) => {
    event.preventDefault();
    const files = event.dataTransfer.files;
    if(files.length < 1) { return; }
    const file = files[0];
    const result = await apiFile(`metadata/${props.records[0].id}/poster`, "POST", "poster", file);
    if((result.status < 200) || (result.status > 299)) {
      props.setError(`Error saving poster: ${(result.body && result.body.error) || result.status}`);
      return;
    }
    props.setError("");
    props.refresh();
  };

  if(props.records.length == 0) { return html`<div class="metadata-properties" style="display:flex; flex-direction:column;"></div>`; }

  if(props.records.length > 1) {
    return html`
      <div class="metadata-properties" style="display:flex; flex-direction:column;">
        <div><button onClick=${doSave}>Save Changes</button></div>
        <div><button onClick=${doDelete}>Delete Records</button></div>
        <table>
          <tr>
            <td>Parent</td>
            <td><${MetadataPathSelector} refresh=${props.refresh} tree=${props.tree} value=${parentId} onInput=${onParentChanged} setError=${props.setError}/></td>
          </tr>
        </table>
      </div>
    `;
  }

  return html`
    <div class="metadata-properties" style="display:flex; flex-direction:column;">
      <div><button onClick=${doSave}>Save Changes</button></div>
      <div><button onClick=${doDelete}>Delete Record</button></div>
      <table>
        <tr>
          <td>Parent</td>
          <td><${MetadataPathSelector} refresh=${props.refresh} tree=${props.tree} value=${parentId} onInput=${onParentChanged} setError=${props.setError} /></td>
        </tr>
        <tr>
          <td>Media Type</td>
          <td>${mediaType}</td>
        </tr>
        <tr>
          <td>Display Name</td>
          <td><input type="text" value=${nameDisplay} onInput=${(event) => { setNameDisplay(event.target.value); }} /></td>
        </tr>
        <tr>
          <td>Sort Name</td>
          <td><input type="text" value=${nameSort} onInput=${(event) => { setNameSort(event.target.value); }} /></td>
        </tr>
        <tr>
          <td>Duration</td>
          <td>${timeString(duration)}</td>
        </tr>
        <tr>
          <td>Size</td>
          <td>${sizeString(size)}</td>
        </tr>
        <tr>
          <td>Poster</td>
          <td>
            <img src="/poster/${props.records[0].id}/small?t=${new Date().getTime()}" onDrop=${onPosterDrop} onDragOver=${onPosterDrag} />
          </td>
        </tr>
      </table>
    </div>
  `;
}

// =============================================================================
// MetadataEditor

function MetadataEditor(props) {
  const [tree, setTree] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [parentId, setParentId] = useState('lost');
  const [parentRecords, setParentRecords] = useState([]);
  const [selectedRecords, setSelectedRecords] = useState([]);

  const refresh = async () => {
    setLoading(true);
    setError("");
    const treeResult = await api("metadata/tree", "GET");
    if((treeResult.status < 200) || (treeResult.status > 299)) {
      setError(`Error retrieving metadata tree: ${(treeResult.body && treeResult.body.error) || treeResult.status}`);
      setLoading(false);
      return;
    }
    setTree(treeResult.body);
    const parentResult = await api(`metadata/by-parent/${parentId}`, "GET");
    if((parentResult.status < 200) || (parentResult.status > 299)) {
      setError(`Error retrieving metadata listing: ${(parentResult.body && parentResult.body.error) || parentResult.status}`);
      setLoading(false);
      return;
    }
    setParentRecords(parentResult.body);
    setLoading(false);
  };
  useEffect(() => { refresh(); }, []);
  useEffect(() => { refresh(); }, [parentId]);

  const onParentChanged = (event) => {
    setSelectedRecords([]);
    setParentId(event.target.value);
  };

  if(loading) {
    return html`<div class="metadata-editor" style="display:flex; flex-direction:column;">Loading...</div>`;
  }

  return html`
    <div class="metadata-editor" style="display:flex; flex-direction:row;">
      <div style="display:flex; flex-direction:column;">
        <div><button onClick=${refresh}>Refresh</button></div>
        <span class="error">${error || ''}</span>
        <${MetadataPathSelector} refresh=${refresh} tree=${tree} value=${parentId} onInput=${onParentChanged} setError=${setError} canCreateGroup=${true} />
        <${MetadataParentListing} refresh=${refresh} records=${parentRecords} selectedRecords=${selectedRecords} setSelectedRecords=${setSelectedRecords} setError=${setError} />
      </div>
      <${MetadataProperties} refresh=${refresh} tree=${tree} records=${selectedRecords} setError=${setError}/>
    </div>
  `;
}

export const MetadataPathSelector = MetadataPathSelector;
export default MetadataEditor;
