const state = {
  categories: [],
  products: [],
  category: "all",
  search: "",
  sort: "newest",
  editing: null,
};

const elements = {
  totalProducts: document.querySelector("#totalProducts"),
  totalCategories: document.querySelector("#totalCategories"),
  averagePrice: document.querySelector("#averagePrice"),
  inStock: document.querySelector("#inStock"),
  navCount: document.querySelector("#navCount"),
  resultCount: document.querySelector("#resultCount"),
  productGrid: document.querySelector("#productGrid"),
  categoryTabs: document.querySelector("#categoryTabs"),
  scrapeCategory: document.querySelector("#scrapeCategory"),
  scrapePages: document.querySelector("#scrapePages"),
  scrapeButton: document.querySelector("#scrapeButton"),
  filterCategory: document.querySelector("#filterCategory"),
  sortSelect: document.querySelector("#sortSelect"),
  searchInput: document.querySelector("#searchInput"),
  dialog: document.querySelector("#productDialog"),
  form: document.querySelector("#productForm"),
  toastRegion: document.querySelector("#toastRegion"),
};

const categoryLabel = (id) => state.categories.find((category) => category.id === id)?.label ?? id;
const currency = new Intl.NumberFormat(undefined, { style: "currency", currency: "USD" });

function escapeHTML(value = "") {
  return String(value).replace(/[&<>"']/g, (character) => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
  })[character]);
}

function safeURL(value) {
  if (!value) return "";
  try {
    const url = new URL(value, window.location.origin);
    return ["https:", "http:"].includes(url.protocol) ? url.href : "";
  } catch {
    return "";
  }
}

async function request(path, options) {
  const response = await fetch(path, options);
  if (response.status === 204) return null;
  const body = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(body.error || `Request failed (${response.status})`);
  return body;
}

function showToast(message, error = false) {
  const toast = document.createElement("div");
  toast.className = `toast${error ? " error" : ""}`;
  toast.textContent = message;
  elements.toastRegion.append(toast);
  window.setTimeout(() => toast.remove(), 3800);
}

function populateCategoryControls() {
  const options = state.categories.map((category) =>
    `<option value="${escapeHTML(category.id)}">${escapeHTML(category.label)}</option>`,
  ).join("");
  elements.scrapeCategory.innerHTML = options;
  document.querySelector("#formCategory").innerHTML = options;
  elements.filterCategory.innerHTML = `<option value="all">All categories</option>${options}`;
  renderCategoryTabs();
}

function renderCategoryTabs() {
  elements.categoryTabs.innerHTML = [
    `<button class="category-tab${state.category === "all" ? " active" : ""}" role="tab" aria-selected="${state.category === "all"}" data-category="all">All products</button>`,
    ...state.categories.map((category) => `<button class="category-tab${state.category === category.id ? " active" : ""}" role="tab" aria-selected="${state.category === category.id}" data-category="${escapeHTML(category.id)}">${escapeHTML(category.label)}</button>`),
  ].join("");
}

function renderStats() {
  const count = state.products.length;
  const average = count ? state.products.reduce((total, item) => total + Number(item.price || 0), 0) / count : 0;
  const inStock = state.products.filter((item) => item.available).length;
  elements.totalProducts.textContent = count.toLocaleString();
  elements.totalCategories.textContent = new Set(state.products.map((item) => item.category)).size.toString();
  elements.averagePrice.textContent = currency.format(average);
  elements.inStock.textContent = inStock.toLocaleString();
  elements.navCount.textContent = count > 99 ? "99+" : count.toString();
}

