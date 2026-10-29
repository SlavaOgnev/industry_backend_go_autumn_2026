package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}
type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	return &inMemoryTaskRepo{mu: sync.RWMutex{}, clock: clock, seq: 0, tasks: make(map[string]Task)}
}
func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return Task{}, ErrInvalidTitle
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	newTask := Task{ID: strconv.FormatUint(r.seq, 10), Title: trimmed, Done: false, UpdatedAt: r.clock.Now()}
	r.tasks[newTask.ID] = newTask
	return newTask, nil
}

func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	if id == "" {
		return Task{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	if t, ok := r.tasks[id]; ok {

		return t, ok
	}
	return Task{}, false
}
func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Task, 0)
	for _, t := range r.tasks {
		if done == t.Done {
			result = append(result, t)
		}
	}
	return result
}
func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	if id == "" {
		return Task{}, ErrNotFound
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	task, ok := r.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}

	var timeUpdate = task.UpdatedAt
	if done != task.Done {
		timeUpdate = r.clock.Now()
	}
	task.Done = done
	task.UpdatedAt = timeUpdate
	r.tasks[id] = task
	return task, nil
}

type httpHandler struct{ repo TaskRepo }

func NewHTTPHandler(repo TaskRepo) http.Handler {
	return &httpHandler{repo: repo}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	switch {
	case path == "/tasks":
		{
			switch r.Method {
			case http.MethodGet:
				h.handleList(w, r)
			case http.MethodPost:
				h.handleCreate(w, r)
			default:
				http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
			}

		}
	case strings.HasPrefix(path, "/tasks/"):
		id := strings.TrimPrefix(path, "/tasks/")
		if id == "" || strings.Contains(id, "/") {
			http.Error(w, "Invalid path", http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.handleGet(w, r, id)
		case http.MethodPatch:
			h.handlePatch(w, r, id)
		default:
			http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		}

	default:
		http.Error(w, "Invalid path", http.StatusNotFound)

	}
}
func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title *string `json:"title"`
	}
	err := decodeStrictJSON(r.Body, &body)
	if err != nil || body.Title == nil {
		http.Error(w, "error parsing", http.StatusBadRequest)
		return
	}
	task, err := h.repo.Create(*body.Title)
	if err != nil {
		if errors.Is(err, ErrInvalidTitle) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, 201, task)
}
func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := h.repo.Get(id)
	if ok == false {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, task)
}
func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	vals := query["done"]
	done := false
	switch len(vals) {
	case 0:
	case 1:
		switch vals[0] {
		case "true":
			done = true
		default:
			http.Error(w, "invalid done value", http.StatusBadRequest)
			return
		}
	default:
		http.Error(w, "done must be passed once", http.StatusBadRequest)
		return
	}
	list := h.repo.List(done)
	if list == nil {
		list = make([]Task, 0)
	}

	slices.SortFunc(list, func(a, b Task) int {
		if c := time.Time.Compare(b.UpdatedAt, a.UpdatedAt); c != 0 {
			return c
		}
		return cmp.Compare(a.ID, b.ID)
	})

	writeJSON(w, http.StatusOK, list)

}
func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Done *bool `json:"done"`
	}

	err := decodeStrictJSON(r.Body, &body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if body.Done == nil {
		http.Error(w, "empty value", http.StatusBadRequest)
		return
	}
	task, err := h.repo.SetDone(id, *body.Done)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, task)

}
func decodeStrictJSON(r io.Reader, v any) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(v)
	if err != nil {
		return err
	}
	if smt := decoder.Decode(v); smt != io.EOF {
		return errors.New("unexpected end of JSON input")
	}
	return nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		return
	}

}
