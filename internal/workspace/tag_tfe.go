package workspace

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	types "github.com/hashicorp/go-tfe"
	"github.com/tofutf/tofutf/internal/http/decode"
	"github.com/tofutf/tofutf/internal/tfeapi"
)

const (
	addTags tagOperation = iota
	removeTags
)

type tagOperation int

func (a *tfe) addTagHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

	r.HandleFunc("/workspaces/{workspace_id}/relationships/tags", a.addTags).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/relationships/tags", a.removeTags).Methods("DELETE")
	r.HandleFunc("/workspaces/{workspace_id}/relationships/tags", a.getTags).Methods("GET")

	r.HandleFunc("/organizations/{organization_name}/tags", a.listTags).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/tags", a.deleteTags).Methods("DELETE")
	r.HandleFunc("/tags/{tag_id}/relationships/workspaces", a.tagWorkspaces).Methods("POST")
}

func (a *tfe) listTags(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params ListTagsOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.ListTags(r.Context(), org, params)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert to
	to := make([]*types.OrganizationTag, len(page.Items))
	for i, from := range page.Items {
		to[i] = a.toTag(from)
	}
	a.Respond(w, r, types.OrganizationTagsList{
		Items:      to,
		Pagination: page.Pagination,
	}, http.StatusOK)
}

func (a *tfe) deleteTags(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params []struct {
		ID string `jsonapi:"primary,tags"`
	}
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	tagIDs := make([]string, len(params))
	for i, p := range params {
		tagIDs[i] = p.ID
	}

	if err := a.DeleteTags(r.Context(), org, tagIDs); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) tagWorkspaces(w http.ResponseWriter, r *http.Request) {
	tagID, err := decode.Param("tag_id", r)
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
	for i, p := range params {
		workspaceIDs[i] = p.ID
	}

	if err := a.TagWorkspaces(r.Context(), tagID, workspaceIDs); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) addTags(w http.ResponseWriter, r *http.Request) {
	a.alterWorkspaceTags(w, r, addTags)
}

func (a *tfe) removeTags(w http.ResponseWriter, r *http.Request) {
	a.alterWorkspaceTags(w, r, removeTags)
}

func (a *tfe) alterWorkspaceTags(w http.ResponseWriter, r *http.Request, op tagOperation) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params []*types.Tag
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	switch op {
	case addTags:
		err = a.AddTags(r.Context(), workspaceID, params)
	case removeTags:
		err = a.RemoveTags(r.Context(), workspaceID, params)
	default:
		err = errors.New("unknown tag operation")
	}
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) getTags(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params ListWorkspaceTagsOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.ListWorkspaceTags(r.Context(), workspaceID, params)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert to
	to := make([]*types.OrganizationTag, len(page.Items))
	for i, from := range page.Items {
		to[i] = a.toTag(from)
	}
	a.Respond(w, r, types.OrganizationTagsList{
		Items:      to,
		Pagination: page.Pagination,
	})
}

func (a *tfe) toTag(from *Tag) *types.OrganizationTag {
	return &types.OrganizationTag{
		ID:            from.ID,
		Name:          from.Name,
		InstanceCount: from.InstanceCount,
		Organization: &types.Organization{
			Name: from.Organization,
		},
	}
}

func toTagSpecs(from []*types.Tag) (to []TagSpec) {
	for _, tag := range from {
		to = append(to, TagSpec{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}
	return
}
