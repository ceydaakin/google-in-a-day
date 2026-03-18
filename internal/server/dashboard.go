package server

const dashboardPage = `<!DOCTYPE html>
<html>
<head><title>Dashboard - Google in a Day</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: sans-serif; background: #f5f5f5; color: #333; }
  .header { background: #fff; border-bottom: 1px solid #ddd; padding: 15px 30px; display: flex; align-items: center; justify-content: space-between; }
  .header h1 { font-size: 1.3em; }
  .header a { color: #4285f4; text-decoration: none; font-size: 14px; }
  .container { max-width: 1000px; margin: 20px auto; padding: 0 20px; }
  .card { background: #fff; border-radius: 8px; padding: 20px; margin-bottom: 16px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
  .card h2 { font-size: 1em; color: #666; margin-bottom: 12px; text-transform: uppercase; letter-spacing: 1px; }
  .metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); gap: 12px; }
  .metric { text-align: center; padding: 12px; background: #f8f9fa; border-radius: 6px; }
  .metric .value { font-size: 1.8em; font-weight: bold; color: #1a73e8; }
  .metric .label { font-size: 0.75em; color: #666; margin-top: 4px; }
  .state-badge { display: inline-block; padding: 4px 12px; border-radius: 12px; font-size: 13px; font-weight: bold; text-transform: uppercase; }
  .state-idle { background: #e8eaed; color: #5f6368; }
  .state-running { background: #e6f4ea; color: #137333; }
  .state-paused { background: #fef7e0; color: #b06000; }
  .state-stopped { background: #fce8e6; color: #c5221f; }
  .state-completed { background: #e8f0fe; color: #1967d2; }
  .controls { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 16px; }
  .btn { padding: 8px 20px; font-size: 13px; border: none; border-radius: 4px; cursor: pointer; font-weight: 500; }
  .btn:disabled { opacity: 0.4; cursor: not-allowed; }
  .btn-start { background: #1a73e8; color: #fff; }
  .btn-pause { background: #f9ab00; color: #fff; }
  .btn-resume { background: #34a853; color: #fff; }
  .btn-stop { background: #ea4335; color: #fff; }
  .form-row { display: flex; gap: 8px; margin-bottom: 10px; flex-wrap: wrap; align-items: end; }
  .form-group { display: flex; flex-direction: column; }
  .form-group label { font-size: 11px; color: #666; margin-bottom: 3px; }
  .form-group input { padding: 7px 10px; border: 1px solid #ddd; border-radius: 4px; font-size: 13px; }
  .form-group input[type=text] { width: 280px; }
  .form-group input[type=number] { width: 70px; }
  table { width: 100%; border-collapse: collapse; font-size: 13px; }
  th { text-align: left; padding: 8px; color: #666; border-bottom: 2px solid #eee; font-weight: 500; }
  td { padding: 8px; border-bottom: 1px solid #f0f0f0; }
  .history-list { max-height: 300px; overflow-y: auto; }
  .error-text { color: #ea4335; }
  .success-text { color: #34a853; }
  #message { padding: 8px 16px; border-radius: 4px; margin-bottom: 10px; display: none; font-size: 13px; }
</style>
</head>
<body>
  <div class="header">
    <h1>Dashboard</h1>
    <a href="/">Search</a>
  </div>
  <div class="container">
    <div id="message"></div>

    <div class="card">
      <h2>Start Crawl</h2>
      <div class="form-row">
        <div class="form-group">
          <label>Seed URL</label>
          <input type="text" id="seedUrl" placeholder="https://example.com" value="">
        </div>
        <div class="form-group">
          <label>Depth</label>
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
      <div style="margin-bottom: 12px;">
        State: <span id="crawlState" class="state-badge state-idle">idle</span>
        <span id="seedDisplay" style="margin-left: 12px; font-size: 13px; color: #666;"></span>
      </div>
      <div class="controls">
        <button class="btn btn-pause" id="btnPause" onclick="pauseCrawl()" disabled>Pause</button>
        <button class="btn btn-resume" id="btnResume" onclick="resumeCrawl()" disabled>Resume</button>
        <button class="btn btn-stop" id="btnStop" onclick="stopCrawl()" disabled>Stop</button>
      </div>
      <div class="metrics">
        <div class="metric"><div class="value" id="urlsProcessed">0</div><div class="label">Processed</div></div>
        <div class="metric"><div class="value" id="urlsQueued">0</div><div class="label">Queued</div></div>
        <div class="metric"><div class="value" id="urlsDropped">0</div><div class="label">Dropped</div></div>
        <div class="metric"><div class="value" id="urlsErrored">0</div><div class="label">Errors</div></div>
        <div class="metric"><div class="value" id="activeWorkers">0</div><div class="label">Active Workers</div></div>
        <div class="metric"><div class="value" id="backPressures">0</div><div class="label">Back Pressure Events</div></div>
        <div class="metric"><div class="value" id="docCount">0</div><div class="label">Indexed Pages</div></div>
        <div class="metric"><div class="value" id="keywordCount">0</div><div class="label">Keywords</div></div>
      </div>
    </div>

    <div class="card">
      <h2>Workers</h2>
      <table>
        <thead><tr><th>ID</th><th>State</th><th>URL</th></tr></thead>
        <tbody id="workerTable"></tbody>
      </table>
    </div>

    <div class="card">
      <h2>Recent Crawl History</h2>
      <div class="history-list">
        <table>
          <thead><tr><th>URL</th><th>Status</th><th>Duration</th><th>Time</th></tr></thead>
          <tbody id="historyTable"></tbody>
        </table>
      </div>
    </div>
  </div>

<script>
var refreshInterval = null;

function showMsg(text, isError) {
  var el = document.getElementById("message");
  el.textContent = text;
  el.style.display = "block";
  el.style.background = isError ? "#fce8e6" : "#e6f4ea";
  el.style.color = isError ? "#c5221f" : "#137333";
  setTimeout(function() { el.style.display = "none"; }, 4000);
}

function startCrawl() {
  var body = {
    url: document.getElementById("seedUrl").value,
    depth: parseInt(document.getElementById("maxDepth").value),
    workers: parseInt(document.getElementById("workers").value),
    queue_size: parseInt(document.getElementById("queueSize").value)
  };
  fetch("/api/index", { method: "POST", headers: {"Content-Type": "application/json"}, body: JSON.stringify(body) })
    .then(function(r) { return r.json(); })
    .then(function(d) {
      if (d.success) { showMsg("Crawl started: " + body.url, false); }
      else { showMsg("Error: " + d.error, true); }
    })
    .catch(function(e) { showMsg("Request failed: " + e, true); });
}

function pauseCrawl() {
  fetch("/api/pause", { method: "POST" })
    .then(function(r) { return r.json(); })
    .then(function(d) {
      if (d.success) { showMsg("Crawler paused", false); }
      else { showMsg("Error: " + d.error, true); }
    });
}

function resumeCrawl() {
  fetch("/api/resume", { method: "POST" })
    .then(function(r) { return r.json(); })
    .then(function(d) {
      if (d.success) { showMsg("Crawler resumed", false); }
      else { showMsg("Error: " + d.error, true); }
    });
}

function stopCrawl() {
  fetch("/api/stop", { method: "POST" })
    .then(function(r) { return r.json(); })
    .then(function(d) {
      if (d.success) { showMsg("Crawler stopped", false); }
      else { showMsg("Error: " + d.error, true); }
    });
}

function updateDashboard() {
  fetch("/api/status")
    .then(function(r) { return r.json(); })
    .then(function(d) {
      var state = d.state || "idle";
      var badge = document.getElementById("crawlState");
      badge.textContent = state;
      badge.className = "state-badge state-" + state;

      document.getElementById("seedDisplay").textContent = d.seed_url ? "Seed: " + d.seed_url : "";
      document.getElementById("urlsProcessed").textContent = d.metrics.urls_processed || 0;
      document.getElementById("urlsQueued").textContent = d.metrics.urls_queued || 0;
      document.getElementById("urlsDropped").textContent = d.metrics.urls_dropped || 0;
      document.getElementById("urlsErrored").textContent = d.metrics.urls_errored || 0;
      document.getElementById("activeWorkers").textContent = d.metrics.active_workers || 0;
      document.getElementById("backPressures").textContent = d.metrics.back_pressures || 0;
      document.getElementById("docCount").textContent = d.docs || 0;
      document.getElementById("keywordCount").textContent = d.keywords || 0;

      // Update control buttons
      var isRunning = state === "running";
      var isPaused = state === "paused";
      document.getElementById("btnStart").disabled = isRunning || isPaused;
      document.getElementById("btnPause").disabled = !isRunning;
      document.getElementById("btnResume").disabled = !isPaused;
      document.getElementById("btnStop").disabled = !isRunning && !isPaused;

      // Workers table
      var wt = document.getElementById("workerTable");
      var workers = d.metrics.workers || [];
      var whtml = "";
      for (var i = 0; i < workers.length; i++) {
        var w = workers[i];
        var urlDisplay = w.url ? (w.url.length > 60 ? w.url.substring(0,60) + "..." : w.url) : "-";
        whtml += "<tr><td>" + w.id + "</td><td>" + w.state + "</td><td style='font-size:12px;word-break:break-all;'>" + urlDisplay + "</td></tr>";
      }
      wt.innerHTML = whtml;

      // History table
      var ht = document.getElementById("historyTable");
      var history = d.metrics.history || [];
      var hhtml = "";
      for (var j = history.length - 1; j >= Math.max(0, history.length - 20); j--) {
        var h = history[j];
        var url = h.url || "";
        var shortUrl = url.length > 50 ? url.substring(0,50) + "..." : url;
        var statusText = h.error ? "<span class='error-text'>" + (h.error.length > 30 ? h.error.substring(0,30) + "..." : h.error) + "</span>" : "<span class='success-text'>" + (h.status_code || 200) + "</span>";
        var dur = h.duration_ms ? (h.duration_ms / 1000000).toFixed(0) + "ms" : "-";
        var time = h.timestamp ? new Date(h.timestamp).toLocaleTimeString() : "-";
        hhtml += "<tr><td style='font-size:12px;word-break:break-all;'>" + shortUrl + "</td><td>" + statusText + "</td><td>" + dur + "</td><td>" + time + "</td></tr>";
      }
      ht.innerHTML = hhtml;
    })
    .catch(function(e) { console.error("Status fetch failed:", e); });
}

// Start polling
updateDashboard();
refreshInterval = setInterval(updateDashboard, 2000);
</script>
</body>
</html>`
