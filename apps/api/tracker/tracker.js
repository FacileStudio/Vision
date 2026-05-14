(function () {
  var src = document.currentScript.src;
  var apiUrl = src.replace(/\/t\.js.*$/, "/api");

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
    var url = apiUrl + "/event/pageview";
    var body = JSON.stringify(data);

    try {
      fetch(url, {
        method: "POST",
        body: body,
        mode: "no-cors",
        keepalive: true,
        headers: { "Content-Type": "text/plain" }
      });
    } catch (e) {
      new Image().src = url + "?data=" + encodeURIComponent(body);
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
