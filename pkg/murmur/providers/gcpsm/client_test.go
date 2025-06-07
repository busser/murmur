package gcpsm_test

import (
	"context"
	"fmt"
	"log"

	"github.com/busser/murmur/pkg/murmur/providers/gcpsm"
)

func Example() {
	c, err := gcpsm.New()
	if err != nil {
		log.Fatal(err)
	}

	ref := "example-project/secret-sauce"
	val, err := c.Resolve(context.Background(), ref)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("The secret sauce is", val)
}
