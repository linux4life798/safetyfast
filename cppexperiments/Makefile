CXXFLAGS += -ggdb -Wall
CXXFLAGS += -march=native
#CXXFLAGS += -march=skylake
CXXFLAGS += -O3
#CXXFLAGS += -O0
LDLIBS+=-mrtm -lpthread

TARGETS = testrtm testhle

all: $(TARGETS)

clean:
	$(RM) $(TARGETS)
