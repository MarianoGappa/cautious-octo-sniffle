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

const keyToRGBA = (key) => {
    const arr = longToByteArray(hash(key))
    return `rgba(${arr[0]}, ${arr[1]}, ${arr[2]}, 0.6)`
}

const minibox = (id) => {
    const color = document.getElementById('component_' + id).style.backgroundColor
    return `<span class="minibox" style="background-color: ${color}">${id}</span>`
}

let documentationModeIterator = 0
const eventQueue = []
const eventLog = []
const state = {}
let rateCalculationCache = []
let rateCalculationInterval = null

const _ = a => document.querySelector(a)

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

const init = (configFile) => {
    if (!_(`init_script_${configFile}`)) {
        const element = document.createElement('script')
        element.setAttribute('id', `init_script_${configFile}`)
        element.setAttribute('src', `configs/${configFile}.js`)
        element.setAttribute('async', false)
        document.head.appendChild(element)
    }
}

const log = (message, _color, from, to, json) => {
    const colors = {
        'severe':'#E53A40',
        'error': '#E53A40',
        'warning': '#FFBC42',
        'info': 'inherit',
        'trace': '#6E7783',
        'debug': '#6E7783',
        'happy': '#2f751b',
        'default': 'inherit'
    }

    const color = colors[_color] || colors['default']
    const header = typeof from !== 'undefined' && typeof to !== 'undefined'
        ? minibox(from) + ` → ` + minibox(to) + `<br/>` : ''

    const prettyJson = typeof json !== 'undefined' ? '<pre>' + syntaxHighlight(json) + '</pre>' : '';

    const element = document.createElement('span')
    element.className = 'logLine'
    element.style.color = color
    element.innerHTML = header + message + '<br/>' + prettyJson
    _('#log').insertBefore(element, _('#log').firstChild)

    if (_('#log').children.length > 1000) {
        _('#log').removeChild(_('#log').lastElementChild)
    }
}

const run = (timeout) => {
    if (typeof config !== 'undefined') {
        if (typeof brokersOverride !== 'undefined' && brokersOverride !== null) {
            config.serverConfig.brokers = brokersOverride
            log(`Overriding brokers to [${brokersOverride}]`)
        }
        doRun()
    } else if (timeout > 0) {
        console.log("not ready; retrying...")
        window.setTimeout(() => run(timeout - 1), 50)
    } else {
        log('Did you add .js to it? (you shouldn\'t)', 'error')
        log('Is the ?config=xxx filename wrong?', 'error')
        log('Did you break the JSON syntax?', 'error')
        log('Cannot load configuration file', 'error')
        _('#title').innerHTML = 'Flowbro is drunk :('
    }
}

const doRun = () => {
    _('#title').innerHTML = config.title
    loadComponents(config)

    window.setInterval(() => showNextUiEvent(), config.eventSeparationIntervalMilliseconds)

    startRateCalculation()

    if (!config.documentationMode) {
        openWebSocket()
        _('#rest').innerHTML = '<button onclick="javascript:replayEventLog()">Replay</button><button onclick="javascript:cleanEventLog()">Clear</button>'
    } else {
        _('#rest').innerHTML = '<button onclick="javascript:resetDocumentationMode()">Reset</button><button onclick="javascript:mockPoll()">Next</button>'
    }
}

const showNextUiEvent = () => {
    if (eventQueue.length > 0) {
        const event = eventQueue.shift()

        if (event.eventType == 'message') {
            animateFromTo(_(`[id='component_${event.sourceId}']`), _(`[id='component_${event.targetId}']`), event.quantity ? event.quantity : 1, event.key)
        }
        if (typeof event.logs !== 'undefined') {
            for (let i in event.logs) { log(event.logs[i].text, event.logs[i].color, event.sourceId, event.targetId, i == 0 ? event.json : undefined) }
        } else if (event.text) {
            log(event.text, event.color, event.sourceId, event.targetId, event.json)
        }

        // Save enqueued animation into event log; keep it <= 100 events
        if (!config.documentationMode) {
            eventLog.push([event])
            if (eventLog.length > 100)
                eventLog.shift()
        }
    }
}

