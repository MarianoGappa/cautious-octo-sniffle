{
  "title": "FlowBro Example",
  "components": [
    {
      "id": "Endpoint",
      "top": 50,
      "left": 300
    },
    {
      "id": "Server",
      "top": 250,
      "left": 300
    },
    {
      "id": "Phone",
      "top": 450,
      "left": 70
    },
    {
      "id": "Desktop",
      "top": 450,
      "left": 300
    },
    {
      "id": "Tablet",
      "top": 450,
      "left": 525
    },
    {
      "id": "Tutorial",
      "top": 650,
      "left": 170
    },
    {
      "id": "You",
      "top": 650,
      "left": 425
    }
  ],
  "rules": [
    {
      "patterns": [
        {
          "field": "{{ .Topic }}",
          "pattern": "requests"
        }
      ],
      "events": [
        {
          "eventType": "message",
          "sourceId": "Endpoint",
          "targetId": "Server",
          "text": "Route push notification to device",
          "aggregate": true
        }
      ]
    },
    {
      "patterns": [
        {
          "field": "{{ .Topic }}",
          "pattern": "notifications"
        },
        {
          "field": "{{ .Value.target }}",
          "pattern": "desktop"
        }
      ],
      "events": [
        {
          "eventType": "message",
          "sourceId": "Server",
          "targetId": "Desktop",
          "text": "{{.Value.message}}",
          "aggregate": true
        }
      ]
    },
    {
      "patterns": [
        {
          "field": "{{ .Topic }}",
          "pattern": "notifications"
        },
        {
          "field": "{{ .Value.target }}",
          "pattern": "phone"
        }
      ],
      "events": [
        {
          "eventType": "message",
          "sourceId": "Server",
          "targetId": "Phone",
          "text": "{{.Value.message}}",
          "aggregate": true
        }
      ]
    },
    {
      "patterns": [
        {
          "field": "{{ .Topic }}",
          "pattern": "notifications"
        },
        {
          "field": "{{ .Value.target }}",
          "pattern": "tablet"
        }
      ],
      "events": [
        {
          "eventType": "message",
          "sourceId": "Server",
          "targetId": "Tablet",
          "text": "{{.Value.message}}",
          "aggregate": true
        }
      ]
    },
    {
      "patterns": [
        {
          "field": "{{ .Topic }}",
          "pattern": "tutorial.messages"
        }
      ],
      "events": [
        {
          "eventType": "message",
          "sourceId": "Tutorial",
          "targetId": "You",
          "text": "{{.Value.log}}",
          "noJSON": true
        }
      ]
    }
  ],
  "defaultMessage": {
    "img": "message",
    "height": "38px",
    "backgroundColor": "transparent"
  },
  "images": {
    "message": "images/message.gif",
    "person": "images/person.png"
  },
  "colourPalette": [
    "#f44336",
    "#9c27b0",
    "#3f51b5",
    "#03a9f4",
    "#009688",
    "#8bc34a",
    "#ffeb3b",
    "#ff9800",
    "#795548",
    "#607d8b",
    "#e91e63",
    "#673ab7",
    "#2196f3",
    "#00bcd4",
    "#4caf50",
    "#cddc39",
    "#ffc107",
    "#ff5722",
    "#9e9e9e"
  ],
  "eventSeparationIntervalMilliseconds": 500,
  "animationLengthMilliseconds": 1000,
  "hideIgnoredMessages": false,
  "webSocketAddress": "localhost:41234",
  "kafka": {
    "brokers": "kafka1.company.com:9092,kafka2.company.com:9092,kafka3.company.com:9092",
    "consumers": [
      {
        "topic": "one"
      }
    ]
  },
  "tutorial": true
}
