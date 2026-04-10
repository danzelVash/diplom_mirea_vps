package main

import (
	"context"

	"edge-bridge-service/internal"
)

func main() {
	internal.New().Run(context.Background())
}

