package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, Go server!")
}

// Модель данных
type Note struct {
	Name string `json:"name"` //Ключ в JSON будет "name"
	Text string `json:"text"` //Ключ в JSON будет "text"
}

// Обработчики
// POST /note - принимает JSON-объект Note и возвращает его обратно
func createNoteHandler(w http.ResponseWriter, r *http.Request) {
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

	//Сформируем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) //201 Created
	_ = json.NewEncoder(w).Encode(in) //Вернем то, что приняли
}

func main() {
	// Используем свой mux(маршрутизатор)
	mux := http.NewServeMux()
	// Регистрируем обработчики
	mux.HandleFunc("GET /", helloHandler)
	mux.HandleFunc("POST /note", createNoteHandler) //Функция-обработчик

	http.ListenAndServe(":8000", mux)
}
