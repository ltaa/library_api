# Состав
Проект содержит следующие образы:    

- nginx_router
- library_api
- library_db

Также необходим образ с app

[library](https://github.com/ltaa/library)

# Сборка
Для сборки текущего проекта необходимо находясь в папке проекта выполнить:
```
docker build -t library_api .

cd docker/library_db && docker build -t library_db .

cd ../nginx_router && docker build -t nginx_router .
```

# Запуск
Запуск с помощью docker-compose
```
docker-compose -f compose.yml up
```
nginx мапится на 4000 порт localhost

# credentials:
```
admin/admin
```
```
test/test
```
