package config

var cfg *Config

type Config struct {
	Broker struct {
		Port        int    `env:"APPLICATION_PORT" env-deafult:"8081" env-description:"Broker app port for gRPC"`
		StorageType string `env:"STORAGE_TYPE" env-deafult:"NOT_PERSISTED" env-description:"it must be one of (POSTGRES, CASSANDRA, NOT_PERSISTED)"`
	}

	PostgresDB struct {
		Host     string `env:"POSTGRES_HOST" env-default:"localhost" env-description:"Database host for service"`
		Port     int    `env:"POSTGRES_PORT" env-default:"5432" env-description:"Database port for service"`
		DbName   string `env:"POSTGRES_DBNAME" env-default:"broker" env-description:"Database name for service"`
		Username string `env:"POSTGRES_USERNAME" env-default:"admin" env-description:"Database username for service"`
		Password string `env:"POSTGRES_PASSWORD" env-default:"admin" env-description:"Database password for service"`
	}

	Prometheus struct {
		Port int `env:"APPLICATION_PROM_PORT" env-deafult:"9091" env-description:"Defined metrics for each RPC"`
	}

	Jaeger struct {
		ServiceName string `env:"JAEGER_SERVICE" env-deafult:"brokerService" env-description:"Jaeger service name for Golang client"`
		Host        string `env:"JAEGER_HOST" env-default:"localhost" env-description:"Jaeger host for service"`
		TraceRate   int    `env:"JAEGER_TRACE_RATE" env-default:"10" env-description:"Jaeger trace rate for broker"`
		Port        int    `env:"JAEGER_PORT3" env-deafult:"6831" env-description:"Jaeger Port for Golang"`
	}

	CassandraDB struct {
		Host          string `env:"CASSANDRA_HOSTS" env-default:"localhost" env-description:"Database host for service"`
		Port          int    `env:"CASSANDRA_PORT" env-default:"9042" env-description:"Cassandra port for service"`
		Keyspace      string `env:"CASSANDRA_KEYSPACE" env-default:"broker" env-description:"Cassandra keyspace for service"`
		Username      string `env:"CASSANDRA_USERNAME" env-default:"admin" env-description:"Cassandra username for service"`
		Password      string `env:"CASSANDRA_PASSWORD" env-default:"admin" env-description:"Cassandra password for service"`
		BatchSize     int    `env:"CASSANDRA_BATCH_SIZE" env-default:"10000" env-description:"Cassandra batch size for batch daemon"`
		TimeThreshold int    `env:"CASSANDRA_TIME" env-default:"10" env-description:"Cassandra time ticker for batch threshold"`
	}

	Graylog struct {
	}
}

func SetConfigInstance(newCfg *Config) {
	cfg = new(Config)
	cfg = newCfg
}

func GetConfigInstance() *Config {
	return cfg
}
