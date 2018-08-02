// JS API for OpenREPL backend services
var openrepl = {};

// openrepl.promiseWS returns a promise that is fulfilled when the WebSocket is opened.
openrepl.promiseWS = function(url) {
    return new Promise(function(s, f) {
        var ws = new WebSocket(url);
        ws.binaryType = "arrayBuffer";
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
                    f(su.error);
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
                    f(su.error);
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

Terminal.applyAddon(fit);
Terminal.applyAddon(attach);

// openrepl.createTerminal creates an xterm with the given WebSocket and container.
// NOTE: requires xterm.js
openrepl.createTerminal = function(ws, container) {
    // TODO: set color theme https://xtermjs.org/docs/api/terminal/interfaces/iterminaloptions/#optional-theme
    var term = new Terminal({
        cursorBlink: true
    });
    term.attach(ws);
    term.open(container);
    term.fit();
    return term;
};
