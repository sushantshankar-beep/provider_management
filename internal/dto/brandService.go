package dto
 
type BrandResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Logo        string `json:"logo"`
	VehicleType string `json:"vehicleType"`
	BrandType   string `json:"brandType"`
}

type ServiceResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	VehicleType string `json:"vehicleType"`
}