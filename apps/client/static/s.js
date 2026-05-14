(function () {
  var base = document.currentScript.src.replace(/\/s\.js.*$/, "/api/e");
  var endpoint = base + "/p";
  var heartbeatEndpoint = base + "/h";

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
      hostname: location.hostname,
      path: path || location.pathname,
      referrer: document.referrer || "",
      language: navigator.language || "",
      screen_width: window.innerWidth,
      visitor_id: visitorId()
    };
    var img = new Image();
    img.src = endpoint + "?data=" + encodeURIComponent(JSON.stringify(data));
  }

  function heartbeat() {
    var data = {
      hostname: location.hostname,
      visitor_id: visitorId()
    };
    var img = new Image();
    img.src = heartbeatEndpoint + "?data=" + encodeURIComponent(JSON.stringify(data));
  }

  var hbTimer = null;

  function startHeartbeat() {
    if (hbTimer) return;
    heartbeat();
    hbTimer = setInterval(heartbeat, 30000);
  }

  function stopHeartbeat() {
    if (hbTimer) {
      clearInterval(hbTimer);
      hbTimer = null;
    }
  }

  send();
  startHeartbeat();

  document.addEventListener("visibilitychange", function () {
    if (document.hidden) {
      stopHeartbeat();
    } else {
      startHeartbeat();
    }
  });

  var pushState = history.pushState;
  history.pushState = function () {
    pushState.apply(history, arguments);
    send();
  };

  window.addEventListener("popstate", function () {
    send();
  });
})();
