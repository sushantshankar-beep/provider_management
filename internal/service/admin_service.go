package service

import (
	"context"
	"errors"
	"os"
	"regexp"
	"strconv"
	"time"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)


type AdminService struct {
	repo *repository.AdminRepository
}

type CreateAdminRequest struct {
	Name          string
	Email         string
	Phone         string
	Password      string
	Role          string
	ServiceZones  []string
	AccessModules []primitive.ObjectID
	ProfileURL    string
}

type UpdateAdminRequest struct {
	Name          string
	Email         string
	Phone         string
	Role          string
	ServiceZones  []string
	AccessModules []primitive.ObjectID
	ProfileURL    string
}

func NewAdminService(repo *repository.AdminRepository) *AdminService{
	return &AdminService{repo: repo}
}

func (s *AdminService) Login(ctx context.Context, email, password string) (string, *domain.Admin, error) {
	admin, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil, errors.New("Invalid email or password")
		}
		return "", nil, err
	}

	if admin.Status == domain.StatusDeactive {
		return "", nil, errors.New("Account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", nil, errors.New("Invalid email or password")
	}

	token, err := generateToken(admin.ID, admin.Role, admin.PowerLevel)
	if err != nil {
		return "", nil, err
	}

	admin.Tokens = append(admin.Tokens, domain.Token{Token: token})
	if err := s.repo.Update(ctx, admin.ID, bson.M{"tokens": admin.Tokens}); err != nil {
		return "", nil, err
	}

	return token, admin, nil
}

func (s *AdminService) Logout(ctx context.Context, admin *domain.Admin, token string) error {
	var newTokens []domain.Token
	for _, t := range admin.Tokens {
		if t.Token != token {
			newTokens = append(newTokens, t)
		}
	}
	return s.repo.Update(ctx, admin.ID, bson.M{"tokens": newTokens})
}

func (s *AdminService) LogoutAll(ctx context.Context, admin *domain.Admin) error {
	return s.repo.Update(ctx, admin.ID, bson.M{"tokens": []domain.Token{}})
}

func (s *AdminService) CreateAdmin(ctx context.Context, req CreateAdminRequest, creatorID primitive.ObjectID) (*domain.Admin, error) {
	existing, _ := s.repo.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"email": req.Email},
			{"phone": req.Phone},
		},
	})
	if existing != nil {
		return nil, errors.New("Admin already exists with this email or phone")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return nil, err
	}

	admin := &domain.Admin{
		Name:          req.Name,
		Email:         req.Email,
		Phone:         req.Phone,
		Password:      string(hashedPassword),
		Role:          req.Role,
		ServiceZones:  req.ServiceZones,
		AccessModules: req.AccessModules,
		ProfileURL:    req.ProfileURL,
		CreatedBy:     creatorID,
		Status:        domain.StatusActive,
		Tokens:        []domain.Token{},
	}

	if err := s.repo.Create(ctx, admin); err != nil {
		return nil, err
	}

	return admin, nil
}

func (s *AdminService) GetAllAdmins(ctx context.Context, limit, offset int64, search, status, role string) ([]domain.Admin, int64, error) {
	query := bson.M{}

	if search != "" {
		orQuery := []bson.M{
			{"name": bson.M{"$regex": search, "$options": "i"}},
			{"email": bson.M{"$regex": search, "$options": "i"}},
			{"phone": bson.M{"$regex": search, "$options": "i"}},
		}
		if id, err := strconv.ParseInt(search, 10, 64); err == nil {
			orQuery = append(orQuery, bson.M{"id": id})
		}
		query["$or"] = orQuery
	}

	if status != "" {
		query["status"] = status
	}

	if role != "" {
		query["role"] = role
	}

	return s.repo.FindAll(ctx, query, limit, offset)
}

func (s *AdminService) GetAdminByID(ctx context.Context, id primitive.ObjectID) (*domain.Admin, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *AdminService) UpdateAdmin(ctx context.Context, id primitive.ObjectID, req UpdateAdminRequest) (*domain.Admin, error) {
	admin, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("Admin not found")
	}

	update := bson.M{}

	if req.Email != "" && req.Email != admin.Email {
		existing, _ := s.repo.FindByEmail(ctx, req.Email)
		if existing != nil {
			return nil, errors.New("Email already in use")
		}
		update["email"] = req.Email
	}

	if req.Phone != "" && req.Phone != admin.Phone {
		existing, _ := s.repo.FindOne(ctx, bson.M{"phone": req.Phone})
		if existing != nil {
			return nil, errors.New("Phone already in use")
		}
		update["phone"] = req.Phone
	}

	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Role != "" {
		update["role"] = req.Role
		update["powerLevel"] = domain.GetPowerLevel(req.Role)
	}
	if req.ServiceZones != nil {
		update["serviceZones"] = req.ServiceZones
	}
	if req.AccessModules != nil {
		update["accessModules"] = req.AccessModules
	}
	if req.ProfileURL != "" {
		update["profileUrl"] = req.ProfileURL
	}

	if err := s.repo.Update(ctx, id, update); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *AdminService) ToggleAdminStatus(ctx context.Context, id primitive.ObjectID) (*domain.Admin, error) {
	admin, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.New("Admin not found")
	}

	newStatus := domain.StatusActive
	if admin.Status == domain.StatusActive {
		newStatus = domain.StatusDeactive
	}

	if err := s.repo.Update(ctx, id, bson.M{"status": newStatus}); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

func (s *AdminService) DeleteAdmin(ctx context.Context, id primitive.ObjectID) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("Admin not found")
	}
	return s.repo.Delete(ctx, id)
}

func (s *AdminService) ResetPasswordBySuperAdmin(ctx context.Context, id primitive.ObjectID, newPassword string) error {
	admin, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return errors.New("Admin not found")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return err
	}

	return s.repo.Update(ctx, admin.ID, bson.M{
		"password": string(hashedPassword),
		"tokens":   []domain.Token{},
	})
}

func (s *AdminService) ChangeOwnPassword(ctx context.Context, admin *domain.Admin, currentPassword, newPassword, currentToken string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(currentPassword)); err != nil {
		return errors.New("Current password is incorrect")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 10)
	if err != nil {
		return err
	}

	var newTokens []domain.Token
	for _, t := range admin.Tokens {
		if t.Token == currentToken {
			newTokens = append(newTokens, t)
		}
	}

	return s.repo.Update(ctx, admin.ID, bson.M{
		"password": string(hashedPassword),
		"tokens":   newTokens,
	})
}

func (s *AdminService) GetDashboardStats(ctx context.Context) (map[string]int64, error) {
	totalAdmin, _ := s.repo.CountDocuments(ctx, bson.M{"role": domain.RoleAdmin})
	totalSubAdmin, _ := s.repo.CountDocuments(ctx, bson.M{"role": domain.RoleSubAdmin})
	totalActive, _ := s.repo.CountDocuments(ctx, bson.M{"status": domain.StatusActive})
	totalDeactive, _ := s.repo.CountDocuments(ctx, bson.M{"status": domain.StatusDeactive})

	return map[string]int64{
		"totalAdmin":    totalAdmin,
		"totalSubAdmin": totalSubAdmin,
		"totalActive":   totalActive,
		"totalDeactive": totalDeactive,
	}, nil
}

func (s *AdminService) GetProfile(ctx context.Context, id primitive.ObjectID) (*domain.Admin, error) {
	return s.repo.FindByID(ctx, id)
}

func generateToken(id primitive.ObjectID, role string, powerLevel int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"_id":        id.Hex(),
		"role":       role,
		"powerLevel": powerLevel,
		"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
	})

	return token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
}

func ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}