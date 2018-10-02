BIN = ./benchmarks
#PARAMS += -lock systemmutex,spinhlemutex,spinrtmnopause,spinrtmwithlibrary

HOSTNAME = $(shell hostname|tr '.' '\n'|head -n1)
OUTPUT = output-$(HOSTNAME).svg
INFOFILE = info-$(HOSTNAME).txt

.PHONY: all clean cleanall watch

all: $(OUTPUT)

$(BIN): $(wildcard *.go)
	go build

$(OUTPUT): $(BIN)
	date >  $(INFOFILE)
	echo >> $(INFOFILE)
	who  >> $(INFOFILE)
	$(BIN) $(PARAMS) -plot $@

outputfilename:
	echo $(OUTPUT)

clean:
	$(RM) $(BIN)
	$(RM) $(OUTPUT)
	$(RM) $(INFOFILE)

cleanall:
	$(RM) *.svg
	$(RM) *.txt

watch:
	while true; do make; sleep 1; done
