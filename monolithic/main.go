package main

import (
	"ecommerce/config"
	"ecommerce/pkg/client"
	"ecommerce/routers"
	"log"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func init() {
	config.Envload()
	config.DBconnect()
	client.InitWishlistClient()
	client.InitWalletClient()
}

func main() {
	r := gin.Default()

	files, err := filepath.Glob("views/**/*.html")
	if err != nil {
		log.Fatal(err)
	}

	wishlistFiles, err := filepath.Glob("../wishlist-service/views/*.html")
	if err != nil {
		log.Fatal(err)
	}

	files = append(files, wishlistFiles...)

	if len(files) > 0 {
		r.LoadHTMLFiles(files...)
	} else {
		log.Println("No templates found")
	}

	routers.AdminRoutes(r)
	routers.UserRoutes(r)
	r.Run(":8080")
}
