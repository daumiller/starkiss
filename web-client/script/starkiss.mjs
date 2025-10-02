import { h, render } from "/web-client/vendor/preact.mjs";
import { useState, useMemo, useEffect } from "/web-client/vendor/preact-hooks.mjs";
import htm from "/web-client/vendor/htm.mjs"; const html = htm.bind(h);
import api from "/web-client/script/api.mjs";

// const SERVER_ROOT = "http://192.168.0.8:4331";
const SERVER_ROOT = "";

/* CATEGORY:
  "id", "name", "sort_index",
  "media_type": ["movie", "series", "music"]
*/
/* LISTING:
    "id", "name", "parent_id', "poster_ratio",
    "listing_type": [],
    // cateogiry listing_types: "invalid", "movies", "series", "artists"
    // metadata listing_types: "invalid", "seasons", "episodes", "albums", 'songs"
    "entry_count", "entries": []
  LISTING-ENTRY:
    "id", "name", "entry_type",
  Poster:  `${SERVER_ROOT}/poster/${id}/small`
*/

function Starkiss(props) {
  const [id,       setId]         = useState("");
  const [parents,  setParents]    = useState([]);
  const [childType, setChildType] = useState("");
  const [children, setChildren]   = useState([]);
  const [error,    setError]      = useState("");

  const onHashChange = function(_event) {
    if(!window.location.hash || (window.location.hash.length < 2)) { window.location.hash = "#home"; return; }
    const hash = window.location.hash.substring(1);
    setId(hash);
    // console.log(`Hash Change: ${hash}`);
  };

  // onLoad()
  useEffect(async () => {
    setError("");
    window.addEventListener("hashchange", onHashChange.bind(this));
    onHashChange();
  }, []);

  // onIdChange()
  useEffect(async () => {
    setError("");
    if(id.length < 1) { return; }

    if(id === "home") {
      const result = await api(`${SERVER_ROOT}/client/categories`, "GET");
      if((result.status < 200) || (result.status > 299)) {
        setError(`Error ${(result.body && result.body.error) || result.status} retrieving categories`);
        return;
      }
      setParents([ { id:"home", name:"Home" } ]);
      const categories = result.body;
      categories.sort((a, b) => (a.sortIndex - b.sortIndex));
      setChildType("categories");
      setChildren(categories);
    } else {
      const result = await api(`${SERVER_ROOT}/client/listing/${id}`, "GET");
      if((result.status < 200) || (result.status > 299)) {
        setError(`Error ${(result.body && result.body.error) || result.status} retrieving listing`);
        return;
      }
      const newParents = result.body.path;
      newParents.push({ id:"home", name:"Home" });
      newParents.reverse();
      setParents(newParents);
      setChildType(result.body.listing_type);
      setChildren(result.body.entries);
    }
  }, [id]);

  const Navigation = useMemo(() => {
    const components = parents.map((parent, index) => {
      return html`<a key="parent-${parent.id}" href="#${parent.id}">${parent.name}</a>`;
    });
    return html`<div id="navigation">${components}</div>`;
  }, [parents]);

  const Listing = useMemo(() => {
    if((childType == "categories") || (childType == "episodes")) {
      const components = children.map((child) => {
        const child_type = child.media_type || child.entry_type || "";
        let a_href = `#${child.id}`;
        if((child_type === "file-video") || (child_type === "file_audio")) { a_href = `${SERVER_ROOT}/media/${child.id}`; }
        return html`<a key="child-${child.id}" href="${a_href}">${child.name}</a>`;
      });
      return html`<div id="listing-text">${components}</div>`;
    }

    const components = children.map((child) => {
      const child_type = child.media_type || child.entry_type || "";
      let a_href = `#${child.id}`;
      if((child_type === "file-video") || (child_type === "file_audio")) { a_href = `${SERVER_ROOT}/media/${child.id}`; }
      let img_src = `${SERVER_ROOT}/poster/${child.id}/small`;
      if(child.hasOwnProperty("media_type")) { img_src = `${SERVER_ROOT}/web-client/image/poster-missing-small.png`; }
      return html`<a key="child-${child.id}" href="${a_href}"><div><img src="${img_src}" /><br />${child.name}</div></a>`;
    });
    return html`<div id="listing-poster">${components}</div>`;
  }, [children]);

  if(!!error) { return html`<span class="error">${error}</span>`; }

  return html`
    <div id="client-root">
      ${Navigation}
      ${Listing}
    </div>
  `;
}

render(html`<${Starkiss} />`, document.body);
