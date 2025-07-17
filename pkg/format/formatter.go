package format

// Formatter defines the interface for converting key-value pairs to different output formats
type Formatter interface {
	Format(vars map[string]string) ([]byte, error)
}

// Formatters is the registry mapping format names to formatter instances
var Formatters = map[string]Formatter{
	"dotenv":     &DotenvFormatter{},
	"properties": &PropertiesFormatter{},
}