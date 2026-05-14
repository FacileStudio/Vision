(function () {
  var src = document.currentScript.src;
  var match = src.match(/\/t\/([^\/]+)\.js/);
  if (!match) return;

  var apiKey = match[1];
  var apiUrl = src.replace(/\/t\/[^\/]+\.js.*$/, "");

  function visitorId() {
    var key = "_vs_id";
    var id = localStorage.getItem(key);
    if (id) return id;
    id = Math.random().toString(36).substring(2) + Date.now().toString(36);
    try { localStorage.setItem(key, id); } catch (e) {}
    return id;
  }

  function send(path) {
    var body = JSON.stringify({
      path: path || location.pathname,
      referrer: document.referrer || "",
      language: navigator.language || "",
      visitor_id: visitorId()
    });

    if (navigator.sendBeacon) {
      var blob = new Blob([body], { type: "application/json" });
      navigator.sendBeacon(apiUrl + "/event/pageview?key=" + apiKey, blob);
    } else {
      var xhr = new XMLHttpRequest();
      xhr.open("POST", apiUrl + "/event/pageview?key=" + apiKey);
      xhr.setRequestHeader("Content-Type", "application/json");
      xhr.send(body);
    }
  }

  send();

  var pushState = history.pushState;
  history.pushState = function () {
    pushState.apply(history, arguments);
    send();
  };

  window.addEventListener("popstate", function () {
    send();
  });
})();
