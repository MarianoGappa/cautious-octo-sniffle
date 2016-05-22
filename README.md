# flowbro
Real-time Kafka consumer visualisation tool, with mock mode for flow documentation.

# Getting started
- Requires Go and setting $GOPATH: https://golang.org/doc/install
```
go get github.com/MarianoGappa/flowbro
cd $GOPATH/src/github.com/MarianoGappa/flowbro
make
./flowbro
```
- Flowbro should be ready on http://localhost:41234

# I don't have nor want to know anything about Go!
- Fine. Use the latest release binary for your OS: https://github.com/MarianoGappa/flowbro/releases
- Unpack the source code on the same folder. Flowbro doesn't need the source code to work, but it needs the `webroot` folder within it.

# Making your first Flowbro configuration
- Clone `webroot/configs/config-example.js`; give it a name according to your project e.g.: `website-requests.js`
- The config file is well-documented; modify it based on your project's needs.
- The new config should be immediately available on http://localhost:41234

# Can I run Flowbro on Docker? (requires Go)
- Build docker image with
```
$ make image
```
- Run docker image with
```
$ make imagerun
```

# Can I contribute?
Yes, please.
