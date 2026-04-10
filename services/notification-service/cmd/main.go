package main

import (
	"context"

	"notification-service/internal"
)

func main() {
	internal.New().Run(context.Background())
}

