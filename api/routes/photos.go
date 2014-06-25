package routes

import (
	"github.com/danjac/photoshare/api/models"
	"github.com/danjac/photoshare/api/session"
	"github.com/danjac/photoshare/api/storage"
	"github.com/danjac/photoshare/api/validation"
	"github.com/zenazn/goji/web"
	"net/http"
	"strconv"
	"strings"
)

var allowedContentTypes = []string{"image/png", "image/jpeg"}

func isAllowedContentType(contentType string) bool {
	for _, value := range allowedContentTypes {
		if contentType == value {
			return true
		}
	}

	return false
}

func deletePhoto(c web.C, w http.ResponseWriter, r *http.Request) {

	photo, err := photoMgr.Get(c.URLParams["id"])
	if err != nil {
		panic(err)
	}
	if photo == nil {
		http.NotFound(w, r)
		return
	}
	user, err := session.GetCurrentUser(c, r)

	if err != nil {
		panic(err)
	}

	if !user.IsAuthenticated {
		writeError(w, http.StatusUnauthorized)
		return
	}

	if !photo.CanDelete(user) {
		writeError(w, http.StatusForbidden)
		return
	}
	if err := photoMgr.Delete(photo); err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
}

func photoDetail(c web.C, w http.ResponseWriter, r *http.Request) {

	user, err := session.GetCurrentUser(c, r)
	if err != nil {
		panic(err)
	}
	photo, err := photoMgr.GetDetail(c.URLParams["id"], user)
	if err != nil {
		panic(err)
	}
	if photo == nil {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, photo, http.StatusOK)
}

func getPhotoToEdit(c web.C, w http.ResponseWriter, r *http.Request) (*models.Photo, bool) {
	photo, err := photoMgr.Get(c.URLParams["id"])
	if err != nil {
		panic(err)
	}

	if photo == nil {
		http.NotFound(w, r)
		return nil, false
	}

	user, err := session.GetCurrentUser(c, r)
	if err != nil {
		panic(err)
	}

	if !user.IsAuthenticated {
		writeError(w, http.StatusUnauthorized)
		return photo, false
	}

	if !photo.CanEdit(user) {
		writeError(w, http.StatusForbidden)
		return photo, false
	}
	return photo, true
}

func editPhotoTitle(c web.C, w http.ResponseWriter, r *http.Request) {

	photo, ok := getPhotoToEdit(c, w, r)

	if !ok {
		return
	}

	s := &struct {
		Title string `json:"title"`
	}{}

	if err := parseJSON(r, s); err != nil {
		panic(err)
	}

	photo.Title = s.Title

	validator := validation.NewPhotoValidator(photo)

	if result, err := validator.Validate(); err != nil || !result.OK {
		if err != nil {
			panic(err)
		}
		writeJSON(w, result, http.StatusBadRequest)
		return
	}

	if err := photoMgr.Update(photo); err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
}

func editPhotoTags(c web.C, w http.ResponseWriter, r *http.Request) {

	photo, ok := getPhotoToEdit(c, w, r)

	if !ok {
		return
	}

	s := &struct {
		Tags []string `json:"tags"`
	}{}

	if err := parseJSON(r, s); err != nil {
		panic(err)
	}

	photo.Tags = s.Tags

	if err := photoMgr.UpdateTags(photo); err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
}

func upload(c web.C, w http.ResponseWriter, r *http.Request) {

	user, err := session.GetCurrentUser(c, r)
	if err != nil {
		panic(err)
	}
	if !user.IsAuthenticated {
		writeError(w, http.StatusUnauthorized)
		return
	}
	title := r.FormValue("title")
	taglist := r.FormValue("taglist")
	tags := strings.Split(taglist, " ")

	src, hdr, err := r.FormFile("photo")
	if err != nil {
		if err == http.ErrMissingFile || err == http.ErrNotMultipart {
			writeString(w, "No image was posted", http.StatusBadRequest)
			return
		}
		panic(err)
	}
	contentType := hdr.Header["Content-Type"][0]

	if !isAllowedContentType(contentType) {
		writeString(w, "No image was posted", http.StatusBadRequest)
		return
	}

	defer src.Close()

	processor := storage.NewImageProcessor()
	filename, err := processor.Process(src, contentType)

	if err != nil {
		panic(err)
	}

	photo := &models.Photo{Title: title,
		OwnerID: user.ID, Filename: filename, Tags: tags}

	validator := validation.NewPhotoValidator(photo)

	if result, err := validator.Validate(); err != nil || !result.OK {
		if err != nil {
			panic(err)
		}
		writeJSON(w, result, http.StatusBadRequest)
	}

	if err := photoMgr.Insert(photo); err != nil {
		panic(err)
	}

	writeJSON(w, photo, http.StatusOK)
}

func searchPhotos(c web.C, w http.ResponseWriter, r *http.Request) {
	photos, err := photoMgr.Search(getPage(r), r.FormValue("q"))
	if err != nil {
		panic(err)
	}
	writeJSON(w, photos, http.StatusOK)
}

func photosByOwnerID(c web.C, w http.ResponseWriter, r *http.Request) {
	photos, err := photoMgr.ByOwnerID(getPage(r), c.URLParams["ownerID"])
	if err != nil {
		panic(err)
	}
	writeJSON(w, photos, http.StatusOK)
}

func getPhotos(c web.C, w http.ResponseWriter, r *http.Request) {
	photos, err := photoMgr.All(getPage(r), r.FormValue("orderBy"))
	if err != nil {
		panic(err)
	}
	writeJSON(w, photos, http.StatusOK)
}

func getTags(c web.C, w http.ResponseWriter, r *http.Request) {
	tags, err := photoMgr.GetTagCounts()
	if err != nil {
		panic(err)
	}
	writeJSON(w, tags, http.StatusOK)
}

func voteDown(c web.C, w http.ResponseWriter, r *http.Request) {
	vote(c, w, r, func(photo *models.Photo) { photo.DownVotes += 1 })
}

func voteUp(c web.C, w http.ResponseWriter, r *http.Request) {
	vote(c, w, r, func(photo *models.Photo) { photo.UpVotes += 1 })
}

func vote(c web.C, w http.ResponseWriter, r *http.Request, fn func(photo *models.Photo)) {
	var (
		photo *models.Photo
		err   error
	)

	user, err := session.GetCurrentUser(c, r)
	if !user.IsAuthenticated {
		writeError(w, http.StatusUnauthorized)
		return
	}
	photo, err = photoMgr.Get(c.URLParams["id"])
	if err != nil {
		panic(err)
	}
	if photo == nil {
		http.NotFound(w, r)
		return
	}

	if !photo.CanVote(user) {
		writeError(w, http.StatusForbidden)
		return
	}

	fn(photo)

	if err = photoMgr.Update(photo); err != nil {
		panic(err)
	}

	user.RegisterVote(photo.ID)

	if err = userMgr.Update(user); err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
}

func getPage(r *http.Request) int64 {
	page, err := strconv.ParseInt(r.FormValue("page"), 10, 64)
	if err != nil {
		page = 1
	}
	return page
}
