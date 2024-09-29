package main

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/itaraxa/effectivepancake/internal/errors"
)

func main() {
	body := []byte(``)

	req, err := http.NewRequest("POST", "http://localhost:8080/update/counter/Agent1/123", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v", errors.ErrRequestCreating)
		return
	}

	req.Header.Set("Content-Type", "txt/html")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v", errors.ErrRequestSending)
		return
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Status code = 200. Succes request")
	} else {
		fmt.Printf("Status code = %d", resp.StatusCode)
	}

	defer resp.Body.Close()

}
