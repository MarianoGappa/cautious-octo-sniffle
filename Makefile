all: build

build: ARTIFACT ?= flowbro
build: GOOS ?= darwin
build: GOARCH ?= amd64
build: clean
		GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -o ${ARTIFACT} -a .

clean: ARTIFACT ?= flowbro
clean: cleanmac
		rm -f ${ARTIFACT}

cleanmac:
		find . -name '*.DS_Store' -type f -delete

image: ARTIFACT ?= flowbro
image: TAG ?= latest
image: clean
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${ARTIFACT} -a .
		sudo docker build -t ${ARTIFACT}:$(TAG) .

test:
		go test

run: ARTIFACT ?= flowbro
run: build
	./${ARTIFACT}

imagerun: ARTIFACT ?= flowbro
imagerun: TAG ?= latest
imagerun: cleanmac
		sudo docker run --name flowbro -p 41234:41234 ${ARTIFACT}:$(TAG)

release-linux: TAG ?= latest
release-linux: ARTIFACT = flowbro-${TAG}-linux
release-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${ARTIFACT} -a .
	tar -cf ${ARTIFACT}.tar ${ARTIFACT}
	git ls-files webroot/ | tr '\n' ' ' | xargs tar -rf ${ARTIFACT}.tar
	gzip ${ARTIFACT}.tar
	rm -rf ${ARTIFACT}

release-darwin: TAG ?= latest
release-darwin: ARTIFACT = flowbro-${TAG}-darwin
release-darwin:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o ${ARTIFACT} -a .
	tar -cf ${ARTIFACT}.tar ${ARTIFACT}
	git ls-files webroot/ | tr '\n' ' ' | xargs tar -rf ${ARTIFACT}.tar
	gzip ${ARTIFACT}.tar
	rm -rf ${ARTIFACT}

release: TAG ?= latest
release: release-linux release-darwin
	git tag ${TAG}
