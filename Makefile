# figure out what GOROOT is supposed to be
GOROOT ?= $(shell printf 't:;@echo $$(GOROOT)\n' | gomake -f -)
include $(GOROOT)/src/Make.inc

TARG=s3me
GOFILES=\
	s3me.go\

include $(GOROOT)/src/Make.cmd
