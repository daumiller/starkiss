import { h } from "/web-admin/vendor/preact.mjs";
import { useState, useMemo, useEffect } from "/web-admin/vendor/preact-hooks.mjs";
import htm from "/web-admin/vendor/htm.mjs"; const html = htm.bind(h);
import api from "/web-admin/script/api.mjs";

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
    <select value=${props.value} onInput=${onInput}>
      ${includeEmptyOptions && html`<option value="---">--</option>`}
      ${props.children.map((child) => { return html`<option value=${child.id}>${child.name}</option>`; })}
    </select>
  `;
}

function MetadataPathSelector(props) {
  const [pathIds, setPathIds] = useState([ props.value ]);
  const [tree, setTree] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const refresh = () => {
    setLoading(true);
    api("metadata/tree", "GET", false, true).then((data) => {
      setError("");
      setTree(data);
      setLoading(false);
    }).catch((error) => {
      setError(`Error \"${error}\" retrieving metadata tree`);
    });
  }
  useEffect(() => { refresh(); }, []);

  useEffect(() => {
    const new_path_ids = treePathForId(tree, props.value);
    if(new_path_ids) {
      setPathIds(new_path_ids);
    } else {
      setError(`Error: could not find path for ${props.value}`);
    }
  }, [props.value]);

  const pathSelectors = useMemo(() => {
    const selectors = [];

    let children = tree;
    for(let idx=0; idx<pathIds.length; ++idx) {
      const currentId = pathIds[idx];
      const parentId = (idx > 0) ? pathIds[idx-1] : '';
      selectors.push(html`<${MetadataPathElementSelector} parentId=${parentId} value=${currentId} onInput=${props.onInput} children=${children}/>`);

      let next_children = null;
      for(let child_idx=0; child_idx<children.length; ++child_idx) {
        const child = children[child_idx];
        if(child.id == currentId) { next_children = child.children; break; }
      }
      if(!next_children) { setError(`Error: could not find path for ${pathIds}`); return null; }
      children = next_children;
    }
    if(children.length > 0) {
      const parentId = children[0].parentId;
      selectors.push(html`<${MetadataPathElementSelector} parentId=${parentId} value=${"---"} onInput=${props.onInput} children=${children}/>`);
    }

    return selectors;
  }, [pathIds, tree]);

  return html`
    <div class="metadata-path-selector" style="display:flex; flex-direction:row;">
      <button onClick=${refresh}>Refresh</button>
      ${ error &&             html`<span>${error}</span>`}
      ${!error &&  loading && html`<span>Loading...</span>`}
      ${!error && !loading && pathSelectors}
    </div>
  `;
}

// =============================================================================
// MetadataParentListing

function MetadataParentListing(props) {
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(true);
  const [records, setRecords] = useState([]);
  const [selectedRecords, setSelectedRecords] = useState({});

  const refresh = () => {
    setLoading(true);
    setRecords([]);
    setSelectedRecords({});
    setError("");
    api(`metadata/by-parent/${props.parentId}`, "GET", false, true).then((data) => {
      setRecords(data);
      setLoading(false);
    }).catch((error) => {
      setError(`Error \"${error}\" retrieving metadata parents`);
    });
  };
  useEffect(() => { refresh(); }, []);
  useEffect(() => { refresh(); }, [props.parentId]);

  const selectAll  = () => {
    const new_selection = {};
    for(let idx=0; idx<records.length; ++idx) { new_selection[records[idx].id] = records[idx]; }
    setSelectedRecords(new_selection);
  };
  const selectNone = () => { setSelectedRecords({}); };
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
      <span class="metadata-row" style="cursor:pointer; background-color:${!!selectedRecords[record.id] ? 'blue' : 'none'};" onClick=${() => { selectRecord(record.id); }}>
        <span class="metadata-row-type">${record.media_type}</span>
        <span> | </span>
        <span class="metadata-row-name">${record.name_display}</span>
      </span>
    `; })
  }, [records, selectedRecords]);

  return html`
    <div class="metadata-parent-listing" style="display:flex; flex-direction:column;">
      <span>
        <button onClick=${refresh}>Refresh</button>
        <button onClick=${selectAll}>Select All</button>
        <button onClick=${selectNone}>Select None</button>
        <span>(${Object.keys(selectedRecords).length} selected)</span>
      </span>
      <br />
      ${ error &&             html`<span>${error}</span>`}  
      ${!error &&  loading && html`<span>Loading...</span>`}
      ${!error && !loading && recordElements}
    </div>
  `;
}

// =============================================================================
// MetadataEditor

function MetadataEditor(props) {
  const [parentId, setParentId] = useState('lost');
  const [selectedRecords, setSelectedRecords] = useState([]);

  const onParentChanged = (event) => {
    setSelectedRecords([]);
    setParentId(event.target.value);
  };
  const onSelectedRecordsChanged = (event) => {
    setSelectedRecords(event.target.value);
  };

  return html`
    <${MetadataPathSelector} value=${parentId} onInput=${onParentChanged} />
    <${MetadataParentListing} parentId=${parentId} value=${selectedRecords} onInput=${onSelectedRecordsChanged} />
  `;
}

export const MetadataPathSelector = MetadataPathSelector;
export default MetadataEditor;
