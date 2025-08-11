SHELL=/bin/zsh

.PHONY: genpb

PROTO_FILES = $(wildcard ./proto/original/*.proto)
PROTO_PLUGIN = protoc-gen-go-futu
PROTO_PLUGIN_FOLDER = ./tools/$(PROTO_PLUGIN)/

all:


$(PROTO_PLUGIN): $(wildcard $(PROTO_PLUGIN_FOLDER)/*.go)
	go build $(PROTO_PLUGIN_FOLDER)


genpb:
	@echo Applying fixproto.awk to originals
	@$(foreach pf, $(PROTO_FILES), \
		$(eval OUTFILE := ./proto/$(basename $(notdir $(pf))).proto);	\
		awk -f ./tools/fixproto.awk $(pf) > $(OUTFILE) ;	\
	)
	@echo Making protoc plugin
	@make $(PROTO_PLUGIN)
	@echo Generating \*.pb.go
	@protoc -I=./proto \
		--go_out=./pb \
		--go_opt=module=github.com/santsai/futu-go/pb \
		--plugin=protoc-gen-go-futu=./${PROTO_PLUGIN} \
		--go-futu_out=./pb	\
		--go-futu_opt=module=github.com/santsai/futu-go/pb \
		./proto/*.proto 2>&1


