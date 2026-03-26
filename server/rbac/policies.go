package rbac

import "github.com/casbin/casbin/v3"

// seedDefaultPolicies seeds the default RBAC policies if the policy store is empty.
// This is idempotent: if policies already exist they are not overwritten, allowing
// administrators to customize policies after initial setup.
func seedDefaultPolicies(e *casbin.Enforcer) error {
	policies, err := e.GetPolicy()
	if err != nil {
		return err
	}
	roles, err := e.GetGroupingPolicy()
	if err != nil {
		return err
	}
	if len(policies) > 0 || len(roles) > 0 {
		return nil // already seeded
	}

	// Role hierarchy: admin → maintainer → user; viewer is standalone
	groupings := [][]string{
		{"maintainer", "user"},
		{"admin", "maintainer"},
	}
	if _, err := e.AddGroupingPolicies(groupings); err != nil {
		return err
	}

	// viewer: read-only access to state and sessions
	viewerPolicies := [][]string{
		{"viewer", "/api/state", "GET"},
		{"viewer", "/api/sessions", "GET"},
		{"viewer", "/api/gridsessions", "GET"},
		{"viewer", "/api/tariff/*", "GET"},
		{"viewer", "/api/loadpoints/*/plan", "GET"},
		{"viewer", "/api/loadpoints/*/plan/static/preview/*/*", "GET"},
		{"viewer", "/ws", "GET"},
	}

	// user: charging parameter adjustments for current session
	userPolicies := [][]string{
		{"user", "/api/state", "GET"},
		{"user", "/api/sessions", "GET"},
		{"user", "/api/gridsessions", "GET"},
		{"user", "/api/tariff/*", "GET"},
		{"user", "/api/loadpoints/*/plan", "GET"},
		{"user", "/api/loadpoints/*/plan/static/preview/*/*", "GET"},
		{"user", "/ws", "GET"},
		{"user", "/api/loadpoints/*/mode/*", "POST"},
		{"user", "/api/loadpoints/*/limitsoc/*", "POST"},
		{"user", "/api/loadpoints/*/limitenergy/*", "POST"},
		{"user", "/api/loadpoints/*/plan/energy/*", "POST"},
		{"user", "/api/loadpoints/*/plan/energy", "DELETE"},
		{"user", "/api/loadpoints/*/plan/strategy", "POST"},
		{"user", "/api/loadpoints/*/vehicle/*", "POST"},
		{"user", "/api/loadpoints/*/vehicle", "DELETE"},
		{"user", "/api/loadpoints/*/vehicle", "PATCH"},
		{"user", "/api/vehicles/*/minsoc/*", "POST"},
		{"user", "/api/vehicles/*/limitsoc/*", "POST"},
		{"user", "/api/vehicles/*/plan/soc/*", "POST"},
		{"user", "/api/vehicles/*/plan/soc", "DELETE"},
		{"user", "/api/vehicles/*/plan/repeating", "POST"},
		{"user", "/api/vehicles/*/plan/strategy", "POST"},
	}

	// maintainer: config changes and session management (inherits user)
	maintainerPolicies := [][]string{
		{"maintainer", "/api/session/*", "PUT"},
		{"maintainer", "/api/session/*", "DELETE"},
		{"maintainer", "/api/config/*", "*"},
		{"maintainer", "/api/system/log", "GET"},
		{"maintainer", "/api/system/log/areas", "GET"},
		{"maintainer", "/api/loadpoints/*/mincurrent/*", "POST"},
		{"maintainer", "/api/loadpoints/*/maxcurrent/*", "POST"},
		{"maintainer", "/api/loadpoints/*/phases/*", "POST"},
		{"maintainer", "/api/loadpoints/*/priority/*", "POST"},
		{"maintainer", "/api/loadpoints/*/enable/*", "POST"},
		{"maintainer", "/api/loadpoints/*/disable/*", "POST"},
		{"maintainer", "/api/loadpoints/*/smartcostlimit/*", "POST"},
		{"maintainer", "/api/loadpoints/*/smartcostlimit", "DELETE"},
		{"maintainer", "/api/loadpoints/*/smartfeedinprioritylimit/*", "POST"},
		{"maintainer", "/api/loadpoints/*/smartfeedinprioritylimit", "DELETE"},
		{"maintainer", "/api/loadpoints/*/batteryboost/*", "POST"},
		{"maintainer", "/api/loadpoints/*/batteryboostlimit/*", "POST"},
		{"maintainer", "/api/buffersoc/*", "POST"},
		{"maintainer", "/api/bufferstartsoc/*", "POST"},
		{"maintainer", "/api/batterydischargecontrol/*", "POST"},
		{"maintainer", "/api/batterygridchargelimit/*", "POST"},
		{"maintainer", "/api/batterygridchargelimit", "DELETE"},
		{"maintainer", "/api/batterymode/*", "POST"},
		{"maintainer", "/api/batterymode", "DELETE"},
		{"maintainer", "/api/prioritysoc/*", "POST"},
		{"maintainer", "/api/residualpower/*", "POST"},
		{"maintainer", "/api/smartcostlimit/*", "POST"},
		{"maintainer", "/api/smartcostlimit", "DELETE"},
		{"maintainer", "/api/smartfeedinprioritylimit/*", "POST"},
		{"maintainer", "/api/smartfeedinprioritylimit", "DELETE"},
		{"maintainer", "/api/settings/telemetry/*", "POST"},
	}

	// admin: user management and destructive ops (inherits maintainer)
	adminPolicies := [][]string{
		{"admin", "/api/users", "*"},
		{"admin", "/api/users/*", "*"},
		{"admin", "/api/system/cache", "DELETE"},
		{"admin", "/api/system/backup", "POST"},
		{"admin", "/api/system/restore", "POST"},
		{"admin", "/api/system/reset", "POST"},
		{"admin", "/api/system/shutdown", "POST"},
		{"admin", "/api/auth/password", "PUT"},
	}

	all := append(viewerPolicies, userPolicies...)
	all = append(all, maintainerPolicies...)
	all = append(all, adminPolicies...)

	if _, err := e.AddPolicies(all); err != nil {
		return err
	}

	return e.SavePolicy()
}
