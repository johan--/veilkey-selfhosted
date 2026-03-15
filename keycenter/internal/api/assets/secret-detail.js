function sEsc(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function sBadge(value) {
  const normalized = String(value || "").toLowerCase();
  let klass = "warn";
  if (normalized === "active" || normalized === "ok") klass = "ok";
  if (normalized === "blocked" || normalized === "error") klass = "err";
  return '<span class="badge ' + klass + '">' + sEsc(value || "unknown") + '</span>';
}

function sCards(rows) {
  return rows.map(([label, value]) =>
    '<div class="event"><div class="event-top"><strong>' + sEsc(label) + '</strong></div><div class="mono">' + value + '</div></div>'
  ).join("");
}

async function sFetchJSON(path) {
  const res = await fetch(path, {headers:{accept:"application/json"}});
  if (!res.ok) {
    const text = await res.text();
    throw new Error(path + " -> " + res.status + " " + text);
  }
  return res.json();
}

function secretRefFromPath() {
  const parts = location.pathname.split("/").filter(Boolean);
  return decodeURIComponent(parts.slice(2).join("/"));
}

function renderSecretBindings(rows) {
  const el = document.getElementById("secret-binding-table");
  if (!rows.length) {
    el.innerHTML = '<tr><td colspan="5"><div class="empty">No binding rows for this ref.</div></td></tr>';
    return;
  }
  el.innerHTML = rows.map((row) =>
    '<tr>' +
      '<td>' + sBadge(row.binding_type || "unknown") + '</td>' +
      '<td class="mono">' + sEsc(row.target_name || "") + '</td>' +
      '<td class="mono">' + sEsc(row.vault_hash || "") + '</td>' +
      '<td>' + sEsc(row.field_key || "-") + '</td>' +
      '<td>' + sEsc(row.updated_at || "-") + '</td>' +
    '</tr>'
  ).join("");
}

function renderSecretAudit(rows) {
  const el = document.getElementById("secret-audit-list");
  if (!rows.length) {
    el.innerHTML = '<div class="empty">No audit events for this secret.</div>';
    return;
  }
  el.innerHTML = rows.map((row) =>
    '<article class="event">' +
      '<div class="event-top">' +
        '<strong>' + sEsc(row.action || "event") + '</strong>' +
        sBadge(row.actor_type || "system") +
      '</div>' +
      '<div class="muted">' + sEsc(row.actor_id || "system") + ' · ' + sEsc(row.source || "runtime") + ' · ' + sEsc(row.created_at || "-") + '</div>' +
      '<div class="mono" style="margin-top:8px">' + sEsc(row.reason || "") + '</div>' +
    '</article>'
  ).join("");
}

async function loadSecretDetail() {
  const ref = secretRefFromPath();
  if (!ref) throw new Error("canonical ref route segment is required");

  document.getElementById("secret-title").textContent = "Secret Detail · " + ref;

  const [detail, bindings, audit] = await Promise.all([
    sFetchJSON("/api/catalog/secrets/" + encodeURIComponent(ref)),
    sFetchJSON("/api/catalog/bindings?ref_canonical=" + encodeURIComponent(ref) + "&limit=20"),
    sFetchJSON("/api/catalog/audit?entity_type=secret&entity_id=" + encodeURIComponent(ref) + "&limit=20"),
  ]);

  document.getElementById("secret-detail-chip").innerHTML = sBadge(detail.status || "active");
  document.getElementById("secret-summary").innerHTML = sCards([
    ["Canonical Ref", '<span class="mono">' + sEsc(detail.ref_canonical || ref) + '</span>'],
    ["Name", sEsc(detail.secret_name || detail.name || "")],
    ["Class", sBadge(detail.class || "secret")],
    ["Status", sBadge(detail.status || "active")],
    ["Binding Count", sEsc(detail.binding_count ?? 0)],
    ["Fields Present", '<span class="mono">' + sEsc((detail.fields_present_json || detail.fields_present || []).toString()) + '</span>'],
  ]);

  document.getElementById("secret-context").innerHTML = sCards([
    ["Vault Hash", '<a href="/ui/vaults/' + encodeURIComponent(detail.vault_hash || "") + '">' + sEsc(detail.vault_hash || "") + '</a>'],
    ["Vault Node UUID", '<span class="mono">' + sEsc(detail.vault_node_uuid || "") + '</span>'],
    ["Vault Runtime Hash", '<span class="mono">' + sEsc(detail.vault_runtime_hash || "") + '</span>'],
    ["Last Rotated", sEsc(detail.last_rotated_at || "never")],
    ["Last Revealed", sEsc(detail.last_revealed_at || "never")],
    ["Updated At", sEsc(detail.updated_at || "-")],
  ]);

  document.getElementById("secret-binding-chip").textContent = (bindings.count ?? bindings.rows?.length ?? 0) + " rows";
  document.getElementById("secret-audit-chip").textContent = (audit.count ?? audit.rows?.length ?? 0) + " rows";
  renderSecretBindings(bindings.rows || []);
  renderSecretAudit(audit.rows || []);
}

loadSecretDetail().catch((err) => {
  document.getElementById("secret-summary").innerHTML = '<div class="empty">' + sEsc(err.message) + '</div>';
  document.getElementById("secret-context").innerHTML = '<div class="empty">Secret detail failed to load.</div>';
  renderSecretBindings([]);
  renderSecretAudit([]);
});
