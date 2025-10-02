<<<<<<< HEAD
# Go-server (HTTP-сервер на Go)

Мини-проект для разбора клиент-серверной архитектуры на практике.
Цель — по шагам построить простой REST API только на стандартной библиотеке Go.

## Что уже есть

- `GET /` — проверочный хендлер, выводит текст `Hello, Go server!`.
- `POST /note` — принимает JSON заметки и возвращает её же.

  - Успех: **201 Created** и тело заметки.
  - Ошибки:
    - **400 Bad Request** — невалидный JSON (`{"detail":"invalid JSON"}`)или пустое поле `name` (`{"detail":"name is required"}`).
    - **405 Method Not Allowed** — если вызвать не POST.
- `GET /ping` — возвращает `{"message":"pong"}` (для быстрой проверки JSON).

## Требования

- Go ≥ 1.18 (подойдёт любая), проект без внешних зависимостей.

## Запуск

```bash
go run .
```
=======
# Go-server (учебный HTTP-сервер на Go)

Мини-проект, чтобы руками понять клиент-серверную архитектуру и REST на чистой стандартной библиотеке Go.

## Что уже есть

* `GET /` — проверка, что сервер жив: возвращает `Hello, Go server!`.
* `POST /note` — принимает JSON заметки и возвращает её же.

  Успех:  **201 Created** .

  Ошибки:

  * **400 Bad Request** — невалидный JSON (`{"detail":"invalid JSON"}`) или пустое `name` (`{"detail":"name is required"}`);
  * **405 Method Not Allowed** — если дернуть не POST.
* `GET /ping` → `{"message":"pong"}` (быстрая проверка JSON-ответа).

## Как запустить

```bash
go run .
# сервер слушает http://localhost:8000
```

---

## API и примеры запросов (curl)

### Проверка жизни

```bash
curl http://localhost:8000/
# Hello, Go server!
```

### Создать заметку (успех)

```bash
curl -i -X POST http://localhost:8000/note \
  -H 'Content-Type: application/json' \
  -d '{"name":"first","text":"hello go"}'
```

Ожидаемо:

```
HTTP/1.1 201 Created
Content-Type: application/json
{"name":"first","text":"hello go"}
```

### Пустое имя (валидация)

```bash
curl -i -X POST http://localhost:8000/note \
  -H 'Content-Type: application/json' \
  -d '{"name":"","text":"no name"}'
```

Ответ:

```
HTTP/1.1 400 Bad Request
{"detail":"name is required"}
```

### Не-JSON

```bash
curl -i -X POST http://localhost:8000/note \
  -H 'Content-Type: application/json' \
  -d 'this is not json'
```

Ответ:

```
HTTP/1.1 400 Bad Request
{"detail":"invalid JSON"}
```

---

## Полезные приёмы работы с `curl`

* `-i` — показать статус и заголовки в ответе.
* `-v` — подробный режим (видно, что реально ушло/пришло).
* `-H 'Content-Type: application/json'` — обязательно для JSON-тел.
* `-d '…'` — тело запроса.  **Используй одинарные кавычки `'…'`** , а не обратные ``…`` — те запускают команду в bash.
* Перенос строки `\` допустим,  **но после слеша не должно быть пробелов

 `net/http` — сервер и маршрутизатор (`http.NewServeMux`).

* Обработчики имеют сигнатуру:

  ```go
  func(w http.ResponseWriter, r *http.Request)
  ```
  где `w` — «куда писать ответ», `r` — «что пришло».
* JSON: пакет `encoding/json` (`json.NewDecoder(r.Body).Decode(&in)` / `json.NewEncoder(w).Encode(out)`).
* Модель данных (ожидаемый JSON):

  ```go
  type Note struct {
      Name string `json:"name"`
      Text string `json:"text"`
  }
  ```
* Статусы: `http.StatusCreated` (201), `http.StatusBadRequest` (400), `http.StatusMethodNotAllowed` (405) и т.д.
>>>>>>> 5853640 (Readme.md)
