package controllers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"os"
	"time"

	"ecommerce/wallet-service/config"
	"ecommerce/wallet-service/models"
	"ecommerce/wallet-service/pkg/pb"

	"github.com/google/uuid"
	"github.com/razorpay/razorpay-go"
	"gorm.io/gorm"
)

type WalletServer struct {
	pb.UnimplementedWalletServiceServer
}

func (s *WalletServer) GetWallet(ctx context.Context, req *pb.GetWalletRequest) (*pb.GetWalletResponse, error) {
	var wallet models.Wallet
	if err := config.DB.Where("user_id = ?", req.UserId).First(&wallet).Error; err != nil {
		wallet = models.Wallet{UserID: uint(req.UserId), Balance: 0.00}
		if err := config.DB.Create(&wallet).Error; err != nil {
			log.Printf("Failed to create wallet: %v", err)
			return nil, err
		}
	}

	var transactions []models.WalletTransaction
	if err := config.DB.Where("wallet_id = ?", wallet.ID).Order("transaction_date DESC").Find(&transactions).Error; err != nil {
		log.Printf("Failed to fetch transactions: %v", err)
		return nil, err
	}

	pbWallet := &pb.Wallet{
		Id:      uint32(wallet.ID),
		UserId:  uint32(wallet.UserID),
		Balance: wallet.Balance,
	}

	var pbTransactions []*pb.Transaction
	for _, t := range transactions {
		pbTransactions = append(pbTransactions, &pb.Transaction{
			TransactionUid: t.TransactionUID,
			Amount:         t.TransactionAmount,
			Type:           t.TransactionType,
			Status:         t.TransactionStatus,
			Date:           t.TransactionDate.Format("Jan 02, 2006"),
			Description:    t.Description,
		})
	}

	return &pb.GetWalletResponse{
		Wallet:       pbWallet,
		Transactions: pbTransactions,
	}, nil
}

func (s *WalletServer) CreateTransaction(ctx context.Context, req *pb.CreateTransactionRequest) (*pb.CreateTransactionResponse, error) {
	if req.Amount <= 0 {
		return &pb.CreateTransactionResponse{Status: "Fail", Message: "Invalid amount"}, nil
	}

	var wallet models.Wallet
	if err := config.DB.Where("user_id = ?", req.UserId).First(&wallet).Error; err != nil {
		wallet = models.Wallet{UserID: uint(req.UserId), Balance: 0.00}
		config.DB.Create(&wallet)
	}

	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))

	shortUUID := uuid.New().String()[:8]
	receipt := "wallet_" + shortUUID

	data := map[string]interface{}{
		"amount":   int(req.Amount * 100),
		"currency": "INR",
		"receipt":  receipt,
	}
	body, err := client.Order.Create(data, nil)
	if err != nil {
		log.Printf("Failed to create Razorpay order: %v", err)
		return &pb.CreateTransactionResponse{Status: "Fail", Message: "Failed to create Razorpay order"}, nil
	}

	razorpayOrderID, ok := body["id"].(string)
	if !ok {
		return &pb.CreateTransactionResponse{Status: "Fail", Message: "Invalid Razorpay order response"}, nil
	}

	transaction := models.WalletTransaction{
		WalletID:          wallet.ID,
		TransactionUID:    "TXN-" + uuid.New().String(),
		TransactionAmount: req.Amount,
		TransactionType:   "credit",
		TransactionStatus: "Pending",
		TransactionDate:   time.Now(),
		Description:       "Wallet Top-up",
	}
	if err := config.DB.Create(&transaction).Error; err != nil {
		return &pb.CreateTransactionResponse{Status: "Fail", Message: "Failed to create transaction"}, nil
	}

	return &pb.CreateTransactionResponse{
		Status:          "Success",
		RazorpayOrderId: razorpayOrderID,
		Amount:          float64(int(req.Amount * 100)),
		Currency:        "INR",
		Key:             os.Getenv("RAZORPAY_KEY_ID"),
		TransactionId:   transaction.TransactionUID,
		Name:            "User",             
		Email:           "user@example.com", 
	}, nil
}

