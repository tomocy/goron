package manager

import (
	"net/http/httptest"
	"os"
	"sync"
	"time"

	"github.com/tomocy/goron/session"
	"github.com/tomocy/goron/session/manager"
)

func setUpManagerTest() (*managerTest, error) {
	m := manager.GetReady()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://192.168.55.55:8080/count", nil)

	return &managerTest{
		m:   m,
		rec: rec,
		req: req,
	}, nil
}

func deleteSession(sessID string) error {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	fname := "storage/sessions/" + sessID
	err := os.Remove(fname)

	return err
}

func makeSessionExpires(m *manager.Manager, sess *session.Session) {
	expiredSess := session.New(sess.ID(), time.Now().Add(-1*time.Hour), sess.Data())
	m.SetSession(expiredSess)
}
