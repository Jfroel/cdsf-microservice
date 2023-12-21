.DEFAULT_GOAL := build
.PHONY: fmt lint vet build clean test test_race proto work

fmt:
	go fmt ./...

lint: fmt
	golint ./...

vet: fmt
	go vet ./...

build: vet
	go build -o cmd/main cmd/main.go 

clean:
	rm cmd/main

test:
	go test -v ./test/* -run TestInsertAndRemoveMax -heap 0
	go test -v ./test/* -run TestInsertAndRemoveMax -heap 1


testcon:
	go test -v ./test/* -run TestCon -heap 0
	go test -v ./test/* -run TestCon -heap 1

testpast:
	go test -v ./test/* -run TestInsertPastCapacity -heap 0
	go test -v ./test/* -run TestInsertPastCapacity -heap 1 
	
# go test test/*

test_race:
	go test -race test/*

secret:
	kubectl create secret docker-registry regcred \
	--docker-server=https://index.docker.io/v1/ \
	--docker-username=jamesfroelich \
	--docker-password=QuickFoxLazyDanny \
	--docker-email=f_james123@hotmail.com

kube_combo:
	$(MAKE) kube_clean 
	$(MAKE) kube_build 
	$(MAKE) apply_manifests

kube_build:
	sudo bash scripts/build_images.sh -u jamesfroelich -t cdsf-microservice

apply_manifests:
	kubectl apply -f manifests/

kube_clean:
	kubectl delete svc,po,deploy --all

kube_restart:
	$(MAKE) kube_clean 
	$(MAKE) apply_manifests

proto:
	protoc --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/filter/filter.proto 


