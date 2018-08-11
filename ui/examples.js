var search = document.getElementById('search');
var cardbox = document.getElementById('cardbox');
var preload = document.getElementById('preload');
var sp = document.getElementById('sp');


function toastErr(err) {
    M.toast({
        html: err,
        displayLength: 10000
    });
};

function exampleCard(ex, cb) {
    // add code element
    var code = document.createElement('code');
    code.appendChild(document.createTextNode(ex.code));

    // outer col
    var col = document.createElement('div');
    col.classList.add('col', 's12', 'm12', 'l6', 'xl4');

    // card div
    var card = document.createElement('div');
    card.classList.add('card');

    // card-content
    var cardContent = document.createElement('div');
    cardContent.classList.add('card-content');

    // card title span
    var cardTitle = document.createElement('span');
    cardTitle.classList.add('card-title');
    cardTitle.appendChild(document.createTextNode(ex.name));

    // card action div
    var cardAction = document.createElement('div');
    cardAction.classList.add('card-action');

    // add language as a chip
    var langchip = document.createElement('div');
    langchip.classList.add('chip');
    langchip.appendChild(document.createTextNode(ex.lang));
    langchip.onclick = () => { search.value = 'lang:' + ex.lang; search.onchange(); };
    cardAction.appendChild(langchip);

    // add tags to card action
    if(ex.tags) {
        for(var i = 0; i < ex.tags.length; i++) {
            var tag = ex.tags[i];
            var chip = document.createElement('div');
            chip.classList.add('chip');
            chip.appendChild(document.createTextNode(tag));
            chip.onclick = () => { search.value = tag; search.onchange(); };
            cardAction.appendChild(chip);
        }
    }

    // add example runner
    var run = document.createElement('a');
    run.href = '#';
    run.onclick = () => { runExample(ex); };
    var runicon = document.createElement('i');
    runicon.classList.add('material-icons', 'small');
    runicon.innerHTML = 'play_arrow';
    run.appendChild(runicon);

    // build card
    col.appendChild(card);
    card.appendChild(cardContent);
    cardContent.appendChild(cardTitle);
    cardContent.appendChild(code);
    card.appendChild(cardAction);
    cardAction.appendChild(run);

    // start async syntax highlight
    openrepl.highlight(ex.code, ex.lang).then((c) => {
        cardContent.replaceChild(c, code);
        cb();
    }, (e) => {
        toastErr('failed to highlight');
        console.log(e);
    });

    return col;
}

function runExample(ex) {
    M.toast({html: 'Opening example. . .'});

    cardbox.innerHTML = '';
    cardbox.appendChild(preload);

    openrepl.store(ex.code, ex.lang).then((key) => {
        window.location.href = '/?key=' + key;
    }, (e) => {
        toastErr('failed to open example');
        console.log(e);
        search.value = '';
    });
}

function runQuery(query) {
    sp.classList.remove('invisible');
    return new Promise((resolve, reject) => {
        openrepl.queryExamples(query).then((es) => {
            cardbox.innerHTML = '';
            var n = es.length;
            var cb = function() {
                n--;
                if(n == 0) {
                    relayout();
                }
            };
            for(var i = 0; i < es.length; i++) {
                cardbox.appendChild(exampleCard(es[i], cb));
            }
            sp.classList.add('invisible');
            resolve();
        }, (e) => {
            sp.classList.add('invisible');
            reject(e);
        });
    });
}

search.onchange = () => {
    runQuery(search.value).then(() => {}, (e) => {
        toastErr('search failed');
        console.log(e);
    });
};

window.onresize = () => { search.onchange(); };

search.onchange();

function relayout() {
    // count columns
    var cards = Array.from(cardbox.childNodes);
    var x = cards[0].offsetTop;
    var cols;
    for(cols = 1; (cols < cards.length) && (cards[cols].offsetTop == x); cols++);

    // extract card divs from card cols
    for(var i = 0; i < cards.length; i++) {
        var col = cards[i];

        cards[i] = col.childNodes[0];

        if(i < cols) {
            // reuse col
            col.removeChild(cards[i]);
        } else {
            // remove col
            cardbox.removeChild(col);
        }
    }

    cols = cardbox.childNodes;

    // add cards to cols
    for(var i = 0; i < cards.length; i++) {
        // find shortest column
        var bestcol = 0;
        var minH = cols[0].offsetHeight;
        for(var j = 1; j < cols.length; j++) {
            var h = cols[j].offsetHeight;
            if(h < minH) {
                bestcol = j;
                minH = h;
            }
        }

        // add to col
        cols[bestcol].appendChild(cards[i]);
    }
}
