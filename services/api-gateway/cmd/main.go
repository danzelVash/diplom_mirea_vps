package main

import (
	"context"

	"api-gateway/internal"
)

func main() {
	internal.New().Run(context.Background())
}
