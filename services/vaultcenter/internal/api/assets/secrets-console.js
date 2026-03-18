function scEsc(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function scBadge(value) {
  const normalized = String(value || "").toLowerCase();
  let klass = "warn";
  if (normalized === "active" || normalized === "ok" || normalized === "secret") klass = "ok";
  if (normalized === "blocked" || normalized === "error") klass = "err";
  return '<span class="badge ' + klass + '">' + scEsc(value || "unknown") + '</span>';
}

function scCards(rows) {
  return rows.map(([label, value]) =>
    '<div class="event"><div class="event-top"><strong>' + scEsc(label) + '</strong></div><div class="mono">' + value + '</div></div>'
  ).join("");
}

async function scFetchJSON(path) {
  const res = await fetch(path, {headers: {accept: "application/json"}});
  if (!res.ok) {
    const text = await res.text();
    throw new Error(path + " -> " + res.status + " " + text);
  }
  return res.json();
}

function renderSecretCatalog(rows) {
  const el = document.getElementById("secrets-table");
  if (!rows.length) {
    el.innerHTML = '<tr><td colspan="6"><div class="empty">No canonical secret rows available.</div></td></tr>';
    return;
  }
  el.innerHTML = rows.map((row) =>
    '<tr>' +
      '<td class="mono"><a href="/ui/secrets/' + encodeURIComponent(row.ref_canonical || row.ref || "") + '">' + scEsc(row.ref_canonical || row.ref || "") + '</a></td>' +
      '<td>' + scEsc(row.secret_name || row.name || "") + '</td>' +
      '<td>' + scBadge(row.class || "secret") + '</td>' +
      '<td class="mono"><a href="/ui/vaults/' + encodeURIComponent(row.vault_hash || "") + '">' + scEsc(row.vault_hash || "") + '</a></td>' +
      '<td>' + scEsc(row.binding_count ?? 0) + '</td>' +
      '<td>' + scBadge(row.status || "active") + '</td>' +
    '</tr>'
  ).join("");
}

async function loadSecretsPage() {
  const list = await scFetchJSON("/api/catalog/secrets?class=key&limit=50");
  const rows = list.secrets || [];
  document.getElementById("secrets-chip").textContent = (list.count ?? rows.length) + " rows";
  renderSecretCatalog(rows);

  const preview = rows[0];
  if (!preview) {
    document.getElementById("secrets-preview-chip").textContent = "Empty";
    document.getElementById("secrets-preview-summary").innerHTML = '<div class="empty">No secret preview available.</div>';
    document.getElementById("secrets-preview-context").innerHTML = '<div class="empty">No vault context available.</div>';
    return;
  }

  document.getElementById("secrets-preview-chip").innerHTML = scBadge(preview.status || "active");
  document.getElementById("secrets-preview-summary").innerHTML = scCards([
    ["Canonical Ref", '<a href="/ui/secrets/' + encodeURIComponent(preview.ref_canonical || preview.ref || "") + '">' + scEsc(preview.ref_canonical || preview.ref || "") + '</a>'],
    ["Name", scEsc(preview.secret_name || preview.name || "")],
    ["Class", scBadge(preview.class || "secret")],
    ["Status", scBadge(preview.status || "active")],
    ["Bindings", scEsc(preview.binding_count ?? 0)],
  ]);
  document.getElementById("secrets-preview-context").innerHTML = scCards([
    ["Vault Hash", '<a href="/ui/vaults/' + encodeURIComponent(preview.vault_hash || "") + '">' + scEsc(preview.vault_hash || "") + '</a>'],
    ["Vault Node UUID", '<span class="mono">' + scEsc(preview.vault_node_uuid || "") + '</span>'],
    ["Vault Runtime Hash", '<span class="mono">' + scEsc(preview.vault_runtime_hash || "") + '</span>'],
    ["Last Rotated", scEsc(preview.last_rotated_at || "never")],
    ["Last Revealed", scEsc(preview.last_revealed_at || "never")],
  ]);
}

loadSecretsPage().catch((err) => {
  document.getElementById("secrets-chip").textContent = "Load failed";
  document.getElementById("secrets-table").innerHTML = '<tr><td colspan="6"><div class="empty">' + scEsc(err.message) + '</div></td></tr>';
  document.getElementById("secrets-preview-summary").innerHTML = '<div class="empty">Secret preview failed to load.</div>';
  document.getElementById("secrets-preview-context").innerHTML = '<div class="empty">Vault context failed to load.</div>';
});
