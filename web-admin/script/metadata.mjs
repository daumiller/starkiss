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
  return html`
    <select value=${props.value} onInput=${props.onInput}>
      ${props.children.map((child) => { return html`<option value=${child.id}>${child.name}</option>`; })}
    </select>
  `;
}

function MetadataPathSelector(props) {
  const [pathIds, setPathIds] = useState([ props.value ]);
  const [tree, setTree] = useState([]);
  const [treeLoaded, setTreeLoaded] = useState(false);
  const [error, setError] = useState("");

  const refresh = () => {
    setTreeLoaded(false);
    api("metadata/tree", "GET", false, true).then((data) => {
      setError("");
      setTree(data);
      setTreeLoaded(true);
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
      selectors.push(html`<${MetadataPathElementSelector} value=${currentId} onInput=${props.onInput} children=${children}/>`);

      let next_children = null;
      for(let child_idx=0; child_idx<children.length; ++child_idx) {
        const child = children[child_idx];
        if(child.id == currentId) { next_children = child.children; break; }
      }
      if(!next_children) { setError(`Error: could not find path for ${pathIds}`); return null; }
      children = next_children;
    }

    return selectors;
  }, [pathIds, tree]);

  return html`
    <div class="metadata-path-selector" style="display:flex; flex-direction:row;">
      <button onClick=${refresh}>Refresh</button>
      ${ error && html`<span>${error}</span>`}
      ${!error && !treeLoaded && html`<span>Loading...</span>`}
      ${!error &&  treeLoaded && pathSelectors}
    </div>
  `;
}

// =============================================================================
// MetadataEditor

function MetadataEditor(props) {
  const [parentId, setParentId] = useState('lost');

  const onParentChanged = (event) => {
    setParentId(event.target.value);
  }

  return html`
    <${MetadataPathSelector} value=${parentId} onInput=${onParentChanged} />
  `;
}

export default MetadataEditor;
