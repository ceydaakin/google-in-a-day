// Dashboard — lifecycle controls and real-time metrics polling.
(function () {
  "use strict";

  // --- Toast notifications ---
  function showMsg(text, isError) {
    var el = document.getElementById("message");
    el.textContent = text;
    el.style.display = "block";
    el.className = "toast " + (isError ? "toast-error" : "toast-success");
    setTimeout(function () { el.style.display = "none"; }, 4000);
  }

  // --- API helpers ---
  function postJSON(url, body) {
    var opts = { method: "POST", headers: { "Content-Type": "application/json" } };
    if (body) { opts.body = JSON.stringify(body); }
    return fetch(url, opts).then(function (r) { return r.json(); });
  }

  function postAction(url, successMsg) {
    postJSON(url)
      .then(function (d) {
        if (d.success) { showMsg(successMsg, false); }
        else { showMsg(d.error, true); }
      })
      .catch(function (e) { showMsg("Request failed: " + e, true); });
  }

  // --- Lifecycle controls ---
  window.startCrawl = function () {
    var origin = document.getElementById("seedUrl").value;
    var k = parseInt(document.getElementById("maxDepth").value, 10);
    postJSON("/api/index", { origin: origin, k: k })
      .then(function (d) {
        if (d.success) { showMsg("Crawl started: " + origin, false); }
        else { showMsg(d.error, true); }
      })
      .catch(function (e) { showMsg("Request failed: " + e, true); });
  };

  window.pauseCrawl = function () { postAction("/api/pause", "Crawler paused"); };
  window.resumeCrawl = function () { postAction("/api/resume", "Crawler resumed"); };
  window.stopCrawl = function () { postAction("/api/stop", "Crawler stopped"); };

  // --- Truncate helper ---
  function trunc(str, max) {
    return str.length > max ? str.substring(0, max) + "\u2026" : str;
  }

  // --- Dashboard update ---
  function updateDashboard() {
    fetch("/api/status")
      .then(function (r) { return r.json(); })
      .then(function (d) {
        var state = d.state || "idle";

        // State badge
        var badge = document.getElementById("crawlState");
        badge.className = "state-badge state-" + state;
        badge.innerHTML = '<span class="pulse"></span> ' + state;

        // Seed URL display
        document.getElementById("seedDisplay").textContent =
          d.seed_url ? "Origin: " + d.seed_url : "";

        // Metrics
        var m = d.metrics || {};
        document.getElementById("urlsProcessed").textContent = m.urls_processed || 0;
        document.getElementById("urlsQueued").textContent = m.urls_queued || 0;
        document.getElementById("urlsDropped").textContent = m.urls_dropped || 0;
        document.getElementById("urlsErrored").textContent = m.urls_errored || 0;
        document.getElementById("activeWorkers").textContent = m.active_workers || 0;
        document.getElementById("backPressures").textContent = m.back_pressures || 0;
        document.getElementById("docCount").textContent = d.docs || 0;
        document.getElementById("keywordCount").textContent = d.keywords || 0;

        // Progress bar
        var processed = m.urls_processed || 0;
        var queued = m.urls_queued || 0;
        var total = processed + queued;
        var pct = total > 0 ? Math.round((processed / total) * 100) : 0;
        var bar = document.getElementById("progressBar");
        if (bar) {
          bar.style.width = pct + "%";
        }

        // Button states
        var isRunning = state === "running";
        var isPaused = state === "paused";
        document.getElementById("btnStart").disabled = isRunning || isPaused;
        document.getElementById("btnPause").disabled = !isRunning;
        document.getElementById("btnResume").disabled = !isPaused;
        document.getElementById("btnStop").disabled = !isRunning && !isPaused;

        // Workers table
        var workers = m.workers || [];
        var wh = "";
        for (var i = 0; i < workers.length; i++) {
          var w = workers[i];
          var cls = "ws ws-" + w.state;
          var u = w.url ? trunc(w.url, 35) : "-";
          wh += "<tr><td style='color:var(--text-muted)'>" + w.id +
            "</td><td><span class='" + cls + "'>" + w.state +
            "</span></td><td style='font-size:12px;color:var(--text-dim);word-break:break-all'>" +
            u + "</td></tr>";
        }
        document.getElementById("workerTable").innerHTML = wh;

        // History table
        var hist = m.history || [];
        var hh = "";
        var start = Math.max(0, hist.length - 20);
        for (var j = hist.length - 1; j >= start; j--) {
          var h = hist[j];
          var url = h.url || "";
          var su = trunc(url, 40);
          var st = h.error
            ? "<span class='status-err'>" + trunc(h.error, 20) + "</span>"
            : "<span class='status-ok'>" + (h.status_code || 200) + "</span>";
          var dur = h.duration_ms
            ? (h.duration_ms / 1000000).toFixed(0) + "ms"
            : "-";
          var tm = h.timestamp
            ? new Date(h.timestamp).toLocaleTimeString()
            : "-";
          hh += "<tr><td style='font-size:12px;word-break:break-all;color:var(--text-dim)'>" +
            su + "</td><td>" + st + "</td><td style='color:var(--text-muted)'>" +
            dur + "</td><td style='color:var(--text-muted)'>" + tm + "</td></tr>";
        }
        document.getElementById("historyTable").innerHTML = hh;
      })
      .catch(function (e) {
        console.error("Status fetch failed:", e);
      });
  }

  // Initial fetch + polling every 2 seconds
  updateDashboard();
  setInterval(updateDashboard, 2000);
})();