function renderProducts(products) {
  elements.resultCount.textContent = products.length.toString();
  if (!products.length) {
    const hasFilters = Boolean(state.search) || state.category !== "all";
    elements.productGrid.innerHTML = `<div class="empty-state"><span class="empty-symbol"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="m4 7 8-4 8 4-8 4-8-4Z"/><path d="m4 12 8 4 8-4"/></svg></span><h3>${hasFilters ? "No matching products" : "Your catalog is ready"}</h3><p>${hasFilters ? "Try another search or category, or scrape a fresh collection." : "Scrape a category from the demo shop or add a product to get started."}</p>${hasFilters ? "" : '<button class="button button-primary" data-action="scrape-first">Browse products</button>'}</div>`;
    return;
  }

  elements.productGrid.innerHTML = products.map((item) => {
    const image = safeURL(item.image_url);
    const productLink = safeURL(item.product_url);
    const productName = escapeHTML(item.title);
    const rating = Number(item.rating || 0);
    const ratingMarkup = rating > 0
      ? `<span class="product-rating"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="m12 3 2.8 5.7 6.2.9-4.5 4.4 1.1 6.2-5.6-3-5.6 3 1.1-6.2L3 9.6l6.2-.9L12 3Z"/></svg>${rating.toFixed(1)}${item.reviews ? ` <span>(${Number(item.reviews).toLocaleString()})</span>` : ""}</span>`
      : '<span class="product-rating rating-empty">No rating yet</span>';
    return `<article class="product-card" data-id="${escapeHTML(item.id)}">
      <div class="product-visual">
        ${image ? `<img src="${escapeHTML(image)}" alt="${productName}" loading="lazy">` : '<span class="product-placeholder"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="m4 7 8-4 8 4-8 4-8-4Z"/><path d="m4 12 8 4 8-4"/></svg></span>'}
        <span class="availability${item.available ? "" : " out"}">${item.available ? "In stock" : "Out of stock"}</span>
        <div class="card-menu"><button class="card-action" data-action="edit" aria-label="Edit ${productName}" title="Edit product"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="m14 5 5 5M4 20l4.5-1 10-10a2.1 2.1 0 0 0-3-3l-10 10L4 20Z" stroke-linecap="round" stroke-linejoin="round"/></svg></button><button class="card-action delete" data-action="delete" aria-label="Delete ${productName}" title="Delete product"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 7h16M10 11v6m4-6v6M6 7l1 13h10l1-13M9 7V4h6v3" stroke-linecap="round" stroke-linejoin="round"/></svg></button></div>
      </div>
      <div class="product-info">
        <span class="product-category">${escapeHTML(categoryLabel(item.category))}</span>
        <h3 class="product-name">${productName}</h3>
        <p class="product-description">${escapeHTML(item.description || "No description provided.")}</p>
        <div class="product-meta"><strong class="product-price">${currency.format(Number(item.price || 0))}</strong>${ratingMarkup}</div>
        <div class="product-source"><span>${escapeHTML(item.source || "Manual entry")}</span>${productLink ? `<a href="${escapeHTML(productLink)}" target="_blank" rel="noopener noreferrer">View source ↗</a>` : ""}</div>
      </div>
    </article>`;
  }).join("");
  elements.productGrid.querySelectorAll(".product-visual img").forEach((image) => {
    image.addEventListener("error", () => {
      const placeholder = document.createElement("span");
      placeholder.className = "product-placeholder";
      placeholder.innerHTML = '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m4 7 8-4 8 4-8 4-8-4Z"/><path d="m4 12 8 4 8-4"/></svg>';
      image.replaceWith(placeholder);
    }, { once: true });
  });
}

function updateVisibleProducts() {
  const search = state.search.toLocaleLowerCase();
  const products = state.products.filter((item) => {
    const matchesCategory = state.category === "all" || item.category === state.category;
    const matchesSearch = !search || `${item.title} ${item.description}`.toLocaleLowerCase().includes(search);
    return matchesCategory && matchesSearch;
  });
  switch (state.sort) {
    case "price-asc": products.sort((a, b) => a.price - b.price); break;
    case "price-desc": products.sort((a, b) => b.price - a.price); break;
    case "rating": products.sort((a, b) => b.rating - a.rating); break;
    default: products.sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at));
  }
  renderProducts(products);
}

async function loadProducts() {
  state.products = await request("/api/products");
  renderStats();
  updateVisibleProducts();
}

function openDialog(item = null) {
  state.editing = item;
  elements.form.reset();
  document.querySelector("#dialogTitle").textContent = item ? "Edit product" : "Add product";
  document.querySelector("#saveProductButton").textContent = item ? "Save changes" : "Save product";
  document.querySelector("#formTitle").value = item?.title ?? "";
  document.querySelector("#formDescription").value = item?.description ?? "";
  document.querySelector("#formPrice").value = item?.price ?? "";
  document.querySelector("#formCategory").value = item?.category ?? state.categories[0]?.id ?? "";
  document.querySelector("#formImage").value = item?.image_url ?? "";
  document.querySelector("#formAvailable").checked = item?.available ?? true;
  elements.dialog.showModal();
}

async function refresh() {
  try {
    await loadProducts();
  } catch (error) {
    showToast(error.message, true);
  }
}

