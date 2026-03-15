const fmt = new Intl.DateTimeFormat(undefined, {dateStyle:"medium", timeStyle:"short"});

function esc(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function badgeClass(value) {
  const normalized = String(value || "").toLowerCase();
  if (normalized === "ok" || normalized === "complete" || normalized === "hkm" || normalized === "active") return "ok";
  if (normalized === "locked" || normalized === "blocked" || normalized === "error") return "err";
  return "warn";
}

function asBadge(value) {
  return '<span class="badge ' + badgeClass(value) + '">' + esc(value || "unknown") + '</span>';
}

function cardRows(rows) {
  return rows.map(([label, value]) =>
    '<div class="event"><div class="event-top"><strong>' + esc(label) + '</strong></div><div class="mono">' + (value || '<span class="muted">n/a</span>') + '</div></div>'
  ).join("");
}

async function fetchJSON(path) {
  const res = await fetch(path, {headers:{accept:"application/json"}});
  if (!res.ok) {
    const text = await res.text();
    throw new Error(path + " -> " + res.status + " " + text);
  }
  return res.json();
}

function setText(id, value) {
  const el = document.getElementById(id);
  if (el) el.textContent = value;
}

function renderEmpty(id, message, tag = "div") {
  const el = document.getElementById(id);
  if (!el) return;
  el.innerHTML = '<' + tag + ' class="empty">' + esc(message) + '</' + tag + '>';
}

function renderVaults(rows) {
  const el = document.getElementById("vault-table");
  if (!rows.length) {
    el.innerHTML = '<tr><td colspan="6"><div class="empty">No vault inventory rows available.</div></td></tr>';
    return;
  }
  el.innerHTML = rows.map((row) =>
    '<tr>' +
      '<td class="mono"><a href="/ui/vaults/' + encodeURIComponent(row.vault_hash || row.vault_runtime_hash || "") + '">' + esc(row.vault_hash || row.vault_runtime_hash || "") + '</a></td>' +
      '<td>' + asBadge(row.vault_type || "unknown") + '</td>' +
      '<td class="mono">' + esc(row.vault_node_uuid || row.node_id || "") + '</td>' +
      '<td class="mono">' + esc(row.managed_path || "-") + '</td>' +
      '<td>' + esc(row.ref_count ?? 0) + '</td>' +
      '<td>' + esc(row.config_count ?? 0) + '</td>' +
    '</tr>'
  ).join("");
}

function renderSecrets(rows, detail) {
  const list = document.getElementById("secret-table");
  if (!rows.length) {
    list.innerHTML = '<tr><td colspan="5"><div class="empty">No secret catalog rows available.</div></td></tr>';
  } else {
    list.innerHTML = rows.map((row) =>
      (() => {
        const ref = row.ref_canonical || row.ref || "";
        const detailPath = String(ref).startsWith("VE:") ? "/ui/configs/" : "/ui/secrets/";
        return (
      '<tr>' +
        '<td class="mono"><a href="' + detailPath + encodeURIComponent(ref) + '">' + esc(ref) + '</a></td>' +
        '<td>' + esc(row.name || "") + '</td>' +
        '<td>' + asBadge(row.class || "secret") + '</td>' +
        '<td>' + esc(row.binding_count ?? 0) + '</td>' +
        '<td>' + asBadge(row.status || "active") + '</td>' +
      '</tr>');
      })()
    ).join("");
  }

  const detailEl = document.getElementById("secret-detail");
  if (!detail) {
    detailEl.innerHTML = '<div class="empty">No secret detail available.</div>';
    return;
  }
  detailEl.innerHTML = cardRows([
    ["Canonical Ref", '<span class="mono">' + esc(detail.ref_canonical || detail.ref || "") + '</span>'],
    ["Name", esc(detail.name || "")],
    ["Class", asBadge(detail.class || "secret")],
    ["Vault Hash", '<span class="mono">' + esc(detail.vault_hash || "") + '</span>'],
    ["Bindings", esc(detail.binding_count ?? 0)],
    ["Last Revealed", esc(detail.last_revealed_at || "never")],
  ]);
}

function renderBindings(rows) {
  const el = document.getElementById("binding-table");
  if (!rows.length) {
    el.innerHTML = '<tr><td colspan="5"><div class="empty">No binding rows available.</div></td></tr>';
    return;
  }
  el.innerHTML = rows.map((row) =>
    '<tr>' +
      '<td>' + asBadge(row.binding_type || "unknown") + '</td>' +
      '<td class="mono">' + esc(row.target_name || "") + '</td>' +
      '<td class="mono">' + esc(row.vault_hash || "") + '</td>' +
      '<td class="mono">' + esc(row.ref_canonical || "") + '</td>' +
      '<td>' + esc(row.updated_at ? fmt.format(new Date(row.updated_at)) : "-") + '</td>' +
    '</tr>'
  ).join("");
}

function renderAudit(rows) {
  const el = document.getElementById("audit-list");
  if (!rows.length) {
    el.innerHTML = '<div class="empty">No audit events available.</div>';
    return;
  }
  el.innerHTML = rows.map((row) =>
    '<article class="event">' +
      '<div class="event-top">' +
        '<strong>' + esc(row.action || "event") + '</strong>' +
        asBadge(row.entity_type || "unknown") +
      '</div>' +
      '<div class="muted">' + esc(row.actor || "system") + ' · ' + esc(row.source || "runtime") + ' · ' + esc(row.created_at ? fmt.format(new Date(row.created_at)) : "-") + '</div>' +
      '<div class="mono" style="margin-top:8px">' + esc(row.entity_id || "") + '</div>' +
    '</article>'
  ).join("");
}

async function boot() {
  try {
    const [status, vaults, secrets, bindings, audit] = await Promise.all([
      fetchJSON("/api/status"),
      fetchJSON("/api/vault-inventory?limit=8"),
      fetchJSON("/api/catalog/secrets?limit=8"),
      fetchJSON("/api/catalog/bindings?limit=8"),
      fetchJSON("/api/catalog/audit?limit=8"),
    ]);

    const secretRows = secrets.rows || [];
    const detailRef = secretRows[0]?.ref_canonical || secretRows[0]?.ref;
    const secretDetail = detailRef ? await fetchJSON("/api/catalog/secrets/" + encodeURIComponent(detailRef)) : null;

    const install = status.install || {};
    setText("metric-install", install.complete ? "Complete" : "Pending");
    setText("metric-mode", String(status.mode || "unknown").toUpperCase());
    setText("metric-vaults", String(vaults.count ?? vaults.rows?.length ?? 0));
    setText("metric-secrets", String(secrets.count ?? secretRows.length));

    const statusChip = document.getElementById("status-chip");
    statusChip.innerHTML = asBadge(status.mode || "unknown") + ' ' + asBadge(install.complete ? "complete" : "pending");

    document.getElementById("state-summary").innerHTML = cardRows([
      ["Mode", asBadge(status.mode || "unknown")],
      ["Node ID", '<span class="mono">' + esc(status.node_id || "") + '</span>'],
      ["Vault Node UUID", '<span class="mono">' + esc(status.vault_node_uuid || "") + '</span>'],
      ["Locked", asBadge(status.locked ? "locked" : "ok")],
    ]);

    document.getElementById("state-install").innerHTML = cardRows([
      ["Install Complete", asBadge(install.complete ? "complete" : "pending")],
      ["Session ID", '<span class="mono">' + esc(install.session_id || "") + '</span>'],
      ["Last Stage", esc(install.last_stage || "-")],
      ["Final Stage", esc(install.final_stage || "-")],
    ]);

    setText("vault-chip", (vaults.count ?? vaults.rows?.length ?? 0) + " rows");
    setText("secret-chip", (secrets.count ?? secretRows.length ?? 0) + " rows");
    setText("binding-chip", (bindings.count ?? bindings.rows?.length ?? 0) + " rows");
    setText("audit-chip", (audit.count ?? audit.rows?.length ?? 0) + " rows");

    renderVaults(vaults.rows || []);
    renderSecrets(secretRows, secretDetail?.row || null);
    renderBindings(bindings.rows || []);
    renderAudit(audit.rows || []);
  } catch (err) {
    setText("status-chip", "Load failed");
    renderEmpty("state-summary", err.message);
    renderEmpty("state-install", "Operator shell could not load runtime state.");
    renderVaults([]);
    renderSecrets([], null);
    renderBindings([]);
    renderAudit([]);
  }
}

boot();
