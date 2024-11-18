package api

import "github.com/gren236/fiskaly-go-challenge/internal/domain"

type CreateDeviceRequest struct {
	Label     *string `json:"label"`
	Algorithm string  `json:"algorithm"`
}

type DeviceResponse struct {
	ID        string  `json:"id"`
	Label     *string `json:"label"`
	Algorithm string  `json:"algorithm"`
}

func DeviceToApi(device domain.Device) DeviceResponse {
	return DeviceResponse{
		ID:        device.ID.String(),
		Label:     device.Label,
		Algorithm: device.Algorithm.String(),
	}
}
