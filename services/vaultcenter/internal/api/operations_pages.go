package api

import "net/http"

func (s *Server) handleOperationsPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(operationsHTML))
}

const operationsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>VeilKey Operations</title>
<link rel="stylesheet" href="/assets/operator-console.css">
</head>
<body>
<div class="shell">
  <section class="hero">
    <div class="hero-main">
      <div class="eyebrow">Operations</div>
      <h1>VaultCenter Operations</h1>
      <p class="hero-copy">
        This page groups high-impact operator actions. The current focus is planned rotation scheduling, tracked-ref audit,
        and tracked-ref cleanup preview/apply.
      </p>
      <div class="quick-links">
        <a href="/">Operator console</a>
        <a href="#rotation">Rotation</a>
        <a href="#audit">Tracked-ref audit</a>
        <a href="#cleanup">Cleanup</a>
      </div>
    </div>
    <aside class="hero-side">
      <h2>Guardrails</h2>
      <ul>
        <li>Actions remain explicit and operator-triggered.</li>
        <li>Tracked-ref cleanup defaults to preview mode.</li>
        <li>Rotation remains global until vault-scoped scheduling exists.</li>
      </ul>
    </aside>
  </section>

  <main class="content">
    <section class="panel" id="rotation">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Planned Rotation</h2>
          <div class="panel-sub">Schedule the current global planned rotation pass.</div>
        </div>
        <div class="chip" id="ops-rotate-chip">Idle</div>
      </div>
      <div class="panel" style="padding:16px;background:#0c1822">
        <label style="display:flex;gap:10px;align-items:flex-start;margin-bottom:12px;color:#d9e6ef">
          <input id="ops-rotate-confirm" type="checkbox" style="margin-top:3px">
          <span>I understand rotate-all is a global runtime action and should only run after catalog review.</span>
        </label>
        <button id="ops-rotate-btn" class="btn">Schedule rotate-all</button>
        <div id="ops-rotate-result" class="result" style="display:block;margin-top:14px"></div>
      </div>
    </section>

    <section class="panel" id="audit">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Tracked-Ref Audit</h2>
          <div class="panel-sub">Blocked refs, stale ownership, and duplicate issues.</div>
        </div>
        <div class="chip" id="ops-audit-chip">Loading</div>
      </div>
      <div class="grid-2">
        <div class="stack" id="ops-audit-summary"></div>
        <div class="stack" id="ops-audit-issues"></div>
      </div>
    </section>

    <section class="panel" id="cleanup">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">Tracked-Ref Cleanup</h2>
          <div class="panel-sub">Preview or apply cleanup for stale tracked refs.</div>
        </div>
        <div class="chip" id="ops-cleanup-chip">Idle</div>
      </div>
      <div class="panel" style="padding:16px;background:#0c1822">
        <label style="display:flex;gap:10px;align-items:flex-start;margin-bottom:12px;color:#d9e6ef">
          <input id="ops-cleanup-confirm" type="checkbox" style="margin-top:3px">
          <span>I understand cleanup apply mutates tracked-ref state and should only follow preview review.</span>
        </label>
        <div style="display:flex;gap:10px;flex-wrap:wrap">
          <button id="ops-cleanup-preview" class="btn" style="background:#19b394">Preview cleanup</button>
          <button id="ops-cleanup-apply" class="btn">Apply cleanup</button>
        </div>
        <div id="ops-cleanup-result" class="result" style="display:block;margin-top:14px"></div>
      </div>
    </section>

    <section class="panel">
      <div class="panel-head">
        <div>
          <h2 class="panel-title">API Surface</h2>
          <div class="panel-sub">Current runtime endpoints behind this operations view.</div>
        </div>
      </div>
      <div class="stack">
        <div class="event"><div class="mono">POST /api/agents/rotate-all</div></div>
        <div class="event"><div class="mono">GET /api/tracked-refs/audit</div></div>
        <div class="event"><div class="mono">POST /api/tracked-refs/cleanup</div></div>
      </div>
    </section>
  </main>
</div>
<script src="/assets/operations-console.js"></script>
</body>
</html>
` + "\n"
