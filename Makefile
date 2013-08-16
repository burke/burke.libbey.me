.PHONY: upload build

upload: build
	rsync -avze 'ssh' public/ root@burke.libbey.me:/var/www/burke.libbey.me

build: bloggy
	./bloggy

bloggy: bloggy.go
	go build -o bloggy