document.querySelector("#scrapeForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const pages = Number(elements.scrapePages.value);
  if (!Number.isInteger(pages) || pages < 1 || pages > 10) {
    showToast("Choose between 1 and 10 pages.", true);
    return;
  }
  elements.scrapeButton.disabled = true;
  elements.scrapeButton.innerHTML = '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 3v4m0 10v4M4.9 4.9l2.8 2.8m8.6 8.6 2.8 2.8M3 12h4m10 0h4M4.9 19.1l2.8-2.8m8.6-8.6 2.8-2.8" stroke-linecap="round"/></svg>Scraping…';
  try {
    const result = await request("/api/scrape", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ category: elements.scrapeCategory.value, pages }),
    });
    await loadProducts();
    showToast(`${result.count} product${result.count === 1 ? "" : "s"} added to your catalog.`);
  } catch (error) {
    showToast(error.message, true);
  } finally {
    elements.scrapeButton.disabled = false;
    elements.scrapeButton.innerHTML = '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="m12 3 1.9 5.8L20 11l-6.1 2.2L12 19l-1.9-5.8L4 11l6.1-2.2L12 3Z"/><path d="m19 14 .9 2.1L22 17l-2.1.9L19 20l-.9-2.1L16 17l2.1-.9L19 14Z"/></svg>Scrape products';
  }
});

document.querySelector("#addProductButton").addEventListener("click", () => openDialog());
document.querySelector("#closeDialog").addEventListener("click", () => elements.dialog.close());
document.querySelector("#cancelDialog").addEventListener("click", () => elements.dialog.close());
elements.dialog.addEventListener("click", (event) => {
  if (event.target === elements.dialog) elements.dialog.close();
});

elements.form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const item = {
    title: document.querySelector("#formTitle").value.trim(),
    description: document.querySelector("#formDescription").value.trim(),
    price: Number(document.querySelector("#formPrice").value),
    category: document.querySelector("#formCategory").value,
    image_url: document.querySelector("#formImage").value.trim(),
    available: document.querySelector("#formAvailable").checked,
  };
  const saveButton = document.querySelector("#saveProductButton");
  saveButton.disabled = true;
  try {
    await request(state.editing ? `/api/products/${encodeURIComponent(state.editing.id)}` : "/api/products", {
      method: state.editing ? "PUT" : "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(item),
    });
    elements.dialog.close();
    await loadProducts();
    showToast(state.editing ? "Product updated." : "Product added to your catalog.");
  } catch (error) {
    showToast(error.message, true);
  } finally {
    saveButton.disabled = false;
  }
});

elements.productGrid.addEventListener("click", async (event) => {
  const button = event.target.closest("[data-action]");
  if (!button) return;
  if (button.dataset.action === "scrape-first") {
    document.querySelector("#scraper").scrollIntoView({ behavior: "smooth" });
    elements.scrapeCategory.focus();
    return;
  }
  const card = button.closest(".product-card");
  const item = state.products.find((product) => product.id === card?.dataset.id);
  if (!item) return;
  if (button.dataset.action === "edit") openDialog(item);
  if (button.dataset.action === "delete" && window.confirm(`Remove "${item.title}" from your catalog?`)) {
    button.disabled = true;
    try {
      await request(`/api/products/${encodeURIComponent(item.id)}`, { method: "DELETE" });
      await loadProducts();
      showToast("Product removed from your catalog.");
    } catch (error) {
      showToast(error.message, true);
      button.disabled = false;
    }
  }
});

elements.categoryTabs.addEventListener("click", (event) => {
  const tab = event.target.closest("[data-category]");
  if (!tab) return;
  state.category = tab.dataset.category;
  elements.filterCategory.value = state.category;
  renderCategoryTabs();
  updateVisibleProducts();
});
elements.filterCategory.addEventListener("change", () => {
  state.category = elements.filterCategory.value;
  renderCategoryTabs();
  updateVisibleProducts();
});
elements.sortSelect.addEventListener("change", () => {
  state.sort = elements.sortSelect.value;
  updateVisibleProducts();
});
elements.searchInput.addEventListener("input", () => {
  state.search = elements.searchInput.value.trim();
  updateVisibleProducts();
});
document.querySelector("#refreshButton").addEventListener("click", refresh);

async function initialize() {
  try {
    state.categories = await request("/api/categories");
    populateCategoryControls();
    await loadProducts();
  } catch (error) {
    showToast(error.message, true);
    elements.productGrid.innerHTML = `<div class="empty-state"><h3>Could not load the catalog</h3><p>${escapeHTML(error.message)}</p><button class="button button-primary" data-action="reload">Try again</button></div>`;
    elements.productGrid.querySelector('[data-action="reload"]').addEventListener("click", initialize, { once: true });
  }
}

initialize();
