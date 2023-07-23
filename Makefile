APP=$(shell basename $(shell git remote get-url origin))
REGISTRY=dletnikov
VERSION=$(shell git describe --tags --abbrev=0)-$(shell git rev-parse --short HEAD)
TARGETOS=linux
TARGETARCH=amd64

#lint:
#		golint
#test:
#		go test -v
#format:
#		gofmt -s -w

log-build:
	cd log && go get && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o log

job-build:
	cd job && go get && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o job

bot-build:
	cd bot && go get && CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o bot


log-image:
	cd log && docker build -t ${REGISTRY}/${APP}-log:${VERSION}-${TARGETOS}-${TARGETARCH} -f Dockerfile.log .

job-image:
	cd job && docker build -t ${REGISTRY}/${APP}-job:${VERSION}-${TARGETOS}-${TARGETARCH} -f Dockerfile.job .

bot-image:
	cd bot && docker build -t ${REGISTRY}/${APP}-bot:${VERSION}-${TARGETOS}-${TARGETARCH} -f Dockerfile.bot .

log-push:
	docker push ${REGISTRY}/${APP}-log:${VERSION}-${TARGETOS}-${TARGETARCH}

job-push:
	docker push ${REGISTRY}/${APP}-job:${VERSION}-${TARGETOS}-${TARGETARCH}

bot-push:
	docker push ${REGISTRY}/${APP}-bot:${VERSION}-${TARGETOS}-${TARGETARCH}




#log-docker-run:
#	docker run -p 8000:8000 ${REGISTRY}/${APP}:${VERSION}-${TARGETOS}-${TARGETARCH}


#log-clean:
#	rm -rf log
#	docker rmi ${REGISTRY}/${APP}:${VERSION}-${TARGETOS}-${TARGETARCH}|| true