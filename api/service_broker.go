package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ajg/form"
	"github.com/tsuru/tsuru/auth"
	"github.com/tsuru/tsuru/errors"
	"github.com/tsuru/tsuru/permission"
	"github.com/tsuru/tsuru/service/v2"
	"github.com/tsuru/tsuru/servicemanager"
	"github.com/tsuru/tsuru/types/service"
)

// title: service broker list
// path: /brokers
// method: GET
// produce: application/json
// responses:
//   200: List service brokers
//   204: No content
//   401: Unauthorized
func serviceBrokerList(w http.ResponseWriter, r *http.Request, t auth.Token) error {
	if !permission.Check(t, permission.PermServiceBrokerRead) {
		return permission.ErrUnauthorized
	}
	brokers, err := servicemanager.ServiceBroker.List()
	if err != nil {
		return err
	}
	if len(brokers) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"brokers": brokers,
	})
}

// title: Add service broker
// path: /brokers
// method: POST
// responses:
//   201: Service broker created
//   401: Unauthorized
//   409: Broker already exists
func serviceBrokerAdd(w http.ResponseWriter, r *http.Request, t auth.Token) error {
	if !permission.Check(t, permission.PermServiceBrokerCreate) {
		return permission.ErrUnauthorized
	}
	var broker service.Broker
	dec := form.NewDecoder(nil)
	dec.IgnoreCase(true)
	dec.IgnoreUnknownKeys(true)
	if err := r.ParseForm(); err != nil {
		return &errors.HTTP{Code: http.StatusBadRequest, Message: fmt.Sprintf("unable to parse form: %v", err)}
	}
	if err := dec.DecodeValues(&broker, r.Form); err != nil {
		return &errors.HTTP{Code: http.StatusBadRequest, Message: fmt.Sprintf("unable to parse broker: %v", err)}
	}
	if err := servicemanager.ServiceBroker.Create(broker); err != nil {
		if err == service.ErrServiceBrokerAlreadyExists {
			return &errors.HTTP{Code: http.StatusConflict, Message: "Broker already exists."}
		}
		return err
	}
	w.WriteHeader(http.StatusCreated)
	return nil
}

// title: Update service broker
// path: /brokers/{broker}
// method: PUT
// responses:
//   200: Service broker updated
//   401: Unauthorized
//	 404: Not Found
func serviceBrokerUpdate(w http.ResponseWriter, r *http.Request, t auth.Token) error {
	if !permission.Check(t, permission.PermServiceBrokerUpdate) {
		return permission.ErrUnauthorized
	}
	brokerName := r.URL.Query().Get(":broker")
	if brokerName == "" {
		return &errors.HTTP{Code: http.StatusBadRequest, Message: "Empty broker name."}
	}
	var broker service.Broker
	dec := form.NewDecoder(nil)
	dec.IgnoreCase(true)
	dec.IgnoreUnknownKeys(true)
	if err := r.ParseForm(); err != nil {
		return &errors.HTTP{Code: http.StatusBadRequest, Message: fmt.Sprintf("unable to parse form: %v", err)}
	}
	if err := dec.DecodeValues(&broker, r.Form); err != nil {
		return &errors.HTTP{Code: http.StatusBadRequest, Message: fmt.Sprintf("unable to parse broker: %v", err)}
	}
	err := servicemanager.ServiceBroker.Update(brokerName, broker)
	if err == service.ErrServiceBrokerNotFound {
		w.WriteHeader(http.StatusNotFound)
	}
	return err
}

// title: Delete service broker
// path: /brokers/{broker}
// method: DELETE
// responses:
//   200: Service broker deleted
//   401: Unauthorized
//	 404: Not Found
func serviceBrokerDelete(w http.ResponseWriter, r *http.Request, t auth.Token) error {
	if !permission.Check(t, permission.PermServiceBrokerDelete) {
		return permission.ErrUnauthorized
	}
	brokerName := r.URL.Query().Get(":broker")
	if brokerName == "" {
		return &errors.HTTP{Code: http.StatusBadRequest, Message: "Empty broker name."}
	}
	err := servicemanager.ServiceBroker.Delete(brokerName)
	if err == service.ErrServiceBrokerNotFound {
		w.WriteHeader(http.StatusNotFound)
	}
	return err
}

// title: Get service broker catalog
// path: /brokers/{broker}/v2/catalog
// method: GET
// responses:
//   200: Service broker catalog
//	 400: Invalid data
//   401: Unauthorized
//	 404: Not Found
func serviceBrokerCatalog(w http.ResponseWriter, r *http.Request, t auth.Token) error {
	brokerName := r.URL.Query().Get(":broker")
	if brokerName == "" {
		return &errors.HTTP{Code: http.StatusBadRequest, Message: "Empty broker name."}
	}
	b, err := servicemanager.ServiceBroker.Find(brokerName)
	if err != nil {
		if err == service.ErrServiceBrokerNotFound {
			w.WriteHeader(http.StatusNotFound)
		}
		return err
	}
	bClient, err := v2.NewClient(b)
	if err != nil {
		return err
	}
	catalog, err := bClient.GetCatalog()
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(catalog)
}