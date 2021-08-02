CERT_PATH=${HOME}/.yass/

.PHONY: init
init:
	mkdir -p ${CERT_PATH}

genproto:
	protoc --go_out=plugins=grpc:. --go_opt=paths=source_relative api/yass.proto

.PHONY: gencert
gencert:
	cfssl gencert -initca certs/ca-csr.json | cfssljson -bare ca
	cfssl gencert \
        -ca=ca.pem \
        -ca-key=ca-key.pem \
        -config=certs/ca-config.json \
        -profile=server \
        certs/server-csr.json | cfssljson -bare server
	cfssl gencert \
        -ca=ca.pem \
        -ca-key=ca-key.pem \
        -config=certs/ca-config.json \
        -profile=client \
        -cn="client" certs/client-csr.json | cfssljson -bare client
	mv *.pem *.csr ${CERT_PATH}

test:
	go test ./... --cover

