package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/isw2-unileon/Grupo-16/backend/internal/model"
	"github.com/isw2-unileon/Grupo-16/backend/internal/service"
	"github.com/isw2-unileon/Grupo-16/backend/internal/transport/middleware"
	"github.com/jackc/pgx/v5"
)

const manualHandlerUserID = "550e8400-e29b-41d4-a716-446655440111"
const manualHandlerRoutineID = "550e8400-e29b-41d4-a716-446655440000"
const manualHandlerExerciseID = "550e8400-e29b-41d4-a716-446655440222"

type manualHandlerTestRepository struct {
	countFunc     func(ctx context.Context, userID string) (int, error)
	createFunc    func(ctx context.Context, routine model.ManualRoutineToSave) (string, error)
	updateFunc    func(ctx context.Context, routineID, userID string, routine model.ManualRoutineToSave) error
	deleteFunc    func(ctx context.Context, routineID, userID string) error
	duplicateFunc func(ctx context.Context, routineID, userID string) (string, error)
}

func (r *manualHandlerTestRepository) CountByUser(ctx context.Context, userID string) (int, error) {
	if r.countFunc != nil {
		return r.countFunc(ctx, userID)
	}
	return 0, nil
}

func (r *manualHandlerTestRepository) CreateManualRoutine(ctx context.Context, routine model.ManualRoutineToSave) (string, error) {
	if r.createFunc != nil {
		return r.createFunc(ctx, routine)
	}
	return "new-routine-id", nil
}

func (r *manualHandlerTestRepository) UpdateManualRoutine(ctx context.Context, routineID, userID string, routine model.ManualRoutineToSave) error {
	if r.updateFunc != nil {
		return r.updateFunc(ctx, routineID, userID, routine)
	}
	return nil
}

func (r *manualHandlerTestRepository) DeleteRoutine(ctx context.Context, routineID, userID string) error {
	if r.deleteFunc != nil {
		return r.deleteFunc(ctx, routineID, userID)
	}
	return nil
}

func (r *manualHandlerTestRepository) DuplicateRoutine(ctx context.Context, routineID, userID string) (string, error) {
	if r.duplicateFunc != nil {
		return r.duplicateFunc(ctx, routineID, userID)
	}
	return "duplicated-routine-id", nil
}

func newManualRoutineHandler(repo *manualHandlerTestRepository) *RoutineHandler {
	routineService := service.NewRoutineService(&routineHandlerTestRepository{})
	manualService := service.NewManualRoutineService(repo)
	return NewRoutineHandler(routineService, nil, manualService)
}

func newAuthedContext(method, target, body string) (*httptest.ResponseRecorder, *gin.Context) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(middleware.ContextUserIDKey, manualHandlerUserID)

	var reader *bytes.Buffer
	if body == "" {
		reader = bytes.NewBufferString("")
	} else {
		reader = bytes.NewBufferString(body)
	}
	c.Request = httptest.NewRequest(method, target, reader)
	c.Request.Header.Set("Content-Type", "application/json")
	return w, c
}

func TestCreateRoutineSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newManualRoutineHandler(&manualHandlerTestRepository{
		createFunc: func(ctx context.Context, routine model.ManualRoutineToSave) (string, error) {
			return "routine-1", nil
		},
	})

	body := `{"name":"Push Day","routine_type":"Fuerza","notes":"[Objetivo] fuerza","exercises":[{"exercise_id":"` + manualHandlerExerciseID + `","sets":[{"set_number":1}]}]}`
	w, c := newAuthedContext("POST", "/api/routines", body)

	handler.CreateRoutine(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp["routine_id"] != "routine-1" {
		t.Fatalf("expected routine_id routine-1, got %q", resp["routine_id"])
	}
}

func TestCreateRoutineNameRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newManualRoutineHandler(&manualHandlerTestRepository{})

	w, c := newAuthedContext("POST", "/api/routines", `{"name":"   "}`)
	handler.CreateRoutine(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCreateRoutineLimitReached(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newManualRoutineHandler(&manualHandlerTestRepository{
		countFunc: func(ctx context.Context, userID string) (int, error) {
			return 9999, nil
		},
	})

	w, c := newAuthedContext("POST", "/api/routines", `{"name":"Routine"}`)
	handler.CreateRoutine(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestCreateRoutineUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newManualRoutineHandler(&manualHandlerTestRepository{})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/routines", bytes.NewBufferString(`{"name":"Routine"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateRoutine(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestCreateRoutineInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newManualRoutineHandler(&manualHandlerTestRepository{})

	w, c := newAuthedContext("POST", "/api/routines", `{"name":`)
	handler.CreateRoutine(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateRoutineSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newManualRoutineHandler(&manualHandlerTestRepository{})

	w, c := newAuthedContext("PUT", "/api/routines/"+manualHandlerRoutineID, `{"name":"Routine"}`)
	c.Params = gin.Params{{Key: "id", Value: manualHandlerRoutineID}}

	handler.UpdateRoutine(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateRoutineInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newManualRoutineHandler(&manualHandlerTestRepository{})

	w, c := newAuthedContext("PUT", "/api/routines/invalid", `{"name":"Routine"}`)
	c.Params = gin.Params{{Key: "id", Value: "invalid"}}

	handler.UpdateRoutine(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateRoutineNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newManualRoutineHandler(&manualHandlerTestRepository{
		updateFunc: func(ctx context.Context, routineID, userID string, routine model.ManualRoutineToSave) error {
			return pgx.ErrNoRows
		},
	})

	w, c := newAuthedContext("PUT", "/api/routines/"+manualHandlerRoutineID, `{"name":"Routine"}`)
	c.Params = gin.Params{{Key: "id", Value: manualHandlerRoutineID}}

	handler.UpdateRoutine(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDeleteRoutineSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newManualRoutineHandler(&manualHandlerTestRepository{})

	_, c := newAuthedContext("DELETE", "/api/routines/"+manualHandlerRoutineID, "")
	c.Params = gin.Params{{Key: "id", Value: manualHandlerRoutineID}}

	handler.DeleteRoutine(c)

	// c.Status only stages the code; in a unit context (no engine flush) read it from the writer.
	if c.Writer.Status() != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, c.Writer.Status())
	}
}

func TestDeleteRoutineNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newManualRoutineHandler(&manualHandlerTestRepository{
		deleteFunc: func(ctx context.Context, routineID, userID string) error {
			return pgx.ErrNoRows
		},
	})

	w, c := newAuthedContext("DELETE", "/api/routines/"+manualHandlerRoutineID, "")
	c.Params = gin.Params{{Key: "id", Value: manualHandlerRoutineID}}

	handler.DeleteRoutine(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestDuplicateRoutineSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newManualRoutineHandler(&manualHandlerTestRepository{
		duplicateFunc: func(ctx context.Context, routineID, userID string) (string, error) {
			return "copy-1", nil
		},
	})

	w, c := newAuthedContext("POST", "/api/routines/"+manualHandlerRoutineID+"/duplicate", "")
	c.Params = gin.Params{{Key: "id", Value: manualHandlerRoutineID}}

	handler.DuplicateRoutine(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp["routine_id"] != "copy-1" {
		t.Fatalf("expected routine_id copy-1, got %q", resp["routine_id"])
	}
}

func TestDuplicateRoutineLimitReached(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := newManualRoutineHandler(&manualHandlerTestRepository{
		countFunc: func(ctx context.Context, userID string) (int, error) {
			return 9999, nil
		},
	})

	w, c := newAuthedContext("POST", "/api/routines/"+manualHandlerRoutineID+"/duplicate", "")
	c.Params = gin.Params{{Key: "id", Value: manualHandlerRoutineID}}

	handler.DuplicateRoutine(c)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestManualRoutineHandlerNilService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewRoutineHandler(service.NewRoutineService(&routineHandlerTestRepository{}), nil)

	w, c := newAuthedContext("POST", "/api/routines", `{"name":"Routine"}`)
	handler.CreateRoutine(c)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
