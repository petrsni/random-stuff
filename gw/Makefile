.PHONY: build install
build:
	go build -o tsgw cmd/tsgw/main.go

install: build
	sudo mv tsgw /usr/sbin
	echo "[Unit]\n\
	Description=Api for writing timeseries data\n\
	After=network.target\n\
	\n\
	[Service]\n\
	ExecStart=/usr/sbin/tsgw $(command) --url=$(url) --port=$(port) --token=$(token) --user=$(user) --pass=$(pass)\n\
	User=root\n\
	\n\
	[Install]\n\
	WantedBy=multi-user.target" | sudo tee /etc/systemd/system/tsgw.service
	sudo systemctl daemon-reload
	sudo systemctl enable tsgw.service
	sudo systemctl start tsgw.service

uninstall:
	sudo systemctl stop tsgw.service
	sudo systemctl disable tsgw.service
	sudo rm /etc/systemd/system/tsgw.service
	sudo systemctl daemon-reload
	sudo systemctl reset-failed
