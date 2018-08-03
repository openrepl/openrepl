Terminal.applyAddon(fit);
Terminal.applyAddon(attach);
var term1 = new Terminal({
    cursorBlink: true
});
var term2;
term1.open(document.getElementById("terminal"));
var t1ws;
var t1c = true;
function updateT1WS(ws) {
    t1ws = ws;
    t1c = false;
    ws.onclose = function() {
        term1.detach(ws);
        t1c = true;
    };
    term1.attach(ws, true, true);
}
function loadTerm1(lang) {
    if(!t1c) {
        t1ws.onclose = function() {};
        term1.detach(t1ws);
        t1ws.close();
    }
    term1.reset();
    openrepl.term(lang).then((ws) => updateT1WS(ws), (e) => console.log(e));
}
var runbtn = document.getElementById("runbtn");
var savebtn = document.getElementById("savebtn");
var loadbtn = document.getElementById("loadbtn");
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
    name: 'open',
    bindKey: {
        win: 'Ctrl-O',
        mac: 'Command-O',
        sender: 'editor'
    },
    exec: function(env, args, request) {
        loadbtn.click();
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
setLanguage("lua");
var t2ws;
var t2c = true;
runbtn.onclick = function() {
    runbtn.classList.add("disabled");
    openrepl.run(editor.getValue(), language).then(function(ws) {
        if(!t2c) {
            t2ws.onclose = function() {};
            t2ws.close();
            term2.detach(t2ws);
            term2.reset();
        }
        termdiv.style.visibility = "visible";
        t2ws = ws;
        if(!term2) {
            term2 = new Terminal({
                cursorBlink: true
            });
            term2.open(document.getElementById("term2"));
        }
        term2.attach(ws);
        window.onresize();
        runbtn.classList.remove("disabled");
    }, function(e) {
        console.log(e);
        runbtn.classList.remove("disabled");
    });
};
savebtn.onclick = function() {
    savebtn.classList.add("disabled");
    var xhr = new XMLHttpRequest();
    xhr.open('PUT', "/api/60s/add", true);
    xhr.send(editor.getValue());
    xhr.onreadystatechange = function(){
        if (xhr.readyState == 4) {
            if (xhr.status == 200) {
                var x2 = new XMLHttpRequest();
                x2.open('POST', "/api/save/save", true);
                x2.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
                x2.onreadystatechange = function() {
                    if(x2.readyState == 4) {
                        if(x2.status == 200) {
                            document.getElementById("savelnk").innerHTML = "Save ID: " + x2.responseText;
                            savebtn.classList.remove("disabled");
                        }
                    }
                }
                x2.send('srcid='+xhr.responseText);
            }
        }
    };
};
loadbtn.onclick = function() {
    var id = prompt("Enter save ID");
    if (id != null) {
        var xhr = new XMLHttpRequest();
        xhr.open('GET', "/api/save/get?id="+id, true);
        xhr.onreadystatechange = function() {
            if (xhr.readyState == 4) {
                if (xhr.status == 200) {
                    editor.setValue(xhr.responseText, -1);
                }
            }
        }
        xhr.send();
    }
};
function attachLang(l) {
    document.getElementById("lang-"+l).onclick = function() {
        setLanguage(l);
    };
}
attachLang("lua");
attachLang("python2");
attachLang("python3");
attachLang("forth");
attachLang("cpp");
attachLang("bash");
attachLang("javascript");
attachLang("typescript");
attachLang("php");
