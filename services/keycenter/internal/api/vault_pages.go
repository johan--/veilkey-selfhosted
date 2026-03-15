package api

import "net/http"

func (s *Server) handleVaultsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(vaultsHTML))
}

func (s *Server) handleVaultDetailPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(vaultDetailHTML))
}

const vaultsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>VeilKey Vaults</title>
<link rel="stylesheet" href="/assets/operator-console.css">
</head>
<body>
<div class="shell">
  <section class="hero">
    <div class="hero-main">
      <div class="eyebrow">Vault inventory</div>
      <h1>Vaults</h1>
      <p class="hero-copy">
        Hash-first vault inventory surface for operator routing. This page keeps ` + "`vault_hash`" + ` visible in list views and leaves ` + "`vault_node_uuid`" + ` for exact detail context.
      </p>
      <div class="quick-links">
        <a href="/">Operator console</a>
        <a href="#inventory">Inventory</a>
        <a href="#preview">Preview</a>
      </div>
    </div>
    <aside class="hero-side">
      <h2>Guardrails</h2>
      <ul>
        <li>Vault tables stay hash-first.</li>
        <li>Detail depth should be reached through explicit vault selection.</li>
        <li>Managed path and vault type stay visible before action routing.</li>
      </ul>
    </aside>
  </section>

  <main class="content">
    <section class="panel" id="inventory">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Vault Inventory</h2>
          <div class="panel-sub">Connected vault rows with type, managed path, and counts.</div>
        </div>
        <div class="chip" id="vaults-page-chip">Loading</div>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Vault Hash</th>
              <th>Type</th>
              <th>Node UUID</th>
              <th>Managed Path</th>
              <th>Refs</th>
              <th>Configs</th>
            </tr>
          </thead>
          <tbody id="vaults-page-table"></tbody>
        </table>
      </div>
    </section>

    <section class="panel" id="preview">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Preview</h2>
          <div class="panel-sub">First vault row preview with direct links into detail and approvals.</div>
        </div>
        <div class="chip" id="vaults-preview-chip">Loading</div>
      </div>
      <div class="grid-2">
        <div class="stack" id="vaults-preview-summary"></div>
        <div class="stack" id="vaults-preview-context"></div>
      </div>
    </section>
  </main>
</div>
<script src="/assets/vaults-console.js"></script>
</body>
</html>
` + "\n"

const vaultDetailHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>VeilKey Vault Detail</title>
<link rel="stylesheet" href="/assets/operator-console.css">
</head>
<body>
<div class="shell">
  <section class="hero">
    <div class="hero-main">
      <div class="eyebrow">Vault detail</div>
      <h1 id="vault-title">Vault Detail</h1>
      <p class="hero-copy" id="vault-subtitle">
        Hash-first detail route for a single vault. This page joins state, catalog, and action entrypoints around one vault hash.
      </p>
      <div class="quick-links">
        <a href="/">Operator console</a>
        <a href="/ui/approvals" id="approvals-link">Approvals</a>
        <a href="#inventory">Inventory</a>
        <a href="#secrets">Secrets</a>
        <a href="#actions">Actions</a>
      </div>
    </div>
    <aside class="hero-side">
      <h2>Route contract</h2>
      <ul>
        <li>Table/list routes should prefer <span class="mono">vault_hash</span>.</li>
        <li>Action and approval flows can still surface <span class="mono">vault_runtime_hash</span> when needed.</li>
        <li>This page is a bridge between state, catalog, detail, and action.</li>
      </ul>
    </aside>
  </section>

  <main class="content">
    <section class="panel" id="inventory">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Inventory</h2>
          <div class="panel-sub">Live inventory row and agent state for this vault hash.</div>
        </div>
        <div class="chip" id="vault-detail-chip">Loading</div>
      </div>
      <div class="grid-2">
        <div class="stack" id="vault-overview"></div>
        <div class="stack" id="vault-agent"></div>
      </div>
    </section>

    <section class="panel" id="secrets">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Secrets</h2>
          <div class="panel-sub">Catalog slice filtered by this vault hash.</div>
        </div>
        <div class="chip" id="vault-secret-chip">Loading secrets</div>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Canonical Ref</th>
              <th>Name</th>
              <th>Class</th>
              <th>Status</th>
              <th>Bindings</th>
            </tr>
          </thead>
          <tbody id="vault-secret-table"></tbody>
        </table>
      </div>
    </section>

    <section class="panel" id="actions">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Actions</h2>
          <div class="panel-sub">Current action entrypoints for this vault.</div>
        </div>
      </div>
      <div class="grid-2">
        <div class="panel" style="padding:16px;background:#0c1822">
          <h3 style="margin:0 0 8px">Planned rotation</h3>
          <p class="muted" style="margin:0 0 12px">This schedules the existing KeyCenter planned rotation pass. Current API is global, so use with operator judgment.</p>
          <button id="rotate-all-btn" class="btn">Schedule planned rotation</button>
          <div id="rotate-result" class="result" style="display:block;margin-top:14px"></div>
        </div>
        <div class="panel" style="padding:16px;background:#0c1822">
          <h3 style="margin:0 0 8px">Approval handoff</h3>
          <p class="muted" style="margin:0 0 12px">Open the approval hub with this vault runtime hash prefilled for rebind work.</p>
          <a id="vault-approval-link" href="/ui/approvals" style="text-decoration:none">
            <span style="display:inline-flex;padding:10px 14px;border-radius:999px;border:1px solid #24384a;background:#0c1822;color:#d4edf0;font-size:13px">Open approval hub</span>
          </a>
        </div>
      </div>
    </section>
  </main>
</div>
<script src="/assets/vault-detail.js"></script>
</body>
</html>
` + "\n"
