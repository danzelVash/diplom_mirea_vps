package main

import (
	"context"

	"context-service/internal"
)

func main() {
	internal.New().Run(context.Background())
}

