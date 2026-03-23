package server

const dashboardPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Dashboard - Google in a Day</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
  <div class="header">
    <div class="header-left">
      <h1 class="logo" style="font-size:1.15em">Google in a Day</h1>
    </div>
    <a href="/" class="nav-link">Search</a>
  </div>

  <div class="container fade-in">
    <div id="message" class="toast"></div>

    <!-- Start Crawl -->
    <div class="hero">
      <div class="hero-blob hero-blob-1"></div>
      <div class="hero-blob hero-blob-2"></div>
      <h2 class="card-title">Start Crawl</h2>
      <div class="form-row">
        <div class="form-group">
          <label>Origin URL</label>
          <input class="input-url" type="text" id="seedUrl" placeholder="https://example.com">
        </div>
        <div class="form-group">
          <label>k (max depth)</label>
          <input class="input-num" type="number" id="maxDepth" value="2" min="0" max="10">
        </div>
        <div class="form-group">
          <label>Workers</label>
          <input class="input-num" type="number" id="workers" value="5" min="1" max="50">
        </div>
        <div class="form-group">
          <label>Queue</label>
          <input class="input-num" type="number" id="queueSize" value="100" min="10" max="10000">
        </div>
        <button class="btn btn-start" id="btnStart" onclick="startCrawl()">Start</button>
      </div>
    </div>

    <!-- Status -->
    <div class="card">
      <h2 class="card-title">Status</h2>
      <div class="status-row">
        <span id="crawlState" class="state-badge state-idle"><span class="pulse"></span> idle</span>
        <span id="seedDisplay" class="seed-url"></span>
      </div>
      <div class="progress-wrap">
        <div class="progress-bar" id="progressBar"></div>
      </div>
      <div class="controls" style="margin-top:16px">
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

    <!-- Workers & History -->
    <div class="two-col">
      <div class="card">
        <h2 class="card-title">Workers</h2>
        <table>
          <thead><tr><th>ID</th><th>State</th><th>URL</th></tr></thead>
          <tbody id="workerTable"></tbody>
        </table>
      </div>
      <div class="card">
        <h2 class="card-title">Crawl History</h2>
        <div class="history-list">
          <table>
            <thead><tr><th>URL</th><th>Status</th><th>Duration</th><th>Time</th></tr></thead>
            <tbody id="historyTable"></tbody>
          </table>
        </div>
      </div>
    </div>
  </div>

  <script src="/static/js/dashboard.js"></script>
</body>
</html>`
