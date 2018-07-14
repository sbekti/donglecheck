build:
	docker build -t shbekti/donglecheck:latest .
	docker tag shbekti/donglecheck:latest shbekti/donglecheck:1.0.0

push:
	docker push shbekti/donglecheck:1.0.0
	docker push shbekti/donglecheck:latest