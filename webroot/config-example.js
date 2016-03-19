var config = {
    /*
        Will be displayed in the header of the page
    */
    "title": "FlowBro Example",

    /*
        Components are the guys that appear on the UI and exchange
        messages between each other.

        The ids are gonna be used in the "logic" section, where
        the interactions are defined.

        The img property is a reference to the "images" section.

        You don't need to set backgroundColor, but it's useful to
        check if you set the height/width properly. The downside
        to larger-than-needed dimensions is that the flying messages
        don't center properly on components.
    */
    "components": [
        {
            "id": "person",
            "top": 50,
            "left": 70,
            "height": "75px",
            "width": "61px",
            "img": "person",
            "backgroundColor": "transparent"
        },
        {
            "id": "server",
            "top": 250,
            "left": 300,
            "width": "50px",
            "height": "80px",
            "img": "server",
            "backgroundColor": "transparent"
        },
        {
            "id": "phone",
            "top": 450,
            "left": 70,
            "width": "50px",
            "img": "phone",
            "backgroundColor": "transparent"
        },
        {
            "id": "tablet",
            "top": 450,
            "left": 500,
            "width": "50px",
            "height": "82px",
            "img": "tablet",
            "backgroundColor": "transparent"
        }
    ],

    /*
        This is where you define the mapping between consumed messages and
        ui interactions between components. Note that it's a 1 to n mapping:
        one event (i.e. one consumed message) can trigger multiple ui
        interactions.

        An "event" is a Kafka Message received on a queue by the server. It
        looks like this:
        (Use your cleverness to infer what each element means in Kafka lingo)

        {"topic": "test", "partition": "0", "offset": "6", "key": "", "value": "tablet"}

        You can use whichever Javascript magic you choose on this function, provided
        that you return an array of ui events as a result (empty array as a default case).

        Ui events are objects that look like the example below. For now, always use
        'streamMessage' as the 'eventType'; this is a namespace so that in the future
        other interactions can be added.

        'sourceId' and 'targetId' are the ids defined in the "components section"
        The 'caption' text is shown in the right hand log section in the UI.
    */
    "logic": function(event) {
        if (event.value.match(/broadcast/i)) {
            return [
                    {
                        'eventType': 'streamMessage',
                        'sourceId': 'person',
                        'targetId': 'server',
                        'caption': 'Person initiates a request to submit content to all devices'
                    }
            ]
        } else if (event.value.match(/tablet/i)) {
            return [
                    {
                        'eventType': 'streamMessage',
                        'sourceId': 'server',
                        'targetId': 'phone',
                        'caption': 'Server produces content to cellphone'
                    }
            ]
        } else if (event.value.match(/cellphone/i)) {
            return [
                    {
                        'eventType': 'streamMessage',
                        'sourceId': 'server',
                        'targetId': 'tablet',
                        'caption': 'Server produces content to tablet'
                    }
            ]
        } else
            return []
    },

    /*
        defaultMessage is the flying message that is passed around components.
        Same property rules as components.
    */
    "defaultMessage": {
        "img": "message",
        "height": "38px",
        "backgroundColor": "inherit"
    },

    /*
        Images used throughout Flowbro. Use these references on components and
        defaultMessage. Will be used in the src attribute of img tags.
    */
    "images" : {
        "message": "images/message.gif",
        "phone": "images/phone.png",
        "person": "images/person.png",
        "tablet": "images/tablet.png",
        "server": "images/server.png"
    },

    /*
        Incoming events (i.e. kafka messages coming from the websocket)
        are buffered such that if 10 events come in a quick burst, you
        can still see the order in which they came. This is useful to
        understand the flow of the event pipeline, but could be
        misleading if there is a lot of traffic on the queues.
    */
    "eventSeparationIntervalMilliseconds": 500,

    /*
        How long does it take for one message to go from component A to
        component b. CSS transition property value.
    */
    "animationLengthMilliseconds": 1000,

    /*
        If documentationMode is set to true, no WebSocket connection is
        established to the server. Instead, the server's incoming events
        are mocked by the "documentationSteps" array, and they can be
        triggered by the "Next" button that will appear on the top-right
        corner of the screen.
        This mode is a very clear way to document a project that is best
        explained by a flowchart.
    */
    "documentationMode": true,
    "documentationSteps": [
        [{"topic": "example", "value": "broadcast request [123456]"}],
        [
            {"topic": "example", "value": "produced content for tablet"},
            {"topic": "example", "value": "produced content for cellphone"},
        ]
    ],

    /*
        Where is the server listening at? Note that the port MUST be equal
        to the one you used when starting the server, as its first argument.
    */
    "webSocketHost": "localhost",
    "webSocketPort": "41234",

    /*
        Please include all Kafka topic/partition pairs that you need to
        listen to.
        Note that there is an "offset" setting. It should always be set
        to "newest" (i.e. read new messages from now on on the queue), but
        you can set it to "oldest" (i.e. read from beginning) for test
        purposes. This can be a very bad idea!

        This configuration will be sent to the server, so please don't add
        extra fields or you will likely break the server!
    */
    "serverConfig": {
        "consumers" : [
            {
                "broker" : "localhost:9092",
                "partition" : 0,
                "topic": "test",
                "offset": "newest"
            }
        ]
    }
}
