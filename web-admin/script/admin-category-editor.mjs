import { h, render } from "/web/vendor/preact.mjs";
import { useState, useMemo } from "/web/vendor/preact-hooks.mjs";
import htm from "/web/vendor/htm.mjs"; const html = htm.bind(h);

function CategoryEditor(props) {
  const [categories, setCategories] = useState([]);
  const [selectedCategory, setSelectedCategory] = useState(null);
  const [formFields, setFormFields] = useState({ name:"", type:"" });
  const createFormHandler = (fieldName) => {
    return (event) => { setFormFields({ ...formFields, [fieldName]:event.target.value }); };
  };

  const categoryOptions = useMemo(() => {
    return categories.map((category) => {
      return html`<option value=${category.id}>${category.name}</option>`;
    });
  }, [categories]);

  const categorySelected = (event) => {
    setSelectedCategory(event.target.value);
    const cat_selected = categories.find((category) => { return category.id == event.target.value; });
    setFormFields(cat_selected || { name:"", type:"" });
  }

  const listCategories = () => {
    fetch("/admin/categories", { method: "GET" }).then((response) => {
      return response.json();
    }).then((data) => {
      setCategories(data);
    }).catch((error) => {
      console.error(error);
    });
  }

  const updateCategory = () => {
    fetch("/admin/category/" + selectedCategory, { method: "POST", headers:{ "Content-Type": "application/json"}, body: JSON.stringify(formFields) }).catch((error) => {
      console.error(error);
    });
  }

  const createCategory = () => {
    fetch("/admin/category", { method: "POST", headers:{ "Content-Type": "application/json"}, body: JSON.stringify(formFields) }).then((response) => {
      return response.json();
    }).then((data) => {
      setFormFields(data);
    }).catch((error) => {
      console.error(error);
    });
  }

  const deleteCategory = () => {
    fetch("/admin/category/" + selectedCategory, { method: "DELETE" }).catch((error) => {
      console.error(error);
    });
  }

  return html`
    <div id="category-editor-root">
      <Button onClick=${listCategories}>List</Button>
      <Select value=${selectedCategory} onChange=${categorySelected}>${categoryOptions}</Select><br />
      <Label for="category-name">Name</Label><Input type="text" id="category-name" value=${formFields.name} onChange=${createFormHandler("name")} /><br />
      <Label for="category-type">Type</Label><Input type="text" id="category-type" value=${formFields.type} onChange=${createFormHandler("type")} /><br />
      <Button onClick=${createCategory}>Create</Button>
      <Button onClick=${updateCategory}>Update</Button>
      <Button onClick=${deleteCategory}>Delete</Button>
    </div>
  `;
}

export default CategoryEditor;