const openWebSocket = () => {
    const wsUrl = "ws://" + config.webSocketAddress + "/ws"
    const ws = new WebSocket(wsUrl)

    ws.onopen = (event) => {
        log(`WebSocket open on [${wsUrl}]!`, 'happy')
        try {
            ws.send(JSON.stringify(config.serverConfig))
            log("Sent configurations to server successfully!", 'happy')
        } catch(e) {
            log("Server is drunk :( can't send him configurations!", 'error')
            console.log(e)
        }
    }

    ws.onmessage = (message) => {
        if (!config.documentationMode) {
            consumedMessages = []
            if (message.data.trim()) {
                lines = cleanArray(message.data.trim().split(/\n/))
                for (i in lines) {
                    try {
                        maybeResult = JSON.parse(lines[i])

                        if (config.rateCalculationEnabled && maybeResult.timestamp) {
                            rateCalculationCache.push(parseInt(maybeResult.timestamp))
                        }

                        consumedMessages.push(maybeResult)
                    } catch (e) {
                        console.log(`Couldn't parse this as JSON: ${lines[i]}`, "\nError: ", e)
                    }
                }
            }

            processUiEvents(consumedMessagesToEvents(consumedMessages))
        } else if (!config.hideIgnoredMessages) {
            console.log('Ignored incoming message', message)
            log('Ignored incoming message.', 'debug')
        }
    }

    ws.onclose = (event) => log("WebSocket closed!", 'error')
    ws.onerror = (event) => log(`WebSocket had error! ${event}`, 'error')
}

const processUiEvents = (events) => { for (event of events) {
    if (config.documentationMode)
        eventQueue.push(event)
    else
        aggregateEventOnEventQueue(event)
} }

const aggregateEventOnEventQueue = (event) => {
    const indexOfSimilarMessage = (event, eventQueue) => {
        let index = undefined
        eventQueue.forEach((v, i) => { if (v.sourceId == event.sourceId && v.targetId == event.targetId) index = i })
        return index
    }

    // if it's an A -> B type of event
    if (event.eventType == 'message') {
        const i = indexOfSimilarMessage(event, eventQueue)

        // if a message from the same A -> B exists, +1 its quantity and add its log if present
        if (typeof i !== 'undefined') {
            eventQueue[i].quantity = eventQueue[i].quantity ? eventQueue[i].quantity + 1 : 2
            if (typeof event.text !== 'undefined') eventQueue[i].logs.push(event)

        // if it's the first message from A -> B, add it to the queue and start a collection of logs for it
        } else {
            let aggregatedEvent = event
            if (typeof event.text !== 'undefined') {
                if (typeof event.logs !== 'undefined') {
                    aggregatedEvent.logs.push(event)
                } else {
                    aggregatedEvent.logs = [event]
                }
            }
            eventQueue.push(aggregatedEvent)
        }

    // if it's a log type of event
    } else if (event.eventType == 'log') {
        let lastId = eventQueue.length - 1

        // if the last event on the queue is a log event, add this log to it
        if (eventQueue[lastId] && eventQueue[lastId].eventType == 'log') {
            eventQueue[lastId].logs.push(event)

        // otherwise, push a new event and start a collection of logs for it
        } else {
            let aggregatedEvent = event
            if (typeof event.logs !== 'undefined') {
                aggregatedEvent.logs.push(event)
            } else {
                aggregatedEvent.logs = [event]
            }
            eventQueue.push(aggregatedEvent)
        }
    }
}

