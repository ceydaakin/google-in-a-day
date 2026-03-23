package server

const homePage = `<!DOCTYPE html>
<html>
<head><title>Google in a Day</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: hsl(0,0%,3.9%); color: hsl(0,0%,98%); min-height: 100vh; display: flex; flex-direction: column; align-items: center; justify-content: center; overflow: hidden; position: relative; }

  /* Decorative blurs */
  .blob1 { position: fixed; top: -80px; right: -80px; width: 300px; height: 300px; background: linear-gradient(135deg, rgba(138,43,226,0.15), rgba(255,20,147,0.1)); border-radius: 50%; filter: blur(80px); pointer-events: none; }
  .blob2 { position: fixed; bottom: -60px; left: -60px; width: 250px; height: 250px; background: linear-gradient(135deg, rgba(255,20,147,0.1), rgba(255,165,0,0.08)); border-radius: 50%; filter: blur(60px); pointer-events: none; }

  .container { text-align: center; padding: 40px 20px; position: relative; z-index: 1; }
  h1 { font-size: 3.2em; font-weight: 700; background: linear-gradient(135deg, #8A2BE2, #FF1493); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; margin-bottom: 8px; }
  .subtitle { color: hsl(0,0%,63.9%); font-size: 15px; margin-bottom: 44px; letter-spacing: 0.3px; }

  .search-box { position: relative; width: 100%; max-width: 540px; margin: 0 auto; }
  .search-box input { width: 100%; padding: 16px 24px; font-size: 16px; border: 1px solid hsl(0,0%,14.9%); border-radius: 12px; background: hsl(0,0%,6%); color: hsl(0,0%,98%); outline: none; transition: all 0.3s ease; backdrop-filter: blur(8px); }
  .search-box input:focus { border-color: rgba(138,43,226,0.5); box-shadow: 0 0 0 3px rgba(138,43,226,0.1), 0 8px 32px rgba(138,43,226,0.1); }
  .search-box input::placeholder { color: hsl(0,0%,45%); }

  .buttons { margin-top: 28px; display: flex; gap: 12px; justify-content: center; }
  .btn-primary { padding: 11px 28px; font-size: 14px; font-weight: 500; border: none; border-radius: 8px; background: linear-gradient(135deg, #8A2BE2, #FF1493); color: #fff; cursor: pointer; transition: all 0.2s ease; box-shadow: 0 4px 16px rgba(138,43,226,0.25); }
  .btn-primary:hover { transform: translateY(-2px); box-shadow: 0 6px 24px rgba(138,43,226,0.35); }
  .btn-secondary { padding: 11px 28px; font-size: 14px; font-weight: 500; border: 1px solid hsl(0,0%,14.9%); border-radius: 8px; background: hsl(0,0%,3.9%); color: hsl(0,0%,98%); cursor: pointer; transition: all 0.2s ease; }
  .btn-secondary:hover { border-color: rgba(138,43,226,0.4); background: hsl(0,0%,6%); }

  .stats-bar { position: fixed; bottom: 0; left: 0; right: 0; background: hsl(0,0%,3.9%); border-top: 1px solid hsl(0,0%,14.9%); padding: 10px 20px; display: flex; justify-content: center; gap: 28px; font-size: 12px; color: hsl(0,0%,45%); z-index: 2; }
  .stats-bar .dot { width: 7px; height: 7px; border-radius: 50%; display: inline-block; margin-right: 6px; vertical-align: middle; }
  .dot-running { background: #34a853; box-shadow: 0 0 8px rgba(52,168,83,0.5); }
  .dot-idle { background: hsl(0,0%,45%); }
</style>
</head>
<body>
  <div class="blob1"></div>
  <div class="blob2"></div>
  <div class="container">
    <h1>Google in a Day</h1>
    <p class="subtitle">Concurrent Web Crawler & Search Engine</p>
    <form action="/search" method="GET">
      <div class="search-box">
        <input type="text" name="q" placeholder="Search the crawled web..." autofocus>
      </div>
      <div class="buttons">
        <button type="submit" class="btn-primary">Search</button>
        <button type="button" class="btn-secondary" onclick="window.location='/dashboard'">Dashboard</button>
      </div>
    </form>
  </div>
  <div class="stats-bar" id="statsBar">
    <span><span class="dot dot-idle"></span> Loading...</span>
  </div>
  <script>
    fetch("/api/status").then(function(r){return r.json();}).then(function(d){
      var docs=d.docs||0, kw=d.keywords||0, state=d.state||"idle";
      var dotClass=state==="running"?"dot-running":"dot-idle";
      document.getElementById("statsBar").innerHTML=
        '<span><span class="dot '+dotClass+'"></span>'+state.toUpperCase()+'</span>'+
        '<span>'+docs+' pages indexed</span>'+
        '<span>'+kw+' keywords</span>';
    }).catch(function(){});
  </script>
</body>
</html>`

