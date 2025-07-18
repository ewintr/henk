deploy:
	go build -o henk .
	scp henk server:dist
	mv henk ~/bin
