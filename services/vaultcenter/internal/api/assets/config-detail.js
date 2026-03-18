function cEsc(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function cBadge(value) {
  const normalized = String(value || "").toLowerCase();
  let klass = "warn";
  if (normalized === "active" || normalized === "ok") klass = "ok";
  if (normalized === "blocked" || normalized === "error") klass = "err";
  return '<span class="badge ' + klass + '">' + cEsc(value || "unknown") + '</span>';
}

function cCards(rows) {
  return rows.map(([label, value]) =>
    '<div class="event"><div class="event-top"><strong>' + cEsc(label) + '</strong></div><div class="mono">' + value + '</div></div>'
  ).join("");
}

async function cFetchJSON(path) {
  const res = await fetch(path, {headers:{accept:"application/json"}});
  if (!res.ok) {
    const text = await res.text();
    throw new Error(path + " -> " + res.status + " " + text);
  }
  return res.json();
}

function configRefFromPath() {
  const parts = location.pathname.split("/").filter(Boolean);
  return decodeURIComponent(parts.slice(2).join("/"));
}

function renderConfigBindings(rows) {
  const el = document.getElementById("config-binding-table");
  if (!rows.length) {
    el.innerHTML = '<tr><td colspan="5"><div class="empty">No binding rows for this config ref.</div></td></tr>';
    return;
  }
  el.innerHTML = rows.map((row) =>
    '<tr>' +
      '<td>' + cBadge(row.binding_type || "unknown") + '</td>' +
      '<td class="mono">' + cEsc(row.target_name || "") + '</td>' +
      '<td class="mono">' + cEsc(row.vault_hash || "") + '</td>' +
      '<td>' + cEsc(row.field_key || "-") + '</td>' +
      '<td>' + cEsc(row.updated_at || "-") + '</td>' +
    '</tr>'
  ).join("");
}

function renderConfigAudit(rows) {
  const el = document.getElementById("config-audit-list");
  if (!rows.length) {
    el.innerHTML = '<div class="empty">No audit events for this config.</div>';
    return;
  }
  el.innerHTML = rows.map((row) =>
    '<article class="event">' +
      '<div class="event-top">' +
        '<strong>' + cEsc(row.action || "event") + '</strong>' +
        cBadge(row.actor_type || "system") +
      '</div>' +
      '<div class="muted">' + cEsc(row.actor_id || "system") + ' · ' + cEsc(row.source || "runtime") + ' · ' + cEsc(row.created_at || "-") + '</div>' +
      '<div class="mono" style="margin-top:8px">' + cEsc(row.reason || "") + '</div>' +
    '</article>'
  ).join("");
}

async function loadConfigDetail() {
  const ref = configRefFromPath();
  if (!ref) throw new Error("canonical config ref route segment is required");

  document.getElementById("config-title").textContent = "Config Detail · " + ref;

  const [detail, bindings, audit] = await Promise.all([
    cFetchJSON("/api/catalog/secrets/" + encodeURIComponent(ref)),
    cFetchJSON("/api/catalog/bindings?ref_canonical=" + encodeURIComponent(ref) + "&limit=20"),
    cFetchJSON("/api/catalog/audit?entity_type=config&entity_id=" + encodeURIComponent(ref) + "&limit=20"),
  ]);

  document.getElementById("config-detail-chip").innerHTML = cBadge(detail.status || "active");
  document.getElementById("config-summary").innerHTML = cCards([
    ["Canonical Ref", '<span class="mono">' + cEsc(detail.ref_canonical || ref) + '</span>'],
    ["Key", cEsc(detail.secret_name || detail.name || "")],
    ["Class", cBadge(detail.class || "config")],
    ["Status", cBadge(detail.status || "active")],
    ["Binding Count", cEsc(detail.binding_count ?? 0)],
    ["Fields Present", '<span class="mono">' + cEsc((detail.fields_present_json || detail.fields_present || []).toString()) + '</span>'],
  ]);

  document.getElementById("config-context").innerHTML = cCards([
    ["Vault Hash", '<a href="/ui/vaults/' + encodeURIComponent(detail.vault_hash || "") + '">' + cEsc(detail.vault_hash || "") + '</a>'],
    ["Vault Node UUID", '<span class="mono">' + cEsc(detail.vault_node_uuid || "") + '</span>'],
    ["Vault Runtime Hash", '<span class="mono">' + cEsc(detail.vault_runtime_hash || "") + '</span>'],
    ["Last Rotated", cEsc(detail.last_rotated_at || "never")],
    ["Last Revealed", cEsc(detail.last_revealed_at || "never")],
    ["Updated At", cEsc(detail.updated_at || "-")],
  ]);

  document.getElementById("config-binding-chip").textContent = (bindings.count ?? bindings.rows?.length ?? 0) + " rows";
  document.getElementById("config-audit-chip").textContent = (audit.count ?? audit.rows?.length ?? 0) + " rows";
  renderConfigBindings(bindings.rows || []);
  renderConfigAudit(audit.rows || []);
}

loadConfigDetail().catch((err) => {
  document.getElementById("config-summary").innerHTML = '<div class="empty">' + cEsc(err.message) + '</div>';
  document.getElementById("config-context").innerHTML = '<div class="empty">Config detail failed to load.</div>';
  renderConfigBindings([]);
  renderConfigAudit([]);
});
