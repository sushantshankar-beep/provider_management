package dto 

type VehicleBrandResponse struct {
	BrandName   string   `json:"brandName"`
	VehicleType string   `json:"vehicleType"`
	ModelName   []string `json:"modelName"`
}