const cleanEventLog = () => { eventLog.length = 0; log('-- Replay event log is now empty --', 'debug'); }
const replayEventLog = () => {
    if (eventLog.length > 0) {
        config.documentationMode = true
        documentationModeIterator = 0
        config.documentationSteps = eventLog
        stopRateCalculation()
        eventQueue.length = 0
        refreshDocumentationModeStepCount()
        log('-- Replay event log mode; ignoring real-time messages --', 'happy')
        _('#rest').innerHTML = '<button onclick="javascript:resetDocumentationMode()">|&lt;&lt;</button><button onclick="javascript:mockPoll()">&gt;</button><button onclick="javascript:restoreRealTime()">Back</button>'
    } else {
        log('-- Replay event log is empty --', 'error')
    }
}
const restoreRealTime = () => {
    config.documentationSteps.length = 0
    documentationModeIterator = 0
    config.documentationMode = false
    eventLog.length = 0
    startRateCalculation()
    log('-- Back to real-time mode --')
    _('#rest').innerHTML = '<button onclick="javascript:replayEventLog()">Replay</button><button onclick="javascript:cleanEventLog()">Clear</button>'
}

const resetDocumentationMode = () => {
    documentationModeIterator = 0
    refreshDocumentationModeStepCount()
    log('-- reset --', 'debug')
}

const mockPoll = () => {
    newEvents = config.documentationSteps[documentationModeIterator] ? config.documentationSteps[documentationModeIterator++] : []
    refreshDocumentationModeStepCount()
    processUiEvents(newEvents)
}
const refreshDocumentationModeStepCount = () => {
    _('#footer').style.display = 'block';
    _('#footer').innerHTML = `${documentationModeIterator}/${config.documentationSteps.length} events`
}

const consumedMessagesToEvents = (consumedMessages) => {
    consumedMessages.sort((a, b) => a.timestamp < b.timestamp ? -1 : a.timestamp > b.timestamp ? 1 : 0)

    const events = []
    for (let i in consumedMessages) {
        if (consumedMessages[i]) {
            const newEvents = config.logic(consumedMessages[i], log, state)
            if (newEvents.length == 0 && !config.hideIgnoredMessages) {
                log(`Ignoring event: ${consumedMessages[i].value}`, 'debug')
            }
            for (let j in newEvents) {
                try {
                    newEvents[j].json = JSON.parse(consumedMessages[i].value) // json specific
                } finally {}
                newEvents[j].key = consumedMessages[i].key
                events.push(newEvents[j])
            }
        }
    }
    return events
}

function* colorGenerator() {
    let i = 0
    while(true) {
        i = i >= config.colourPalette.length - 1 ? 0 : i + 1
        yield config.colourPalette[i]
    }
}

const loadComponents = (config) => {
    let colorRing = colorGenerator()
    for (let i in config.components) {
        const component = config.components[i]

        let element = document.createElement('div')
        element.id = `component_${component.id}`
        element.className = 'detached component'

        _('#container').appendChild(element)

        if (component.backgroundColor) {
            element.style.backgroundColor = component.backgroundColor
        }

        element.style.width = component.width ? component.width : "150px"
        element.style.height = component.height ? component.height : "100px"

        const position = componentPosition(config.components, i)
        element.style.left = position.left
        element.style.top = position.top

        if (component.img) {
            const img = document.createElement('img')
            img.src = config.images[component.img]
            element.appendChild(img)
        } else {
            const title = document.createElement('span')
            title.className = 'component_title'
            title.innerHTML = component.id
            element.appendChild(title)
            element.style.backgroundColor = component.backgroundColor ? component.backgroundColor : colorRing.next().value
            title.style.marginTop = "-" + (parseInt(title.offsetHeight) / 2) + "px"
            title.style.width = parseInt(element.style.width) - 20 - 2 // 20 = padding
            element.style.boxShadow = "0 2px 5px rgba(0,0,0,0.26)"
            element.style.textTransform = "uppercase"
        }
    }
}

