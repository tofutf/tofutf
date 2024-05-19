package variable

import (
	"errors"
	"net/http"

	types "github.com/hashicorp/go-tfe"
	"github.com/tofutf/tofutf/internal"
	"github.com/tofutf/tofutf/internal/tfeapi"

	"github.com/gorilla/mux"
	"github.com/tofutf/tofutf/internal/http/decode"
)

type tfe struct {
	*tfeapi.Responder
	*Service
}

// Implements TFC workspace variables and variable set APIs:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspace-variables
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/variable-sets
func (a *tfe) addHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

	r.HandleFunc("/workspaces/{workspace_id}/vars", a.createWorkspaceVariable).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/vars", a.list).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", a.get).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", a.update).Methods("PATCH")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", a.delete).Methods("DELETE")

	r.HandleFunc("/organizations/{organization_name}/varsets", a.createVariableSet).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/varsets", a.listVariableSets).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/varsets", a.listWorkspaceVariableSets).Methods("GET")
	r.HandleFunc("/varsets/{varset_id}", a.getVariableSet).Methods("GET")
	r.HandleFunc("/varsets/{varset_id}", a.updateVariableSet).Methods("PATCH")
	r.HandleFunc("/varsets/{varset_id}", a.deleteVariableSet).Methods("DELETE")

	r.HandleFunc("/varsets/{varset_id}/relationships/vars", a.listVariableSetVariables).Methods("GET")
	r.HandleFunc("/varsets/{varset_id}/relationships/vars", a.addVariableToSet).Methods("POST")
	r.HandleFunc("/varsets/{varset_id}/relationships/vars/{variable_id}", a.getVariableSetVariable).Methods("GET")
	r.HandleFunc("/varsets/{varset_id}/relationships/vars/{variable_id}", a.updateVariableSetVariable).Methods("PATCH")
	r.HandleFunc("/varsets/{varset_id}/relationships/vars/{variable_id}", a.deleteVariableFromSet).Methods("DELETE")
	r.HandleFunc("/varsets/{varset_id}/relationships/workspaces", a.applySetToWorkspaces).Methods("POST")
	r.HandleFunc("/varsets/{varset_id}/relationships/workspaces", a.deleteSetFromWorkspaces).Methods("DELETE")
}

func (a *tfe) createWorkspaceVariable(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts types.VariableCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	v, err := a.Service.CreateWorkspaceVariable(r.Context(), workspaceID, CreateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		variableError(w, err)
		return
	}
	a.Respond(w, r, a.convertWorkspaceVariable(v, true, workspaceID), http.StatusCreated)
}

func (a *tfe) get(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	wv, err := a.GetWorkspaceVariable(r.Context(), variableID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertWorkspaceVariable(wv.Variable, true, wv.WorkspaceID), http.StatusOK)
}

func (a *tfe) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	variables, err := a.ListWorkspaceVariables(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to := make([]*types.Variable, len(variables))
	for i, from := range variables {
		to[i] = a.convertWorkspaceVariable(from, true, workspaceID)
	}

	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) update(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts types.VariableUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		variableError(w, err)
		return
	}
	updated, err := a.UpdateWorkspaceVariable(r.Context(), variableID, UpdateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertWorkspaceVariable(updated.Variable, true, updated.WorkspaceID), http.StatusOK)
}

func (a *tfe) delete(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	_, err = a.DeleteWorkspaceVariable(r.Context(), variableID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
}

func (a *tfe) createVariableSet(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params types.VariableSetCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	opts := CreateVariableSetOptions{}
	if params.Name != nil {
		opts.Name = *params.Name
	}
	if params.Description != nil {
		opts.Description = *params.Description
	}
	if params.Global != nil {
		opts.Global = *params.Global
	}
	set, err := a.Service.createVariableSet(r.Context(), org, opts)

	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.convertVariableSet(set), http.StatusCreated)
}

func (a *tfe) updateVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params types.VariableSetUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	set, err := a.Service.updateVariableSet(r.Context(), setID, UpdateVariableSetOptions{
		Name:        params.Name,
		Description: params.Description,
		Global:      params.Global,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.convertVariableSet(set), http.StatusOK)
}

func (a *tfe) listVariableSets(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	sets, err := a.Service.listVariableSets(r.Context(), org)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to := make([]*types.VariableSet, len(sets))
	for i, from := range sets {
		to[i] = a.convertVariableSet(from)
	}

	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) listWorkspaceVariableSets(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	sets, err := a.Service.listWorkspaceVariableSets(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to := make([]*types.VariableSet, len(sets))
	for i, from := range sets {
		to[i] = a.convertVariableSet(from)
	}

	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) getVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	set, err := a.Service.getVariableSet(r.Context(), setID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertVariableSet(set), http.StatusOK)
}

