package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	fetchsms "github.com/belaldev/fetchsms-go"
	"log"
	"os"
	"time"
)

func main() {
	spend := flag.Bool("spend", false, "authorize one paid rental")
	flag.Parse()
	if !*spend {
		log.Fatal("This example spends wallet funds; explicitly pass -spend after review.")
	}
	key := os.Getenv("FETCHSMS_API_KEY")
	if key == "" {
		log.Fatal("set FETCHSMS_API_KEY")
	}
	c, err := fetchsms.NewClient(key)
	if err != nil {
		log.Fatal(err)
	}
	service, err := fetchsms.ServiceID(1)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	r, err := c.CreateRental(ctx, fetchsms.CreateRentalRequest{Service: service, Days: 1})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Rental %s: enter %s into a supported service and request its SMS. Press Enter afterward.\n", r.ID, r.Number)
	if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
		log.Fatal(err)
	}
	code, err := c.WaitForRentalCode(ctx, r.ID)
	if err != nil {
		log.Fatal(err)
	}
	_ = code
	fmt.Println("Code received (not printed).")
}
