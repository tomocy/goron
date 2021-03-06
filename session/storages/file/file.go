package file

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tomocy/goron/log"
	"github.com/tomocy/goron/session"
	"github.com/tomocy/goron/settings"
)

type file struct {
	path string
	mu   sync.Mutex
}

const (
	DstDir       = "storage/sessions"
	Delimiter    = ":"
	ExpiresAtKey = "expiresAt"
	TimeLayout   = time.RFC3339Nano
)

func init() {
	// make session dir if it does not exist
	_, err := os.Stat(DstDir)
	if err != nil {
		os.MkdirAll(DstDir, 0744)
	}
}

func New() *file {
	return &file{path: DstDir}
}

func (f *file) InitSession(sessionID string) *session.Session {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.initSession(sessionID)
}

func (f *file) GetSession(sessionID string) (*session.Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.getSession(sessionID)
}

func (f *file) SetSession(session *session.Session) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.setSession(session)
}

func (f *file) DeleteSession(sessionID string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.deleteSession(sessionID)
}

func (f *file) DeleteExpiredSessions() {
	f.mu.Lock()
	defer f.mu.Unlock()

	ids := f.getIDs()
	for _, id := range ids {
		session, err := f.getSession(id)
		if err != nil {
			panic(err)
		}

		if session.DoesExpire() {
			log.Debug("Session " + session.ID() + " expired, so deleted")
			f.deleteSession(id)
		}
	}
}

func (f *file) initSession(sessionID string) *session.Session {
	name := f.path + "/" + sessionID
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	dat := make(map[string]string)
	session := session.New(sessionID, time.Now().Add(settings.Session.ExpiresIn()), dat)

	// Write when it expires
	fmt.Fprintln(file, ExpiresAtKey+Delimiter+session.ExpiresAt().Format(TimeLayout))

	return session
}

func (f *file) getSession(sessionID string) (*session.Session, error) {
	file, err := os.Open(f.path + "/" + sessionID)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var expiresAt time.Time
	dat := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ss := strings.SplitN(scanner.Text(), Delimiter, 2)
		if len(ss) < 2 {
			continue
		}

		if ss[0] == ExpiresAtKey {
			expiresAt, err = time.Parse(TimeLayout, ss[1])
			if err != nil {
				panic(err)
			}

			continue
		}

		dat[ss[0]] = ss[1]
	}

	return session.New(sessionID, expiresAt, dat), nil
}

func (f *file) setSession(session *session.Session) {
	file, err := os.OpenFile(f.path+"/"+session.ID(), os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write when the session expires
	fmt.Fprintln(file, ExpiresAtKey+Delimiter+session.ExpiresAt().Format(TimeLayout))

	// Write other keies and values
	for k, v := range session.Data() {
		fmt.Fprintln(file, fmt.Sprintf("%s:%s", k, v))
	}
}

func (f *file) deleteSession(sessionID string) {
	os.Remove(f.path + "/" + sessionID)
}

func (f *file) getIDs() []string {
	files, err := ioutil.ReadDir(f.path)
	if err != nil {
		panic(err)
	}

	ids := make([]string, 0, 50)
	for _, file := range files {
		ids = append(ids, file.Name())
	}

	return ids
}
