mpush:
	@git add .
	@git commit -m "$(MSG)"
	@git push

push:
	@git tag "$(TAG)"
	@git push origin "$(TAG)"
