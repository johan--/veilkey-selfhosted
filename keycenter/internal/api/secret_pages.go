package api

import "net/http"

func (s *Server) handleSecretsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(secretsHTML))
}

func (s *Server) handleSecretDetailPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(secretDetailHTML))
}

const secretsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>VeilKey Secrets</title>
<link rel="stylesheet" href="/assets/operator-console.css">
</head>
<body>
<div class="shell">
  <section class="hero">
    <div class="hero-main">
      <div class="eyebrow">Secrets catalog</div>
      <h1>Secrets</h1>
      <p class="hero-copy">
        Canonical secret list surface anchored on ` + "`VK:{SCOPE}:{REF}`" + ` refs. This page keeps the list masked-by-default and pairs the catalog with one live preview pane.
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
        <li>Secret lists stay metadata-first.</li>
        <li>Canonical ref is the primary route key.</li>
        <li>Vault impact stays visible before action depth.</li>
      </ul>
    </aside>
  </section>

  <main class="content">
    <section class="panel" id="catalog">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Secret Catalog</h2>
          <div class="panel-sub">Canonical secret rows with vault and binding counts.</div>
        </div>
        <div class="chip" id="secrets-chip">Loading</div>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Canonical Ref</th>
              <th>Name</th>
              <th>Class</th>
              <th>Vault Hash</th>
              <th>Bindings</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody id="secrets-table"></tbody>
        </table>
      </div>
    </section>

    <section class="panel" id="preview">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Preview</h2>
          <div class="panel-sub">One canonical secret preview with direct links into detail and vault context.</div>
        </div>
        <div class="chip" id="secrets-preview-chip">Loading</div>
      </div>
      <div class="grid-2">
        <div class="stack" id="secrets-preview-summary"></div>
        <div class="stack" id="secrets-preview-context"></div>
      </div>
    </section>
  </main>
</div>
<script src="/assets/secrets-console.js"></script>
</body>
</html>
` + "\n"

const secretDetailHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>VeilKey Secret Detail</title>
<link rel="stylesheet" href="/assets/operator-console.css">
</head>
<body>
<div class="shell">
  <section class="hero">
    <div class="hero-main">
      <div class="eyebrow">Secret detail</div>
      <h1 id="secret-title">Secret Detail</h1>
      <p class="hero-copy">
        Canonical secret detail route anchored on ref. This page joins the catalog row, binding visibility, and recent audit for one secret.
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
        <li>Detail routes prefer canonical ref, not plaintext identity.</li>
        <li>Catalog remains masked-by-default.</li>
        <li>Binding and audit context should stay visible before any future reveal action.</li>
      </ul>
    </aside>
  </section>

  <main class="content">
    <section class="panel" id="summary">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Summary</h2>
          <div class="panel-sub">Catalog row and inventory context for this canonical ref.</div>
        </div>
        <div class="chip" id="secret-detail-chip">Loading</div>
      </div>
      <div class="grid-2">
        <div class="stack" id="secret-summary"></div>
        <div class="stack" id="secret-context"></div>
      </div>
    </section>

    <section class="panel" id="bindings">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Bindings</h2>
          <div class="panel-sub">Target visibility for this ref.</div>
        </div>
        <div class="chip" id="secret-binding-chip">Loading</div>
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
          <tbody id="secret-binding-table"></tbody>
        </table>
      </div>
    </section>

    <section class="panel" id="audit">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Recent Audit</h2>
          <div class="panel-sub">Latest audit slice for this secret entity.</div>
        </div>
        <div class="chip" id="secret-audit-chip">Loading</div>
      </div>
      <div class="stack" id="secret-audit-list"></div>
    </section>
  </main>
</div>
<script src="/assets/secret-detail.js"></script>
</body>
</html>
` + "\n"
