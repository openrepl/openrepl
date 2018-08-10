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

function exampleCard(ex) {
    // add code element
    var code = document.createElement('code');
    code.appendChild(document.createTextNode(ex.code));

    // outer col
    var col = document.createElement('div');
    col.classList.add('col', 's12', 'm6', 'l4', 'xl3');

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
            for(var i = 0; i < es.length; i++) {
                cardbox.appendChild(exampleCard(es[i]));
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

search.onchange();
