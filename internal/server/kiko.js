(function(){
  var d = {
    host: location.hostname,
    path: location.pathname + location.search,
    referrer: document.referrer || '',
    title: document.title,
    width: screen.width
  };
  var b = new Blob([JSON.stringify(d)], {type:'application/json'});
  try {
    if (!navigator.sendBeacon('/hit', b)) throw 0;
  } catch(e) {
    (new Image()).src = '/hit.gif?p=' + encodeURIComponent(d.path) +
      '&r=' + encodeURIComponent(d.referrer) +
      '&t=' + encodeURIComponent(d.title) +
      '&w=' + d.width +
      '&h=' + encodeURIComponent(d.host);
  }
})();
