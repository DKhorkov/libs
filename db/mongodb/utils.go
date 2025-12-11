package mongodb

import (
	"fmt"
)

func BuildDsn(config Config) string {
	dsn := fmt.Sprintf(
		"mongodb://%s:%d",
		config.Host,
		config.Port,
	)

	return dsn
}
