package awssm_test

import (
	"context"
	"fmt"
	"log"

	"github.com/busser/whisper/internal/whisper/clients/awssm"
)

func Example() {
	c, err := awssm.New()
	if err != nil {
		log.Fatal(err)
	}

	ref := "secret-sauce"
	val, err := c.Resolve(context.Background(), ref)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("The secret sauce is", val)
}
