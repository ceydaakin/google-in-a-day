package server

const dashboardPage = `<!DOCTYPE html>
<html>
<head><title>Dashboard - Google in a Day</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: hsl(0,0%,3.9%); color: hsl(0,0%,98%); }

  /* Header */
  .header { background: hsl(0,0%,3.9%); border-bottom: 1px solid hsl(0,0%,14.9%); padding: 14px 24px; display: flex; align-items: center; justify-content: space-between; position: sticky; top: 0; z-index: 10; backdrop-filter: blur(12px); }
  .header .left { display: flex; align-items: center; gap: 14px; }
  .header h1 { font-size: 1.15em; font-weight: 600; background: linear-gradient(135deg, #8A2BE2, #FF1493); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
  .header .nav-link { color: hsl(0,0%,63.9%); text-decoration: none; font-size: 13px; padding: 6px 16px; border: 1px solid hsl(0,0%,14.9%); border-radius: 8px; transition: all 0.2s; }
  .header .nav-link:hover { border-color: rgba(138,43,226,0.4); }

  .container { max-width: 1100px; margin: 20px auto; padding: 0 20px; }

  /* Cards */
  .card { background: hsl(0,0%,6%); border: 1px solid hsl(0,0%,14.9%); border-radius: 12px; padding: 20px 24px; margin-bottom: 16px; transition: border-color 0.3s; }
  .card:hover { border-color: hsl(0,0%,20%); }
  .card h2 { font-size: 11px; color: hsl(0,0%,45%); margin-bottom: 16px; text-transform: uppercase; letter-spacing: 1.5px; font-weight: 600; }

  /* Hero banner */
  .hero { position: relative; overflow: hidden; background: linear-gradient(135deg, rgba(138,43,226,0.08), rgba(255,20,147,0.05), rgba(255,165,0,0.03)); border-radius: 12px; padding: 28px 28px; margin-bottom: 16px; border: 1px solid hsl(0,0%,14.9%); }
  .hero h2 { font-size: 11px; color: hsl(0,0%,45%); text-transform: uppercase; letter-spacing: 1.5px; font-weight: 600; margin-bottom: 14px; }
  .hero .blob { position: absolute; border-radius: 50%; filter: blur(40px); pointer-events: none; }
  .hero .blob1 { top: -30px; right: -30px; width: 120px; height: 120px; background: rgba(138,43,226,0.12); }
  .hero .blob2 { bottom: -20px; left: -20px; width: 100px; height: 100px; background: rgba(255,20,147,0.08); }

  /* Form */
  .form-row { display: flex; gap: 10px; flex-wrap: wrap; align-items: end; position: relative; z-index: 1; }
  .form-group { display: flex; flex-direction: column; }
  .form-group label { font-size: 11px; color: hsl(0,0%,45%); margin-bottom: 5px; letter-spacing: 0.3px; }
  .form-group input { padding: 9px 12px; border: 1px solid hsl(0,0%,14.9%); border-radius: 8px; font-size: 13px; background: hsl(0,0%,3.9%); color: hsl(0,0%,98%); outline: none; transition: all 0.2s; }
  .form-group input:focus { border-color: rgba(138,43,226,0.5); box-shadow: 0 0 0 2px rgba(138,43,226,0.1); }
  .form-group input[type=text] { width: 280px; }
  .form-group input[type=number] { width: 72px; }

  /* Status */
  .status-row { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
  .state-badge { display: inline-flex; align-items: center; gap: 6px; padding: 5px 14px; border-radius: 20px; font-size: 12px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; }
  .state-badge .pulse { width: 8px; height: 8px; border-radius: 50%; }
  .state-idle { background: hsl(0,0%,9%); color: hsl(0,0%,45%); border: 1px solid hsl(0,0%,14.9%); }
  .state-idle .pulse { background: hsl(0,0%,45%); }
  .state-running { background: rgba(52,168,83,0.1); color: #81c995; border: 1px solid rgba(52,168,83,0.25); }
  .state-running .pulse { background: #34a853; box-shadow: 0 0 8px rgba(52,168,83,0.5); animation: pulse-glow 2s infinite; }
  .state-paused { background: rgba(251,188,4,0.1); color: #fdd663; border: 1px solid rgba(251,188,4,0.25); }
  .state-paused .pulse { background: #f9ab00; }
  .state-stopped { background: rgba(234,67,53,0.1); color: #f28b82; border: 1px solid rgba(234,67,53,0.25); }
  .state-stopped .pulse { background: #ea4335; }
  .state-completed { background: rgba(138,43,226,0.1); color: #c084fc; border: 1px solid rgba(138,43,226,0.25); }
  .state-completed .pulse { background: #8A2BE2; }
  @keyframes pulse-glow { 0%,100%{opacity:1} 50%{opacity:0.4} }
  .seed-url { font-size: 13px; color: hsl(0,0%,45%); }

  /* Buttons */
  .controls { display: flex; gap: 8px; margin-bottom: 18px; }
  .btn { padding: 8px 20px; font-size: 13px; border: none; border-radius: 8px; cursor: pointer; font-weight: 500; transition: all 0.2s ease; }
  .btn:hover:not(:disabled) { transform: translateY(-1px); }
  .btn:disabled { opacity: 0.25; cursor: not-allowed; transform: none !important; }
  .btn-start { background: linear-gradient(135deg, #8A2BE2, #FF1493); color: #fff; box-shadow: 0 4px 16px rgba(138,43,226,0.25); }
  .btn-start:hover:not(:disabled) { box-shadow: 0 6px 24px rgba(138,43,226,0.35); }
  .btn-pause { background: rgba(251,188,4,0.15); color: #fdd663; border: 1px solid rgba(251,188,4,0.25); }
  .btn-pause:hover:not(:disabled) { background: rgba(251,188,4,0.25); }
  .btn-resume { background: rgba(52,168,83,0.15); color: #81c995; border: 1px solid rgba(52,168,83,0.25); }
  .btn-resume:hover:not(:disabled) { background: rgba(52,168,83,0.25); }
  .btn-stop { background: rgba(234,67,53,0.15); color: #f28b82; border: 1px solid rgba(234,67,53,0.25); }
  .btn-stop:hover:not(:disabled) { background: rgba(234,67,53,0.25); }

  /* Metrics */
  .metrics { display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px; }
  @media (max-width: 700px) { .metrics { grid-template-columns: repeat(2, 1fr); } }
  .metric { text-align: center; padding: 18px 8px; background: hsl(0,0%,3.9%); border-radius: 10px; border: 1px solid hsl(0,0%,14.9%); transition: all 0.3s ease; }
  .metric:hover { border-color: hsl(0,0%,20%); transform: translateY(-2px); box-shadow: 0 4px 16px rgba(0,0,0,0.2); }
  .metric .value { font-size: 2em; font-weight: 700; line-height: 1.2; }
  .metric .label { font-size: 10px; color: hsl(0,0%,45%); margin-top: 6px; text-transform: uppercase; letter-spacing: 0.8px; font-weight: 500; }
  .metric .icon { font-size: 16px; margin-bottom: 6px; }
  .c-purple .value { background: linear-gradient(135deg, #8A2BE2, #c084fc); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
  .c-green .value { background: linear-gradient(135deg, #34a853, #81c995); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
  .c-amber .value { background: linear-gradient(135deg, #f9ab00, #fdd663); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
  .c-red .value { background: linear-gradient(135deg, #ea4335, #f28b82); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }

  /* Tables */
  .two-col { display: grid; grid-template-columns: 1fr 1.6fr; gap: 16px; }
  @media (max-width: 800px) { .two-col { grid-template-columns: 1fr; } }
  table { width: 100%; border-collapse: collapse; font-size: 13px; }
  th { text-align: left; padding: 10px 12px; color: hsl(0,0%,45%); border-bottom: 1px solid hsl(0,0%,14.9%); font-weight: 500; font-size: 11px; text-transform: uppercase; letter-spacing: 0.5px; }
  td { padding: 10px 12px; border-bottom: 1px solid hsl(0,0%,9%); }
  tr:hover td { background: hsl(0,0%,6%); }
  .history-list { max-height: 360px; overflow-y: auto; }
  .history-list::-webkit-scrollbar { width: 4px; }
  .history-list::-webkit-scrollbar-track { background: transparent; }
  .history-list::-webkit-scrollbar-thumb { background: hsl(0,0%,14.9%); border-radius: 2px; }

  .ws { display: inline-block; padding: 2px 8px; border-radius: 6px; font-size: 11px; font-weight: 500; }
  .ws-idle { background: hsl(0,0%,9%); color: hsl(0,0%,45%); }
  .ws-fetching { background: rgba(138,43,226,0.1); color: #c084fc; }
  .ws-parsing { background: rgba(52,168,83,0.1); color: #81c995; }
  .ws-paused { background: rgba(251,188,4,0.1); color: #fdd663; }
  .status-ok { color: #81c995; font-weight: 500; }
  .status-err { color: #f28b82; font-weight: 500; }

  /* Message toast */
  #message { padding: 12px 20px; border-radius: 10px; margin-bottom: 12px; display: none; font-size: 13px; font-weight: 500; backdrop-filter: blur(8px); }
</style>
</head>
<body>
  <div class="header">
    <div class="left">
      <h1>Google in a Day</h1>
    </div>
    <a href="/" class="nav-link">Search</a>
  </div>
  <div class="container">
    <div id="message"></div>

    <div class="hero">
      <div class="blob blob1"></div>
      <div class="blob blob2"></div>
      <h2>Start Crawl</h2>
      <div class="form-row">
        <div class="form-group">
          <label>Origin URL</label>
          <input type="text" id="seedUrl" placeholder="https://example.com">
        </div>
        <div class="form-group">
          <label>k (max depth)</label>
          <input type="number" id="maxDepth" value="2" min="0" max="10">
        </div>
        <div class="form-group">
          <label>Workers</label>
          <input type="number" id="workers" value="5" min="1" max="50">
        </div>
        <div class="form-group">
          <label>Queue</label>
          <input type="number" id="queueSize" value="100" min="10" max="10000">
        </div>
        <button class="btn btn-start" id="btnStart" onclick="startCrawl()">Start</button>
      </div>
    </div>

    <div class="card">
      <h2>Status</h2>
      <div class="status-row">
        <span id="crawlState" class="state-badge state-idle"><span class="pulse"></span> idle</span>
        <span id="seedDisplay" class="seed-url"></span>
      </div>
      <div class="controls">
        <button class="btn btn-pause" id="btnPause" onclick="pauseCrawl()" disabled>Pause</button>
        <button class="btn btn-resume" id="btnResume" onclick="resumeCrawl()" disabled>Resume</button>
        <button class="btn btn-stop" id="btnStop" onclick="stopCrawl()" disabled>Stop</button>
      </div>
      <div class="metrics">
        <div class="metric c-green"><div class="icon">&#9650;</div><div class="value" id="urlsProcessed">0</div><div class="label">Processed</div></div>
        <div class="metric c-purple"><div class="icon">&#9654;</div><div class="value" id="urlsQueued">0</div><div class="label">Queued</div></div>
        <div class="metric c-amber"><div class="icon">&#9660;</div><div class="value" id="urlsDropped">0</div><div class="label">Dropped</div></div>
        <div class="metric c-red"><div class="icon">&#10006;</div><div class="value" id="urlsErrored">0</div><div class="label">Errors</div></div>
        <div class="metric c-purple"><div class="icon">&#9881;</div><div class="value" id="activeWorkers">0</div><div class="label">Active Workers</div></div>
        <div class="metric c-amber"><div class="icon">&#9888;</div><div class="value" id="backPressures">0</div><div class="label">Back Pressure</div></div>
        <div class="metric c-green"><div class="icon">&#9733;</div><div class="value" id="docCount">0</div><div class="label">Indexed Pages</div></div>
        <div class="metric c-purple"><div class="icon">&#128273;</div><div class="value" id="keywordCount">0</div><div class="label">Keywords</div></div>
      </div>
    </div>

    <div class="two-col">
      <div class="card">
        <h2>Workers</h2>
        <table>
          <thead><tr><th>ID</th><th>State</th><th>URL</th></tr></thead>
          <tbody id="workerTable"></tbody>
        </table>
      </div>
      <div class="card">
        <h2>Crawl History</h2>
        <div class="history-list">
          <table>
            <thead><tr><th>URL</th><th>Status</th><th>Duration</th><th>Time</th></tr></thead>
            <tbody id="historyTable"></tbody>
          </table>
        </div>
      </div>
    </div>
  </div>

<script>
function showMsg(text, isError) {
  var el = document.getElementById("message");
  el.textContent = text;
  el.style.display = "block";
  el.style.background = isError ? "rgba(234,67,53,0.1)" : "rgba(52,168,83,0.1)";
  el.style.color = isError ? "#f28b82" : "#81c995";
  el.style.border = "1px solid " + (isError ? "rgba(234,67,53,0.25)" : "rgba(52,168,83,0.25)");
  setTimeout(function() { el.style.display = "none"; }, 4000);
}

function startCrawl() {
  var origin = document.getElementById("seedUrl").value;
  fetch("/api/index", { method: "POST", headers: {"Content-Type": "application/json"}, body: JSON.stringify({ origin: origin, k: parseInt(document.getElementById("maxDepth").value) }) })
    .then(function(r){return r.json();})
    .then(function(d){ if(d.success) showMsg("Crawl started: "+origin,false); else showMsg(d.error,true); })
    .catch(function(e){ showMsg("Request failed: "+e,true); });
}
function pauseCrawl(){ fetch("/api/pause",{method:"POST"}).then(function(r){return r.json();}).then(function(d){ if(d.success) showMsg("Crawler paused",false); else showMsg(d.error,true); }); }
function resumeCrawl(){ fetch("/api/resume",{method:"POST"}).then(function(r){return r.json();}).then(function(d){ if(d.success) showMsg("Crawler resumed",false); else showMsg(d.error,true); }); }
function stopCrawl(){ fetch("/api/stop",{method:"POST"}).then(function(r){return r.json();}).then(function(d){ if(d.success) showMsg("Crawler stopped",false); else showMsg(d.error,true); }); }

function updateDashboard() {
  fetch("/api/status").then(function(r){return r.json();}).then(function(d) {
    var state = d.state || "idle";
    var badge = document.getElementById("crawlState");
    badge.className = "state-badge state-" + state;
    badge.innerHTML = '<span class="pulse"></span> ' + state;

    document.getElementById("seedDisplay").textContent = d.seed_url ? "Origin: " + d.seed_url : "";
    document.getElementById("urlsProcessed").textContent = d.metrics.urls_processed || 0;
    document.getElementById("urlsQueued").textContent = d.metrics.urls_queued || 0;
    document.getElementById("urlsDropped").textContent = d.metrics.urls_dropped || 0;
    document.getElementById("urlsErrored").textContent = d.metrics.urls_errored || 0;
    document.getElementById("activeWorkers").textContent = d.metrics.active_workers || 0;
    document.getElementById("backPressures").textContent = d.metrics.back_pressures || 0;
    document.getElementById("docCount").textContent = d.docs || 0;
    document.getElementById("keywordCount").textContent = d.keywords || 0;

    var isRunning = state === "running", isPaused = state === "paused";
    document.getElementById("btnStart").disabled = isRunning || isPaused;
    document.getElementById("btnPause").disabled = !isRunning;
    document.getElementById("btnResume").disabled = !isPaused;
    document.getElementById("btnStop").disabled = !isRunning && !isPaused;

    var wt = document.getElementById("workerTable"), workers = d.metrics.workers || [], wh = "";
    for (var i = 0; i < workers.length; i++) {
      var w = workers[i], cls = "ws ws-" + w.state;
      var u = w.url ? (w.url.length > 35 ? w.url.substring(0,35) + "..." : w.url) : "-";
      wh += "<tr><td style='color:hsl(0,0%,45%);'>" + w.id + "</td><td><span class='" + cls + "'>" + w.state + "</span></td><td style='font-size:12px;color:hsl(0,0%,63.9%);word-break:break-all;'>" + u + "</td></tr>";
    }
    wt.innerHTML = wh;

    var ht = document.getElementById("historyTable"), hist = d.metrics.history || [], hh = "";
    for (var j = hist.length - 1; j >= Math.max(0, hist.length - 20); j--) {
      var h = hist[j], url = h.url || "", su = url.length > 40 ? url.substring(0,40) + "..." : url;
      var st = h.error ? "<span class='status-err'>" + (h.error.length > 20 ? h.error.substring(0,20) + "..." : h.error) + "</span>" : "<span class='status-ok'>" + (h.status_code || 200) + "</span>";
      var dur = h.duration_ms ? (h.duration_ms / 1000000).toFixed(0) + "ms" : "-";
      var tm = h.timestamp ? new Date(h.timestamp).toLocaleTimeString() : "-";
      hh += "<tr><td style='font-size:12px;word-break:break-all;color:hsl(0,0%,63.9%);'>" + su + "</td><td>" + st + "</td><td style='color:hsl(0,0%,45%);'>" + dur + "</td><td style='color:hsl(0,0%,45%);'>" + tm + "</td></tr>";
    }
    ht.innerHTML = hh;
  }).catch(function(e){ console.error("Status fetch failed:", e); });
}

updateDashboard();
setInterval(updateDashboard, 2000);
</script>
</body>
</html>`
