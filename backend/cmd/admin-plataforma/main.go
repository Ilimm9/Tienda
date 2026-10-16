package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	application "tienda/backend/internal/application/plataforma"
	"tienda/backend/internal/config"
	"tienda/backend/internal/database"
	infrastructure "tienda/backend/internal/infrastructure/plataforma"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Init(db); err != nil {
		log.Fatal(err)
	}
	service := application.NewAssignmentService(infrastructure.NewRepository(db))

	switch os.Args[1] {
	case "grant":
		runMutation(service, true, os.Args[2:])
	case "revoke":
		runMutation(service, false, os.Args[2:])
	case "list":
		items, err := service.Listar(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		for _, item := range items {
			fmt.Printf("%s\t%s\n", item.UsuarioID, item.Correo)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func runMutation(service *application.AssignmentService, grant bool, args []string) {
	name := "revoke"
	if grant {
		name = "grant"
	}
	flags := flag.NewFlagSet(name, flag.ExitOnError)
	email := flags.String("email", "", "correo de la cuenta")
	reason := flags.String("reason", "", "motivo auditable")
	_ = flags.Parse(args)
	if *email == "" || *reason == "" {
		flags.Usage()
		os.Exit(2)
	}
	var result application.AdminAsignado
	var err error
	if grant {
		result, err = service.Asignar(context.Background(), *email, *reason)
	} else {
		result, err = service.Retirar(context.Background(), *email, *reason)
	}
	if err != nil {
		log.Fatal(err)
	}
	action := "sin cambios"
	if result.Asignado {
		if grant {
			action = "rol asignado"
		} else {
			action = "rol retirado"
		}
	}
	fmt.Printf("%s: %s (%s)\n", action, result.Correo, result.UsuarioID)
}

func usage() {
	fmt.Fprintln(os.Stderr, "uso: admin-plataforma grant --email correo --reason motivo | revoke --email correo --reason motivo | list")
}
