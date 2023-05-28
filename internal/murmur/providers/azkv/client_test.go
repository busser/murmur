package azkv_test

import (
	"context"
	"fmt"
	"log"

	"github.com/busser/murmur/internal/murmur/providers/azkv"
)

func Example() {
	c, err := azkv.New()
	if err != nil {
		log.Fatal(err)
	}

	ref := "example.vault.azure.net/secret-sauce"
	val, err := c.Resolve(context.Background(), ref)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("The secret sauce is", val)
}
