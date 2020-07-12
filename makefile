DIR=$(shell pwd)
BIN=bin
TAR=main

run: compile
	@${BIN}/${TAR}

compile: gobin ${BIN}
	@GOOP=linux GOARCH=386
	@go install
	@mv ${BIN}/TwitchLib ${BIN}/${TAR}

gobin:
	@go env -w GOBIN="${DIR}/${BIN}"
	
${BIN}:
	@mkdir $@