func (a *tfe) deleteVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if _, err := a.Service.deleteVariableSet(r.Context(), setID); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) listVariableSetVariables(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	set, err := a.Service.getVariableSet(r.Context(), setID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to := make([]*types.VariableSetVariable, len(set.Variables))
	for i, from := range set.Variables {
		to[i] = a.convertVariableSetVariable(from, true, setID)
	}

	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) addVariableToSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts types.VariableCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	v, err := a.Service.createVariableSetVariable(r.Context(), setID, CreateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		variableError(w, err)
		return
	}

	a.Respond(w, r, a.convertVariableSetVariable(v, true, setID), http.StatusOK)
}

func (a *tfe) updateVariableSetVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts types.VariableUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	set, err := a.Service.updateVariableSetVariable(r.Context(), variableID, UpdateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		variableError(w, err)
		return
	}

	v := set.getVariable(variableID)
	a.Respond(w, r, a.convertVariableSetVariable(v, true, set.ID), http.StatusOK)
}

func (a *tfe) getVariableSetVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	set, err := a.Service.getVariableSetByVariableID(r.Context(), variableID)
	if err != nil {
		variableError(w, err)
		return
	}

	v := set.getVariable(variableID)
	a.Respond(w, r, a.convertVariableSetVariable(v, true, set.ID), http.StatusOK)
}

func (a *tfe) deleteVariableFromSet(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	_, err = a.Service.deleteVariableSetVariable(r.Context(), variableID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
}

func (a *tfe) applySetToWorkspaces(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params []*types.Workspace
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	workspaceIDs := make([]string, len(params))
	for i, ws := range params {
		workspaceIDs[i] = ws.ID
	}

	err = a.Service.applySetToWorkspaces(r.Context(), setID, workspaceIDs)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) deleteSetFromWorkspaces(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params []*types.Workspace
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	workspaceIDs := make([]string, len(params))
	for i, ws := range params {
		workspaceIDs[i] = ws.ID
	}

	err = a.Service.deleteSetFromWorkspaces(r.Context(), setID, workspaceIDs)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) convertWorkspaceVariable(from *Variable, scrubSensitiveValue bool, workspaceID string) *types.Variable {

	to := &types.Variable{
		ID:          from.ID,
		Key:         from.Key,
		Value:       from.Value,
		Description: from.Description,
		Category:    types.CategoryType(from.Category),
		Sensitive:   from.Sensitive,
		HCL:         from.HCL,
		VersionID:   from.VersionID,
		Workspace: &types.Workspace{
			ID: workspaceID,
		},
	}
	if from.Sensitive && scrubSensitiveValue {
		to.Value = "" // scrub sensitive values
	}
	return to

}

func (a *tfe) convertVariableSet(from *VariableSet) *types.VariableSet {
	to := &types.VariableSet{
		ID:          from.ID,
		Name:        from.Name,
		Description: from.Description,
		Global:      from.Global,
		Organization: &types.Organization{
			Name: from.Organization,
		},
	}
	to.Variables = make([]*types.VariableSetVariable, len(from.Variables))
	for i, v := range from.Variables {
		to.Variables[i] = a.convertVariableSetVariable(v, true, v.ID)
	}
	to.Workspaces = make([]*types.Workspace, len(from.Workspaces))
	for i, workspaceID := range from.Workspaces {
		to.Workspaces[i] = &types.Workspace{
			ID: workspaceID,
		}
	}
	return to
}

func (a *tfe) convertVariableSetVariable(from *Variable, scrubSensitiveValue bool, setID string) *types.VariableSetVariable {
	to := &types.VariableSetVariable{
		ID:          from.ID,
		Key:         from.Key,
		Value:       from.Value,
		Description: from.Description,
		Category:    types.CategoryType(from.Category),
		Sensitive:   from.Sensitive,
		HCL:         from.HCL,
		VersionID:   from.VersionID,
		VariableSet: &types.VariableSet{ID: setID},
	}
	if from.Sensitive && scrubSensitiveValue {
		to.Value = "" // scrub sensitive values
	}
	return to
}

func (a *tfe) convertVariable(from *Variable, scrubSensitiveValue bool) *types.Variable {
	to := &types.Variable{
		ID:          from.ID,
		Key:         from.Key,
		Value:       from.Value,
		Description: from.Description,
		Category:    types.CategoryType(from.Category),
		Sensitive:   from.Sensitive,
		HCL:         from.HCL,
		VersionID:   from.VersionID,
	}
	if to.Sensitive && scrubSensitiveValue {
		to.Value = "" // scrub sensitive values
	}
	return to
}

func variableError(w http.ResponseWriter, err error) {
	var isUnprocessableError bool
	if errors.Is(err, ErrVariableDescriptionMaxExceeded) {
		isUnprocessableError = true
	}
	if errors.Is(err, ErrVariableKeyMaxExceeded) {
		isUnprocessableError = true
	}
	if errors.Is(err, ErrVariableValueMaxExceeded) {
		isUnprocessableError = true
	}
	if isUnprocessableError {
		tfeapi.Error(w, &internal.HTTPError{
			Message: err.Error(),
			Code:    http.StatusUnprocessableEntity,
		})
	} else {
		tfeapi.Error(w, err)
	}
}
