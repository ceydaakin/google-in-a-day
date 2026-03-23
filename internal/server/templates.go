package server

const homePage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Google in a Day</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="/static/css/style.css">
</head>
<body class="home-center">
  <div class="blob blob-1"></div>
  <div class="blob blob-2"></div>
  <div class="home-content fade-in">
    <h1 class="logo">Google in a Day</h1>
    <p class="subtitle">Concurrent Web Crawler &amp; Search Engine</p>
    <form action="/search" method="GET">
      <div class="search-box">
        <input class="search-input" type="text" name="q" placeholder="Search the crawled web..." autofocus>
      </div>
      <div class="btn-group" style="margin-top:28px">
        <button type="submit" class="btn btn-primary">Search</button>
        <button type="button" class="btn btn-secondary" onclick="window.location='/dashboard'">Dashboard</button>
      </div>
    </form>
  </div>
  <div class="stats-bar" id="statsBar">
    <span><span class="dot dot-idle"></span> Loading...</span>
  </div>
  <script src="/static/js/home.js"></script>
</body>
</html>`

const resultsPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>{{.Query}} - Search</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
  <div class="header">
    <a href="/" class="logo">Google in a Day</a>
    <form action="/search" method="GET">
      <input class="header-input" type="text" name="q" value="{{.Query}}">
      <button type="submit">Search</button>
    </form>
    <a href="/dashboard" class="nav-link">Dashboard</a>
  </div>
  <div class="results-content fade-in">
    <div class="results-stats">{{.Count}} results &middot; {{.Docs}} pages indexed &middot; {{.Keywords}} keywords</div>
    {{if .Results}}
      {{range .Results}}
      <div class="result">
        <div class="url">{{.RelevantURL}}</div>
        <div class="title"><a href="{{.RelevantURL}}">{{if .Title}}{{.Title}}{{else}}{{.RelevantURL}}{{end}}</a></div>
        <div class="meta">
          <span class="tag tag-score">Score {{printf "%.0f" .Score}}</span>
          <span class="tag">Origin: {{.OriginURL}}</span>
          <span class="tag">k = {{.Depth}}</span>
        </div>
      </div>
      {{end}}
    {{else}}
      <div class="no-results">No results found for "{{.Query}}"</div>
    {{end}}
  </div>
</body>
</html>`
