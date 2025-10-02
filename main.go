package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, Go server!")
}

// Структура для ответ /ping
type Ping struct {
	Message string `json:"message"` //Тег говорит JSON-эндкодеру использовать ключ "message"
}

// GET /ping -> {"message": "pong"}
func pingHandler(w http.ResponseWriter, r *http.Request) {
	//Сообщаем клиенту формат
	w.Header().Set("Content-Type", "application/json") // Гооворим клиенту, что щас придет JSON
	//Ставим статус 200
	w.WriteHeader(http.StatusOK)
	//Кодируем структуру в JSON прямо в поток ответа
	// Создаем энкодер, который пишет в w, и кодируем структуру в JSON
	_ = json.NewEncoder(w).Encode(Ping{Message: "pong"})

}

func main() {
	// Используем свой mux(маршрутизатор)
	mux := http.NewServeMux()
	// Регистрируем обработчики
	mux.HandleFunc("GET /", helloHandler)
	mux.HandleFunc("GET /ping", pingHandler) //Функция-обработчик

	http.ListenAndServe(":8000", mux)

}
