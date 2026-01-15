SHELL=/bin/zsh

.PHONY: runtest genkey genpb gen_possible_enums
.PHONY: start_opend build_opend install_brew_tools

PROTO_FILES = $(wildcard ./pb/proto/original/*.proto)
PROTO_PLUGIN = protoc-gen-go-futu
PROTO_PLUGIN_FOLDER = ./tools/$(PROTO_PLUGIN)/

all:


$(PROTO_PLUGIN): $(wildcard $(PROTO_PLUGIN_FOLDER)/*.go)
	go build $(PROTO_PLUGIN_FOLDER)


genpb:
	@echo Applying fixproto.awk to originals
	@$(foreach pf, $(PROTO_FILES), \
		$(eval OUTFILE := ./pb/proto/$(basename $(notdir $(pf))).proto);	\
		awk -f ./tools/fixenum.awk -f ./tools/fixproto.awk $(pf) > $(OUTFILE) ;	\
	)
	@echo Making protoc plugin
	@make $(PROTO_PLUGIN)
	@echo Generating \*.pb.go
	@protoc -I=./pb/proto \
		--go_out=./pb \
		--go_opt=module=github.com/santsai/futu-go/pb \
		--plugin=protoc-gen-go-futu=./${PROTO_PLUGIN} \
		--go-futu_out=./pb	\
		--go-futu_opt=module=github.com/santsai/futu-go/pb \
		./pb/proto/*.proto 2>&1

# % container system start
# building opend container image
build_opend:
	cd opend && \
	container build --platform=linux/amd64 . -t opend

# starting opend container
start_opend:
	cd opend && \
	container run --rm -it --name opend --platform linux/amd64 \
		-p 11111:11111 \
		-c 2 -m 1024M \
		-v `pwd`/data:/root/.com.futunn.FutuOpenD \
		opend

install_brew_tools:
	brew install \
		go container \
		protobuf protoc-gen-go \
		uni2ascii

genkey:
	openssl genrsa -out ./opend/data/opend-dev-key.pem 1024

runtest:
	go test -v . -- -privateKey=./opend/data/opend-dev-key.pem

#
# awk:
# 1. replace leading space & trailing space & \r
# 2. double quote string
# 3. print in awk format for easy adding
#
gen_possible_enums:
	grep -h ' int32' pb/proto/original/*proto | \
		awk '{ \
			gsub(/^[[:space:]]+|[[:space:]\r]+$$/, ""); \
			gsub(/"/, "\\\"", $$0); \
			print "enum_replaces[\"" $$0 "\"] = \"\"" \
		}' | \
		uni2ascii -q | sort | uniq | ascii2uni -q > ./tools/possible_enum.txt
