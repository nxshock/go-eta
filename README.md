# go-eta

ETA calculator for Go.

## Example

```go
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/nxshock/go-eta"
)

func main() {
	stepsCount := 1000

	eta := eta.New(time.Minute, stepsCount)

	processed := 0

	// Emulate work
	go func() {
		for processed < stepsCount {
			time.Sleep(time.Second)
			r := rand.Intn(30)

			processed += r
			eta.Increment(r)
		}
	}()

	// Print progress
	for processed < stepsCount {
		time.Sleep(time.Second) // Update progress every second
		fmt.Fprintf(os.Stderr, "\rProcessed %d of %d, ETA: %s", processed, stepsCount, eta.Eta().Format("15:04:05"))
	}

}
```