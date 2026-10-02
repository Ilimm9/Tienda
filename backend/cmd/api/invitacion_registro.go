package main

import (
	"context"
	"errors"

	cuentaapplication "tienda/backend/internal/application/cuenta"
	negocioapplication "tienda/backend/internal/application/negocio"
)

// registroInvitacion adapta el registro de `cuenta` al contrato que consume `negocio`,
// para que ningún dominio importe la implementación del otro.
type registroInvitacion struct {
	verificacion *cuentaapplication.VerificationService
}

func (r registroInvitacion) RegistrarCuentaPendiente(ctx context.Context, nombres, apellidos, correo, telefono, contrasena, ip string) (negocioapplication.DesafioRegistro, error) {
	resultado, err := r.verificacion.RegisterWithNames(ctx, nombres, apellidos, correo, telefono, contrasena, ip)
	desafio := negocioapplication.DesafioRegistro{
		DesafioID: resultado.ChallengeID, CorreoEnmascarado: resultado.MaskedEmail,
		ReenviarEnSegundos: int(resultado.ResendAfter.Seconds()),
	}
	switch {
	case err == nil:
		return desafio, nil
	case errors.Is(err, cuentaapplication.ErrVerificationTooSoon), errors.Is(err, cuentaapplication.ErrVerificationLimited):
		return negocioapplication.DesafioRegistro{}, negocioapplication.ErrInvitacionLimite
	case errors.Is(err, cuentaapplication.ErrEmailDelivery):
		return desafio, negocioapplication.ErrInvitacionEnvioCodigo
	default:
		return negocioapplication.DesafioRegistro{}, err
	}
}
