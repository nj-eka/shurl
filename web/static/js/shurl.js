"use strict";

let generateBtn = document.getElementById("generate");
let targetUrl = document.getElementById("targetUrl");
let expiredIn = document.getElementById("expiredIn");

async function postData(url = '', data = {}) {
    const response = await fetch(url, {
        method: 'POST',
        mode: 'cors', // no-cors, *cors, same-origin
        cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
        credentials: 'same-origin', // include, *same-origin, omit
        headers: {
            'Content-Type': 'application/json'
        },
        redirect: 'follow', // manual, *follow, error
        referrerPolicy: 'no-referrer', // no-referrer, *client
        body: JSON.stringify(data)
    });
    return await response.json();
}

function renderResponse(data) {
    let htmlSegment = `<p>Target URL: <i>${targetUrl.value}</i></p>
                       <p>Short Link: <a href="${data.shortUrl}">${data.shortUrl}</a></p>
                       <p>Short Link Info: <a href="${data.shortUrlInfo}">${data.shortUrlInfo}</a></p>
                       <br>`
    let answers = document.getElementById("answers");
    answers.innerHTML += htmlSegment;
}

generateBtn.onclick = function() {
    const requestShurl = { targetUrl: targetUrl.value }
    if (expiredIn.value) {
        requestShurl.expiredInDays = parseInt(expiredIn.value)
    }
    postData('/', requestShurl)
        .then((data) => {
            renderResponse(data);
        }
    );
        // .catch((err) => {
        //     console.error(err);
        // });
}
