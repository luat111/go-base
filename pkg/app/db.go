package app

import "go-base/pkg/datasource/postgres"

func (a *App[EnvInterface]) DB() *postgres.DB {
	return a.container.DB
}
