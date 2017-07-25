Проект содержит следующие контейнеры:

nginx_router
library_api
library_db

Также необходим контейнер с app
library (https://github.com/ltaa/library)


Для сборки текущего проекта необходимо находясь в папке проекта выполнить:

docker build -t library_api .

cd docker/library_db
docker build -t library_db .

cd ../nginx_router
docker build -t nginx_router .

Также необхоимо собрать контейнер library(https://github.com/ltaa/library)

Запуск с помощью docker-compose

docker-compose -f compose.yml up
