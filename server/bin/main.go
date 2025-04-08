package main

import (
	"context"

	gopxgrid "github.com/vkumov/go-pxgrid"
	"github.com/vkumov/go-pxgrider/server/internal"
	"github.com/vkumov/go-pxgrider/server/internal/connection"
	"github.com/vkumov/go-pxgrider/server/shared"
)

func someTest(app shared.App) {
	<-app.Ready()

	ctx := context.Background()
	app.Log().Debug().Msg("App is ready")
	users := app.Users()
	app.Log().Debug().Msg("Users handler created")
	user := users.GetUser(ctx, "test")
	app.Log().Debug().Msg("User created")
	var c *connection.Connection
	var err error

	if cns := user.GetConnections(); len(cns) > 0 {
		c = cns[0]
		app.Log().Debug().Msg("Connection found")
	} else {
		c, err = user.AddConnection(ctx, connection.ConnectionCreate{
			FriendlyName: "test",
			PrimaryNode: connection.Node{
				FQDN:        "test.com",
				ControlPort: 8910,
			},
			Credentials: connection.Credentials{
				Type:     connection.CredentialsTypePassword,
				NodeName: "test",
			},
			Description: "",
			DNS:         "",
			DNSStrategy: gopxgrid.IPv4,
			ClientName:  "test",
			InsecureTLS: true,
			CA:          []string{},
		})
		if err != nil {
			app.Log().Fatal().Err(err).Msg("Failed to create connection")
		}
		app.Log().Debug().Msg("Connection created")
	}
	app.Log().Debug().Str("id", c.ID()).Msg("Connection ID")

	wrong, err := c.GetMethodsOfService("bruh")
	app.Log().Debug().Interface("methods", wrong).AnErr("err", err).Msg("Methods of service")

	methods, err := c.GetMethodsOfService("com.cisco.ise.mdm")
	app.Log().Debug().Interface("methods", methods).AnErr("err", err).Msg("Methods of service")

	topics, err := c.GetTopicsOfService("com.cisco.ise.mdm")
	app.Log().Debug().Interface("topics", topics).AnErr("err", err).Msg("Topics of service")

	methods, err = c.GetMethodsOfService("SessionDirectory")
	app.Log().Debug().Interface("methods", methods).AnErr("err", err).Msg("Methods of service")

	all, err := c.GetAllTopics()
	app.Log().Debug().Interface("all", all).AnErr("err", err).Msg("All topics")
}

func main() {
	app := internal.NewApp()

	if !app.IsProd() {
		go someTest(app)
	}

	if err := app.Start(); err != nil {
		app.Log().Fatal().Err(err).Msg("Failed to start server")
	}
}
