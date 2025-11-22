package main

import (
	"context"
	"log"
	"time"

	"grpc/pkg/grpcutils"
	pb "grpc/protobuf/bookstore"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

func main() {
	// Connect to server
	conn := grpcutils.ConnectOrPanic(grpcutils.GRPCClientConfig{
		Address: "localhost:9090",
	})
	defer conn.Close()

	client := pb.NewBookstoreServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test 1: Add a book
	log.Println("=== Test 1: Add Book ===")
	addResp, err := client.AddBook(ctx, &pb.AddBookRequest{
		Title:    "The Go Programming Language",
		Author:   "Alan Donovan & Brian Kernighan",
		Isbn:     "978-0134190440",
		Price:    39.99,
		Quantity: 10,
		Category: "Programming",
	})
	if err != nil {
		handleError(err)
	} else {
		log.Printf("Book added with ID: %s\n", addResp.Id)
	}

	// Test 2: Get the book
	if addResp != nil {
		log.Println("\n=== Test 2: Get Book ===")
		getResp, err := client.GetBook(ctx, &pb.GetBookRequest{
			Id: addResp.Id,
		})
		if err != nil {
			handleError(err)
		} else {
			log.Printf("Retrieved book: %+v\n", getResp.Book)
		}
	}

	// Test 3: List books
	log.Println("\n=== Test 3: List Books ===")
	listResp, err := client.ListBooks(ctx, &pb.ListBooksRequest{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		handleError(err)
	} else {
		log.Printf("Found %d books (total: %d)\n", len(listResp.Books), listResp.TotalCount)
		for _, book := range listResp.Books {
			log.Printf("  - %s by %s ($%.2f)\n", book.Title, book.Author, book.Price)
		}
	}

	// Test 4: Invalid input (should fail)
	log.Println("\n=== Test 4: Add Invalid Book (should fail) ===")
	_, err = client.AddBook(ctx, &pb.AddBookRequest{
		Title: "",  // Empty title should fail validation
		Price: -10, // Negative price should fail
	})
	if err != nil {
		handleError(err)
	}
}

func handleError(err error) {
	st, ok := status.FromError(err)
	if !ok {
		log.Printf("Unknown error: %v\n", err)
		return
	}

	log.Printf("Error code: %s\n", st.Code())
	log.Printf("Error message: %s\n", st.Message())

	// Check for detailed validation errors
	for _, detail := range st.Details() {
		if badReq, ok := detail.(*errdetails.BadRequest); ok {
			log.Println("Validation errors:")
			for _, violation := range badReq.FieldViolations {
				log.Printf("  - %s: %s\n", violation.Field, violation.Description)
			}
		}
	}
}
