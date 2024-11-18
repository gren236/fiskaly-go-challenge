package api

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gren236/fiskaly-go-challenge/internal/domain"
	"net/http"
)

func (s *Server) CreateDevice(response http.ResponseWriter, request *http.Request) {
	var req CreateDeviceRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		WriteErrorResponse(response, http.StatusBadRequest, []string{err.Error()})

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