const animateFromTo = (source, target, quantity, key) => {
    const element = document.createElement('div')
    element.id = guid()
    element.className = 'detached message'

    _('#container').appendChild(element)
    element.style.top = parseInt(source.offsetTop) + (parseInt(source.offsetHeight) / 2) - (parseInt(element.offsetHeight) / 2)
    element.style.left = parseInt(source.offsetLeft) + (parseInt(source.offsetWidth) / 2) - (parseInt(element.offsetWidth) / 2)

    element.style.zIndex = -1

    var rgb = undefined
    if (config.colorBasedOnKey && typeof key !== 'undefined' && key !== '') {
        rgb = keyToRGBA(key)
    }

    if (quantity > 1) {
        const q = document.createElement('h2')
        q.innerHTML = quantity
        element.appendChild(q)
    }

    if (typeof rgb !== 'undefined') {
        element.style.background = `linear-gradient(${rgb}, ${rgb}), url(images/message.gif)`
    } else {
        element.style.background = 'url(images/message.gif)'
    }
    element.style.backgroundSize = 'cover'

    const newTop = target.offsetHeight / 2 - parseInt(element.offsetHeight) / 2 + parseInt(target.offsetTop)
    const newLeft = target.offsetWidth / 2 - parseInt(element.offsetWidth) / 2 + parseInt(target.offsetLeft)

    style = document.createElement('style')
    style.type = 'text/css'
    const styleId = `style_${guid()}`
    const length = config.animationLengthMilliseconds
    style.appendChild(document.createTextNode(`.${styleId} { top: ${newTop}px !important; left: ${newLeft}px !important; -webkit-transition: top${length}ms, left ${length}ms; /* Safari */ transition: top ${length}ms, left ${length}ms;}`))
    document.body.appendChild(style)

    element.className = `${styleId} detached message`

    const removeNodes = (element, style) => () => {
        element.parentNode.removeChild(element)
        style.parentNode.removeChild(style)
    }

    window.setTimeout(removeNodes(element, style), length)
}

const calculateRate = (rateCalculationCache) => {
    rateCalculationCache.sort()
    const first = rateCalculationCache.shift()
    const last = rateCalculationCache.pop()
    const size = rateCalculationCache.length

    return size < 2 ? size : parseInt(size / (last - first + 1))
}

const startRateCalculation = () => {
    if (config.rateCalculationEnabled && !config.documentationMode) {
        _('#footer').style.display = 'block';

        rateCalculationInterval = window.setInterval(() => {
            _('#footer').innerHTML = `${calculateRate(rateCalculationCache)} messages/s (${eventLog.length} events logged)`
            rateCalculationCache = []
        }, config.rateCalculationIntervalMilliseconds)
    }
}
const stopRateCalculation = () => {
    if (rateCalculationInterval !== 'undefined') {
        window.clearInterval(rateCalculationInterval)
    }
}

const s4 = () => Math.floor((1 + Math.random()) * 0x10000).toString(16).substring(1)
const guid = () => s4() + s4() + '-' + s4() + '-' + s4() + '-' + s4() + '-' + s4() + s4() + s4()
const cleanArray = (actual) => actual.filter((elem) => Boolean(elem))

const componentPosition = (components, i) => {
    const defaultPositions = [
        [],
        [{left: 50, top: 50}],
        [{left: 50, top: 50}, {left: 450, top: 450}],
        [{left: 50, top: 50}, {left: 50, top: 450}, {left: 450, top: 450}],
        [{left: 50, top: 50}, {left: 50, top: 450}, {left: 450, top: 50}, {left: 450, top: 450}],
        [{left: 50, top: 250}, {left: 200, top: 50}, {left: 100, top: 450}, {left: 450, top: 220}, {left: 450, top: 450}],
    ]

    const position = {}

    if (components[i].top != undefined) {
        position.top = components[i].top
    } else if (defaultPositions[components.length] != undefined) {
        position.top = defaultPositions[components.length][i].top
    } else {
        position.top = 0
    }

    if (components[i].left != undefined) {
        position.left = components[i].left
    } else if (defaultPositions[components.length] != undefined) {
        position.left = defaultPositions[components.length][i].left
    } else {
        position.left = 0
    }

    return position
}

let brokersOverride = undefined
const brokerOverrideParam = getParameterByName('brokers')
if (typeof brokerOverrideParam !== 'undefined') {
    brokersOverride = brokerOverrideParam
}
try {
    const inlineConfigParam = getParameterByName('inlineConfig', 'no_lowercase')
    if (inlineConfigParam !== null) {
        const inlineConfig = atob(inlineConfigParam)
        eval(inlineConfig)
    }
} finally {
    if (typeof config === 'undefined') {
        init(getParameterByName('config') || 'config-example')
    }
}
