package api

import "net/http"

func (s *Server) handleConfigsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(configsHTML))
}

func (s *Server) handleConfigDetailPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(configDetailHTML))
}

const configsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>VeilKey Configs</title>
<link rel="stylesheet" href="/assets/operator-console.css">
</head>
<body>
<div class="shell">
  <section class="hero">
    <div class="hero-main">
      <div class="eyebrow">Config catalog</div>
      <h1>Configs</h1>
      <p class="hero-copy">
        Canonical config list surface using ` + "`VE:{SCOPE}:{KEY}`" + ` refs from the shared operator catalog. This keeps config spread visible before mutation or bulk update depth.
      </p>
      <div class="quick-links">
        <a href="/">Operator console</a>
        <a href="#catalog">Catalog</a>
        <a href="#preview">Preview</a>
      </div>
    </div>
    <aside class="hero-side">
      <h2>Guardrails</h2>
      <ul>
        <li>Config refs stay scoped and canonical.</li>
        <li>Bulk change should follow list and vault review.</li>
        <li>Binding visibility stays attached to each key.</li>
      </ul>
    </aside>
  </section>

  <main class="content">
    <section class="panel" id="catalog">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Config Catalog</h2>
          <div class="panel-sub">Config rows filtered from the shared catalog read-model.</div>
        </div>
        <div class="chip" id="configs-chip">Loading</div>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Canonical Ref</th>
              <th>Key</th>
              <th>Class</th>
              <th>Vault Hash</th>
              <th>Bindings</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody id="configs-table"></tbody>
        </table>
      </div>
    </section>

    <section class="panel" id="preview">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Preview</h2>
          <div class="panel-sub">One canonical config preview with direct links into detail and vault context.</div>
        </div>
        <div class="chip" id="configs-preview-chip">Loading</div>
      </div>
      <div class="grid-2">
        <div class="stack" id="configs-preview-summary"></div>
        <div class="stack" id="configs-preview-context"></div>
      </div>
    </section>
  </main>
</div>
<script src="/assets/configs-console.js"></script>
</body>
</html>
` + "\n"

const configDetailHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>VeilKey Config Detail</title>
<link rel="stylesheet" href="/assets/operator-console.css">
</head>
<body>
<div class="shell">
  <section class="hero">
    <div class="hero-main">
      <div class="eyebrow">Config detail</div>
      <h1 id="config-title">Config Detail</h1>
      <p class="hero-copy">
        Canonical config detail route anchored on ` + "`VE:`" + ` ref. This page joins the catalog row, binding visibility, and recent audit for one config key.
      </p>
      <div class="quick-links">
        <a href="/">Operator console</a>
        <a href="#summary">Summary</a>
        <a href="#bindings">Bindings</a>
        <a href="#audit">Audit</a>
      </div>
    </div>
    <aside class="hero-side">
      <h2>Guardrails</h2>
      <ul>
        <li>Canonical config refs stay in the ` + "`VE:{SCOPE}:{KEY}`" + ` contract.</li>
        <li>Operator detail should show config spread before any bulk mutation.</li>
        <li>Binding and audit context stay visible alongside the canonical key.</li>
      </ul>
    </aside>
  </section>

  <main class="content">
    <section class="panel" id="summary">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Summary</h2>
          <div class="panel-sub">Catalog row and vault context for this config ref.</div>
        </div>
        <div class="chip" id="config-detail-chip">Loading</div>
      </div>
      <div class="grid-2">
        <div class="stack" id="config-summary"></div>
        <div class="stack" id="config-context"></div>
      </div>
    </section>

    <section class="panel" id="bindings">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Bindings</h2>
          <div class="panel-sub">Target visibility for this config ref.</div>
        </div>
        <div class="chip" id="config-binding-chip">Loading</div>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Type</th>
              <th>Target</th>
              <th>Vault Hash</th>
              <th>Field</th>
              <th>Updated</th>
            </tr>
          </thead>
          <tbody id="config-binding-table"></tbody>
        </table>
      </div>
    </section>

    <section class="panel" id="audit">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Recent Audit</h2>
          <div class="panel-sub">Latest audit slice for this config entity.</div>
        </div>
        <div class="chip" id="config-audit-chip">Loading</div>
      </div>
      <div class="stack" id="config-audit-list"></div>
    </section>
  </main>
</div>
<script src="/assets/config-detail.js"></script>
</body>
</html>
` + "\n"
