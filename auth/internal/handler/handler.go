package handler

import "net/http"

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}