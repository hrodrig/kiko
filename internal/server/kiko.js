(function(){
  var s=document.currentScript;
  var u=s?s.src.split('/'):[];
  u.pop();
  var base=u.join('/');
  var payload=function(){
    return {
      host: location.hostname,
      path: location.pathname + location.search,
      referrer: document.referrer || '',
      title: document.title,
      width: screen.width
    };
  };
  var send=function(){
    var d=payload();
    var b=new Blob([JSON.stringify(d)],{type:'application/json'});
    try{
      if(!navigator.sendBeacon(base+'/hit',b)) throw 0;
    }catch(e){
      (new Image()).src=base+'/hit.gif?p='+encodeURIComponent(d.path)+
        '&r='+encodeURIComponent(d.referrer)+
        '&t='+encodeURIComponent(d.title)+
        '&w='+d.width+
        '&h='+encodeURIComponent(d.host);
    }
  };
  send();
  var push=history.pushState;
  history.pushState=function(){push.apply(history,arguments);send();};
  addEventListener('popstate',send);
  if(s&&/[\?&]hash=1(?:&|$)/.test(s.src)) addEventListener('hashchange',send);
})();
