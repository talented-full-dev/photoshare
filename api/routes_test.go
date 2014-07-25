package api

import (
	"database/sql"
	"encoding/json"
	"github.com/zenazn/goji/web"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockCache struct{}

func (m *mockCache) set(key string, obj interface{}) ([]byte, error) {
	return json.Marshal(obj)
}

func (m *mockCache) clear() error {
	return nil
}

func (m *mockCache) get(key string, fn func() (interface{}, error)) (interface{}, error) {
	return fn()
}

func (m *mockCache) render(w http.ResponseWriter, status int, key string, fn func() (interface{}, error)) error {
	obj, err := fn()
	if err != nil {
		return err
	}
	value, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return writeBody(w, value, status, "application/json")
}

type mockSessionManager struct {
}

func (m *mockSessionManager) readToken(r *http.Request) (int64, error) {
	return 0, nil
}

func (m *mockSessionManager) writeToken(w http.ResponseWriter, userID int64) error {
	return nil
}

type mockPhotoDataStore struct {
}

func (m *mockPhotoDataStore) get(photoID int64) (*photo, error) {
	return nil, sql.ErrNoRows
}

func (m *mockPhotoDataStore) getDetail(photoID int64, user *user) (*photoDetail, error) {
	canEdit := user.ID == 1
	photo := &photoDetail{
		photo: photo{
			ID:      1,
			Title:   "test",
			OwnerID: 1,
		},
		OwnerName: "tester",
		Permissions: &permissions{
			Edit: canEdit,
		},
	}
	return photo, nil
}

func (m *mockPhotoDataStore) all(page *page, orderBy string) (*photoList, error) {
	item := &photo{
		ID:      1,
		Title:   "test",
		OwnerID: 1,
	}
	photos := []photo{*item}
	return newPhotoList(photos, 1, 1), nil
}

func (m *mockPhotoDataStore) byOwnerID(page *page, ownerID int64) (*photoList, error) {
	return &photoList{}, nil
}

func (m *mockPhotoDataStore) search(page *page, q string) (*photoList, error) {
	return &photoList{}, nil
}

func (m *mockPhotoDataStore) updateTags(photo *photo) error {
	return nil
}

func (m *mockPhotoDataStore) getTagCounts() ([]tagCount, error) {
	return []tagCount{}, nil
}

func (m *mockPhotoDataStore) delete(photo *photo) error {
	return nil
}

func (m *mockPhotoDataStore) insert(photo *photo) error {
	return nil
}

func (m *mockPhotoDataStore) update(photo *photo) error {
	return nil
}

type emptyPhotoDataStore struct {
	mockPhotoDataStore
}

func (m *emptyPhotoDataStore) all(page *page, orderBy string) (*photoList, error) {
	var photos []photo
	return &photoList{photos, 0, 1, 0}, nil
}

func (m *emptyPhotoDataStore) getDetail(photoID int64, user *user) (*photoDetail, error) {
	return nil, sql.ErrNoRows
}

// should return a 404
func TestGetPhotoDetailIfNone(t *testing.T) {
	req := &http.Request{}
	res := httptest.NewRecorder()
	c := web.C{}
	c.Env = make(map[string]interface{})

	a := &appContext{
		sessionMgr: &mockSessionManager{},
		ds:         &dataStores{photos: &emptyPhotoDataStore{}},
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
	c.Env = make(map[string]interface{})
	c.URLParams = make(map[string]string)
	c.URLParams["id"] = "1"

	a := &appContext{
		sessionMgr: &mockSessionManager{},
		ds:         &dataStores{photos: &mockPhotoDataStore{}},
	}

	c.Env["user"] = &user{}

	a.photoDetail(c, res, req)
	value := &photoDetail{}
	parseJSONBody(res, value)
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

	a := &appContext{
		ds:    &dataStores{photos: &mockPhotoDataStore{}},
		cache: &mockCache{},
	}

	a.getPhotos(web.C{}, res, req)
	value := &photoList{}
	parseJSONBody(res, value)
	if value.Total != 1 {
		t.Fail()
	}

}
