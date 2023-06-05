builder := go
builddir := bin
exe := $(builddir)/pvevncproxy
app := $(builddir)/app
config := $(builddir)/config.yaml
install := $(builddir)/install.sh
systemd := $(builddir)/pvevncproxy.service

all: $(exe) $(app) $(config) $(systemd) $(install)

$(config): config.yaml
		cp config.yaml $(config)

$(install): install.sh
		cp install.sh $(install)

$(systemd): pvevncproxy.service
		cp pvevncproxy.service $(systemd)

$(exe): main.go go.mod go.sum models middlewares router utils
		$(builder) build -o $(exe) $<

$(app): app $(builddir)
	    cp -r $< $(app)

$(builddir):
		mkdir $(builddir)

.PHONY = clean

clean: 
		rm -r $(builddir)
