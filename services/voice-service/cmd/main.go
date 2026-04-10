package main

import (
	"context"

	"voice-service/internal"
)

func main() {
	internal.New().Run(context.Background())
}

