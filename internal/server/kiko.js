(function () {
  'use strict';

  var script = document.currentScript;
  var base = '';

  // === DETERMINE THE BASE ENDPOINT (First-Party + CSP Support) ===
  if (script && script.dataset.endpoint) {
    // Priority: data-endpoint (recommended)
    base = script.dataset.endpoint.replace(/\/$/, '');
  } else if (script && script.src) {
    // Fallback: extract from the script src
    var u = script.src.split('/');
    u.pop();
    base = u.join('/');
  } else {
    // Last resort: same origin
    base = window.location.protocol + '//' + window.location.host;
  }

  var payload = function () {
    return {
      host: location.hostname,
      path: location.pathname + location.search,
      referrer: document.referrer || '',
      title: document.title,
      width: screen.width,
      height: screen.height,
      language: navigator.language || ''
    };
  };

  var send = function () {
    var d = payload();
    var url = base + '/api';

    // Attempt sendBeacon (best method)
    var blob = new Blob([JSON.stringify(d)], { type: 'application/json' });

    try {
      if (navigator.sendBeacon(url, blob)) return;
    } catch (e) {}

    // Fallback with GIF (works even under strict CSP)
    var gifUrl = base + '/api.gif?' + new URLSearchParams({
      p: d.path,
      r: d.referrer,
      t: d.title,
      w: d.width,
      h: d.host,
      l: d.language
    }).toString();

    new Image().src = gifUrl;
  };

  // Send current page
  send();

  // SPA support (Single Page Applications)
  var pushState = history.pushState;
  history.pushState = function () {
    pushState.apply(history, arguments);
    send();
  };

  addEventListener('popstate', send);

  // Hash routing support (when ?hash=1 is passed)
  if (script && /[\?&]hash=1(?:&|$)/.test(script.src)) {
    addEventListener('hashchange', send);
  }

  // Expose manual tracking function for modern SPAs
  window.kiko = window.kiko || {};
  window.kiko.track = send;

})();