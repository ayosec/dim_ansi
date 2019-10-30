JAIL=docker run --rm -ti -v $(PWD):/host -w /host -e GO111MODULE=on -e HOME=/tmp -u `id -u` golang:1.13.2

all: dim_ansi

clean:
	rm -f dim_ansi

jail:
	$(JAIL)

dim_ansi: dim_ansi.go
	$(JAIL) go build $<

.PHONY: all clean jail
