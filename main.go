package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, Go server!")
}

type server struct {
	db *notesDB
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
	if r.Method != http.MethodPost {
		http.Error(w, `{"detail": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.db.list())
}

func main() {
	s := &server{db: newDB()} //Инициалзируем сервер с пустой БД
	// Используем свой mux(маршрутизатор)
	mux := http.NewServeMux()
	// Регистрируем обработчики
	mux.HandleFunc("GET /", helloHandler)
	mux.HandleFunc("POST /note", s.createNote) //Функция-обработчик
	mux.HandleFunc("GET /notes", s.listNotes)  //Функция-обработчик

	http.ListenAndServe(":8000", mux)
}
