package api

import "net/http"

func (s *Server) handleFunctionsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(functionsHTML))
}

func (s *Server) handleFunctionDetailPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(functionDetailHTML))
}

const functionsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>VeilKey Functions</title>
<link rel="stylesheet" href="/assets/operator-console.css">
</head>
<body>
<div class="shell">
  <section class="hero">
    <div class="hero-main">
      <div class="eyebrow">Functions</div>
      <h1>Global Functions</h1>
      <p class="hero-copy">
        Inspect global function templates, canonical command shapes, and variable contracts without dropping into raw JSON first.
      </p>
      <div class="quick-links">
        <a href="/">Operator console</a>
        <a href="#catalog">Catalog</a>
        <a href="#detail">Detail</a>
      </div>
      <div style="margin-top:14px"><span class="badge warn">Experimental prototype until DB-backed execution is complete</span></div>
    </div>
    <aside class="hero-side">
      <h2>Guardrails</h2>
      <ul>
        <li>Global functions stay template-first and masked-by-default.</li>
        <li>Variable contracts should point to canonical ` + "`VK:{SCOPE}:{REF}`" + ` or ` + "`VE:{SCOPE}:{KEY}`" + ` refs.</li>
        <li>Detail pages should expose command shape and vars, not resolved plaintext.</li>
      </ul>
    </aside>
  </section>

  <main class="content">
    <section class="panel" id="catalog">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Catalog</h2>
          <div class="panel-sub">Current global function rows ordered by name.</div>
        </div>
        <div class="chip" id="functions-chip">Loading</div>
      </div>
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Category</th>
              <th>Function Hash</th>
              <th>Updated</th>
            </tr>
          </thead>
          <tbody id="functions-table"></tbody>
        </table>
      </div>
    </section>

    <section class="panel" id="detail">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Detail Preview</h2>
          <div class="panel-sub">First function row preview to shorten operator path.</div>
        </div>
        <div class="chip" id="function-detail-chip">Loading</div>
      </div>
      <div class="grid-2">
        <div class="stack" id="function-summary"></div>
        <div class="stack" id="function-vars"></div>
      </div>
    </section>
  </main>
</div>
<script src="/assets/functions-console.js"></script>
</body>
</html>
` + "\n"

const functionDetailHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>VeilKey Function Detail</title>
<link rel="stylesheet" href="/assets/operator-console.css">
</head>
<body>
<div class="shell">
  <section class="hero">
    <div class="hero-main">
      <div class="eyebrow">Function detail</div>
      <h1 id="function-title">Function Detail</h1>
      <p class="hero-copy">
        Canonical function detail route anchored on the function name. This page keeps the command template and variable contract visible together.
      </p>
      <div class="quick-links">
        <a href="/">Operator console</a>
        <a href="/ui/functions">Functions</a>
        <a href="#summary">Summary</a>
        <a href="#command">Command</a>
      </div>
      <div style="margin-top:14px"><span class="badge warn">Experimental prototype until DB-backed execution is complete</span></div>
    </div>
    <aside class="hero-side">
      <h2>Guardrails</h2>
      <ul>
        <li>Function names are canonical operator entrypoints.</li>
        <li>Vars JSON should stay inspectable without resolving secrets.</li>
        <li>Command templates remain auditable before any execution surface is added.</li>
      </ul>
    </aside>
  </section>

  <main class="content">
    <section class="panel" id="summary">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Summary</h2>
          <div class="panel-sub">Function metadata and canonical refs contract.</div>
        </div>
        <div class="chip" id="function-page-chip">Loading</div>
      </div>
      <div class="grid-2">
        <div class="stack" id="function-page-summary"></div>
        <div class="stack" id="function-page-vars"></div>
      </div>
    </section>

    <section class="panel" id="command">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Command Template</h2>
          <div class="panel-sub">Raw command template stored in the global catalog.</div>
        </div>
      </div>
      <pre class="result show" id="function-command" style="display:block"></pre>
    </section>
  </main>
</div>
<script src="/assets/function-detail.js"></script>
</body>
</html>
` + "\n"
