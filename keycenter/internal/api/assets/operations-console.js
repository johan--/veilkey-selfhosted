function oEsc(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function oBadge(value) {
  const normalized = String(value || "").toLowerCase();
  let klass = "warn";
  if (normalized === "ok" || normalized === "scheduled" || normalized === "preview" || normalized === "applied") klass = "ok";
  if (normalized === "blocked" || normalized === "error") klass = "err";
  return '<span class="badge ' + klass + '">' + oEsc(value || "unknown") + '</span>';
}

function oCards(rows) {
  return rows.map(([label, value]) =>
    '<div class="event"><div class="event-top"><strong>' + oEsc(label) + '</strong></div><div class="mono">' + value + '</div></div>'
  ).join("");
}

async function oFetchJSON(path, options = {}) {
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

function oResult(id, value, ok) {
  const el = document.getElementById(id);
  el.textContent = typeof value === "string" ? value : JSON.stringify(value, null, 2);
  el.className = "result show " + (ok ? "ok" : "err");
}

function oResultHTML(id, html, ok) {
  const el = document.getElementById(id);
  el.innerHTML = html;
  el.className = "result show " + (ok ? "ok" : "err");
}

function oChecked(id) {
  return Boolean(document.getElementById(id)?.checked);
}

function oRefLink(ref) {
  const value = String(ref || "");
  if (!value) return '<span class="muted">n/a</span>';
  const path = value.startsWith("VE:") ? "/ui/configs/" : "/ui/secrets/";
  return '<a href="' + path + encodeURIComponent(value) + '">' + oEsc(value) + '</a>';
}

function oApprovalLink(agentHash) {
  const value = String(agentHash || "");
  if (!value) return '<span class="muted">n/a</span>';
  return '<a href="/ui/approvals?agent=' + encodeURIComponent(value) + '">' + oEsc(value) + '</a>';
}

function oVaultLink(vaultHash) {
  const value = String(vaultHash || "");
  if (!value) return '<span class="muted">n/a</span>';
  return '<a href="/ui/vaults/' + encodeURIComponent(value) + '">' + oEsc(value) + '</a>';
}

function renderRotateResult(result) {
  const agents = result.agents || [];
  const summary = oCards([
    ["Status", oBadge(result.status || "scheduled")],
    ["Scheduled Agents", oEsc(result.count ?? agents.length)],
  ]);
  const rows = agents.length
    ? agents.slice(0, 10).map((agent, index) =>
        '<div class="event">' +
          '<div class="event-top"><strong>Agent ' + oEsc(index + 1) + '</strong>' + oBadge(agent.rotation_required ? "scheduled" : "clear") + '</div>' +
          '<div class="mono">vault: ' + oVaultLink(agent.vault_id || agent.vault_runtime_hash || "") + '</div>' +
          '<div class="mono">approval: ' + oApprovalLink(agent.vault_runtime_hash || "") + '</div>' +
          '<div class="mono">node: ' + oEsc(agent.vault_node_uuid || agent.node_id || "") + '</div>' +
          '<div class="mono">key version: ' + oEsc(agent.key_version ?? "") + '</div>' +
        '</div>'
      ).join("")
    : '<div class="empty">No agent rows were returned.</div>';
  return summary + rows;
}

function renderCleanupResult(result) {
  const actions = (result.actions || []).slice(0, 10);
  const summary = oCards([
    ["Status", oBadge(result.status || "preview")],
    ["Apply", oBadge(result.apply ? "applied" : "preview")],
    ["Actions", oEsc(result.counts?.actions ?? actions.length ?? 0)],
    ["Delete Candidates", oEsc(result.counts?.delete_candidates ?? 0)],
    ["Deleted", oEsc(result.counts?.deleted ?? 0)],
  ]);
  const rows = actions.length
    ? actions.map((action, index) => {
        const deletes = (action.delete || []).map((ref) => '<div class="mono">delete: ' + oRefLink(ref) + '</div>').join("");
        const keeps = (action.keep || []).map((ref) => '<div class="mono">keep: ' + oRefLink(ref) + '</div>').join("");
        return (
          '<div class="event">' +
            '<div class="event-top"><strong>Action ' + oEsc(index + 1) + '</strong>' + oBadge(action.reason || "cleanup") + '</div>' +
            '<div class="mono">key: ' + oEsc(action.key || "") + '</div>' +
            '<div class="mono">manual: ' + oEsc(action.manual ? "yes" : "no") + '</div>' +
            deletes +
            keeps +
          '</div>'
        );
      }).join("")
    : '<div class="empty">No cleanup actions were returned.</div>';
  return summary + rows;
}

async function loadAudit() {
  const report = await oFetchJSON("/api/tracked-refs/audit");
  document.getElementById("ops-audit-chip").innerHTML = oBadge("loaded");
  document.getElementById("ops-audit-summary").innerHTML = oCards([
    ["Total refs", oEsc(report.counts?.total_refs ?? 0)],
    ["Blocked", oEsc(report.counts?.blocked ?? 0)],
    ["Stale issues", oEsc(report.counts?.stale ?? 0)],
  ]);

  const stale = report.stale || [];
  const blocked = report.blocked || [];
  const issueRows = [];
  if (blocked.length) {
    issueRows.push([
      "Blocked refs",
      blocked.slice(0, 8).map((row) => {
        const ref = oRefLink(row.ref);
        const owner = row.vault_runtime_hash ? ' via ' + oApprovalLink(row.vault_runtime_hash) : '';
        return '<span class="mono">' + ref + owner + '</span>';
      }).join('<br>')
    ]);
  }
  stale.slice(0, 8).forEach((issue, index) => {
    const refs = (issue.refs || []).slice(0, 4).map((row) => {
      const ref = oRefLink(row.ref);
      const owner = row.vault_runtime_hash ? ' via ' + oApprovalLink(row.vault_runtime_hash) : '';
      return ref + owner;
    }).join('<br>');
    issueRows.push([
      "Issue " + (index + 1),
      oBadge(issue.reason) + ' <span class="mono">' + oEsc(issue.key) + '</span>' + (refs ? '<div style="margin-top:8px" class="mono">' + refs + '</div>' : ''),
    ]);
  });
  document.getElementById("ops-audit-issues").innerHTML = issueRows.length
    ? oCards(issueRows)
    : '<div class="empty">No tracked-ref issues detected.</div>';
}

async function runRotateAll() {
  if (!oChecked("ops-rotate-confirm")) {
    throw new Error("confirm rotate-all before scheduling the global action");
  }
  const result = await oFetchJSON("/api/agents/rotate-all", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: "{}",
  });
  document.getElementById("ops-rotate-chip").innerHTML = oBadge(result.status);
  oResultHTML("ops-rotate-result", renderRotateResult(result), true);
}

async function runCleanup(apply) {
  if (apply && !oChecked("ops-cleanup-confirm")) {
    throw new Error("confirm cleanup apply before mutating tracked-ref state");
  }
  const result = await oFetchJSON("/api/tracked-refs/cleanup", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({apply}),
  });
  document.getElementById("ops-cleanup-chip").innerHTML = oBadge(result.status);
  oResultHTML("ops-cleanup-result", renderCleanupResult(result), true);
}

document.getElementById("ops-rotate-btn")?.addEventListener("click", () => {
  runRotateAll().catch((err) => {
    document.getElementById("ops-rotate-chip").innerHTML = oBadge("error");
    oResult("ops-rotate-result", err.message, false);
  });
});

document.getElementById("ops-cleanup-preview")?.addEventListener("click", () => {
  runCleanup(false).catch((err) => {
    document.getElementById("ops-cleanup-chip").innerHTML = oBadge("error");
    oResult("ops-cleanup-result", err.message, false);
  });
});

document.getElementById("ops-cleanup-apply")?.addEventListener("click", () => {
  runCleanup(true).catch((err) => {
    document.getElementById("ops-cleanup-chip").innerHTML = oBadge("error");
    oResult("ops-cleanup-result", err.message, false);
  });
});

loadAudit().catch((err) => {
  document.getElementById("ops-audit-chip").innerHTML = oBadge("error");
  document.getElementById("ops-audit-summary").innerHTML = '<div class="empty">' + oEsc(err.message) + '</div>';
  document.getElementById("ops-audit-issues").innerHTML = '<div class="empty">Audit report failed to load.</div>';
});
