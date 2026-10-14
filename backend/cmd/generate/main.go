package main

import (
	"github.com/AiEmStylix/ecommerce/backend/internal/models"
	"gorm.io/gen"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./internal/query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g.ApplyBasic(models.Item{}, models.Category{}, models.ItemImage{}, models.University{}, models.User{}, models.Offer{})

	g.Execute()
}
