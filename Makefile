# Setup name variables for the package/tool
NAME := coffinjoe
PKG := github.com/vitor0/$(NAME)

CGO_ENABLED := 0

# Set any default go build tags.
BUILDTAGS :=

include basic.mk

.PHONY: prebuild
prebuild:
