package plataforma

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrCuentaNoEncontrada = errors.New("cuenta no encontrada")
	ErrCuentaNoDisponible = errors.New("la cuenta debe estar activa y verificada")
	ErrCuentaConNegocios  = errors.New("la cuenta tiene membresías de negocio y no puede ser administradora de plataforma")
	ErrMotivoRequerido    = errors.New("el motivo es obligatorio")
)

type AdminAsignado struct {
	UsuarioID uuid.UUID
	Correo    string
	Asignado  bool
}

type AssignmentRepository interface {
	AsignarAdminLectura(ctx context.Context, correo, motivo string) (AdminAsignado, error)
	RetirarAdminLectura(ctx context.Context, correo, motivo string) (AdminAsignado, error)
	ListarAdmins(ctx context.Context) ([]AdminAsignado, error)
}

type AssignmentService struct{ repository AssignmentRepository }

func NewAssignmentService(repository AssignmentRepository) *AssignmentService {
	return &AssignmentService{repository: repository}
}

func (s *AssignmentService) Asignar(ctx context.Context, correo, motivo string) (AdminAsignado, error) {
	correo = strings.ToLower(strings.TrimSpace(correo))
	motivo = strings.TrimSpace(motivo)
	if motivo == "" {
		return AdminAsignado{}, ErrMotivoRequerido
	}
	return s.repository.AsignarAdminLectura(ctx, correo, motivo)
}

func (s *AssignmentService) Retirar(ctx context.Context, correo, motivo string) (AdminAsignado, error) {
	correo = strings.ToLower(strings.TrimSpace(correo))
	motivo = strings.TrimSpace(motivo)
	if motivo == "" {
		return AdminAsignado{}, ErrMotivoRequerido
	}
	return s.repository.RetirarAdminLectura(ctx, correo, motivo)
}

func (s *AssignmentService) Listar(ctx context.Context) ([]AdminAsignado, error) {
	return s.repository.ListarAdmins(ctx)
}
