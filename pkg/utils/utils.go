package utils

import(
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func ParseBody(r *http.Request, x interface{}) error {
	reqBody, _ := io.ReadAll(r.Body)
	r.Body.Close()

	err := json.Unmarshal(reqBody, x)
	if err != nil {
		log.Printf("❌ Error decoding body: %v", err.Error())
		return fmt.Errorf("Error decoding JSON")
	}

	return nil
}
