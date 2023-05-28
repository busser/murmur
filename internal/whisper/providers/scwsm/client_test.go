package scwsm_test

import (
	"context"
	"fmt"
	"log"

	"github.com/busser/whisper/internal/whisper/providers/scwsm"
)

func Example() {
	c, err := scwsm.New()
	if err != nil {
		log.Fatal(err)
	}

	ref := "fr-par/secret-sauce"
	val, err := c.Resolve(context.Background(), ref)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("The secret sauce is", val)
}
