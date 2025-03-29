package postgres

import "gorm.io/gorm"

var extensionArr = []string{
	`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
	// `CREATE EXTENSION IF NOT EXISTS earthdistance CASCADE;`,
}

func setupExtension(DB *gorm.DB) {
	for _, query := range extensionArr {
		DB.Exec(query)
	}
}
