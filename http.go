package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	// "strings"
)

const letters = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

var code_map = make(map[string]string)

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

func CheckParam(w http.ResponseWriter, r *http.Request, param string) string {
	value := r.URL.Query().Get(param)
	if value == "" {
		http.Error(w, "Incorrect query parameters", http.StatusBadRequest)
		return ""
	}
	return value
}

func CreateMatch(w http.ResponseWriter, r *http.Request) {
	if !CompareMethod(w, r.Method, http.MethodPost) {
		return
	}

	max_players_str := CheckParam(w, r, "max_players")
	if max_players_str == "" {
		log.Println("missing max_players")
		return
	}
	private_str := CheckParam(w, r, "private")
	if private_str == "" {
		log.Println("missing private")
		return
	}
	max_players, err := strconv.Atoi(max_players_str)
	private, err := strconv.ParseBool(private_str)
	if err != nil {
		http.Error(w, "Incorrect query parameters", http.StatusBadRequest)
		return
	}

	id, err := StartContainer(max_players, private)
	if err != nil {
		http.Error(w, "Failed to start container", http.StatusInternalServerError)
		return
	}

	ip := strings.Split(r.RemoteAddr, ":")[0]
	// AddForwardRule(ip, id)
	proxy.AddForwardRule(ip, container_map[id].ip.String())

	code := GenerateCode()

	for ok := true; ok; _, ok = code_map[code] {
		code = GenerateCode()
	}
	code_map[code] = id
	AddMatch(code, max_players, private)

	json := fmt.Sprintf("{\"code\":\"%s\"}\n", code)
	w.WriteHeader(200)
	w.Write([]byte(json))
	log.Println("[Client] created match")
}

func JoinMatch(w http.ResponseWriter, r *http.Request) {
	if !CompareMethod(w, r.Method, http.MethodPost) {
		return
	}

	code := CheckParam(w, r, "code")
	if _, ok := code_map[code]; ok {
		http.Error(w, "Invalid code", http.StatusInternalServerError)
		return
	}

	// ip := strings.Split(r.RemoteAddr, ":")[0]
	// AddForwardRule(ip, code_map[code])

	if val, ok := container_map[code_map[code]]; ok {
		val.num_players += 1
		container_map[code_map[code]] = val
	}

	w.WriteHeader(200)
	w.Write([]byte("")) // TODO: do i need to respond with anything?
	log.Println("[Client] joined match")
}

func LeaveMatch(w http.ResponseWriter, r *http.Request) {
	if !CompareMethod(w, r.Method, http.MethodPost) {
		return
	}

	code := r.URL.Query().Get("code")

	if val, ok := container_map[code_map[code]]; ok {
		val.num_players -= 1
		if val.num_players == 0 {
			log.Println("[Client] last player left removing container")
			StopContainer(code_map[code])
			delete(code_map, code)
			RemoveMatch(code)
			return
		}
		container_map[code_map[code]] = val
	}
	w.WriteHeader(200)
	w.Write([]byte(""))
	log.Println("[Client] left match")
}

func GetMatches(w http.ResponseWriter, r *http.Request) {
	if !CompareMethod(w, r.Method, http.MethodGet) {
		return
	}

	json, err := json.Marshal(matches)
	if err != nil {
		log.Println("[Server] failed to serialize json")
		return
	}
	json = append(json, '\n')

	w.WriteHeader(200)
	w.Write(json)
	log.Println("[Client] get matches")
}

func HandleEndpoints() {
	http.HandleFunc("/create-match", CreateMatch)
	http.HandleFunc("/join-match", JoinMatch)
	http.HandleFunc("/leave-match", LeaveMatch)
	http.HandleFunc("/get-matches", GetMatches)

	http.ListenAndServe(":8000", nil)
}
