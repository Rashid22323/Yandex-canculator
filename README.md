Это проект на Go состоит из двух частей: агента и оркестратора. Агент отвечает за вычисление математических выражений, а оркестратор управляет агентами, хранит результаты вычислений и предоставляет API для взаимодействия.

https://bytepix.ru/ib/DH4GYkPcVw.png
вот ссылка на объяснение и снизу поясню ещё раз.

Вот кратко как работает сам процесс:

Этот код(Агент и оркестратор.go) является реализацией распределенного калькулятора, состоящего из двух частей: агента и оркестратора. Агент отвечает за вычисление математических выражений, а оркестратор управляет агентами, хранит результаты вычислений и предоставляет API для взаимодействия.

Агент реализован в виде gRPC-сервера, который ожидает запросы на вычисление математических выражений и возвращает результат. Оркестратор реализован в виде HTTP-сервера, который предоставляет API для регистрации и авторизации пользователей, добавления новых математических выражений для вычисления, получения результатов вычислений и списка всех вычисленных выражений. Оркестратор также хранит результаты вычислений и информацию о пользователях в базе данных SQLite.

Общение между агентом и оркестратором происходит с помощью gRPC. Оркестратор выбирает случайным образом одного из доступных агентов и отправляет ему запрос на вычисление математического выражения. Агент вычисляет результат и возвращает его оркестратору, который сохраняет его в базе данных и обновляет статус вычисления.


Теперь мы можем хранить данные в SQLite.Также я добавил регистрацию нового пользователя и авторизацию пользователя и получение JWT-токена.

Также  общение вычислителя и сервера вычислений теперь реализовано с помощью GRPC.(не доделал из-за этого проект не запускается)

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

Для проверки работы  вы можете выполнить модульные и интеграционные тесты, расположенные в файлах проекта( они так и называются).

Также в проекте предусмотрена работа с базой данных SQLite. Все результаты вычислений и информация о пользователях хранится в базе данных expressions.db.(оркестратор)


В проекте используются следующие технологии:

Go - язык программирования.

gRPC - система удаленных вызовов процедур.

Protocol Buffers - язык описания данных и формат сериализации.

SQLite - реляционная база данных.

JWT - стандарт авторизации и аутентификации.

HTTP - протокол передачи гипертекста.

JSON - формат обмена данными.

Git - система управления версиями.

SQL - язык запросов к базам данных.

HTTP-запросы - методы взаимодействия с API.

Модульные тесты - тесты отдельных функций и методов.

Интеграционные тесты - тесты взаимодействия между компонентами системы.

Пример работы:

1)Пользователь отправляет запрос на регистрацию в системе, передавая логин и пароль. Например:

POST /api/v1/register
{
  "login": "testuser",
  "password": "testpassword"
}

2) После успешной регистрации пользователь может войти в систему, отправив запрос на авторизацию. Например:

POST /api/v1/login
{
  "login": "testuser",
  "password": "testpassword"
}

3) В ответ на успешную авторизацию пользователь получает токен JWT, который необходимо использовать для последующего взаимодействия с API.

4)Пользователь может добавить новое математическое выражение для вычисления, отправив запрос на соответствующий endpoint. Например:

GET /add?expression=1+2

5) После того, как выражение будет вычислено, пользователь может получить результат, отправив запрос на получение результата по идентификатору выражения. Например:

GET /expression?id=1

6) Пользователь также может получить список всех вычисленных выражений, отправив запрос на соответствующий endpoint. Например:

GET /list


Также я вроде как реализовал интерфейс

Если что у меня возможно что-то съехало,надеюсь ты поймёшь!!!Удачи!!!(по вопросам писать в телегу:https://t.me/Propandatist 
я возможно что-то не так написал запустить я просто только учусь в гитхабе(никогда им  до этого не пользовался)


Уважаемый проверяющий!(у меня оказалось что код не запускается  так как я неправильно реализовал GRPC но я уже ничего не смогу поменять так как нахожусь в деревне без интернета(пишу с телефона)(пожалуйста посмотри просто структуру кода без grpc  я вроде остальное всё реализовал я ещё только новичок в гитхабе только учусь/спасибо за понимание(пожалуйста не ставьте 0 если проект не запуститься просто я уже ничего изменить не могу)(но вроде всё остальное я реализовал)Удачи ещё раз!!!

