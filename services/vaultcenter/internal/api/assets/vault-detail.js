const vaultFmt = new Intl.DateTimeFormat(undefined, {dateStyle:"medium", timeStyle:"short"});

function vEsc(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function vBadge(value) {
  const normalized = String(value || "").toLowerCase();
  let klass = "warn";
  if (normalized === "ok" || normalized === "active") klass = "ok";
  if (normalized === "blocked" || normalized === "error") klass = "err";
  return '<span class="badge ' + klass + '">' + vEsc(value || "unknown") + '</span>';
}

function vCardRows(rows) {
  return rows.map(([label, value]) =>
    '<div class="event"><div class="event-top"><strong>' + vEsc(label) + '</strong></div><div class="mono">' + value + '</div></div>'
  ).join("");
}

async function vFetchJSON(path, options = {}) {
  const res = await fetch(path, {
    headers: {accept: "application/json", ...(options.headers || {})},
    ...options,
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(path + " -> " + res.status + " " + text);
  }
  return res.json();
}

function vaultHashFromPath() {
  const parts = location.pathname.split("/").filter(Boolean);
  return decodeURIComponent(parts[parts.length - 1] || "");
}

function setVaultText(id, value) {
  const el = document.getElementById(id);
  if (el) el.textContent = value;
}

function renderVaultSecrets(rows) {
  const el = document.getElementById("vault-secret-table");
  if (!rows.length) {
    el.innerHTML = '<tr><td colspan="5"><div class="empty">No secret catalog rows for this vault.</div></td></tr>';
    return;
  }
  el.innerHTML = rows.map((row) =>
    (() => {
      const ref = row.ref_canonical || row.ref || "";
      const detailPath = String(ref).startsWith("VE:") ? "/ui/configs/" : "/ui/secrets/";
      return (
    '<tr>' +
      '<td class="mono"><a href="' + detailPath + encodeURIComponent(ref) + '">' + vEsc(ref) + '</a></td>' +
      '<td>' + vEsc(row.name || "") + '</td>' +
      '<td>' + vBadge(row.class || "secret") + '</td>' +
      '<td>' + vBadge(row.status || "active") + '</td>' +
      '<td>' + vEsc(row.binding_count ?? 0) + '</td>' +
    '</tr>');
    })()
  ).join("");
}

async function loadVaultDetail() {
  const vaultHash = vaultHashFromPath();
  if (!vaultHash) throw new Error("vault_hash route segment is required");

  setVaultText("vault-title", "Vault Detail · " + vaultHash);

  const [inventory, agents, secrets] = await Promise.all([
    vFetchJSON("/api/vault-inventory?vault_hash=" + encodeURIComponent(vaultHash) + "&limit=1"),
    vFetchJSON("/api/agents"),
    vFetchJSON("/api/catalog/secrets?vault_hash=" + encodeURIComponent(vaultHash) + "&limit=20"),
  ]);

  const row = (inventory.vaults || [])[0];
  if (!row) throw new Error("vault not found for " + vaultHash);
  const agent = (agents.agents || []).find((item) => item.vault_hash === vaultHash) || null;

  document.getElementById("vault-overview").innerHTML = vCardRows([
    ["Vault Hash", '<span class="mono">' + vEsc(row.vault_hash || row.vault_runtime_hash || vaultHash) + '</span>'],
    ["Vault Type", vBadge(row.vault_type || "unknown")],
    ["Vault Node UUID", '<span class="mono">' + vEsc(row.vault_node_uuid || row.node_id || "") + '</span>'],
    ["Managed Path", '<span class="mono">' + vEsc(row.managed_path || "-") + '</span>'],
    ["Ref Count", vEsc(row.ref_count ?? 0)],
    ["Config Count", vEsc(row.config_count ?? 0)],
  ]);

  document.getElementById("vault-agent").innerHTML = agent ? vCardRows([
    ["Runtime Hash", '<span class="mono">' + vEsc(agent.vault_runtime_hash || "") + '</span>'],
    ["Status", vBadge(agent.status || "unknown")],
    ["Key Version", vEsc(agent.key_version ?? "")],
    ["Rotation Required", vBadge(agent.rotation_required ? "active" : "clear")],
    ["Rebind Required", vBadge(agent.rebind_required ? "active" : "clear")],
    ["Last Seen", vEsc(agent.last_seen || "-")],
  ]) : '<div class="empty">No live agent row for this vault hash.</div>';

  document.getElementById("vault-detail-chip").innerHTML = vBadge(agent?.status || "detail");
  setVaultText("vault-secret-chip", (secrets.count ?? secrets.secrets?.length ?? 0) + " rows");
  renderVaultSecrets(secrets.secrets || []);

  const approvalLink = document.getElementById("vault-approval-link");
  if (approvalLink && agent?.vault_runtime_hash) {
    approvalLink.href = "/ui/approvals?agent=" + encodeURIComponent(agent.vault_runtime_hash);
  }
}

document.getElementById("rotate-all-btn")?.addEventListener("click", async () => {
  const el = document.getElementById("rotate-result");
  try {
    const result = await vFetchJSON("/api/agents/rotate-all", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: "{}",
    });
    el.textContent = JSON.stringify({status: result.status, count: result.count}, null, 2);
    el.className = "result show ok";
  } catch (err) {
    el.textContent = err.message;
    el.className = "result show err";
  }
});

loadVaultDetail().catch((err) => {
  document.getElementById("vault-overview").innerHTML = '<div class="empty">' + vEsc(err.message) + '</div>';
  document.getElementById("vault-agent").innerHTML = '<div class="empty">Vault detail failed to load.</div>';
  renderVaultSecrets([]);
});
