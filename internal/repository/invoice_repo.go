package repository

import (
	"context"
	"log"
	"provider_management/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type InvoiceRepo struct {
	col        *mongo.Collection
}

func NewInvoiceRepo(db *mongo.Database) *InvoiceRepo {
	return &InvoiceRepo{
		col:  db.Collection("invoices"),
	}
}

func (r *InvoiceRepo) GetByID(ctx context.Context, id string) (*domain.Invoice, error) {
	var invoice domain.Invoice
	log.Println("Searching for serviceId:", id)
	
	filter := bson.M{"serviceId": id}
	log.Println("Filter:", filter)
	
	err := r.col.FindOne(ctx, filter).Decode(&invoice)
	if err != nil {
		log.Println("MongoDB error:", err)
		return nil, err
	}
	
	log.Println("Found invoice:", invoice.InvoiceNumber)
	return &invoice, nil
}