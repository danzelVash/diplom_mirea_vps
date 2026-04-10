package main

import (
	"context"

	"scenario-service/internal"
)

func main() {
	internal.New().Run(context.Background())
}

