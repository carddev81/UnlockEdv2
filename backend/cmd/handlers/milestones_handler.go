package handlers

import (
	"Go-Prototype/backend/cmd/models"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

func (srv *Server) registerMilestonesRoutes() {
	srv.Mux.Handle("GET /api/milestones", srv.applyMiddleware(http.HandlerFunc(srv.HandleIndexMilestones)))
	srv.Mux.Handle("POST /api/milestones", srv.applyMiddleware(http.HandlerFunc(srv.HandleCreateMilestone)))
	srv.Mux.Handle("DELETE /api/milestones", srv.applyMiddleware(http.HandlerFunc(srv.HandleDeleteMilestone)))
	srv.Mux.Handle("PATCH /api/milestones/{id}", srv.applyMiddleware(http.HandlerFunc(srv.HandleUpdateMilestone)))
}

func (srv *Server) HandleIndexMilestones(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	orderBy := r.URL.Query().Get("order_by")
	page, perPage := srv.GetPaginationInfo(r)
	total, milestones, err := srv.Db.GetMilestones(page, perPage, search, orderBy)
	if err != nil {
		slog.Debug("IndexMilestones Database Error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	last := srv.CalculateLast(total, perPage)
	paginationData := models.PaginationMeta{
		PerPage:     perPage,
		LastPage:    int(last),
		CurrentPage: page,
		Total:       total,
	}
	response := models.PaginatedResource[models.Milestone]{
		Meta: paginationData,
		Data: milestones,
	}
	if err = srv.WriteResponse(w, http.StatusOK, response); err != nil {
		srv.LogError("Error writing response: " + err.Error())
		srv.ErrorResponse(w, http.StatusBadRequest, err.Error())
	}
}

func (srv *Server) HandleCreateMilestone(w http.ResponseWriter, r *http.Request) {
	miles := &models.Milestone{}
	if err := json.NewDecoder(r.Body).Decode(miles); err != nil {
		srv.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()
	if _, err := srv.Db.CreateMilestone(miles); err != nil {
		srv.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := srv.WriteResponse(w, http.StatusCreated, miles); err != nil {
		srv.LogError("Error writing response: " + err.Error())
		srv.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
}

func (srv *Server) HandleDeleteMilestone(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		srv.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := srv.Db.DeleteMilestone(id); err != nil {
		srv.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (srv *Server) HandleUpdateMilestone(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		srv.LogInfo("No ID provided in URL, checking request body json")
	}
	miles := &models.Milestone{}
	if err := json.NewDecoder(r.Body).Decode(miles); err != nil {
		srv.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	defer r.Body.Close()
	msId := 0
	if id != 0 {
		msId = int(id)
	} else if miles.ID != 0 {
		msId = int(miles.ID)
	} else {
		srv.ErrorResponse(w, http.StatusBadRequest, "No ID provided in URL or request body")
	}
	toUpdate, err := srv.Db.GetMilestoneByID(msId)
	if err != nil {
		srv.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	models.UpdateStruct(&toUpdate, &miles)
	if _, err := srv.Db.UpdateMilestone(toUpdate); err != nil {
		srv.ErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := srv.WriteResponse(w, http.StatusOK, toUpdate); err != nil {
		srv.LogError("Error writing response: " + err.Error())
		srv.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
}
