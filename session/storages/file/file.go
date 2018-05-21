package file

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tomocy/goron/session"
	"github.com/tomocy/goron/settings"
)

type file struct {
	path string
	mu   sync.Mutex
}

var dstDir string
var delimiter string
var expiresAtKey string
var timeLayout string

func init() {
	dstDir = "storage/sessions"
	delimiter = ":"
	expiresAtKey = "expiresAt"
	timeLayout = time.RFC3339Nano
}

func New() *file {
	return &file{path: dstDir}
}

func (f *file) InitSession(sessionID string) session.Session {
	f.mu.Lock()
	defer f.mu.Unlock()

	name := f.path + "/" + sessionID
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}

	dat := make(map[string]string)
	session := session.New(sessionID, time.Now().Add(settings.Session.ExpiresIn), dat)

	// Write when it expires
	fmt.Fprintln(file, expiresAtKey+delimiter+session.ExpiresAt().Format(timeLayout))

	return session
}

func (f *file) GetSession(sessionID string) (session.Session, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	file, err := os.Open(f.path + "/" + sessionID)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var expiresAt time.Time
	dat := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ss := strings.SplitN(scanner.Text(), ":", 2)
		if len(ss) < 2 {
			continue
		}

		if ss[0] == expiresAtKey {
			expiresAt, err = time.Parse(time.RFC3339Nano, ss[1])
			if err != nil {
				panic(err)
			}

			continue
		}

		dat[ss[0]] = ss[1]
	}

	return session.New(sessionID, expiresAt, dat), nil
}

func (f *file) SetSession(session session.Session) {
	file, err := os.OpenFile(f.path+"/"+session.ID(), os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write when the session expires
	fmt.Fprintln(file, expiresAtKey+delimiter+session.ExpiresAt().Format(timeLayout))

	// Write other keies and values
	for k, v := range session.Data() {
		fmt.Fprintln(file, fmt.Sprintf("%s:%s", k, v))
	}
}

func (f *file) DeleteSession(sessionID string) {
	os.Remove(f.path + "/" + sessionID)
}
