{
    "title": "FlowBro Example",
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
     "rules": [
        {
            "patterns": [
                {"field": "{{ .Topic }}", "pattern": "one"}
            ],
            "events": [
                {
                    "eventType": "message",
                    "sourceId": "Server",
                    "targetId": "Phone",
                    "text": "Send SMS"
                }
            ]
        }
    ],
    "defaultMessage": {
        "img": "message",
        "height": "38px",
        "backgroundColor": "transparent"
    },
    "images" : {
        "message": "images/message.gif",
        "person": "images/person.png"
    },
    "colourPalette" : ["#f44336", "#e91e63", "#9c27b0", "#673ab7", "#3f51b5", "#2196f3", "#03a9f4", "#00bcd4", "#009688", "#4caf50", "#8bc34a", "#cddc39", "#ffeb3b", "#ffc107", "#ff9800", "#ff5722","#795548", "#9e9e9e", "#607d8b"],
    "eventSeparationIntervalMilliseconds": 500,
    "animationLengthMilliseconds": 1000,
    "hideIgnoredMessages": false,
    "webSocketAddress": "localhost:41234",
    "kafka": {
        "brokers" : "kafka1.company.com:9092,kafka2.company.com:9092,kafka3.company.com:9092",
        "consumers" : [
            {
                "topic": "one"
            }
        ]
    },
    "tutorial": true
}
