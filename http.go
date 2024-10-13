package main

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
)

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var code_map map[string]string

func GenerateCode() string {
	code := make([]byte, 5)
	for i := range code {
		code[i] = letters[rand.IntN(len(letters))]
	}

	return string(code)
}

func CompareMethod(w http.ResponseWriter, method1 string, method2 string) bool {
	if method1 != method2 {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func CreateMatch(w http.ResponseWriter, r *http.Request) {
	fmt.Print("test")
	if !CompareMethod(w, r.Method, http.MethodPost) {
		return
	}

	id, err := StartContainer()
	if err != nil {
		http.Error(w, "Failed to start container", http.StatusInternalServerError)
		return
	}

	ip := strings.Split(r.RemoteAddr, ":")[0]
	AddForwardRule(ip, id)

	code := GenerateCode()

	for ok := true; ok; _, ok = code_map[code] {
		code = GenerateCode()
	}
	code_map[code] = id

	json := fmt.Sprintf("{\"code\":\"%s\"}", code)
	w.Write([]byte(json))
}

func JoinMatch(w http.ResponseWriter, r *http.Request) {
	if !CompareMethod(w, r.Method, http.MethodPost) {
		return
	}

	// ip := strings.Split(r.RemoteAddr, ":")[0]
	// AddForwardRule(ip, id)
}

func LeaveMatch(w http.ResponseWriter, r *http.Request) {
	if !CompareMethod(w, r.Method, http.MethodPost) {
		return
	}

}

func GetMatches(w http.ResponseWriter, r *http.Request) {
	if !CompareMethod(w, r.Method, http.MethodGet) {
		return
	}
}

func HandleEndpoints() {
	http.HandleFunc("/create-match", CreateMatch)
	http.HandleFunc("/join-match", JoinMatch)
	http.HandleFunc("/leave-match", LeaveMatch)
	http.HandleFunc("/get-matches", GetMatches)

	http.ListenAndServe(":8000", nil)
}
