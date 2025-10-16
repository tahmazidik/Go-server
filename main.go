package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, Go server!")
}

// ожидаем путь вида /note/123
func parseIDFromPath(path, prefix string) (int, bool) {
	if !strings.HasPrefix(path, prefix) {
		return 0, false
	}
	raw := strings.TrimPrefix(path, prefix)
	raw = strings.TrimSuffix(raw, "/")
	id, err := strconv.Atoi(raw)
	return id, err == nil
}

type server struct {
	db *notesDB
}

// GET /note/{id} - возвращает заметку по ID
func (s *server) getNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"detail": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id, ok := parseIDFromPath(r.URL.Path, "/note/")
	if !ok {
		http.NotFound(w, r) //404 страницы не найдена
		return
	}

	n, ok := s.db.get(id)
	if !ok {
		http.Error(w, `{"detail": "note not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(n)
}

func (s *server) delNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"detail": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	id, ok := parseIDFromPath(r.URL.Path, "/note/")
	if !ok {
		http.NotFound(w, r)
		return
	}

	if ok := s.db.del(id); !ok {
		http.Error(w, `{"detail": "note not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// Модель данных
type Note struct {
	ID   int    `json:"id"`   //серверный id
	Name string `json:"name"` //Ключ в JSON будет "name"
	Text string `json:"text"` //Ключ в JSON будет "text"
}

// Маленький объект БД
type notesDB struct {
	mu   sync.RWMutex //защита от гонок: Rlock для чтнения, Lock для записи(много одновременных чтений, но запись только одна)
	data []Note       // таблица заметок
	next int          // счетчик для ID
}

// Конструктор БД
func newDB() *notesDB {
	return &notesDB{next: 1}
}

// Используем RLock() и RUnlock() для чтения
func (db *notesDB) list() []Note {
	db.mu.RLock()
	defer db.mu.RUnlock()
	out := make([]Note, len(db.data)) //Создаем срез нужного размера
	copy(out, db.data)
	return out
}

// Используем Lock() и Unlock() для записи
func (db *notesDB) create(n Note) Note {
	db.mu.Lock()
	defer db.mu.Unlock()
	n.ID = db.next
	db.next++
	db.data = append(db.data, n)
	return n
}

func (db *notesDB) get(id int) (*Note, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	for _, n := range db.data {
		if n.ID == id {
			nc := n //Копия, чтобы наружу не отдавать ссылку на внутренний объект
			return &nc, true
		}
	}
	return nil, false
}

func (db *notesDB) del(id int) bool {
	db.mu.Lock()
	defer db.mu.Unlock()
	for i := range db.data {
		if db.data[i].ID == id {
			db.data = append(db.data[:i], db.data[i+1:]...) // Вырезаем i-й элемент
			return true
		}
	}
	return false
}

func (db *notesDB) update(id int, upd Note) (*Note, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()
	for i := range db.data {
		if db.data[i].ID == id {
			if upd.Name != "" {
				db.data[i].Name = upd.Name
			}
			if upd.Text != "" {
				db.data[i].Text = upd.Text
			}
			nc := db.data[i]
			return &nc, true
		}
	}
	return nil, false
}

// Обработчики
// POST /note - принимает JSON-объект Note и возвращает его обратно
func (s *server) createNote(w http.ResponseWriter, r *http.Request) {
	// Базовая валидация - полезно, если кто-то
	// случайно отправит не-Post на этот путь
	if r.Method != http.MethodPost {
		http.Error(w, `{"detail": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	// Огранич размер тела на всякий случай (защита от гигабайтов)
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

	// Распарсим JSON из тела в структуру
	var in Note
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, `{"detail": "invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Примитивная бизнес-валидация
	if in.Name == "" || in.Text == "" {
		http.Error(w, `{"detail": "name is required"}`, http.StatusBadRequest)
		return
	}

	created := s.db.create(Note{Name: in.Name, Text: in.Text}) //Сохраним в БД

	//Сформируем ответ
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("/note/%d", created.ID))
	w.WriteHeader(http.StatusCreated)      //201 Created
	_ = json.NewEncoder(w).Encode(created) //Вернем то, что приняли
}

func (s *server) listNotes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"detail": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.db.list())
}

// PUT /note/{id} - обновляет заметку по ID
func (s *server) updateNote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"detail": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	//Парсим id из пути
	id, ok := parseIDFromPath(r.URL.Path, "/note/")
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Читаем JSON из тела
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var in Note
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, `{"detail": "invalid JSON"}`, http.StatusBadRequest)
		return
	}

	// Обновляем
	updated, ok := s.db.update(id, in)
	if !ok {
		http.Error(w, `{"detail":"Note not found"}`, http.StatusNotFound)
		return
	}

	//Отвечаем
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(updated)
}

func main() {
	s := &server{db: newDB()} //Инициалзируем сервер с пустой БД
	// Используем свой mux(маршрутизатор)
	mux := http.NewServeMux()
	// Регистрируем обработчики
	mux.HandleFunc("GET /", helloHandler)
	mux.HandleFunc("POST /note", s.createNote) //Функция-обработчик
	mux.HandleFunc("GET /notes", s.listNotes)  //Функция-обработчик
	mux.HandleFunc("GET /note/", s.getNote)    //Функция-обработчик
	mux.HandleFunc("DELETE /note/", s.delNote) //Функция-обработчик
	mux.HandleFunc("PUT /note/", s.updateNote) //Функция-обработчик

	http.ListenAndServe(":8000", mux)
}
