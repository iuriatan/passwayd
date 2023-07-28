test:
	go build
	go test

clean:
	rm ./paswayd

install-client:
	cp passway /usr/local/bin
	cp passway.conf /etc/
	cp passway.service passway.timer /etc/systemd/system

install-server:
	mv passwayd /usr/local/bin
	cp passwayd.service /etc/systemd/system/
	