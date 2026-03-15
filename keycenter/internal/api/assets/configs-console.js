function ccEsc(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function ccBadge(value) {
  const normalized = String(value || "").toLowerCase();
  let klass = "warn";
  if (normalized === "active" || normalized === "ok" || normalized === "config") klass = "ok";
  if (normalized === "blocked" || normalized === "error") klass = "err";
  return '<span class="badge ' + klass + '">' + ccEsc(value || "unknown") + '</span>';
}

function ccCards(rows) {
  return rows.map(([label, value]) =>
    '<div class="event"><div class="event-top"><strong>' + ccEsc(label) + '</strong></div><div class="mono">' + value + '</div></div>'
  ).join("");
}

async function ccFetchJSON(path) {
  const res = await fetch(path, {headers: {accept: "application/json"}});
  if (!res.ok) {
    const text = await res.text();
    throw new Error(path + " -> " + res.status + " " + text);
  }
  return res.json();
}

function renderConfigCatalog(rows) {
  const el = document.getElementById("configs-table");
  if (!rows.length) {
    el.innerHTML = '<tr><td colspan="6"><div class="empty">No canonical config rows available.</div></td></tr>';
    return;
  }
  el.innerHTML = rows.map((row) =>
    '<tr>' +
      '<td class="mono"><a href="/ui/configs/' + encodeURIComponent(row.ref_canonical || row.ref || "") + '">' + ccEsc(row.ref_canonical || row.ref || "") + '</a></td>' +
      '<td>' + ccEsc(row.secret_name || row.name || "") + '</td>' +
      '<td>' + ccBadge(row.class || "config") + '</td>' +
      '<td class="mono"><a href="/ui/vaults/' + encodeURIComponent(row.vault_hash || "") + '">' + ccEsc(row.vault_hash || "") + '</a></td>' +
      '<td>' + ccEsc(row.binding_count ?? 0) + '</td>' +
      '<td>' + ccBadge(row.status || "active") + '</td>' +
    '</tr>'
  ).join("");
}

async function loadConfigsPage() {
  const list = await ccFetchJSON("/api/catalog/secrets?class=config&limit=50");
  const rows = list.secrets || [];
  document.getElementById("configs-chip").textContent = (list.count ?? rows.length) + " rows";
  renderConfigCatalog(rows);

  const preview = rows[0];
  if (!preview) {
    document.getElementById("configs-preview-chip").textContent = "Empty";
    document.getElementById("configs-preview-summary").innerHTML = '<div class="empty">No config preview available.</div>';
    document.getElementById("configs-preview-context").innerHTML = '<div class="empty">No vault context available.</div>';
    return;
  }

  document.getElementById("configs-preview-chip").innerHTML = ccBadge(preview.status || "active");
  document.getElementById("configs-preview-summary").innerHTML = ccCards([
    ["Canonical Ref", '<a href="/ui/configs/' + encodeURIComponent(preview.ref_canonical || preview.ref || "") + '">' + ccEsc(preview.ref_canonical || preview.ref || "") + '</a>'],
    ["Key", ccEsc(preview.secret_name || preview.name || "")],
    ["Class", ccBadge(preview.class || "config")],
    ["Status", ccBadge(preview.status || "active")],
    ["Bindings", ccEsc(preview.binding_count ?? 0)],
  ]);
  document.getElementById("configs-preview-context").innerHTML = ccCards([
    ["Vault Hash", '<a href="/ui/vaults/' + encodeURIComponent(preview.vault_hash || "") + '">' + ccEsc(preview.vault_hash || "") + '</a>'],
    ["Vault Node UUID", '<span class="mono">' + ccEsc(preview.vault_node_uuid || "") + '</span>'],
    ["Vault Runtime Hash", '<span class="mono">' + ccEsc(preview.vault_runtime_hash || "") + '</span>'],
    ["Last Rotated", ccEsc(preview.last_rotated_at || "never")],
    ["Last Revealed", ccEsc(preview.last_revealed_at || "never")],
  ]);
}

loadConfigsPage().catch((err) => {
  document.getElementById("configs-chip").textContent = "Load failed";
  document.getElementById("configs-table").innerHTML = '<tr><td colspan="6"><div class="empty">' + ccEsc(err.message) + '</div></td></tr>';
  document.getElementById("configs-preview-summary").innerHTML = '<div class="empty">Config preview failed to load.</div>';
  document.getElementById("configs-preview-context").innerHTML = '<div class="empty">Vault context failed to load.</div>';
});
