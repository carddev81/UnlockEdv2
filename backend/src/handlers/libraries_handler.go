package handlers

import (
	"UnlockEdv2/src/models"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
)

func (srv *Server) registerLibraryRoutes() []routeDef {
	axx := models.Feature(models.OpenContentAccess)
	return []routeDef{
		{"GET /api/libraries/{id}/search", srv.handleSearchLibraries, false, axx},
		{"GET /api/libraries", srv.handleIndexLibraries, false, axx},
		{"GET /api/libraries/{id}", srv.handleGetLibrary, false, axx},
		{"PUT /api/libraries/{id}/toggle", srv.handleToggleLibraryVisibility, true, axx},
		{"PUT /api/libraries/{id}/favorite", srv.handleToggleFavoriteLibrary, false, axx},
	}
}

func (srv *Server) handleIndexLibraries(w http.ResponseWriter, r *http.Request, log sLog) error {
	page, perPage := srv.getPaginationInfo(r)
	search := r.URL.Query().Get("search")
	orderBy := r.URL.Query().Get("order_by")
	days, err := strconv.Atoi(r.URL.Query().Get("days"))
	if err != nil {
		days = -1
	}
	showHidden := "visible"
	if !userIsAdmin(r) && r.URL.Query().Get("visibility") == "hidden" {
		return newUnauthorizedServiceError()
	} else if !userIsAdmin(r) && r.URL.Query().Get("visibility") == "featured" {
		showHidden = "featured"
	} else if userIsAdmin(r) {
		showHidden = r.URL.Query().Get("visibility")
	}
	claims := r.Context().Value(ClaimsKey).(*Claims)
	total, libraries, err := srv.Db.GetAllLibraries(page, perPage, days, claims.UserID, claims.FacilityID, showHidden, orderBy, search, claims.isAdmin())
	if err != nil {
		return newDatabaseServiceError(err)
	}
	paginationData := models.NewPaginationInfo(page, perPage, total)
	return writePaginatedResponse(w, http.StatusOK, libraries, paginationData)
}

func (srv *Server) handleGetLibrary(w http.ResponseWriter, r *http.Request, log sLog) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return newInvalidIdServiceError(err, "library id")
	}
	library, err := srv.Db.GetLibraryByID(id)
	if err != nil {
		log.add("library_id", id)
		return newDatabaseServiceError(err)
	}
	return writeJsonResponse(w, http.StatusOK, library)
}

func (srv *Server) handleSearchLibraries(w http.ResponseWriter, r *http.Request, log sLog) error {
	page, perPage := srv.getPaginationInfo(r)
	fmt.Println(page)
	id, err := strconv.Atoi(r.PathValue("id")) //library id
	if err != nil {                            //start  --//pageLength
		return newInvalidIdServiceError(err, "library id")
	}
	pattern := r.URL.Query().Get("pattern")
	library, err := srv.Db.GetLibraryByID(id)
	if err != nil {
		log.add("library_id", id)
		return newDatabaseServiceError(err)
	}
	navigate := (page-1)*perPage + 1
	// `libraries/${libraryId}/search?pattern=${searchTerm}&page=${
	//                     (page - 1) * perPage + 1
	//                 }&per_page=${perPage}`
	//build url request
	//"/search?books.name=askubuntu.com_en_all_2024-10pattern=the&format=xml
	kiwixSearchURL := fmt.Sprintf("%s/search?books.name=%s&pattern=%s&format=xml&start=%d&pageLength=%d", models.KiwixLibraryUrl, path.Base(library.Url), pattern, navigate, perPage)
	fmt.Println(kiwixSearchURL)
	// Make the HTTP GET request
	resp, err := http.Get(kiwixSearchURL)
	if err != nil {
		fmt.Printf("Error making the request: %v\n", err)
		return newInternalServerServiceError(err, "unable to make get request to kiwix")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Received status code %d\n", resp.StatusCode)
		return newInternalServerServiceError(err, "bad request going to add status code to this message!!!!")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading the response body: %v\n", err)
		return newInternalServerServiceError(err, "unable to read response body")
	}


	var rss models.RSS
	err = xml.Unmarshal(body, &rss)
	if err != nil {
		fmt.Printf("Error parsing XML: %v\n", err)
		return newInternalServerServiceError(err, "unable to parse xml")
	}

	fmt.Println(rss.Channel.ItemsPerPage, "vvvv", rss.Channel.StartIndex, "vvvv", rss.Channel.TotalResults)
	total, err := strconv.ParseInt(strings.ReplaceAll(rss.Channel.TotalResults, ",", ""), 10, 64)
	if err != nil {
		fmt.Printf("Error parsing XML: %v\n", err)
		return newInternalServerServiceError(err, "unable to make get request to kiwi")
	}
	paginationData := models.NewPaginationInfo(page, perPage, int64(total))

	channels := make([]*models.KiwixChannel, 0, 1)
	fmt.Println("made it")
	channels = append(channels, rss.IntoKiwixChannel(library))//going to change this once multiple libraries can be searched through

	return writePaginatedResponse(w, http.StatusOK, channels, paginationData)
}

func (srv *Server) handleToggleLibraryVisibility(w http.ResponseWriter, r *http.Request, log sLog) error {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return newInvalidIdServiceError(err, "library id")
	}
	library, err := srv.Db.ToggleVisibilityAndRetrieveLibrary(id)
	if err != nil {
		log.add("library_id", id)
		return newDatabaseServiceError(err)
	}
	if srv.buckets != nil { //make sure to update value in bucket if exists
		srv.updateLibraryBucket(r.PathValue("id"), library, log)
	}
	return writeJsonResponse(w, http.StatusOK, "Library visibility updated successfully")
}

func (srv *Server) updateLibraryBucket(key string, library *models.Library, log sLog) {
	var proxyParams *models.LibraryProxyPO
	libraryBucket := srv.buckets[LibraryPaths]
	entry, err := libraryBucket.Get(key)
	if err == nil {
		err = json.Unmarshal(entry.Value(), &proxyParams)
		if err != nil {
			log.warn("unable to unmarshal value from LibaryPaths bucket")
			return
		}
		proxyParams.VisibilityStatus = library.VisibilityStatus
	} else { //build a one for the bucket
		proxyParams = library.IntoProxyPO()
	}
	marshaledParams, err := json.Marshal(proxyParams)
	if err != nil {
		log.warn("unable to marshal value to put into the LibaryPaths bucket")
		return
	}
	if _, err := libraryBucket.Put(key, marshaledParams); err != nil {
		log.warnf("unable to update value within LibaryPaths bucket, error is %v", err)
	}
}

func (srv *Server) handleToggleFavoriteLibrary(w http.ResponseWriter, r *http.Request, log sLog) error {
	claims := r.Context().Value(ClaimsKey).(*Claims)
	libraryID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return newInternalServerServiceError(err, "error converting content id to int")
	}
	var facilityID *uint = nil
	if userIsAdmin(r) {
		// an admin toggling this will save the facilityID as a 'featured' library for that facility
		facilityID = &claims.FacilityID
	}
	library, err := srv.Db.GetLibraryByID(libraryID)
	if err != nil {
		return newDatabaseServiceError(err)
	}
	if _, err := srv.Db.FavoriteOpenContent(libraryID, library.OpenContentProviderID, claims.UserID, facilityID); err != nil {
		return newDatabaseServiceError(err)
	}
	return writeJsonResponse(w, http.StatusOK, "Favorite toggled successfully")
}