const resultsPage = `<!DOCTYPE html>
<html>
<head><title>{{.Query}} - Search</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: hsl(0,0%,3.9%); color: hsl(0,0%,98%); min-height: 100vh; }
  .header { background: hsl(0,0%,3.9%); border-bottom: 1px solid hsl(0,0%,14.9%); padding: 14px 24px; display: flex; align-items: center; gap: 20px; position: sticky; top: 0; z-index: 10; backdrop-filter: blur(12px); }
  .logo { font-size: 1.2em; font-weight: 700; background: linear-gradient(135deg, #8A2BE2, #FF1493); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; text-decoration: none; white-space: nowrap; }
  .header form { flex: 1; max-width: 520px; }
  .header input { width: 100%; padding: 10px 18px; font-size: 14px; border: 1px solid hsl(0,0%,14.9%); border-radius: 10px; background: hsl(0,0%,6%); color: hsl(0,0%,98%); outline: none; transition: border 0.2s; }
  .header input:focus { border-color: rgba(138,43,226,0.5); }
  .header button { display: none; }
  .header .dash-link { color: hsl(0,0%,63.9%); text-decoration: none; font-size: 13px; padding: 6px 14px; border: 1px solid hsl(0,0%,14.9%); border-radius: 8px; transition: all 0.2s; white-space: nowrap; }
  .header .dash-link:hover { border-color: rgba(138,43,226,0.4); }

  .content { max-width: 700px; margin: 0 auto; padding: 24px; }
  .stats { color: hsl(0,0%,45%); font-size: 13px; margin-bottom: 28px; padding-bottom: 16px; border-bottom: 1px solid hsl(0,0%,14.9%); }

  .result { padding: 16px; margin-bottom: 12px; border-radius: 12px; border: 1px solid hsl(0,0%,14.9%); background: hsl(0,0%,6%); transition: all 0.2s ease; }
  .result:hover { border-color: rgba(138,43,226,0.3); box-shadow: 0 4px 20px rgba(138,43,226,0.08); transform: translateY(-1px); }
  .result .url { color: hsl(0,0%,63.9%); font-size: 12px; margin-bottom: 6px; word-break: break-all; }
  .result .title a { color: hsl(0,0%,98%); text-decoration: none; font-size: 17px; font-weight: 500; line-height: 1.4; }
  .result .title a:hover { color: #c084fc; }
  .result .meta { margin-top: 10px; display: flex; gap: 8px; flex-wrap: wrap; }
  .result .tag { background: hsl(0,0%,9%); border: 1px solid hsl(0,0%,14.9%); padding: 3px 10px; border-radius: 6px; font-size: 11px; color: hsl(0,0%,63.9%); }
  .tag-score { border-color: rgba(138,43,226,0.3); color: #c084fc; }

  .no-results { color: hsl(0,0%,45%); font-size: 16px; text-align: center; padding: 60px 0; }
  .nav { margin-top: 32px; padding-top: 16px; border-top: 1px solid hsl(0,0%,14.9%); }
  .nav a { color: hsl(0,0%,63.9%); text-decoration: none; font-size: 13px; }
  .nav a:hover { color: #c084fc; }
</style>
</head>
<body>
  <div class="header">
    <a href="/" class="logo">Google in a Day</a>
    <form action="/search" method="GET">
      <input type="text" name="q" value="{{.Query}}">
      <button type="submit">Search</button>
    </form>
    <a href="/dashboard" class="dash-link">Dashboard</a>
  </div>
  <div class="content">
    <div class="stats">{{.Count}} results &middot; {{.Docs}} pages indexed &middot; {{.Keywords}} keywords</div>
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
