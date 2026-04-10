package main

import (
	"context"

	"edge-agent/internal"
)

func main() {
	internal.New().Run(context.Background())
}

