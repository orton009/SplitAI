package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"testing"
)

func runCommand(t *testing.T, cmd string, token string) {
	fullCmd := fmt.Sprintf("TOKEN=%s; %s", token, cmd)
	out, err := exec.Command("bash", "-c", fullCmd).CombinedOutput()
	if err != nil {
		t.Error("test failed withe error: ", err)
	} else {
		t.Log("test successful! output in next line \n", string(out))
	}
}

func TestIntegration(t *testing.T) {
	// Run the user signup and login commands to get the token
	// signupCmd := `curl -X POST 'http://localhost:8888/v1/user/signup' -H 'Content-Type: application/json' -d '{"name": "Test User","email": "test@example.com","password": "Pass@12345"}'`
	// runCommand(signupCmd, "")

	loginCmd := `curl -X POST 'http://localhost:8888/v1/user/login' -H 'Content-Type: application/json' -d '{"email":"test@example.com",  "password":"Pass@12345"}'`
	tokenBytes, err := exec.Command("bash", "-c", loginCmd).Output()

	type tokenReponse struct {
		Token string `json:"token"`
	}
	var response tokenReponse
	err = json.Unmarshal(tokenBytes, &response)
	if err != nil {
		t.Error(err)
	}
	token := response.Token
	tests := []string{
		fmt.Sprintf(`curl -X POST 'http://localhost:8888/v1/expenses' -H "Content-Type: application/json" -H "Authorization: Bearer %s " -d '{"description":"test expense","amount": 100,"split": {"type": "equal", "data": {"payee": ["f684518d-9843-44b4-b509-093159d51710"], "totalAmount": 100}},"payee": {"type": "single", "data": {"payer": "f684518d-9843-44b4-b509-093159d51710", "amount": 100}}}'`, token),
		fmt.Sprintf(`curl -X POST 'http://localhost:8888/api/groups' -H "Content-Type: application/json" -H "Authorization: Bearer %s" -d '{"name":  "test group","description":  "This is a test group","admin_id":  "userId"}'`, token),
		fmt.Sprintf(`curl -X POST 'http://localhost:8888/api/groups/{groupId}/invite' -H "Content-Type: application/json" -H "Authorization: Bearer %s" -d '{"user_id":  "userId"}'`, token),
	}

	for _, test := range tests {
		fmt.Printf("Running command:\n%s\n", test)
		runCommand(t, test, token)
	}
}
