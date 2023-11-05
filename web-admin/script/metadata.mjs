import { h } from "/web-admin/vendor/preact.mjs";
import { useState, useMemo, useEffect } from "/web-admin/vendor/preact-hooks.mjs";
import htm from "/web-admin/vendor/htm.mjs"; const html = htm.bind(h);
import api, { timeString, sizeString } from "/web-admin/script/api.mjs";

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

  useEffect(() => {
    const new_path_ids = treePathForId(props.tree, props.value);
    if(new_path_ids) {
      setPathIds(new_path_ids);
    } else {
      props.setError(`Error: could not find path for ${props.value}`);
    }
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
      if(!next_children) { props.setError(`Error: could not find path for ${pathIds}`); return null; }
      children = next_children;
    }
    if(children.length > 0) {
      const parentId = children[0].parentId;
      selectors.push(html`<${MetadataPathElementSelector} key=${parentId} parentId=${parentId} value=${"---"} onInput=${props.onInput} children=${children}/>`);
    }

    return selectors;
  }, [pathIds, props.tree]);

  return html`
    <div class="metadata-path-selector" style="display:flex; flex-direction:row;">
      ${props.tree && pathSelectors}
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
      <br />
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

  const save = async () => {
    if(props.records.length < 1) { return; }

    const api_parent_id = (parentId == "lost") ? "" : parentId;
    if(props.records.length == 1) {
      const result = await api(`metadata/${props.records[0].id}`, "POST", { parent_id:api_parent_id, name_display:nameDisplay, name_sort:nameSort });
      if((result.status < 200) || (result.status > 299)) {
        props.setError(`Error saving metadata: ${(result.body && result.body.error) || result.status}`);
        return;
      }
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
  };

  if(props.records.length == 0) { return html`<div class="metadata-properties" style="display:flex; flex-direction:column;"></div>`; }

  if(props.records.length > 1) {
    return html`
      <div class="metadata-properties" style="display:flex; flex-direction:column;">
        <div><button onClick=${save}>Save Changes</button></div>
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
      <div><button onClick=${save}>Save Changes</button></div>
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
            <img src="/poster/${props.records[0].id}/small" />
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
        <${MetadataPathSelector} refresh=${refresh} tree=${tree} value=${parentId} onInput=${onParentChanged} setError=${setError} />
        <${MetadataParentListing} refresh=${refresh} records=${parentRecords} selectedRecords=${selectedRecords} setSelectedRecords=${setSelectedRecords} setError=${setError} />
      </div>
      <${MetadataProperties} refresh=${refresh} tree=${tree} records=${selectedRecords} setError=${setError}/>
    </div>
  `;
}

export const MetadataPathSelector = MetadataPathSelector;
export default MetadataEditor;
