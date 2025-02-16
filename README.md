# Балюк Андрей - Тестовое задание на стажировку - Авито 2025
## Запуск
Для запуска кластера в docker используйте команду: `docker compose up -d --build`\
Необходим файл .env в корне проекта со следующими данными(в качестве примера беруться мои тестовые данные):
```
DB_HOST=localhost
DB_NAME=avito
DB_PORT=5432
DB_USER=postgres
DB_PASS=qaleka123

BACKEND_URL=0.0.0.0:8080

DB_HOST_TEST=localhost
DB_NAME_TEST=test
DB_PORT_TEST=5432
DB_USER_TEST=postgres
DB_PASS_TEST=qaleka123
```
Происходит автоматическая миграция бд и запуск сервера по адресу `http://localhost:8080/`\
Если нужно запустить сам сервер, то используйте: `go run .\cmd\webapp`\
## Проблемы и особенности, с которыми я стоклнулся
* В задании не указано, куда следует вставлять `jwt-token` после получения, поэтому я решил указывать его в заголовки запрос:
  ```
  JWT-Token: Bearer <your-jwt-token>
  ```
* На `username`(от 3, до 20 символов: `^[A-Za-zА-Яа-яЁё0-9][A-Za-zА-Яа-яЁё0-9-_.!@#$%^&*()+=-]{3,20}[A-Za-zА-Яа-яЁё0-9]$`) и `password`(от 8 до 16 символов: `^[a-zA-ZА-Яа-яЁё0-9!@#$%^&*()_+=-]{8,16}$`) наложены ограничения, чтобы валидировать несоответсвующие данные(Пример:username из пробелов)
* В сервисе подключено логирование вызываемых методов и всех вызовов в `repository`, а также ошибок.
* Изначально пароль хэшировался, то т.к. эта операция занимает много времени, пришлось убрать.
* Изначально хотел покупки также класть в транзакции пользователя, но в условии указано именно имя пользователя, так что от этой идеи пришлось отказаться
* Присутствуют санитайзер для предотвращения XSS атак и встроенные методы gorm, которые защищают от SQL-инъекций
## Тестирование бизнес сценариев и E2E тесты
Тесты написаны для каждого из слоев архитектуры, общее покрытее = 63,4%\
Для E2E-тестов создает отдельная бд - `test`
Команда для запуска тестов - `go test -coverpkg=./... -coverprofile=cover ./... && cat cover | grep -v "mock" | grep -v  "easyjson" | grep -v "proto" > cover.out && go tool cover -func=cover.out`\
Написны E2E-тесты для каждого сценария. Тесты можно найти в `internal/auth/e2e_tests` и `internal/merch/e2e_tests`\
![image](https://github.com/user-attachments/assets/8fd3434d-b7c9-4a77-8af9-87489dcfe579)\

## Нагрузочное тестирование
Для тестирование используется инструмент k6\
Тестирование проводиться 2 мин. Просходит постепенно повышение числа запросов до 1k RPS, которое удерживается в течении минуты
1. Тестирование `http://localhost:8080/api/info/`:
![image](https://github.com/user-attachments/assets/ad53565f-6fdc-48f2-80fc-083446f19ad6)\
Как видим процент ошибок = 0%, а 90% запросов выполняются быстрее 50ms

3. Тестирование `http://localhost:8080/api/buy/{item}`:
![image](https://github.com/user-attachments/assets/486b537f-8c91-4064-bb0d-d656a04a0766)\
Как видим, процент ошибок = 1.59%, однако это происходит из-за недостатка монет у пользователя при покупке предмета, что вполне разумно. Также 90% запросов выполняются быстрее 50ms

5. Тестирование `http://localhost:8080/api/sendCoins`:
![image](https://github.com/user-attachments/assets/07c606a1-9ffb-41b4-8442-da651a3379ce)\
Как видим, процент ошибок = 0%, а 90% запросов выполняются быстрее 50ms

5. Тестирование `http://localhost:8080/api/auth`:
![image](https://github.com/user-attachments/assets/ee69e422-b392-4b7a-8f98-b713b910cb14)\
Как видим, процент ошибок = 0%, а 90% запросов выполняются быстрее 50ms
