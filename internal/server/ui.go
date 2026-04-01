package server

import "net/http"

const uiHTML = `<!DOCTYPE html><html lang="en"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Stampede — Stockyard</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital,wght@0,400;0,700;1,400&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
<style>:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#c45d2c;--rust-light:#e8753a;--rust-dark:#8b3d1a;--leather:#a0845c;--leather-light:#c4a87a;--cream:#f0e6d3;--cream-dim:#bfb5a3;--cream-muted:#7a7060;--gold:#d4a843;--green:#5ba86e;--red:#c0392b;--blue:#4a90d9;--font-serif:'Libre Baskerville',Georgia,serif;--font-mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--font-serif);min-height:100vh}a{color:var(--rust-light);text-decoration:none}a:hover{color:var(--gold)}
.hdr{background:var(--bg2);border-bottom:2px solid var(--rust-dark);padding:.9rem 1.8rem;display:flex;align-items:center;justify-content:space-between}.hdr-left{display:flex;align-items:center;gap:1rem}.hdr-brand{font-family:var(--font-mono);font-size:.75rem;color:var(--leather);letter-spacing:3px;text-transform:uppercase}.hdr-title{font-family:var(--font-mono);font-size:1.1rem;color:var(--cream);letter-spacing:1px}.badge{font-family:var(--font-mono);font-size:.6rem;padding:.2rem .6rem;letter-spacing:1px;text-transform:uppercase;border:1px solid;color:var(--green);border-color:var(--green)}
.main{max-width:1000px;margin:0 auto;padding:2rem 1.5rem}.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(120px,1fr));gap:1rem;margin-bottom:2rem}.card{background:var(--bg2);border:1px solid var(--bg3);padding:1rem 1.2rem}.card-val{font-family:var(--font-mono);font-size:1.6rem;font-weight:700;color:var(--cream);display:block}.card-lbl{font-family:var(--font-mono);font-size:.55rem;letter-spacing:2px;text-transform:uppercase;color:var(--leather);margin-top:.2rem}
.section{margin-bottom:2rem}.section-title{font-family:var(--font-mono);font-size:.68rem;letter-spacing:3px;text-transform:uppercase;color:var(--rust-light);margin-bottom:.8rem;padding-bottom:.5rem;border-bottom:1px solid var(--bg3)}table{width:100%;border-collapse:collapse;font-family:var(--font-mono);font-size:.72rem}th{background:var(--bg3);padding:.4rem .6rem;text-align:left;color:var(--leather-light);font-weight:400;letter-spacing:1px;font-size:.6rem;text-transform:uppercase}td{padding:.4rem .6rem;border-bottom:1px solid var(--bg3);color:var(--cream-dim)}tr:hover td{background:var(--bg2)}.empty{color:var(--cream-muted);text-align:center;padding:2rem;font-style:italic}
.btn{font-family:var(--font-mono);font-size:.7rem;padding:.3rem .8rem;border:1px solid var(--leather);background:transparent;color:var(--cream);cursor:pointer;transition:all .2s}.btn:hover{border-color:var(--rust-light);color:var(--rust-light)}.btn-rust{border-color:var(--rust);color:var(--rust-light)}.btn-rust:hover{background:var(--rust);color:var(--cream)}.btn-sm{font-size:.62rem;padding:.2rem .5rem}
.pill{display:inline-block;font-family:var(--font-mono);font-size:.55rem;padding:.1rem .4rem;border-radius:2px;text-transform:uppercase}.pill-running{background:#1a2a3a;color:var(--blue)}.pill-completed{background:#1a3a2a;color:var(--green)}.pill-failed{background:#2a1a1a;color:var(--red)}
.lbl{font-family:var(--font-mono);font-size:.62rem;letter-spacing:1px;text-transform:uppercase;color:var(--leather)}input{font-family:var(--font-mono);font-size:.78rem;background:var(--bg3);border:1px solid var(--bg3);color:var(--cream);padding:.4rem .7rem;outline:none}input:focus{border-color:var(--leather)}.row{display:flex;gap:.8rem;align-items:flex-end;flex-wrap:wrap;margin-bottom:1rem}.field{display:flex;flex-direction:column;gap:.3rem}
.tabs{display:flex;gap:0;margin-bottom:1.5rem;border-bottom:1px solid var(--bg3)}.tab{font-family:var(--font-mono);font-size:.72rem;padding:.6rem 1.2rem;color:var(--cream-muted);cursor:pointer;border-bottom:2px solid transparent;letter-spacing:1px;text-transform:uppercase}.tab:hover{color:var(--cream-dim)}.tab.active{color:var(--rust-light);border-bottom-color:var(--rust-light)}.tab-content{display:none}.tab-content.active{display:block}
pre{background:var(--bg3);padding:.8rem 1rem;font-family:var(--font-mono);font-size:.72rem;color:var(--cream-dim);overflow-x:auto}
.live-panel{background:var(--bg2);border:2px solid var(--rust);padding:1.5rem;margin-bottom:2rem;display:none}
.live-panel .cards{margin-bottom:0}
</style></head><body>
<div class="hdr"><div class="hdr-left">
<svg viewBox="0 0 64 64" width="22" height="22" fill="none"><rect x="8" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="28" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="48" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="8" y="27" width="48" height="7" rx="2.5" fill="#c4a87a"/></svg>
<span class="hdr-brand">Stockyard</span><span class="hdr-title">Stampede</span></div>
<div style="display:flex;gap:.8rem;align-items:center"><span class="badge">Free</span></div></div>
<div class="main">

<div id="live-panel" class="live-panel">
  <div class="section-title" style="color:var(--gold);border-color:var(--rust)">Live Test Running</div>
  <div class="cards" style="margin-top:.8rem">
    <div class="card"><span class="card-val" id="l-total">0</span><span class="card-lbl">Requests</span></div>
    <div class="card"><span class="card-val" id="l-rps">0</span><span class="card-lbl">RPS</span></div>
    <div class="card"><span class="card-val" id="l-success">0</span><span class="card-lbl">Success</span></div>
    <div class="card"><span class="card-val" id="l-errors">0</span><span class="card-lbl">Errors</span></div>
    <div class="card"><span class="card-val" id="l-elapsed">0s</span><span class="card-lbl">Elapsed</span></div>
  </div>
</div>

<div class="cards">
  <div class="card"><span class="card-val" id="s-tests">—</span><span class="card-lbl">Tests</span></div>
  <div class="card"><span class="card-val" id="s-runs">—</span><span class="card-lbl">Runs</span></div>
  <div class="card"><span class="card-val" id="s-active">—</span><span class="card-lbl">Active</span></div>
</div>

<div class="tabs">
  <div class="tab active" onclick="switchTab('tests')">Tests</div>
  <div class="tab" onclick="switchTab('create')">Create</div>
  <div class="tab" onclick="switchTab('usage')">Usage</div>
</div>

<div id="tab-tests" class="tab-content active">
  <div class="section">
    <div class="section-title">Tests</div>
    <table><thead><tr><th>Name</th><th>URL</th><th>Workers</th><th>Duration</th><th></th></tr></thead>
    <tbody id="tests-body"></tbody></table>
  </div>
  <div class="section">
    <div class="section-title">Recent Runs</div>
    <table><thead><tr><th>Test</th><th>Status</th><th>Requests</th><th>RPS</th><th>p50</th><th>p99</th><th>Errors</th><th>Time</th></tr></thead>
    <tbody id="runs-body"></tbody></table>
  </div>
</div>

<div id="tab-create" class="tab-content">
  <div class="section">
    <div class="section-title">Create Test</div>
    <div class="row">
      <div class="field"><span class="lbl">URL</span><input id="c-url" placeholder="https://api.example.com/health" style="width:300px"></div>
      <div class="field"><span class="lbl">Method</span><input id="c-method" placeholder="GET" value="GET" style="width:70px"></div>
    </div>
    <div class="row">
      <div class="field"><span class="lbl">Workers</span><input id="c-conc" placeholder="10" value="10" type="number" style="width:70px"></div>
      <div class="field"><span class="lbl">Duration (s)</span><input id="c-dur" placeholder="30" value="10" type="number" style="width:80px"></div>
      <div class="field"><span class="lbl">Name</span><input id="c-name" placeholder="API Health" style="width:160px"></div>
      <button class="btn btn-rust" onclick="createTest()">Create</button>
    </div>
    <div id="c-result"></div>
  </div>
</div>

<div id="tab-usage" class="tab-content">
  <div class="section"><div class="section-title">Quick Start</div>
    <pre>
# Create a test
curl -X POST http://localhost:8880/api/tests \
  -H "Content-Type: application/json" \
  -d '{"name":"API Health","url":"https://api.example.com/health","concurrency":10,"duration_seconds":30}'

# Run the test
curl -X POST http://localhost:8880/api/tests/{id}/run

# Check live stats (while running)
curl http://localhost:8880/api/runs/{run_id}/live

# Get results
curl http://localhost:8880/api/runs/{run_id}
    </pre>
  </div>
</div>

</div>
<script>
let tests=[],activeRunId=null;
function switchTab(n){document.querySelectorAll('.tab').forEach(t=>t.classList.toggle('active',t.textContent.toLowerCase()===n));document.querySelectorAll('.tab-content').forEach(t=>t.classList.toggle('active',t.id==='tab-'+n));}
async function refresh(){
  try{const s=await(await fetch('/api/status')).json();document.getElementById('s-tests').textContent=s.tests||0;document.getElementById('s-runs').textContent=s.runs||0;document.getElementById('s-active').textContent=s.active_runs||0;}catch(e){}
  try{const d=await(await fetch('/api/tests')).json();tests=d.tests||[];const tb=document.getElementById('tests-body');
  if(!tests.length){tb.innerHTML='<tr><td colspan="5" class="empty">No tests yet.</td></tr>';}
  else{tb.innerHTML=tests.map(t=>'<tr><td style="color:var(--cream);font-weight:600">'+esc(t.name)+'</td><td style="font-size:.65rem">'+esc(t.url)+'</td><td>'+t.concurrency+'</td><td>'+t.duration_seconds+'s</td><td><button class="btn btn-sm btn-rust" onclick="runTest(\''+t.id+'\')">Run</button> <button class="btn btn-sm" onclick="deleteTest(\''+t.id+'\')">Delete</button></td></tr>').join('');}
  // Load runs for all tests
  let runsHTML='';for(const t of tests){const r=await(await fetch('/api/tests/'+t.id+'/runs')).json();const runs=r.runs||[];
  for(const run of runs.slice(0,5)){
    runsHTML+='<tr><td>'+esc(t.name)+'</td><td><span class="pill pill-'+run.status+'">'+run.status+'</span></td><td>'+run.total_requests+'</td><td>'+run.rps+'</td><td>'+run.p50_ms+'ms</td><td>'+run.p99_ms+'ms</td><td>'+run.error_count+'</td><td style="font-size:.62rem;color:var(--cream-muted)">'+timeAgo(run.created_at)+'</td></tr>';
    if(run.status==='running')activeRunId=run.id;
  }}
  document.getElementById('runs-body').innerHTML=runsHTML||'<tr><td colspan="8" class="empty">No runs yet.</td></tr>';
  }catch(e){}
  if(activeRunId)pollLive();
}
async function pollLive(){
  try{const d=await(await fetch('/api/runs/'+activeRunId+'/live')).json();
  document.getElementById('live-panel').style.display='block';
  document.getElementById('l-total').textContent=fmt(d.total||0);
  document.getElementById('l-rps').textContent=d.rps||'0';
  document.getElementById('l-success').textContent=fmt(d.success||0);
  document.getElementById('l-errors').textContent=d.errors||0;
  document.getElementById('l-elapsed').textContent=(d.elapsed||'0')+'s';
  if(!d.running){activeRunId=null;document.getElementById('live-panel').style.display='none';refresh();}
  }catch(e){activeRunId=null;document.getElementById('live-panel').style.display='none';}
}
async function createTest(){const url=document.getElementById('c-url').value.trim();if(!url)return;const r=await fetch('/api/tests',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name:document.getElementById('c-name').value.trim()||url,url,method:document.getElementById('c-method').value.trim()||'GET',concurrency:parseInt(document.getElementById('c-conc').value)||10,duration_seconds:parseInt(document.getElementById('c-dur').value)||10})});const d=await r.json();if(r.ok){document.getElementById('c-result').innerHTML='<span style="color:var(--green)">Created</span>';refresh();}else{document.getElementById('c-result').innerHTML='<span style="color:var(--red)">'+esc(d.error)+'</span>';}}
async function runTest(id){const r=await fetch('/api/tests/'+id+'/run',{method:'POST'});const d=await r.json();if(r.ok){activeRunId=d.run.id;document.getElementById('live-panel').style.display='block';}else{alert(d.error);}}
async function deleteTest(id){if(!confirm('Delete?'))return;await fetch('/api/tests/'+id,{method:'DELETE'});refresh();}
function fmt(n){if(n>=1e6)return(n/1e6).toFixed(1)+'M';if(n>=1e3)return(n/1e3).toFixed(1)+'K';return n;}
function esc(s){const d=document.createElement('div');d.textContent=s||'';return d.innerHTML;}
function timeAgo(s){if(!s)return'—';const d=new Date(s);const diff=Date.now()-d.getTime();if(diff<60000)return'now';if(diff<3600000)return Math.floor(diff/60000)+'m';if(diff<86400000)return Math.floor(diff/3600000)+'h';return Math.floor(diff/86400000)+'d';}
refresh();setInterval(()=>{if(activeRunId)pollLive();else refresh();},activeRunId?1000:8000);
</script></body></html>`

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(uiHTML))
}
