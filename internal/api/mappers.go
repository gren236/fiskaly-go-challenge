package api

import "github.com/gren236/fiskaly-go-challenge/internal/domain"

type CreateDeviceRequest struct {
	Label     *string `json:"label" validate:"omitempty"`
	Algorithm string  `json:"algorithm" validate:"required,oneof=RSA ECC"`
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

type SignTransactionRequest struct {
	Data string `json:"data" validate:"required"`
}

type SignatureResponse struct {
	Signature  string `json:"signature"`
	SignedData string `json:"signed_data"`
}

func SignatureToApi(signature domain.SignedData) SignatureResponse {
	return SignatureResponse{
		Signature:  signature.Signature,
		SignedData: signature.OriginalData,
	}
}
