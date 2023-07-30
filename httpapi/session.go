package httpapi

import (
	"context"
	"fmt"
	"net/http"

	session "github.com/go-session/session/v3"
)

type SessionMiddleWare struct {
}

func (x SessionMiddleWare) Handle(w http.ResponseWriter, r *http.Request) (Abort bool) {
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	store.Set("sessionstart", true)
	err = store.Save()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	return
}
func (x SessionMiddleWare) Enable() bool {
	return true
}
func (x SessionMiddleWare) Name() string {
	return "SessionMiddleWare"
}

type SesssionStore interface {
	// Get a session storage context
	Context() context.Context
	// Get the current session id
	SessionID() string
	// Set session value, call save function to take effect
	Set(key string, value interface{})
	// Get session value
	Get(key string) (interface{}, bool)
	// Delete session value, call save function to take effect
	Delete(key string) interface{}
	// Save session data
	Save() error
	// Clear all session data
	Flush() error
}
