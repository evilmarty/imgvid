APP_NAME=imgvid
BUILD_PATH=bin/${APP_NAME}
INSTALL_PATH=/usr/bin/${APP_NAME}

compile: *.go
	go build -o ${BUILD_PATH}

run:
	go run *.go

install: compile
	cp -f ${BUILD_PATH} ${INSTALL_PATH}

clean:
	rm -rf ${BUILD_PATH}
