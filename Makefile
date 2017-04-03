release:
	docker-compose build app
	docker-compose run app release
	docker-compose build serve
	docker tag pushkiwi_serve:latest lukin0110/push.kiwi
	docker push lukin0110/push.kiwi
