# Go-server (HTTP-сервер на Go)

Мини-проект, чтобы руками понять клиент-серверную архитектуру и REST на чистой стандартной библиотеке Go.

## Что уже есть

* `GET /` — проверка, что сервер жив: возвращает `Hello, Go server!`.
* `POST /note` — принимает JSON заметки и возвращает её же.

  Успех:  **201 Created** .

  Ошибки:

  * **400 Bad Request** — невалидный JSON (`{"detail":"invalid JSON"}`) или пустое `name` (`{"detail":"name is required"}`);
  * **405 Method Not Allowed** — если дернуть не POST.
* `GET /ping` → `{"message":"pong"}` (быстрая проверка JSON-ответа).
* `GET /notes` — вернуть список заметок.
* `GET /note/{id}` — вернуть заметку по id.
* `DELETE /note/{id}` — удалить заметку по id.

<h2>Модель данных </h2>

```{
  "id":   1,           // назначается сервером
  "name": "string",    // обязательно
  "text": "string"     // опционально
}
```

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

<h3> Получить список </h3>

```
curl -i http://localhost:8000/note/1
# HTTP/1.1 200 OK
# {"id":1,"name":"first","text":"hello"}

```

<h3>Получить заметку по id </h3>

```
curl -i http://localhost:8000/note/1
# HTTP/1.1 200 OK
# {"id":1,"name":"first","text":"hello"}

```

<h3> Удалить заметку </h3>

```
curl -i -X DELETE http://localhost:8000/note/1
# HTTP/1.1 200 OK
# {"status":"ok"}

curl -i -X DELETE http://localhost:8000/note/1
# HTTP/1.1 404 Not Found
# {"detail":"Note not found"}

```

---

## Полезные приёмы работы с `curl`

* `-i` — показать статус и заголовки в ответе.
* `-v` — подробный режим (видно, что реально ушло/пришло).
* `-H 'Content-Type: application/json'` — обязательно для JSON-тел.
* `-d '…'` — тело запроса.  **Используй одинарные кавычки `'…'`** , а не обратные ``…`` — те запускают команду в bash.
* Перенос строки `\` допустим,  но после слеша не должно быть пробелов

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

  ---

  ## Как это устроено


  * **Маршрутизатор** — `http.NewServeMux()`. Регистрируем пути через `HandleFunc`.
  * **Хендлер** — обычная функция

    ```go
    func(w http.ResponseWriter, r *http.Request)
    ```
    где `r` — всё про запрос, `w` — куда писать ответ.
  * **JSON** :
  * чтение: `json.NewDecoder(r.Body).Decode(&in)`
  * запись: `json.NewEncoder(w).Encode(out)`
  * **Память/БД** :

  ```go
    type notesDB struct {
        mu   sync.RWMutex // замок
        data []Note       // "таблица" заметок
        next int          // автоинкремент id
    }
  ```
  * для  **чтения** : `RLock()` / `RUnlock()` — много одновременных читателей OK;
  * для  **записи** : `Lock()` / `Unlock()` — только один писатель;
  * в `list()` отдаём **копию** среза, чтобы внешние изменения не портили внутреннее состояние:
    ```go
    out := make([]Note, len(db.data))
    copy(out, db.data)
    return out
    ```
  * **Парсинг `{id}` из пути** : отрезаем префикс `"/note/"`, убираем хвостовой `/`, преобразуем строку в число (`strconv.Atoi`). Если не вышло — 404.

  ---

  ## Типовые статусы

  * `201 Created` — создан ресурс (`POST /note`), в заголовке `Location: /note/{id}`.
  * `200 OK` — обычные успешные ответы.
  * `400 Bad Request` — невалидный JSON или нарушена валидация.
  * `404 Not Found` — заметка не найдена (неверный `id`).
  * `405 Method Not Allowed` — метод не подходит для маршрута.
