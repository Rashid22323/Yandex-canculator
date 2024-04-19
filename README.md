Это проект на Go состоит из двух частей: агента и оркестратора. Агент отвечает за вычисление математических выражений, а оркестратор управляет агентами, хранит результаты вычислений и предоставляет API для взаимодействия.

https://bytepix.ru/ib/DH4GYkPcVw.png

 вот ссылка на объяснение как вообще работает всё.Но добавилось много чего нового.Теперь мы можем хранить данные в SQLite.Также я добавил регистрацию нового пользователя и вторизацию пользователя и получение JWT-токена.
Как запустить и проверить:
Убедитесь, что у вас установлен Go (версия 1.16 или выше), SQLite и git.

Установите советую  go get -u github.com/dgrijalva/jwt-go, go get -u github.com/mattn/go-sqlite3, go get -u google.golang.org/grpc.

Запустите агентов: go run agent/main.go. Запустите оркестратор: go run orchestrator/main.go. Оркестратор будет запущен на порту 8080.

Теперь вы можете взаимодействовать с приложением через API. Для этого можно использовать любой HTTP-клиент, например, curl или Postman.

API предоставляет следующие endpoint'ы:

POST /api/v1/register - регистрация нового пользователя.

POST /api/v1/login - авторизация пользователя и получение JWT-токена.

GET /add?expression={expression} - добавление нового математического выражения для вычисления.

GET /expression?id={id} - получение результата вычисления выражения по его идентификатору.

GET /list - получение списка всех вычисленных выражений.

GET /operations - получение списка доступных математических операций и их времени выполнения.

GET /getTask - получение задачи для вычисления (используется агентами).

POST /receiveResult - отправка результата вычисления агентом оркестратору.

Для проверки работы  вы можете выполнить модульные и интеграционные тесты, расположенные в файлах проекта( они так и назыаются). Для этого необходимо выполнить команду go test.

Также в проекте предусмотрена работа с базой данных SQLite. Все результаты вычислений и информация о пользователях хранится в базе данных expressions.db.
