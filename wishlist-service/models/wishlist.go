package models

import "gorm.io/gorm"

type Wishlist struct {
	gorm.Model
	UserID uint `gorm:"uniqueIndex"`
}

type WishlistItem struct {
	gorm.Model
	WishlistID uint `gorm:"index"`
	ProductID  uint
	VariantID  uint
}
