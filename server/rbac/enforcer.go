package rbac

import (
	"sync"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/evcc-io/evcc/server/db"
)

var (
	once     sync.Once
	enforcer *casbin.Enforcer
)

// Init initializes the casbin enforcer using the embedded model and a gorm adapter
// backed by the application's database instance.
func Init() error {
	var initErr error
	once.Do(func() {
		adapter, err := gormadapter.NewAdapterByDB(db.Instance)
		if err != nil {
			initErr = err
			return
		}

		m, err := model.NewModelFromString(modelConf)
		if err != nil {
			initErr = err
			return
		}

		e, err := casbin.NewEnforcer(m, adapter)
		if err != nil {
			initErr = err
			return
		}

		if err := seedDefaultPolicies(e); err != nil {
			initErr = err
			return
		}

		enforcer = e
	})
	return initErr
}

// Enforcer returns the initialized casbin enforcer
func Enforcer() *casbin.Enforcer {
	return enforcer
}
