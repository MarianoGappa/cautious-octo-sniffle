var config = {
    /*
        Will be displayed in the header of the page
    */
    "title": "FlowBro Example",

    /*
        Components are the guys that appear on the UI and exchange
        messages between each other.

        Components can appear in the UI as either images or flowchart
        rectangles with round corners; the rectangles will be used
        by default if you don't specify an 'img' property.

        The ids are mainly gonna be used in the "logic" section, where
        the interactions are defined, but also as the text inside the
        UI rectangles if you don't specify an image, so feel free to
        capitalise and use spaces.

        The img property is a reference to the "images" section below.

        Every property is optional except for the id, even top & left.
        If you don't specify top & left, flowbro will choose them for
        up to 5 components, but don't specify it for some but not all,
        as flowbro won't be smart in that case.

        You don't need to set backgroundColor, but it's useful to
        check if you set the height/width properly for an image. The
        downside to larger-than-needed dimensions is that the flying messages
        don't center properly on components.
    */
    "components": [
        {
            "id": "Person",
            "top": 50,
            "left": 70,
            "height": "75px",
            "width": "61px",
            "img": "person",
            "backgroundColor": "transparent"
        },
        {
            "id": "Server",
            "top": 250,
            "left": 250
        },
        {
            "id": "Phone",
            "top": 450,
            "left": 70
        },
        {
            "id": "Tablet",
            "top": 450,
            "left": 425,
            "backgroundColor": "rgb(150, 150, 200)"
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

        {"topic": "test", "partition": "0", "offset": "6", "key": "",
        "value": "tablet", "consumedUnixTimestamp": "1460189668"}

        You can use whichever Javascript magic you choose on this function, provided
        that you return an array of ui events as a result (empty array as a default case).

        For now, there are two types of UI interations: 'message' and 'log'.
        'message' is a message flowing from 'sourceId' to 'targetId', and 'log'
        is an entry on the right panel of the UI.

        'sourceId' and 'targetId' are the ids defined in the "components" section
    */
    "logic": function(event) {
        if (event.topic == "test" && event.value.match(/broadcast/i)) {
            return [
                    {
                        'eventType': 'message',
                        'sourceId': 'Person',
                        'targetId': 'Server',
                        'text': 'Person initiates a request to submit content to all devices'
                    }
            ]
        } else if (event.topic == "test" && event.value.match(/tablet/i)) {
            return [
                    {
                        'eventType': 'message',
                        'sourceId': 'Server',
                        'targetId': 'Phone',
                        'text': 'Server produces content to cellphone',
                        'color': 'happy'
                    }
            ]
        } else if (event.topic == "test" && event.value.match(/cellphone/i)) {
            return [
                    {
                        'eventType': 'message',
                        'sourceId': 'Server',
                        'targetId': 'Tablet',
                        'text': 'Server produces content to tablet',
                        'color': 'happy'
                    }
            ]
        } else
            return []
    },

    /*
        Images used throughout Flowbro. Use these references on components and
        defaultMessage. Will be used in the src attribute of img tags.
    */
    "images" : {
        "person": "images/person.png"
    },

    /*
        These colours will be used for components with no image specified.
        An algorithm will cycle through these colours for each component.
    */
    "colourPalette" : ["#4DD0E1","#AED581","#FFD54F","#DCE775","#FFB74D","#81C784","#FFF176","#4DB6AC","#FF8A65"],

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
        component B. CSS transition property value.
    */
    "animationLengthMilliseconds": 1000,

    /*
        When an incoming message goes through the logic function and doesn't
        produce any events, a log will be issued unless this setting is set
        to true.
    */
    "hideIgnoredMessages": false,

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
        [
            {"eventType":"message","sourceId":"Person","targetId":"Server","text":"Person initiates a request to submit content to all devices"}
        ],
        [
            {"eventType":"message","sourceId":"Server","targetId":"Phone","text":"Server produces content to cellphone","color":"happy"},
            {"eventType":"message","sourceId":"Server","targetId":"Tablet","text":"Server produces content to tablet","color":"happy"}
        ]
    ],

    /*
        When enabled, shows the rate at which messages are being consumed.
        Doesn't work on documentationMode.
    */
    "rateCalculationEnabled": true,
    "rateCalculationIntervalMilliseconds": 1000,

    /*
        Where is the server listening at? ws://[this_is_the_address]/ws
        e.g. localhost:41234
        Flowbro starts on port 41234 by default.
    */
    "webSocketAddress": "localhost:41234",

    /*
        When true, the flying messages will be color-coded based on a hash
        of the Kafka message's key. Same key => same color.
    */
    "colorBasedOnKey": true,

    /*
        Please include all Kafka topic/partition pairs that you need to
        listen to.
        Note that there is an "offset" setting. It should always be set
        to "newest" (i.e. read new messages from now on on the queue), but
        you can set it to "oldest" (i.e. read from beginning) for test
        purposes. This can be a very bad idea!

        This configuration will be sent to the server, so please don't add
        extra fields or you will likely break the server!

        Use -1 on the partition property to listen to all partitions in the
        topic. Otherwise, specify the partition number.
    */
    "serverConfig": {
        "consumers" : [
            {
                "broker" : "localhost:9092",
                "partition" : -1,
                "topic": "test",
                "offset": "newest"
            }
        ]
    }
}
