package api

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/gopal-lohar/hackathon-2025/internal/api/db"
	"github.com/gopal-lohar/hackathon-2025/internal/shared/protocol"
	"github.com/gopal-lohar/hackathon-2025/internal/shared/store"
	"github.com/gopal-lohar/hackathon-2025/internal/shared/utils"
	"github.com/gopal-lohar/hackathon-2025/internal/shared/utils/logger"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type APIServer struct {
	endpointStore *store.EndpointStore
	ruleStore     *store.RuleStore
	logger        *logrus.Logger
	endpointMap   map[string]net.Conn // endpointID -> conn
}

func NewAPIServer() *APIServer {
	db, err := db.NewDB()
	if err != nil {
		logrus.Fatalf("Error creating db connection: %v", err)
	}

	return &APIServer{
		logger:        logger.NewLogger(),
		endpointStore: store.NewEndpointStore(db),
		ruleStore:     store.NewRuleStore(db),
		endpointMap:   make(map[string]net.Conn),
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (as *APIServer) handlePolicyOptions(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
func (as *APIServer) Run() {
	router := mux.NewRouter()

	router.Use(corsMiddleware)
	router.HandleFunc("/api/v1/endpoints", as.handleGetEndpoints).Methods("GET")
	router.HandleFunc("/api/v1/policy", as.handlePostPolicy).Methods("POST")
	router.HandleFunc("/api/v1/policy", as.handlePolicyOptions).Methods("OPTIONS")
	router.HandleFunc("/api/v1/policies", as.handleGetPolicies).Methods("GET")
	router.HandleFunc("/api/v1/policy", as.handleDeletePolicy).Methods("DELETE")
	as.logger.Info("Listening on port 8080")
	http.ListenAndServe(":8080", router)
}

type Endpoints struct {
	Endpoints []Endpoint `json:"endpoints"`
}
type Endpoint struct {
	ID   uint   `json:"id"`
	IP   string `json:"ip"`
	Port string `json:"port"`
}

func (as *APIServer) handleDeletePolicy(w http.ResponseWriter, r *http.Request) {
	ruleID := r.URL.Query().Get("rule_id")
	if ruleID == "" {
		http.Error(w, "rule_id is required", http.StatusBadRequest)
		return
	}
	as.logger.Infof("DELETE /api/v1/policy: rule_id=%s", ruleID)
	err := as.ruleStore.DeleteRule(ruleID)
	if err != nil {
		as.logger.Warnf("Error deleting rule: %v", err)
		utils.WriteErrorResponse(w, "Error deleting rule", http.StatusInternalServerError)
		return
	}
	utils.WriteSuccessResponse(w, "Rule successfully deleted")
}

func (as *APIServer) handleGetPolicies(w http.ResponseWriter, r *http.Request) {
	as.logger.Info("GET /api/v1/policies")
	rules, err := as.ruleStore.GetRules()
	if err != nil {
		as.logger.Warnf("Error getting rules: %v", err)
		utils.WriteErrorResponse(w, "Error getting rules", http.StatusInternalServerError)
		return
	}

	sendRules := []store.RealRule{}
	for _, rule := range rules {
		sendRules = append(sendRules, store.RealRule{
			ID:         rule.ID,
			EndpointID: rule.EndpointID,
			Program:    rule.Program,
			Protocol:   rule.Protocol,
			RemoteIP:   rule.RemoteIP,
			Action:     rule.Action,
			Enabled:    rule.Enabled,
		})
	}
	utils.WriteJSONResponse(w, sendRules)
}

func (as *APIServer) handleGetEndpoints(w http.ResponseWriter, r *http.Request) {
	endpoints, err := as.endpointStore.GetEndpoints()
	if err != nil {
		as.logger.Warnf("Error getting endpoints: %v", err)
		utils.WriteErrorResponse(w, "Error getting endpoints", http.StatusInternalServerError)
		return
	}

	var endpointList []Endpoint
	for _, endpoint := range endpoints {
		endpointList = append(endpointList, Endpoint{
			ID:   endpoint.ID,
			IP:   endpoint.IP,
			Port: endpoint.Port,
		})
	}
	response := &Endpoints{
		Endpoints: endpointList,
	}
	utils.WriteJSONResponse(w, response)
}

func (as *APIServer) handlePostPolicy(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}
	if r.Body == nil {
		http.Error(w, "Request body is required", http.StatusBadRequest)
		return
	}

	var requestData struct {
		EndpointID string `json:"endpoint_id"`
		Program    string `json:"program"`
		RemoteIP   string `json:"remote_ip"`
		Action     string `json:"action"`
		Protocol   string `json:"protocol"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestData.EndpointID == "" {
		http.Error(w, "endpoint_id is required", http.StatusBadRequest)
		return
	}
	if requestData.Program == "" {
		http.Error(w, "program is required", http.StatusBadRequest)
		return
	}
	if requestData.RemoteIP == "" {
		http.Error(w, "remote_ip is required", http.StatusBadRequest)
		return
	}
	if requestData.Action == "" {
		http.Error(w, "action is required", http.StatusBadRequest)
		return
	}
	if requestData.Protocol == "" {
		http.Error(w, "protocol is required", http.StatusBadRequest)
		return
	}
	as.logger.Infof("POST /api/v1/policy: %+v", requestData)
	conn := as.endpointMap[requestData.EndpointID]
	if conn == nil {
		http.Error(w, "endpoint not found", http.StatusNotFound)
		return
	}
	netMsg := &protocol.NetworkMessage{
		MessageType: &protocol.NetworkMessage_Policy{
			Policy: &protocol.PolicyMessage{
				AppPath:  requestData.Program,
				RemoteIp: requestData.RemoteIP,
				Action:   requestData.Action,
				Protocol: requestData.Protocol,
			},
		},
	}
	// Save the rule to the database
	temp := store.Temp{
		EndpointID: requestData.EndpointID,
		Enabled:    true,
	}
	_, err = as.ruleStore.AddRule(requestData.Program, requestData.Protocol, requestData.RemoteIP, requestData.Action, true, temp)
	if err != nil {
		as.logger.Warnf("Error adding rule to db: %v", err)
	}
	utils.SendNetMsg(conn, netMsg)
	as.logger.Infof("Sent policy message to endpoint id: %s", requestData.EndpointID)
	utils.WriteSuccessResponse(w, "Successfully added policy")
}
