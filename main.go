package main

import (
	"log"
	"net/http"
)

func main() {
	// 1. Create a new http.ServeMux
	mux := http.NewServeMux()

	// Register a handler function with the multiplexer
	//	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//		fmt.Fprintf(w, "Hello, Go Web Server!\\n")
	//	})

	// 2. Create a new http.Server struct
	server := &http.Server{
		Addr:    ":8080", // 3. Set the .Addr field to ":8080"
		Handler: mux,     // 4. Use the new "ServeMux" as the server's handler
	}

	log.Printf("Server starting on %s\n", server.Addr)

	// 5. Use the server's ListenAndServe method to start the server
	if err := server.ListenAndServe(); err != nil {
		// Log a fatal error if the server fails to start
		log.Fatalf("Server error: %v", err)
	}
}
