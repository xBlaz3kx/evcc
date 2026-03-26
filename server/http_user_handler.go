package server

import (
	"encoding/json"
	"net/http"

	dbuser "github.com/evcc-io/evcc/server/db/user"
	"github.com/gorilla/mux"
)

type createUserRequest struct {
	Username string      `json:"username"`
	Password string      `json:"password"`
	Role     dbuser.Role `json:"role"`
}

type updateUserRequest struct {
	Password *string      `json:"password,omitempty"`
	Role     *dbuser.Role `json:"role,omitempty"`
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := dbuser.All()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonWrite(w, users)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	u, err := getUserByID(r)
	if err != nil {
		if notFound(err) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonWrite(w, u)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	if req.Role == "" {
		if isFirstSetup() {
			req.Role = dbuser.RoleAdmin
		} else {
			req.Role = dbuser.RoleViewer
		}
	}

	u := dbuser.User{
		Username: req.Username,
		Role:     req.Role,
	}
	if err := u.SetPassword(req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := dbuser.Create(&u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	jsonWrite(w, u)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	u, err := getUserByID(r)
	if err != nil {
		if notFound(err) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if u.Role == dbuser.RoleAdmin && isLastAdmin(u.ID) {
		http.Error(w, "Cannot delete the last admin user", http.StatusBadRequest)
		return
	}

	if err := dbuser.Delete(u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	u, err := getUserByID(r)
	if err != nil {
		if notFound(err) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req updateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Password != nil {
		if err := u.SetPassword(*req.Password); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if req.Role != nil {
		if u.Role == dbuser.RoleAdmin && *req.Role != dbuser.RoleAdmin && isLastAdmin(u.ID) {
			http.Error(w, "Cannot remove admin role from the last admin user", http.StatusBadRequest)
			return
		}
		u.Role = *req.Role
	}

	if err := dbuser.Save(u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonWrite(w, u)
}

func listUserVehiclesHandler(w http.ResponseWriter, r *http.Request) {
	u, err := getUserByID(r)
	if err != nil {
		if notFound(err) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	names, err := dbuser.VehiclesForUser(u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonWrite(w, names)
}

func addUserVehicleHandler(w http.ResponseWriter, r *http.Request) {
	u, err := getUserByID(r)
	if err != nil {
		if notFound(err) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var req struct {
		VehicleName string `json:"vehicleName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.VehicleName == "" {
		http.Error(w, "vehicleName is required", http.StatusBadRequest)
		return
	}
	if err := dbuser.AddVehicle(u.ID, req.VehicleName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func removeUserVehicleHandler(w http.ResponseWriter, r *http.Request) {
	u, err := getUserByID(r)
	if err != nil {
		if notFound(err) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vehicleName := mux.Vars(r)["vehicle"]
	if vehicleName == "" {
		http.Error(w, "vehicle name is required", http.StatusBadRequest)
		return
	}
	if err := dbuser.RemoveVehicle(u.ID, vehicleName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func listUserLoadpointsHandler(w http.ResponseWriter, r *http.Request) {
	u, err := getUserByID(r)
	if err != nil {
		if notFound(err) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	names, err := dbuser.LoadpointsForUser(u.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonWrite(w, names)
}

func addUserLoadpointHandler(w http.ResponseWriter, r *http.Request) {
	u, err := getUserByID(r)
	if err != nil {
		if notFound(err) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var req struct {
		LoadpointName string `json:"loadpointName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.LoadpointName == "" {
		http.Error(w, "loadpointName is required", http.StatusBadRequest)
		return
	}
	if err := dbuser.AddLoadpoint(u.ID, req.LoadpointName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func removeUserLoadpointHandler(w http.ResponseWriter, r *http.Request) {
	u, err := getUserByID(r)
	if err != nil {
		if notFound(err) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	loadpointName := mux.Vars(r)["loadpoint"]
	if loadpointName == "" {
		http.Error(w, "loadpoint name is required", http.StatusBadRequest)
		return
	}
	if err := dbuser.RemoveLoadpoint(u.ID, loadpointName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
