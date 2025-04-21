package model

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
}

type PVZ struct {
	ID               string    `json:"id"`
	City             string    `json:"city"`
	RegistrationDate time.Time `json:"registrationDate,omitempty"`
}

type Reception struct {
	ID       string    `json:"id"`
	DateTime time.Time `json:"dateTime,omitempty"`
	PVZID    string    `json:"pvzId"`
	Status   string    `json:"status"`
}

type Product struct {
	ID          string    `json:"id"`
	DateTime    time.Time `json:"dateTime,omitempty"`
	Type        string    `json:"type"`
	ReceptionID string    `json:"receptionId"`
}
