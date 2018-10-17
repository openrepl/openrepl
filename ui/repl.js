M.AutoInit();
Terminal.applyAddon(fit);
Terminal.applyAddon(attach);
function toastErr(err) {
    M.toast({
        html: err,
        displayLength: 10000
    });
};
var term1 = new Terminal({
    cursorBlink: true
});
var term2;
term1.open(document.getElementById("terminal"));
var t1pre = document.getElementById('tpre');
var t1ws;
var t1c = true;
function updateT1WS(ws) {
    t1ws = ws;
    t1c = false;
    ws.onclose = function() {
        toastErr('Interactive terminal disconnected.');
        term1.detach(ws);
        t1c = true;
    };
    term1.attach(ws, true, true);
}
function loadTerm1(lang) {
    t1pre.classList.remove('invisible');
    if(!t1c) {
        t1ws.onclose = function() {};
        term1.detach(t1ws);
        t1ws.close();
    }
    term1.reset();
    openrepl.term(lang).then((ws) => {
        updateT1WS(ws);
        t1pre.classList.add('invisible');
    }, (e) => {
        toastErr('Failed to load repl terminal.');
        console.log(e);
        t1pre.classList.add('invisible');
    });
}
var runbtn = document.getElementById("runbtn");
var savebtn = document.getElementById("savebtn");
var stopbtn = document.getElementById("stopbtn");
ace.require("ace/ext/language_tools");
var editor = ace.edit("editor");
editor.setTheme("ace/theme/monokai");
editor.setFontSize(20);
editor.setOptions({
    enableBasicAutocompletion: true,
    enableLiveAutocompletion: true,
    copyWithEmptySelection: true,
    fadeFoldWidgets: true
});
editor.commands.addCommand({
    name: 'save',
    bindKey: {
        win: 'Ctrl-S',
        mac: 'Command-S',
        sender: 'editor'
    },
    exec: function(env, args, request) {
        savebtn.click();
    }
});
editor.commands.addCommand({
    name: 'run',
    bindKey: {
        win: 'Ctrl-R',
        mac: 'Command-R',
        sender: 'editor'
    },
    exec: function(env, args, request) {
        runbtn.click();
    }
});
var termdiv = document.getElementById("termdiv");
var t2 = document.getElementById("term2");
window.onresize=function() {
    document.getElementById("editor").style.top = document.getElementById("buttons").offsetTop + document.getElementById("buttons").offsetHeight;
    document.getElementById("terminal").style.top = document.getElementById("container").offsetTop;
    document.getElementById("terminal").style.left = document.getElementById("editor").offsetWidth;
    document.getElementById("terminal").style.height = window.innerHeight - document.getElementById("container").offsetTop;
    document.getElementById("buttons").style.width = document.getElementById("editor").offsetWidth;
    if(termdiv.style.visibility == "visible") {
        var eh = 0.65 * window.innerHeight;
        var et = document.getElementById("editor").offsetTop;
        document.getElementById("editor").style.height = eh;
        t2.style.top = et + eh;
        t2.style.height = window.innerHeight - (et + eh);
        term2.fit();
    }
    term1.fit();
};
window.onresize();
var language = "lua";
function setLanguage(lang) {
    if(lang == "bash") {
        editor.getSession().setMode("ace/mode/sh");
    } else if(lang == "cpp") {
        editor.getSession().setMode("ace/mode/c_cpp");
    } else {
        editor.getSession().setMode("ace/mode/"+lang);
    }
    editor.setValue(demos[lang], -1);
    loadTerm1(lang);
    language = lang;
}
var t2ws;
var t2c = true;
var closecancel;
stopbtn.onclick = function() {
    stopbtn.classList.add('disabled');
    closecancel = true;
    t2ws.close();
};
runbtn.onclick = function() {
    if(runbtn.classList.contains('disabled')) return;
    closecancel = false;
    runbtn.classList.add("disabled");
    if(term2) {
        term2.reset();
    }
    openrepl.run(editor.getValue(), language).then(function(ws) {
        ws.onclose = function() {
            term2.detach(ws);
            runbtn.classList.remove("disabled");
            runbtn.classList.remove('invisible');
            stopbtn.classList.add('invisible');
            stopbtn.classList.remove("disabled");
            if(closecancel) {
                toastErr('Sucessfully stopped run.');
            } else {
                M.toast({html: 'Run finished.'});
            }
        };
        runbtn.classList.add('invisible');
        stopbtn.classList.remove('invisible');
        termdiv.style.visibility = "visible";
        t2ws = ws;
        if(!term2) {
            term2 = new Terminal({
                cursorBlink: true
            });
            term2.open(document.getElementById("term2"));
            window.onresize();
        }
        term2.attach(ws);
    }, function(e) {
        toastErr('Failed to load run session.');
        console.log(e);
        runbtn.classList.remove("disabled");
    });
};
savebtn.onclick = function() {
    savebtn.classList.add("disabled");
    openrepl.store(editor.getValue(), language).then(function(key) {
        var u = new URL(window.location.href);
        u.searchParams.set('key', key);
        window.location.replace(u.toString());
    }, function(e) {
        toastErr('Failed to save.');
        console.log(e);
        savebtn.classList.remove("disabled");
    })
};
function attachLang(l) {
    document.getElementById("lang-"+l).onclick = function() {
        setLanguage(l);
    };
}
attachLang("lua");
attachLang("python");
attachLang("forth");
attachLang("cpp");
attachLang("bash");
attachLang("javascript");
attachLang("typescript");
attachLang("php");
attachLang("golang");
attachLang("haskell");

(function() {
    var url = new URL(window.location.href);
    var key = url.searchParams.get('key');
    if(key == null) {
        setLanguage("lua");
        return;
    }
    openrepl.load(key).then(function(code) {
        setLanguage(code.language);
        editor.setValue(code.code, -1);
    }, function(e) {
        toastErr('Failed to load saved code.');
        console.log(e);
        setLanguage("lua");
    });
})();
