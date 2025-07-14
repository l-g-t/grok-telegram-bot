host=$(shell grep '^HOST=' init/artifacts/.env | cut -d '=' -f 2-)
ver = `cat ./VERSION`
app = grok-telegram-bot
fullname = $(app)-$(ver)

configure-vm:
	sh init/init.sh $(host) $(app)

build-osx:
	@echo "Building osx binary version $(ver)"
	go build -a -trimpath -ldflags="-w -s" -o $(app) .
	@mkdir -p $(fullname)-osx
	@cp $(app) $(fullname)-osx/
	echo "Release $(ver) ($(shell date)): $(notes)" > $(fullname)-osx/$(ver).txt
	@tar czf build/$(fullname)-osx.tar.gz -C $(fullname)-osx .
	@rm -r $(fullname)-osx
	@rm $(app)
	@echo "Built kukoldini binary for osx at build/$(fullname)-osx.tar.gz"

build-linux:
	sh build/build.sh $(app) $(fullname)

test:
	cd ../../ && go test -cover ./...

release: build-linux
	scp build/$(fullname).tar.gz $(host):/tmp/$(app).tar.gz
	ssh $(host) "sudo tar xvf /tmp/$(app).tar.gz -C /opt/$(app)/ --no-overwrite-dir \
	    && sudo systemctl restart $(app).service \
		&& sudo systemctl status $(app).service" # \
		# && sudo journalctl -u $(app) -f"
