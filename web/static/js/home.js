// Home page — fetch crawler status and update the stats bar.
(function () {
  "use strict";

  fetch("/api/status")
    .then(function (r) { return r.json(); })
    .then(function (d) {
      var docs = d.docs || 0;
      var kw = d.keywords || 0;
      var state = d.state || "idle";
      var dotClass = state === "running" ? "dot-running" : "dot-idle";

      document.getElementById("statsBar").innerHTML =
        '<span><span class="dot ' + dotClass + '"></span>' + state.toUpperCase() + "</span>" +
        "<span>" + docs + " pages indexed</span>" +
        "<span>" + kw + " keywords</span>";
    })
    .catch(function () {
      // Silently ignore — stats bar stays in loading state.
    });
})();
