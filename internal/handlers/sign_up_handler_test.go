package handlers

import (
	"net/http"
	"net/http/httptest"
	"receipt-wrangler/api/internal/models"
	"receipt-wrangler/api/internal/repositories"
	"receipt-wrangler/api/internal/utils"
	"strings"
	"testing"
)

func teardownSignUpTests() {
	repositories.TruncateTestDb()
}

func TestShouldNotAllowUserToSignUpIfDisabled(t *testing.T) {
	defer teardownSignUpTests()
	reader := strings.NewReader("")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api", reader)

	db := repositories.GetDB()
	db.Create(&models.SystemSettings{EnableLocalSignUp: false})

	expectedResponseCode := http.StatusNotFound

	SignUp(w, r)

	if w.Result().StatusCode != expectedResponseCode {
		utils.PrintTestError(t, w.Result().StatusCode, expectedResponseCode)
	}
}

// TODO: Fix how setting data for this endpoint works, then implement this tests
/*func TestShouldProcessSignUpCommand(t *testing.T) {
	defer teardownSignUpTests()
	db := repositories.GetDB()

	db.Create(&models.SystemSettings{EnableLocalSignUp: true})

	tests := map[string]struct {
		input  commands.SignUpCommand
		expect int
	}{
		"empty body": {
			expect: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		bytes, _ := json.Marshal(test.input)
		reader := strings.NewReader(string(bytes))
		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/api", reader)

		SignUp(w, r)

		if w.Result().StatusCode != test.expect {
			utils.PrintTestError(t, w.Result().StatusCode, test.expect)
		}
	}
}
*/
