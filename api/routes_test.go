package api

import (
	"database/sql"
	"github.com/zenazn/goji/web"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockSessionManager struct {
}

func (m *mockSessionManager) GetCurrentUser(r *http.Request) (*User, error) {
	return &User{}, nil
}

func (m *mockSessionManager) Login(w http.ResponseWriter, user *User) (string, error) {
	return "OK", nil
}

func (m *mockSessionManager) Logout(w http.ResponseWriter) (string, error) {
	return "OK", nil
}

type mockPhotoManager struct {
}

func (m *mockPhotoManager) Get(photoID int64) (*Photo, error) {
	return nil, sql.ErrNoRows
}

func (m *mockPhotoManager) GetDetail(photoID int64, user *User) (*PhotoDetail, error) {
	canEdit := user.ID == 1
	photo := &PhotoDetail{
		Photo: Photo{
			ID:      1,
			Title:   "test",
			OwnerID: 1,
		},
		OwnerName: "tester",
		Permissions: &Permissions{
			Edit: canEdit,
		},
	}
	return photo, nil
}

func (m *mockPhotoManager) All(page *Page, orderBy string) (*PhotoList, error) {
	item := &Photo{
		ID:      1,
		Title:   "test",
		OwnerID: 1,
	}
	photos := []Photo{*item}
	return NewPhotoList(photos, 1, 1), nil
}

func (m *mockPhotoManager) ByOwnerID(page *Page, ownerID int64) (*PhotoList, error) {
	return &PhotoList{}, nil
}

func (m *mockPhotoManager) Search(page *Page, q string) (*PhotoList, error) {
	return &PhotoList{}, nil
}

func (m *mockPhotoManager) UpdateTags(photo *Photo) error {
	return nil
}

func (m *mockPhotoManager) GetTagCounts() ([]TagCount, error) {
	return []TagCount{}, nil
}

func (m *mockPhotoManager) Delete(photo *Photo) error {
	return nil
}

func (m *mockPhotoManager) Insert(photo *Photo) error {
	return nil
}

func (m *mockPhotoManager) Update(photo *Photo) error {
	return nil
}

type emptyPhotoManager struct {
	mockPhotoManager
}

func (m *emptyPhotoManager) All(page *Page, orderBy string) (*PhotoList, error) {
	var photos []Photo
	return &PhotoList{photos, 0, 1, 0}, nil
}

func (m *emptyPhotoManager) GetDetail(photoID int64, user *User) (*PhotoDetail, error) {
	return nil, sql.ErrNoRows
}

// should return a 404
func TestGetPhotoDetailIfNone(t *testing.T) {
	req := &http.Request{}
	res := httptest.NewRecorder()
	c := web.C{}

	a := &AppContext{
		sessionMgr: &mockSessionManager{},
		photoMgr:   &emptyPhotoManager{},
	}

	err := a.photoDetail(c, res, req)
	if err != sql.ErrNoRows {
		t.Fail()
	}
}

func TestGetPhotoDetail(t *testing.T) {

	req, _ := http.NewRequest("GET", "http://localhost/api/photos/1", nil)
	res := httptest.NewRecorder()
	c := web.C{}
	c.URLParams = make(map[string]string)
	c.URLParams["id"] = "1"

	a := &AppContext{
		sessionMgr: &mockSessionManager{},
		photoMgr:   &mockPhotoManager{},
	}

	a.photoDetail(c, res, req)
	value := &PhotoDetail{}
	parseJsonBody(res, value)
	if res.Code != 200 {
		t.Fatal("Photo not found")
	}
	if value.Title != "test" {
		t.Fatal("Title should be test")
	}
	if value.Permissions.Edit {
		t.Fatal("User should have edit permission")
	}
}

func TestGetPhotos(t *testing.T) {

	req := &http.Request{}
	res := httptest.NewRecorder()

	a := &AppContext{
		photoMgr: &mockPhotoManager{},
	}

	a.getPhotos(web.C{}, res, req)
	value := &PhotoList{}
	parseJsonBody(res, value)
	if value.Total != 1 {
		t.Fail()
	}

}