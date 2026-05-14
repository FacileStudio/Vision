(function () {
  var script = document.currentScript;
  var endpoint = script.getAttribute("data-api") ||
    script.src.replace(/\/t\.js.*$/, "/api/event/pageview");

  function visitorId() {
    var key = "_vs_id";
    var id = localStorage.getItem(key);
    if (id) return id;
    id = Math.random().toString(36).substring(2) + Date.now().toString(36);
    try { localStorage.setItem(key, id); } catch (e) {}
    return id;
  }

  function send(path) {
    var data = {
      path: path || location.pathname,
      referrer: document.referrer || "",
      language: navigator.language || "",
      visitor_id: visitorId()
    };
    var img = new Image();
    img.src = endpoint + "?data=" + encodeURIComponent(JSON.stringify(data));
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
