(function () {
  var endpoint = document.currentScript.src.replace(/\/t\.js.*$/, "/api/event/pageview");

  function visitorId() {
    var key = "_vs_id";
    var id = localStorage.getItem(key);
    if (id) return id;
    id = Math.random().toString(36).substring(2) + Date.now().toString(36);
    try { localStorage.setItem(key, id); } catch (e) {}
    return id;
  }

  function send(path) {
    fetch(endpoint, {
      method: "POST",
      headers: { "Content-Type": "text/plain" },
      credentials: "omit",
      keepalive: true,
      body: JSON.stringify({
        path: path || location.pathname,
        referrer: document.referrer || "",
        language: navigator.language || "",
        visitor_id: visitorId()
      })
    });
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
