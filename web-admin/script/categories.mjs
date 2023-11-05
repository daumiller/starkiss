import { h } from "/web-admin/vendor/preact.mjs";
import { useState, useMemo, useEffect } from "/web-admin/vendor/preact-hooks.mjs";
import htm from "/web-admin/vendor/htm.mjs"; const html = htm.bind(h);
import api from "/web-admin/script/api.mjs";

// =============================================================================
// MediaTypeSelector

function MediaTypeSelector(props) {
  return html`
    <select name="media_type" value=${props.value} onInput=${props.onInput}>
      <option value=${"movie"} >Movie</option>
      <option value=${"series"}>Series</option>
      <option value=${"music"} >Music</option>
    </select>
  `;
}

// =============================================================================
// CategoryEntryCreator

function CategoryEntryCreator(props) {
  const [name, setName] = useState("");
  const [mediaType, setMediaType] = useState("movie");

  const createCategoryEntry = async () => {
    const result = await api("category", "POST", { name:name, media_type:mediaType });
    if((result.status < 200) || (result.status > 299)) {
      props.onError(`Error ${(result.body && result.body.error) || result.status} creating category ${name}`);
      return;
    }
    props.onCategoryCreated(result.body);
    setName("");
    setMediaType("movie");
  }

  return html`
    <tr>
      <td><input type="text" name="name" value=${name} onInput=${(event) => { setName(event.target.value); }} /></td>
      <td><${MediaTypeSelector} value=${mediaType} onInput=${(event) => { setMediaType(event.target.value); }} /></td>
      <td>
        <button onClick=${createCategoryEntry} disabled=${name.trim().length < 1}>Create</button>
      </td>
    </tr>
  `;
}

// =============================================================================
// CategoryEntryEditor

function CategoryEntryEditor(props) {
  const [name, setName] = useState(props.name);
  const [mediaType, setMediaType] = useState(props.mediaType);

  const updateCategoryEntry = async () => {
    const result = await api(`category/${props.catId}`, "POST", { name:name, media_type:mediaType });
    if((result.status < 200) || (result.status > 299)) {
      props.onError(`Error ${(result.body && result.body.error) || result.status} updating category`);
      return;
    }
    props.onCategoryUpdated(props.catId, name, mediaType);
  }

  const cancelUpdate = () => {
    setName(props.name);
    setMediaType(props.mediaType);
  }

  const deleteCategoryEntry = () => {
    if(confirm(`Delete category ${props.name}?`) == false) { return; }
    const result = api(`category/${props.catId}`, "DELETE");
    if((result.status < 200) || (result.status > 299)) {
      props.onError(`Error ${(result.body && result.body.error) || result.status} updating category ${props.name}`);
      return;
    }
    props.onCategoryDeleted(props.catId);
  };

  const isChanged = useMemo(() => {
    return (name != props.name) || (mediaType != props.mediaType);
  }, [name, mediaType, props.name, props.mediaType]);

  return html`
    <tr>
      <td><input type="text" name="name" value=${name} onInput=${(event) => { setName(event.target.value); }} /></td>
      <td><${MediaTypeSelector} value=${mediaType} onInput=${(event) => { setMediaType(event.target.value); }} /></td>
      <td>
        <button onClick=${updateCategoryEntry} disabled=${!isChanged}>Update</button>
        <button onClick=${cancelUpdate}        disabled=${!isChanged}>Cancel</button>
        <button onClick=${deleteCategoryEntry}>Delete</button>
      </td>
    </tr>
  `;
}

// =============================================================================
// CategoryEditor

function PropertyEditor(props) {
  const [categories, setCategories] = useState({});
  const [error, setError] = useState("");

  const onCategoryCreated = (category) => {
    setError("");
    setCategories({ ...categories, [category.id]:category });
  }

  const onCategoryUpdated = (catId, name, mediaType) => {
    setError("");
    setCategories({ ...categories, [catId]:{ id:catId, name:name, media_type:mediaType } });
  }

  const onCategoryDeleted = (catId) => {
    setError("");
    const newCategories = { ...categories };
    delete newCategories[catId];
    setCategories(newCategories);
  }

  const onError = (message) => {
    setError(message);
  }

  const categoryEditors = useMemo(() => {
    return Object.entries(categories).map(([key, value]) => {
      return html`
        <${CategoryEntryEditor}
          key=${key}
          catId=${key}
          name=${value.name}
          mediaType=${value.media_type}
          onCategoryUpdated=${onCategoryUpdated}
          onCategoryDeleted=${onCategoryDeleted}
          onError=${onError}
        />`;
    });
  }, [categories]);

  useEffect(async () => {
    const result = await api("categories", "GET")
    if((result.status < 200) || (result.status > 299)) {
      setError(`Error ${(result.body && result.body.error) || result.status} retrieving categories`);
      return;
    }
    const new_categories = {};
    for(let idx=0; idx<result.body.length; ++idx) {
      const category = result.body[idx];
      new_categories[category.id] = category;
    }
    setCategories(new_categories);
    setError("");
  }, []);

  return html`
    <div id="category-editor-root">
      <span class="error">${error}</span>
      <table>
        <thead><tr>
          <th>Name</th>
          <th>Media Type</th>
          <th></th>
        </tr></thead>
        <tbody>
          ${categoryEditors}
          <${CategoryEntryCreator} onCategoryCreated=${onCategoryCreated} onError=${onError} />
        </tbody>
      </table>
    </div>
  `;
}

export default PropertyEditor;
