package server

const homePage = `<!DOCTYPE html>
<html>
<head><title>Google in a Day</title>
<style>
  body { font-family: sans-serif; max-width: 700px; margin: 80px auto; text-align: center; }
  h1 { font-size: 2.5em; margin-bottom: 10px; }
  .subtitle { color: #666; margin-bottom: 30px; }
  input[type=text] { width: 400px; padding: 12px; font-size: 16px; border: 1px solid #ddd; border-radius: 24px; outline: none; }
  input[type=text]:focus { border-color: #4285f4; }
  button { padding: 12px 24px; margin: 20px 8px; font-size: 14px; border: 1px solid #ddd; border-radius: 4px; background: #f8f8f8; cursor: pointer; }
  button:hover { border-color: #999; }
  .nav { margin-top: 30px; }
  .nav a { color: #4285f4; text-decoration: none; }
  .nav a:hover { text-decoration: underline; }
</style>
</head>
<body>
  <h1>Google in a Day</h1>
  <p class="subtitle">Concurrent Web Crawler &amp; Search Engine</p>
  <form action="/search" method="GET">
    <input type="text" name="q" placeholder="Search the crawled web..." autofocus>
    <br>
    <button type="submit">Search</button>
  </form>
  <div class="nav"><a href="/dashboard">Dashboard</a></div>
</body>
</html>`

const resultsPage = `<!DOCTYPE html>
<html>
<head><title>{{.Query}} - Search</title>
<style>
  body { font-family: sans-serif; max-width: 700px; margin: 20px auto; }
  .header { margin-bottom: 20px; }
  .header a { text-decoration: none; font-size: 1.5em; font-weight: bold; }
  .header form { display: inline; margin-left: 20px; }
  .header input[type=text] { width: 350px; padding: 8px; font-size: 14px; border: 1px solid #ddd; border-radius: 20px; }
  .stats { color: #666; font-size: 13px; margin-bottom: 20px; }
  .result { margin-bottom: 20px; }
  .result .url { color: #006621; font-size: 13px; }
  .result .title { font-size: 18px; }
  .result .title a { color: #1a0dab; text-decoration: none; }
  .result .title a:hover { text-decoration: underline; }
  .result .meta { color: #666; font-size: 12px; }
  .no-results { color: #666; font-size: 16px; }
  .nav { margin-top: 20px; }
  .nav a { color: #4285f4; text-decoration: none; font-size: 13px; }
</style>
</head>
<body>
  <div class="header">
    <a href="/">Google in a Day</a>
    <form action="/search" method="GET">
      <input type="text" name="q" value="{{.Query}}">
      <button type="submit">Search</button>
    </form>
  </div>
  <div class="stats">{{.Count}} results found ({{.Docs}} pages indexed, {{.Keywords}} keywords)</div>
  {{if .Results}}
    {{range .Results}}
    <div class="result">
      <div class="url">{{.RelevantURL}}</div>
      <div class="title"><a href="{{.RelevantURL}}">{{if .Title}}{{.Title}}{{else}}{{.RelevantURL}}{{end}}</a></div>
      <div class="meta">Origin: {{.OriginURL}} | Depth: {{.Depth}} | Score: {{printf "%.1f" .Score}}</div>
    </div>
    {{end}}
  {{else}}
    <div class="no-results">No results found for "{{.Query}}"</div>
  {{end}}
  <div class="nav"><a href="/dashboard">Dashboard</a></div>
</body>
</html>`