func (s *WalletServer) ConfirmPayment(ctx context.Context, req *pb.ConfirmPaymentRequest) (*pb.ConfirmPaymentResponse, error) {
	if req.TransactionId == "" {
		return &pb.ConfirmPaymentResponse{Status: "Fail", Message: "Missing transaction ID"}, nil
	}

	var wallet models.Wallet
	if err := config.DB.Where("user_id = ?", req.UserId).First(&wallet).Error; err != nil {
		return &pb.ConfirmPaymentResponse{Status: "Fail", Message: "Wallet not found"}, nil
	}

	var transaction models.WalletTransaction
	if err := config.DB.Where("transaction_uid = ? AND wallet_id = ?", req.TransactionId, wallet.ID).First(&transaction).Error; err != nil {
		return &pb.ConfirmPaymentResponse{Status: "Fail", Message: "Transaction not found"}, nil
	}

	paymentStatus := "Failed"
	redirectURL := "/wallet?status=failed" 
	message := req.Error

	if req.Error == "" {
		payload := req.RazorpayOrderId + "|" + req.RazorpayPaymentId
		if !verifySignature(payload, req.RazorpaySignature, os.Getenv("RAZORPAY_KEY_SECRET")) {
			message = "Invalid Razorpay signature"
		} else {
			paymentStatus = "Completed"
			redirectURL = "/wallet?status=success"
			message = "Payment Confirmed"
		}
	}

	tx := config.DB.Begin()
	transaction.TransactionStatus = paymentStatus

	if paymentStatus == "Completed" {
		if err := tx.Model(&wallet).Update("Balance", gorm.Expr("balance + ?", transaction.TransactionAmount)).Error; err != nil {
			tx.Rollback()
			return nil, errors.New("failed to update balance")
		}
	}

	if err := tx.Save(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to save transaction")
	}
	tx.Commit()

	if message != "" && paymentStatus == "Failed" {
		redirectURL += "&error=" + message
	}

	return &pb.ConfirmPaymentResponse{
		Status:      paymentStatus,
		Message:     message,
		RedirectUrl: redirectURL,
	}, nil
}

func verifySignature(payload, signature, secret string) bool {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil)) == signature
}

func (s *WalletServer) DebitWallet(ctx context.Context, req *pb.DebitWalletRequest) (*pb.DebitWalletResponse, error) {
	if req.Amount <= 0 {
		return &pb.DebitWalletResponse{Status: "Fail", Message: "Invalid amount"}, nil
	}

	var wallet models.Wallet
	if err := config.DB.Where("user_id = ?", req.UserId).First(&wallet).Error; err != nil {
		return &pb.DebitWalletResponse{Status: "Fail", Message: "Wallet not found"}, nil
	}

	if wallet.Balance < req.Amount {
		return &pb.DebitWalletResponse{Status: "Fail", Message: "Insufficient wallet balance"}, nil
	}

	tx := config.DB.Begin()

	if err := tx.Model(&wallet).Update("balance", gorm.Expr("balance - ?", req.Amount)).Error; err != nil {
		tx.Rollback()
		return &pb.DebitWalletResponse{Status: "Fail", Message: "Failed to update balance"}, nil
	}

	transaction := models.WalletTransaction{
		WalletID:          wallet.ID,
		TransactionUID:    "TXN-" + uuid.New().String(),
		TransactionAmount: req.Amount,
		TransactionType:   "debit",
		TransactionStatus: "Completed",
		TransactionDate:   time.Now(),
		Description:       "Payment for Order " + req.OrderId,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return &pb.DebitWalletResponse{Status: "Fail", Message: "Failed to record transaction"}, nil
	}

	tx.Commit()

	return &pb.DebitWalletResponse{
		Status:        "Success",
		Message:       "Payment Successful",
		TransactionId: transaction.TransactionUID,
	}, nil
}
