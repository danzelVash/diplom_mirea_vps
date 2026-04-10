package main

import (
	"context"

	"vision-service/internal"
)

func main() {
	internal.New().Run(context.Background())
}
