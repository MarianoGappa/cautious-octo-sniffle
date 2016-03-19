# flowbro
Real-time Kafka consumer visualisation tool, with mock mode for flow documentation.

# Documentation mode
![Blah](demo.gif?raw=true)

# Getting started
```
git clone git@github.com:MarianoGappa/flowbro.git    // go get github.com/MarianoGappa/flowbro
```

- Clone `webroot/config-example.js`; give it a name according to your project e.g.: `website-requests.js`
- The config file is well-documented; modify it based on your project's needs.
- You don't need Go; run the binary with the port you configured in your config file
- Open in your browser, e.g.: `localhost:41234?config=your-config-file` (don't include the .js)
