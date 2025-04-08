package config

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/ettle/strcase"
	"github.com/spf13/viper"

	"github.com/vkumov/go-pxgrider/server/internal/db"
)

var (
	BuildStamp = "undefined"
	GitHash    = "undefined"
	V          = "undefined"
)

type (
	AuthSpecs struct {
		Token string `env:"AUTH_TOKEN"`
	}

	LoggerSpecs struct {
		Level string `env:"LOG_LEVEL" default:""`
	}

	KeepaliveSpecs struct {
		MaxConnectionIdle     time.Duration `env:"GRPC_MAX_CONN_IDLE" default:"30s"`
		MaxConnectionAge      time.Duration `env:"GRPC_MAX_CONN_AGE" default:"30m"`
		MaxConnectionAgeGrace time.Duration `env:"GRPC_MAX_CONN_AGE_GRACE" default:"15s"`
		Time                  time.Duration `env:"GRPC_TIME" default:"15s"`
		Timeout               time.Duration `env:"GRPC_TIMEOUT" default:"2s"`
	}

	EnforcementPolicySpecs struct {
		MinTime             time.Duration `env:"GRPC_MIN_TIME" default:"5s"`
		PermitWithoutStream bool          `env:"GRPC_PERMIT_WITHOUT_STREAM" default:"true"`
	}

	ServerSpecs struct {
		Port              int `env:"PORT" default:"50051"`
		Keepalive         KeepaliveSpecs
		EnforcementPolicy EnforcementPolicySpecs
	}

	VersionSpecs struct {
		BuildStamp string `ignored:"true"`
		GitHash    string `ignored:"true"`
		GitVersion string `ignored:"true"`
		V          string `ignored:"true"`
	}

	Specs struct {
		sync.Mutex
		Env     string `env:"ENV" default:"dev"`
		Auth    AuthSpecs
		DB      db.DBSpecs
		Log     LoggerSpecs
		Server  ServerSpecs
		Version VersionSpecs `ignored:"true"`
	}
)

func traverseStruct(prefix string, rt reflect.Type, rv reflect.Value) {
	for i := 0; i < rt.NumField(); i++ {
		rf := rt.Field(i)
		ymlName := prefixed(prefix, strcase.ToKebab(rf.Name))
		if rf.Type.Kind() == reflect.Struct {
			traverseStruct(ymlName, rf.Type, rv.FieldByName(rf.Name))
			continue
		}

		def := rf.Tag.Get("default")
		if def != "" {
			viper.SetDefault(ymlName, def)
		}

		if env := rf.Tag.Get("env"); env != "" {
			viper.BindEnv(ymlName, env)
		}

		lower := strings.ReplaceAll(ymlName, "-", "")
		if lower != ymlName {
			viper.RegisterAlias(lower, ymlName)
		}
	}
}

func prefixed(prefix, name string) string {
	if prefix == "" {
		return name
	}

	return prefix + "." + name
}

func (s *Specs) prepareViper() {
	traverseStruct("", reflect.TypeOf(*s), reflect.ValueOf(s).Elem())
}

func (s *Specs) loadFromViper() error {
	s.Lock()
	defer s.Unlock()

	return viper.Unmarshal(s)
}

func (s *Specs) setSpec(key string, value any) (err error) {
	s.Lock()
	defer s.Unlock()

	rv := reflect.ValueOf(s)
	parts := strings.Split(key, ".")
	for i, part := range parts {
		if rv.Kind() != reflect.Ptr {
			return fmt.Errorf("object must be a pointer")
		}
		rv = rv.Elem()
		if rv.Kind() != reflect.Struct {
			err = fmt.Errorf("field not struct: %s, left: %s", strings.Join(parts[:i], "."), strings.Join(parts[i:], "."))
			return
		}

		fieldName := strcase.ToCamel(part)
		rv = rv.FieldByName(fieldName)
		if !rv.IsValid() {
			err = fmt.Errorf("field not found: %s, left: %s", strings.Join(parts[:i], "."), strings.Join(parts[i:], "."))
			return
		}
		rv = rv.Addr()
		if i == len(parts)-1 {
			break
		}
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unable to update value: %v", r)
		}
	}()
	rv.Elem().Set(reflect.ValueOf(value))
	viper.Set(key, value)

	return
}

func (s *Specs) getSpec(key string) any {
	return viper.Get(key)
}
