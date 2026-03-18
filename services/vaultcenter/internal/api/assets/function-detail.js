function fdEsc(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function fdBadge(value) {
  const normalized = String(value || "").toLowerCase();
  let klass = "warn";
  if (normalized === "ok" || normalized === "gitlab" || normalized === "active") klass = "ok";
  if (normalized === "error" || normalized === "blocked") klass = "err";
  return '<span class="badge ' + klass + '">' + fdEsc(value || "unknown") + '</span>';
}

function fdCards(rows) {
  return rows.map(([label, value]) =>
    '<div class="event"><div class="event-top"><strong>' + fdEsc(label) + '</strong></div><div class="mono">' + value + '</div></div>'
  ).join("");
}

async function fdFetchJSON(path) {
  const res = await fetch(path, {headers: {accept: "application/json"}});
  if (!res.ok) {
    const text = await res.text();
    throw new Error(path + " -> " + res.status + " " + text);
  }
  return res.json();
}

function parseFunctionVars(raw) {
  if (!raw) return {};
  try {
    const parsed = JSON.parse(raw);
    return parsed && typeof parsed === "object" ? parsed : {};
  } catch (_) {
    return {};
  }
}

function functionRefLink(ref) {
  const value = String(ref || "");
  if (!value) return '<span class="muted">n/a</span>';
  const path = value.startsWith("VE:") ? "/ui/configs/" : "/ui/secrets/";
  return '<a href="' + path + encodeURIComponent(value) + '">' + fdEsc(value) + '</a>';
}

function renderFunctionVars(raw) {
  const vars = parseFunctionVars(raw);
  const names = Object.keys(vars);
  if (!names.length) {
    return '<div class="empty">No vars contract declared.</div>';
  }
  return names.map((name) => {
    const row = vars[name] || {};
    return (
      '<div class="event">' +
        '<div class="event-top"><strong>' + fdEsc(name) + '</strong>' + fdBadge(row.class || "var") + '</div>' +
        '<div class="mono">' + functionRefLink(row.ref) + '</div>' +
      '</div>'
    );
  }).join("");
}

function functionNameFromPath() {
  const parts = location.pathname.split("/").filter(Boolean);
  return decodeURIComponent(parts.slice(2).join("/"));
}

async function loadFunctionDetail() {
  const name = functionNameFromPath();
  if (!name) throw new Error("function name route segment is required");

  const detail = await fdFetchJSON("/api/functions/global/" + encodeURIComponent(name));
  document.getElementById("function-title").textContent = "Function Detail · " + name;
  document.getElementById("function-page-chip").innerHTML = fdBadge(detail.category || "detail");
  document.getElementById("function-page-summary").innerHTML = fdCards([
    ["Name", '<span class="mono">' + fdEsc(detail.name || name) + '</span>'],
    ["Category", fdBadge(detail.category || "unknown")],
    ["Function Hash", '<span class="mono">' + fdEsc(detail.function_hash || "") + '</span>'],
    ["Created At", fdEsc(detail.created_at || "-")],
    ["Updated At", fdEsc(detail.updated_at || "-")],
  ]);
  document.getElementById("function-page-vars").innerHTML = renderFunctionVars(detail.vars_json || "{}");
  document.getElementById("function-command").textContent = detail.command || "";
}

loadFunctionDetail().catch((err) => {
  document.getElementById("function-page-summary").innerHTML = '<div class="empty">' + fdEsc(err.message) + '</div>';
  document.getElementById("function-page-vars").innerHTML = '<div class="empty">Function detail failed to load.</div>';
  document.getElementById("function-command").textContent = "";
});
