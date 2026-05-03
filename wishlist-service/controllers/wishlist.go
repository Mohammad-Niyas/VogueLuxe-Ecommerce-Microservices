package controllers

import (
	"context"

	"ecommerce/wishlist-service/models"
	"ecommerce/wishlist-service/pkg/pb"

	"gorm.io/gorm"
)

type Server struct {
	pb.UnimplementedWishlistServiceServer 
	DB                                    *gorm.DB
}


func (s *Server) AddToWishlist(ctx context.Context, req *pb.AddRequest) (*pb.AddResponse, error) {
	var wishlist models.Wishlist

	if err := s.DB.FirstOrCreate(&wishlist, models.Wishlist{UserID: uint(req.UserId)}).Error; err != nil {
		return &pb.AddResponse{Status: "Error", Message: "Failed to get wishlist"}, nil
	}

	var existingItem models.WishlistItem
	if err := s.DB.Where("wishlist_id = ? AND product_id = ? AND variant_id = ?", wishlist.ID, req.ProductId, req.VariantId).First(&existingItem).Error; err == nil {
		return &pb.AddResponse{Status: "Exists", Message: "Item already in wishlist"}, nil
	}

	item := models.WishlistItem{
		WishlistID: wishlist.ID,
		ProductID:  uint(req.ProductId),
		VariantID:  uint(req.VariantId),
	}

	if err := s.DB.Create(&item).Error; err != nil {
		return &pb.AddResponse{Status: "Error", Message: "Failed to add item"}, nil
	}

	return &pb.AddResponse{Status: "Success", Message: "Item added to wishlist"}, nil
}


func (s *Server) RemoveFromWishlist(ctx context.Context, req *pb.RemoveRequest) (*pb.RemoveResponse, error) {
	var wishlist models.Wishlist
	if err := s.DB.Where("user_id = ?", req.UserId).First(&wishlist).Error; err != nil {
		return &pb.RemoveResponse{Status: "Error", Message: "Wishlist not found"}, nil
	}

	if err := s.DB.Delete(&models.WishlistItem{}, "id = ? AND wishlist_id = ?", req.WishlistItemId, wishlist.ID).Error; err != nil {
		return &pb.RemoveResponse{Status: "Error", Message: "Failed to delete item"}, nil
	}

	return &pb.RemoveResponse{Status: "Success", Message: "Item removed"}, nil
}

func (s *Server) GetWishlist(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	var wishlist models.Wishlist
	if err := s.DB.Where("user_id = ?", req.UserId).First(&wishlist).Error; err != nil {
		return &pb.GetResponse{Items: []*pb.WishlistItem{}}, nil
	}

	var items []models.WishlistItem
	if err := s.DB.Where("wishlist_id = ?", wishlist.ID).Find(&items).Error; err != nil {
		return &pb.GetResponse{Items: []*pb.WishlistItem{}}, nil
	}

	var pbItems []*pb.WishlistItem
	for _, item := range items {
		pbItems = append(pbItems, &pb.WishlistItem{
			Id:        uint32(item.ID),
			ProductId: uint32(item.ProductID),
			VariantId: uint32(item.VariantID),
		})
	}

	return &pb.GetResponse{Items: pbItems}, nil
}
