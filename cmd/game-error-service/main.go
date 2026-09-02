package main

import (
	"log"
	"net/http"
	"os"

	"github.com/infrai-examples/game-backend-error-groups/internal/gameerrors"
)

func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		log.Fatal("INFRAI_API_KEY is required")
	}
	handler := gameerrors.NewHandler(gameerrors.NewCaptureClient(key))
	server := &http.Server{Addr: ":8080", Handler: handler, ReadHeaderTimeout: 5 * 1e9}
	log.Println("game error service listening on :8080")
	log.Fatal(server.ListenAndServe())
}
