function fEsc(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function fBadge(value) {
  const normalized = String(value || "").toLowerCase();
  let klass = "warn";
  if (normalized === "ok" || normalized === "gitlab" || normalized === "active") klass = "ok";
  if (normalized === "error" || normalized === "blocked") klass = "err";
  return '<span class="badge ' + klass + '">' + fEsc(value || "unknown") + '</span>';
}

function fCards(rows) {
  return rows.map(([label, value]) =>
    '<div class="event"><div class="event-top"><strong>' + fEsc(label) + '</strong></div><div class="mono">' + value + '</div></div>'
  ).join("");
}

async function fFetchJSON(path) {
  const res = await fetch(path, {headers: {accept: "application/json"}});
  if (!res.ok) {
    const text = await res.text();
    throw new Error(path + " -> " + res.status + " " + text);
  }
  return res.json();
}

function parseVarsJSON(raw) {
  if (!raw) return {};
  try {
    const parsed = JSON.parse(raw);
    return parsed && typeof parsed === "object" ? parsed : {};
  } catch (_) {
    return {};
  }
}

function refDetailLink(ref) {
  const value = String(ref || "");
  if (!value) return '<span class="muted">n/a</span>';
  const path = value.startsWith("VE:") ? "/ui/configs/" : "/ui/secrets/";
  return '<a href="' + path + encodeURIComponent(value) + '">' + fEsc(value) + '</a>';
}

function varsContractCards(raw) {
  const vars = parseVarsJSON(raw);
  const names = Object.keys(vars);
  if (!names.length) {
    return '<div class="empty">No vars contract declared.</div>';
  }
  return names.map((name) => {
    const row = vars[name] || {};
    return (
      '<div class="event">' +
        '<div class="event-top"><strong>' + fEsc(name) + '</strong>' + fBadge(row.class || "var") + '</div>' +
        '<div class="mono">' + refDetailLink(row.ref) + '</div>' +
      '</div>'
    );
  }).join("");
}

function renderFunctions(rows) {
  const el = document.getElementById("functions-table");
  if (!rows.length) {
    el.innerHTML = '<tr><td colspan="4"><div class="empty">No global functions available.</div></td></tr>';
    return;
  }
  el.innerHTML = rows.map((row) =>
    '<tr>' +
      '<td class="mono"><a href="/ui/functions/' + encodeURIComponent(row.name || "") + '">' + fEsc(row.name || "") + '</a></td>' +
      '<td>' + fBadge(row.category || "unknown") + '</td>' +
      '<td class="mono">' + fEsc(row.function_hash || "") + '</td>' +
      '<td>' + fEsc(row.updated_at || "-") + '</td>' +
    '</tr>'
  ).join("");
}

async function loadFunctionsPage() {
  const list = await fFetchJSON("/api/functions/global");
  const rows = list.functions || [];
  document.getElementById("functions-chip").textContent = (list.count ?? rows.length) + " rows";
  renderFunctions(rows);

  if (!rows.length) {
    document.getElementById("function-detail-chip").textContent = "Empty";
    document.getElementById("function-summary").innerHTML = '<div class="empty">No function detail preview available.</div>';
    document.getElementById("function-vars").innerHTML = '<div class="empty">No vars contract available.</div>';
    return;
  }

  const detail = await fFetchJSON("/api/functions/global/" + encodeURIComponent(rows[0].name));
  document.getElementById("function-detail-chip").innerHTML = fBadge(detail.category || "detail");
  document.getElementById("function-summary").innerHTML = fCards([
    ["Name", '<a href="/ui/functions/' + encodeURIComponent(detail.name || "") + '">' + fEsc(detail.name || "") + '</a>'],
    ["Category", fBadge(detail.category || "unknown")],
    ["Function Hash", '<span class="mono">' + fEsc(detail.function_hash || "") + '</span>'],
    ["Updated At", fEsc(detail.updated_at || "-")],
  ]);
  document.getElementById("function-vars").innerHTML = varsContractCards(detail.vars_json || "{}");
}

loadFunctionsPage().catch((err) => {
  document.getElementById("functions-chip").textContent = "Load failed";
  document.getElementById("functions-table").innerHTML = '<tr><td colspan="4"><div class="empty">' + fEsc(err.message) + '</div></td></tr>';
  document.getElementById("function-summary").innerHTML = '<div class="empty">Function preview failed to load.</div>';
  document.getElementById("function-vars").innerHTML = '<div class="empty">Vars contract failed to load.</div>';
});
