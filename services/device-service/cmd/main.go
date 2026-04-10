package main

import (
	"context"

	"device-service/internal"
)

func main() {
	internal.New().Run(context.Background())
}

