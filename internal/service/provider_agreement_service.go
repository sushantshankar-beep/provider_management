// service/agreement_service.go
package service

import (
	"context"
	"html/template"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"time"
)

type AgreementService struct {
	agreementRepo *repository.AgreementRepo
}

func NewAgreementService(ar *repository.AgreementRepo) *AgreementService {
	return &AgreementService{
		agreementRepo: ar,
	}
}

type ParagraphResponse struct {
	Number  string `json:"number"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type AgreementResponse struct {
	ID              string              `json:"_id"`
	Title           string              `json:"title"`
	Paragraphs      []ParagraphResponse `json:"paragraphs"`
	AgreementOf     string              `json:"agreementOf"`
	AgreementFor    string              `json:"agreementFor"`
	CompanyName     string              `json:"companyName"`
	Annexures       string              `json:"annexures"`
	CommercialTerms string              `json:"commercialTerms"`
	MarketplaceFee  string              `json:"marketplaceFee"`
	PaymentGateway  string              `json:"paymentGateway"`
	AdditionalNotes string              `json:"additionalNotes"`
	CreatedAt       string              `json:"createdAt"`
	UpdatedAt       string              `json:"updatedAt"`
}

type SafeHTMLAgreementResponse struct {
	ID              string                  `json:"_id"`
	Title           string                  `json:"title"`
	Paragraphs      []SafeParagraphResponse `json:"paragraphs"`
	AgreementOf     template.HTML           `json:"agreementOf"`
	AgreementFor    template.HTML           `json:"agreementFor"`
	CompanyName     template.HTML           `json:"companyName"`
	Annexures       template.HTML           `json:"annexures"`
	CommercialTerms template.HTML           `json:"commercialTerms"`
	MarketplaceFee  template.HTML           `json:"marketplaceFee"`
	PaymentGateway  template.HTML           `json:"paymentGateway"`
	AdditionalNotes template.HTML           `json:"additionalNotes"`
	CreatedAt       string                  `json:"createdAt"`
	UpdatedAt       string                  `json:"updatedAt"`
}

type SafeParagraphResponse struct {
	Number  string        `json:"number"`
	Title   string        `json:"title"`
	Content template.HTML `json:"content"`
}

type CreateAgreementRequest struct {
	Title           string              `json:"title" binding:"required"`
	Paragraphs      []ParagraphResponse `json:"paragraphs" binding:"required"`
	AgreementOf     string              `json:"agreementOf"`
	AgreementFor    string              `json:"agreementFor"`
	CompanyName     string              `json:"companyName"`
	Annexures       string              `json:"annexures"`
	CommercialTerms string              `json:"commercialTerms"`
	MarketplaceFee  string              `json:"marketplaceFee"`
	PaymentGateway  string              `json:"paymentGateway"`
	AdditionalNotes string              `json:"additionalNotes"`
}

type UpdateAgreementRequest struct {
	Title           string              `json:"title"`
	Paragraphs      []ParagraphResponse `json:"paragraphs"`
	AgreementOf     string              `json:"agreementOf"`
	AgreementFor    string              `json:"agreementFor"`
	CompanyName     string              `json:"companyName"`
	Annexures       string              `json:"annexures"`
	CommercialTerms string              `json:"commercialTerms"`
	MarketplaceFee  string              `json:"marketplaceFee"`
	PaymentGateway  string              `json:"paymentGateway"`
	AdditionalNotes string              `json:"additionalNotes"`
}

func (s *AgreementService) GetAgreement(ctx context.Context, id string) (*AgreementResponse, error) {
	a, err := s.agreementRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	paragraphs := make([]ParagraphResponse, len(a.Paragraphs))
	for i, p := range a.Paragraphs {
		paragraphs[i] = ParagraphResponse{
			Number:  p.Number,
			Title:   p.Title,
			Content: p.Content,
		}
	}

	return &AgreementResponse{
		ID:              a.ID.Hex(),
		Title:           a.Title,
		Paragraphs:      paragraphs,
		AgreementOf:     a.AgreementOf,
		AgreementFor:    a.AgreementFor,
		CompanyName:     a.CompanyName,
		Annexures:       a.Annexures,
		CommercialTerms: a.CommercialTerms,
		MarketplaceFee:  a.MarketplaceFee,
		PaymentGateway:  a.PaymentGateway,
		AdditionalNotes: a.AdditionalNotes,
		CreatedAt:       a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       a.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *AgreementService) GetAgreementSafeHTML(ctx context.Context, id string) (*SafeHTMLAgreementResponse, error) {
	a, err := s.agreementRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	paragraphs := make([]SafeParagraphResponse, len(a.Paragraphs))
	for i, p := range a.Paragraphs {
		paragraphs[i] = SafeParagraphResponse{
			Number:  p.Number,
			Title:   p.Title,
			Content: template.HTML(p.Content),
		}
	}

	return &SafeHTMLAgreementResponse{
		ID:              a.ID.Hex(),
		Title:           a.Title,
		Paragraphs:      paragraphs,
		AgreementOf:     template.HTML(a.AgreementOf),
		AgreementFor:    template.HTML(a.AgreementFor),
		CompanyName:     template.HTML(a.CompanyName),
		Annexures:       template.HTML(a.Annexures),
		CommercialTerms: template.HTML(a.CommercialTerms),
		MarketplaceFee:  template.HTML(a.MarketplaceFee),
		PaymentGateway:  template.HTML(a.PaymentGateway),
		AdditionalNotes: template.HTML(a.AdditionalNotes),
		CreatedAt:       a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       a.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *AgreementService) CreateAgreement(ctx context.Context, req CreateAgreementRequest) (*AgreementResponse, error) {
	paragraphs := make([]domain.Paragraph, len(req.Paragraphs))
	for i, p := range req.Paragraphs {
		paragraphs[i] = domain.Paragraph{
			Number:  p.Number,
			Title:   p.Title,
			Content: p.Content,
		}
	}

	agreement := &domain.Agreement{
		Title:           req.Title,
		Paragraphs:      paragraphs,
		AgreementOf:     req.AgreementOf,
		AgreementFor:    req.AgreementFor,
		CompanyName:     req.CompanyName,
		Annexures:       req.Annexures,
		CommercialTerms: req.CommercialTerms,
		MarketplaceFee:  req.MarketplaceFee,
		PaymentGateway:  req.PaymentGateway,
		AdditionalNotes: req.AdditionalNotes,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.agreementRepo.Create(ctx, agreement); err != nil {
		return nil, err
	}

	respParagraphs := make([]ParagraphResponse, len(paragraphs))
	for i, p := range paragraphs {
		respParagraphs[i] = ParagraphResponse{
			Number:  p.Number,
			Title:   p.Title,
			Content: p.Content,
		}
	}

	return &AgreementResponse{
		ID:              agreement.ID.Hex(),
		Title:           agreement.Title,
		Paragraphs:      respParagraphs,
		AgreementOf:     agreement.AgreementOf,
		AgreementFor:    agreement.AgreementFor,
		CompanyName:     agreement.CompanyName,
		Annexures:       agreement.Annexures,
		CommercialTerms: agreement.CommercialTerms,
		MarketplaceFee:  agreement.MarketplaceFee,
		PaymentGateway:  agreement.PaymentGateway,
		AdditionalNotes: agreement.AdditionalNotes,
		CreatedAt:       agreement.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       agreement.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *AgreementService) UpdateAgreement(ctx context.Context, id string, req UpdateAgreementRequest) (*AgreementResponse, error) {
	existing, err := s.agreementRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Paragraphs != nil {
		paragraphs := make([]domain.Paragraph, len(req.Paragraphs))
		for i, p := range req.Paragraphs {
			paragraphs[i] = domain.Paragraph{
				Number:  p.Number,
				Title:   p.Title,
				Content: p.Content,
			}
		}
		existing.Paragraphs = paragraphs
	}
	if req.AgreementFor != "" {
		existing.AgreementFor = req.AgreementFor
	}
	if req.AgreementOf != "" {
		existing.AgreementOf = req.AgreementOf
	}
	if req.CompanyName != "" {
		existing.CompanyName = req.CompanyName
	}
	if req.Annexures != "" {
		existing.Annexures = req.Annexures
	}
	if req.CommercialTerms != "" {
		existing.CommercialTerms = req.CommercialTerms
	}
	if req.MarketplaceFee != "" {
		existing.MarketplaceFee = req.MarketplaceFee
	}
	if req.PaymentGateway != "" {
		existing.PaymentGateway = req.PaymentGateway
	}
	if req.AdditionalNotes != "" {
		existing.AdditionalNotes = req.AdditionalNotes
	}
	existing.UpdatedAt = time.Now()

	if err := s.agreementRepo.Update(ctx, id, existing); err != nil {
		return nil, err
	}

	respParagraphs := make([]ParagraphResponse, len(existing.Paragraphs))
	for i, p := range existing.Paragraphs {
		respParagraphs[i] = ParagraphResponse{
			Number:  p.Number,
			Title:   p.Title,
			Content: p.Content,
		}
	}

	return &AgreementResponse{
		ID:              existing.ID.Hex(),
		Title:           existing.Title,
		Paragraphs:      respParagraphs,
		AgreementOf:     existing.AgreementOf,
		AgreementFor:    existing.AgreementFor,
		CompanyName:     existing.CompanyName,
		Annexures:       existing.Annexures,
		CommercialTerms: existing.CommercialTerms,
		MarketplaceFee:  existing.MarketplaceFee,
		PaymentGateway:  existing.PaymentGateway,
		AdditionalNotes: existing.AdditionalNotes,
		CreatedAt:       existing.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       existing.UpdatedAt.Format(time.RFC3339),
	}, nil
}
