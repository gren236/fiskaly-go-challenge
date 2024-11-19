package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gren236/fiskaly-go-challenge/internal/domain"
	"net/http"
)

func (s *Server) CreateDevice(response http.ResponseWriter, request *http.Request) {
	// Parse request
	var req CreateDeviceRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		WriteErrorResponse(response, http.StatusBadRequest, []string{err.Error()})

		return
	}

	// Validate request
	err := s.validate.Struct(req)
	if err != nil { // Probably the error handling here could be a bit more sophisticated, but for the sake of test challenge it's enough IMO
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s: %s", err.Field(), err.Tag()))
		}

		WriteErrorResponse(response, http.StatusBadRequest, errors)

		return
	}

	device, err := s.deviceService.CreateDevice(request.Context(), req.Label, domain.Algorithm(req.Algorithm))
	if err != nil {
		WriteErrorResponse(response, http.StatusInternalServerError, []string{err.Error()})

		return
	}

	WriteAPIResponse(response, http.StatusCreated, DeviceToApi(device))
}

func (s *Server) GetDevices(response http.ResponseWriter, request *http.Request) {
	devices, err := s.deviceService.GetDevices(request.Context())
	if err != nil {
		WriteErrorResponse(response, http.StatusInternalServerError, []string{err.Error()})

		return
	}

	var res []DeviceResponse
	for _, device := range devices {
		res = append(res, DeviceToApi(device))
	}

	WriteAPIResponse(response, http.StatusOK, res)
}

func (s *Server) GetDevice(response http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	if id == "" {
		WriteErrorResponse(response, http.StatusBadRequest, []string{"missing id parameter"})

		return
	}

	device, err := s.deviceService.GetDevice(request.Context(), uuid.MustParse(id))
	if err != nil {
		WriteErrorResponse(response, http.StatusInternalServerError, []string{err.Error()})

		return
	}

	WriteAPIResponse(response, http.StatusOK, DeviceToApi(device))
}

func (s *Server) SignTransaction(response http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	if id == "" {
		WriteErrorResponse(response, http.StatusBadRequest, []string{"missing id parameter"})

		return
	}

	// Parse request
	var req SignTransactionRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		WriteErrorResponse(response, http.StatusBadRequest, []string{err.Error()})

		return
	}

	// Validate request
	err := s.validate.Struct(req)
	if err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s: %s", err.Field(), err.Tag()))
		}

		WriteErrorResponse(response, http.StatusBadRequest, errors)

		return
	}

	signature, err := s.signatureService.SignTransaction(request.Context(), uuid.MustParse(id), req.Data)
	if err != nil {
		WriteErrorResponse(response, http.StatusInternalServerError, []string{err.Error()})

		return
	}

	WriteAPIResponse(response, http.StatusCreated, SignatureToApi(signature))
}

func (s *Server) GetSignatures(response http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	if id == "" {
		WriteErrorResponse(response, http.StatusBadRequest, []string{"missing id parameter"})

		return
	}

	signatures, err := s.signatureService.GetSignatures(request.Context(), uuid.MustParse(id))
	if err != nil {
		WriteErrorResponse(response, http.StatusInternalServerError, []string{err.Error()})

		return
	}

	var res []SignatureResponse
	for _, signature := range signatures {
		res = append(res, SignatureToApi(signature))
	}

	WriteAPIResponse(response, http.StatusOK, res)
}
