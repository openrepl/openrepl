// JS API for OpenREPL backend services
var openrepl = {};

// openrepl.promiseWS returns a promise that is fulfilled when the WebSocket is opened.
openrepl.promiseWS = function(url) {
    return new Promise(function(s, f) {
        var ws = new WebSocket(url);
        var opened = false;
        ws.onopen = function() {
            s(ws);
            opened = true;
        };
        ws.onerror = function() {
            if(!opened) f();
        };
    });
};

// openrepl.wsurl converts a standard URL to a WebSocket URL.
openrepl.wsurl = function(url) {
    url.protocol = {"http:":"ws:","https:":"wss:","ws:":"ws:","wss:":"wss:"}[url.protocol] || "ws:";
};

// openrepl.run starts a code run session and returns a promise to a corresponding WebSocket.
openrepl.run = function(code, lang) {
    return new Promise(function(s, f) {
        // build target url
        var targurl = new URL('/api/exec/run', window.location.href);
        targurl.searchParams.set('lang', lang);
        openrepl.wsurl(targurl);

        // connect WebSocket
        openrepl.promiseWS(targurl.toString()).then(function(ws) {
            var finished = false;

            // handle messages
            ws.onmessage = function(ev) {
                if(finished) return;
                var su;
                try {
                    su = JSON.parse(ev.data);
                } catch(e) {
                    finished = true;
                    f(e);
                    ws.close();
                    return;
                }
                switch(su.status) {
                case 'ready':
                    // send code
                    ws.send(code);
                    break;
                case 'running':
                    // done - pass off WebSocket
                    finished = true;
                    s(ws);
                    break;
                case 'error':
                    // error - fail
                    finished = true;
                    ws.close();
                    f(su.err);
                    break;
                }
            };

            // handle premature close
            ws.onclose = function() {
                if(finished) return;
                f("premature close");
            };
        }, f);
    });
};

// openrepl.term starts an interactive terminal session and returns a promise to a corresponding WebSocket.
openrepl.term = function(lang) {
    return new Promise(function(s, f) {
        // build target url
        var targurl = new URL('/api/exec/term', window.location.href);
        targurl.searchParams.set('lang', lang);
        openrepl.wsurl(targurl);

        // connect WebSocket
        openrepl.promiseWS(targurl.toString()).then(function(ws) {
            var finished = false;

            // handle messages
            ws.onmessage = function(ev) {
                if(finished) return;
                var su;
                try {
                    su = JSON.parse(ev.data);
                } catch(e) {
                    finished = true;
                    f(e);
                    ws.close();
                    return;
                }
                switch(su.status) {
                case 'running':
                    // done - pass off WebSocket
                    finished = true;
                    s(ws);
                    break;
                case 'error':
                    // error - fail
                    finished = true;
                    ws.close();
                    f(su.err);
                    break;
                }
            };

            // handle premature close
            ws.onclose = function() {
                if(finished) return;
                f("premature close");
            };
        }, f);
    });
};

openrepl.xhrpromise = function(xhr, body) {
    return new Promise(function(resolve, reject) {
        xhr.onload = function() {
            if(xhr.status == 200) {
                resolve(xhr.response);
            } else {
                reject(xhr.statusText);
            }
        };
        xhr.onerror = function(e) {
            reject(e);
        };
        if(body) {
            xhr.send(body);
        } else {
            xhr.send();
        }
    });
}

openrepl.store = function(code, lang) {
    return new Promise(function(resolve, reject) {
        var xhr = new XMLHttpRequest();
        xhr.open('POST', '/api/store/store');
        xhr.responseType = 'text';
        openrepl.xhrpromise(xhr, JSON.stringify({"code": code, "language": lang})).then(function(key) {
            resolve(key);
        }, function(e) {
            reject(e);
        })
    });
}

openrepl.load = function(key) {
    return new Promise(function(resolve, reject) {
        var xhr = new XMLHttpRequest();
        var targ = new URL('/api/store/load', window.location.href);
        targ.searchParams.set('key', key);
        xhr.open('GET', targ.toString());
        xhr.responseType = 'json';
        openrepl.xhrpromise(xhr).then(function(code) {
            resolve(code);
        }, function(e) {
            reject(e);
        })
    });
}
