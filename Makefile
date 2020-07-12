build: build_maclist_manager build_mac_ethers

build_maclist_manager:
	mkdir -p bin
	mkdir -p bin/maclist_manager
	go build -o bin/maclist_manager/maclist_manager cmd/maclist_manager/main.go
	test -f cmd/maclist_manager/config/settings.yml && \
		cp cmd/maclist_manager/config/settings.yml bin/maclist_manager/ \
	|| \
		cp cmd/maclist_manager/config/settings.yml.sample bin/maclist_manager/settings.yml

build_mac_ethers:
	mkdir -p bin/mac_ethers
	go build -o bin/mac_ethers/mac_ethers cmd/mac_ethers/main.go
	test -f cmd/mac_ethers/config/settings.yml && \
		cp cmd/mac_ethers/config/settings.yml bin/mac_ethers/ \
	|| \
		cp cmd/mac_ethers/config/settings.yml.sample bin/mac_ethers/settings.yml

clean:
	rm -Rf bin/maclist_manager/
	rm -Rf bin/mac_ethers/
