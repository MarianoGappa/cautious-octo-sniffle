// http://stackoverflow.com/questions/4810841/how-can-i-pretty-print-json-using-javascript
function syntaxHighlight(json) {
    if (typeof json === 'undefined') {
        return '';
    }
    if (typeof json != 'string') {
         json = JSON.stringify(json, undefined, 2);
    }
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        var cls = 'number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'key';
            } else {
                cls = 'string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'boolean';
        } else if (/null/.test(match)) {
            cls = 'null';
        }
        return '<span class="' + cls + '">' + match + '</span>';
    });
}

// http://stackoverflow.com/questions/7616461/generate-a-hash-from-string-in-javascript-jquery
const hash = (s) => {
    var hash = 0, i, chr, len;
    if (s.length === 0) return hash;
    for (i = 0, len = s.length; i < len; i++) {
        chr   = s.charCodeAt(i);
        hash  = ((hash << 5) - hash) + chr;
        hash |= 0; // Convert to 32bit integer
    }
    return hash;
}

// http://stackoverflow.com/questions/8482309/converting-javascript-integer-to-byte-array-and-back
const longToByteArray = (long) => {
    // we want to represent the input as a 8-bytes array
    var byteArray = [0, 0, 0, 0, 0, 0, 0, 0];

    for ( var index = 0; index < byteArray.length; index ++ ) {
        var byte = long & 0xff;
        byteArray [ index ] = byte;
        long = (long - byte) / 256 ;
    }

    return byteArray;
};

const stringToRGBA = (s) => {
    const arr = longToByteArray(hash(s))
    return `rgba(${arr[0]}, ${arr[1]}, ${arr[2]}, 0.6)`
}

const minibox = (id, label) => {
    const color = document.getElementById(id).style.backgroundColor
    const safeLabel = textLimit(label, 20)
    return `<span class="minibox" style="background-color: ${color}">${safeLabel}</span>`
}

const _ = a => document.querySelector(a)
const __ = a => document.querySelectorAll(a)

const safeId = (id) => id.replace(/[^a-zA-Z0-9-_]/g, '_')

// http://stackoverflow.com/questions/901115/how-can-i-get-query-string-values-in-javascript
const getParameterByName = (name, noLowercase) => {
    url = window.location.href;
    name = name.replace(/[\[\]]/g, "\\$&")
    if (!noLowercase) {
        name = name.toLowerCase(); // This is just to avoid case sensitiveness for query parameter name
        url = url.toLowerCase(); // This is just to avoid case sensitiveness
    }
    const regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
        results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return '';
    return decodeURIComponent(results[2].replace(/\+/g, " "));
}


function* colorGenerator(colors) {
    let i = 0
    while(true) {
        i = i >= colors.length - 1 ? 0 : i + 1
        yield colors[i]
    }
}

const s4 = () => Math.floor((1 + Math.random()) * 0x10000).toString(16).substring(1)
const guid = () => s4() + s4() + '-' + s4() + '-' + s4() + '-' + s4() + '-' + s4() + s4() + s4()
const cleanArray = (actual) => actual.filter((elem) => Boolean(elem))
const textLimit = (s, l) => (l < 3 || s.length <= l) ? s : s.substring(0,l-3) + "..."
