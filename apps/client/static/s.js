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

  function getUTM() {
    var params = new URLSearchParams(location.search);
    return {
      utm_source: params.get('utm_source') || '',
      utm_medium: params.get('utm_medium') || '',
      utm_campaign: params.get('utm_campaign') || '',
      utm_term: params.get('utm_term') || '',
      utm_content: params.get('utm_content') || ''
    };
  }

  function send(path) {
    var utm = getUTM();
    var data = {
      hostname: location.hostname,
      path: path || location.pathname,
      referrer: document.referrer || "",
      language: navigator.language || "",
      screen_width: window.innerWidth,
      visitor_id: visitorId(),
      utm_source: utm.utm_source,
      utm_medium: utm.utm_medium,
      utm_campaign: utm.utm_campaign,
      utm_term: utm.utm_term,
      utm_content: utm.utm_content
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

  function getPerformance() {
    try {
      var nav = performance.getEntriesByType('navigation')[0];
      if (!nav) return null;
      return {
        dns: Math.round(nav.domainLookupEnd - nav.domainLookupStart),
        tcp: Math.round(nav.connectEnd - nav.connectStart),
        ttfb: Math.round(nav.responseStart - nav.requestStart),
        dom_load: Math.round(nav.domContentLoadedEventEnd - nav.startTime),
        page_load: Math.round(nav.loadEventEnd - nav.startTime)
      };
    } catch (e) { return null; }
  }

  function sendPerformance() {
    var perf = getPerformance();
    if (!perf || perf.page_load <= 0) return;
    var data = {
      hostname: location.hostname,
      path: location.pathname,
      visitor_id: visitorId(),
      performance: perf
    };
    var img = new Image();
    img.src = endpoint + "?data=" + encodeURIComponent(JSON.stringify(data)) + "&type=perf";
  }

  window.vision = {
    track: function (name, props) {
      var data = {
        hostname: location.hostname,
        path: location.pathname,
        visitor_id: visitorId(),
        event_name: name,
        event_props: props || {}
      };
      var img = new Image();
      img.src = endpoint.replace('/e/p', '/e/t') + "?data=" + encodeURIComponent(JSON.stringify(data));
    }
  };

  send();
  startHeartbeat();

  if (document.readyState === 'complete') {
    setTimeout(sendPerformance, 100);
  } else {
    window.addEventListener('load', function () {
      setTimeout(sendPerformance, 100);
    });
  }

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
