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
	spend := flag.Bool("spend", false, "authorize one paid verification")
	id := flag.Int("service", 4, "numeric service ID from the current catalog")
	flag.Parse()
	if !*spend {
		log.Fatal("This example spends wallet funds. Review it, then explicitly pass -spend.")
	}
	key := os.Getenv("FETCHSMS_API_KEY")
	if key == "" {
		log.Fatal("set FETCHSMS_API_KEY")
	}
	client, err := fetchsms.NewClient(key)
	if err != nil {
		log.Fatal(err)
	}
	service, err := fetchsms.ServiceID(*id)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	v, err := client.CreateVerification(ctx, fetchsms.CreateVerificationRequest{Service: service})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Verification %s: enter %s into the selected service and request an SMS there.\nPress Enter after triggering the SMS.\n", v.ID, v.Number)
	if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
		log.Fatal(err)
	}
	code, err := client.WaitForCode(ctx, v.ID)
	if err != nil {
		log.Fatal(err)
	}
	// Deliver code to your application securely; do not log it.
	_ = code
	fmt.Println("Code received (not printed).")
}
