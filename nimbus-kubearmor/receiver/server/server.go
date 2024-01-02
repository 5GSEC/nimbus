package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

var (
	// 저장된 Nimbus Policy를 저장하는 메모리 저장소
	nimbusPolicies []interface{}
	lock           sync.Mutex
)

func main() {
	http.HandleFunc("/api/v1/nimbus/export", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Only POST method is accepted", http.StatusMethodNotAllowed)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var data interface{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
			return
		}

		// 수신된 Nimbus Policy 저장
		lock.Lock()
		nimbusPolicies = append(nimbusPolicies, data)
		lock.Unlock()

		fmt.Printf("Received Nimbus Policy: %+v\n", data)
		w.WriteHeader(http.StatusOK)
	})

	// 저장된 Nimbus Policies를 조회하는 API
	http.HandleFunc("/api/v1/nimbus/policies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Only GET method is accepted", http.StatusMethodNotAllowed)
			return
		}

		lock.Lock()
		defer lock.Unlock()
		if err := json.NewEncoder(w).Encode(nimbusPolicies); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	})

	log.Println("Server starting on port 13000...")
	log.Fatal(http.ListenAndServe(":13000", nil))
}
