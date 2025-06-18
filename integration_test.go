package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	apiServer "splitExpense/api"
	"splitExpense/expense"
	"strconv"
	"testing"
	"time"

	"github.com/goombaio/namegenerator"
)

const HOST = "http://localhost:8888/v1"
const NumUsers = 3

func runCommand(t *testing.T, cmd string, cmdName string) ([]byte, error) {
	t.Logf("\n <<      Running Command:  %s    >>> \n", cmdName)
	t.Log(cmd, "\n")
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		t.Error(err)
	}
	t.Log("\n ####  CMD Output ###\n")
	fmt.Println(string(out))
	return out, nil
}

func Curl(route string, body string, method apiServer.Method, token string) string {
	return fmt.Sprintf(`curl -X %s '%s%s' -H 'Content-Type: application/json' -H 'Authorization: Bearer %s' -d '%s'`, method, HOST, route, token, body)
}

func createUsers(t *testing.T) (string, expense.User, []expense.User, error) {
	const password = "Pass@12345"
	var users []expense.User
	ng := namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())
	if NumUsers < 3 {
		t.Error("minimum 3 users required")
	}
	// create users
	for i := range NumUsers {
		var user expense.User
		name := ng.Generate()
		signupCmd := Curl("/user/signup",
			fmt.Sprintf(`{"name": "%s","email": "%s@example.com","password": "%s"}`, name, name, password),
			apiServer.POST,
			"",
		)
		bytes, _ := runCommand(t, signupCmd, "User signup "+strconv.Itoa(i))
		err := json.Unmarshal(bytes, &user)
		if err != nil {
			t.Error(err)
		}
		users = append(users, user)
	}
	// primary user
	// user := users[0]
	user := expense.User{
		Email:    "cold-pond@example.com",
		Password: "Pass@12345",
		Name:     "cold-pond",
		ID:       "992a2caf-0daa-47ad-bf89-6453b5f30fac",
	}

	t.Log("Primary User: ", user)
	// perform user Login
	loginCmd := Curl("/user/login",
		fmt.Sprintf(`{"email": "%s", "password": "%s"}`, user.Email, password),
		apiServer.POST,
		"",
	)
	bytes, _ := runCommand(t, loginCmd, " Primary User login ")

	// retrieve token
	type tokenReponse struct {
		Token string `json:"token"`
	}
	var response tokenReponse
	err := json.Unmarshal(bytes, &response)
	if err != nil {
		t.Error(err)
	}
	token := response.Token

	// use token to add rest of the users as friends of user
	// friends := lo.Slice(users, 1, len(users))
	friends := users
	for _, friend := range friends {
		cmd := Curl("/friend/"+friend.ID+"/add", "", apiServer.PUT, token)
		_, err := runCommand(t, cmd, "Add friend")
		if err != nil {
			t.Error(err)
		}
	}

	return token, user, friends, nil
}

func TestIntegration(t *testing.T) {
	// Run the user signup and login commands to get the token
	token, user, friends, _ := createUsers(t)

	ng := namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())
	groupName := ng.Generate()
	createGroupCmd := Curl("/group",
		fmt.Sprintf(`{"name": "%s", "description": "%s"}`, groupName, "this is a test group"),
		apiServer.POST,
		token,
	)
	bytes, _ := runCommand(t, createGroupCmd, "Create Group")
	var group expense.Group
	err := json.Unmarshal(bytes, &group)
	if err != nil {
		t.Error(err)
	}

	t.Log("friends: ", friends)
	// invite all users to group
	for _, friend := range friends {

		joinGroupCmd := Curl("/group/"+group.Id+"/invite",
			fmt.Sprintf(`{"new_member_id": "%s"}`, friend.ID),
			apiServer.PUT,
			token,
		)
		_, _ = runCommand(t, joinGroupCmd, "Join Group")
	}

	// create expense with friends

	equalSplit, err := json.Marshal(expense.SplitJson{
		Type:        "equal",
		TotalAmount: 100,
		EqualSplit:  []string{user.ID, friends[0].ID, friends[1].ID},
	})
	if err != nil {
		t.Error(err)
	}
	percentSplit, err := json.Marshal(expense.SplitJson{
		Type:        "percentage",
		TotalAmount: 100,
		PercentageSplit: map[string]float64{
			user.ID:       20.0,
			friends[0].ID: 40.5,
			friends[1].ID: 39.5,
		},
	})
	if err != nil {
		t.Error(err)
	}
	shareSplit, err := json.Marshal(expense.SplitJson{
		Type:        "share",
		TotalAmount: 100,
		ShareSplit: map[string]int{
			user.ID:       1,
			friends[0].ID: 3,
			friends[1].ID: 2,
		},
	})
	if err != nil {
		t.Error(err)
	}
	payee, err := json.Marshal(expense.PayerJson{
		Type: "single",
		PayerSplit: map[string]float64{
			user.ID: 100.0,
			// friends[0].ID: 50.0,
		}})
	if err != nil {
		t.Error(err)
	}

	for _, split := range []([]byte){equalSplit, percentSplit, shareSplit} {

		createExpenseCmd := Curl("/expense",
			fmt.Sprintf(`{"description":"test expense","amount": 100,"split": %s,"payee": %s, "groupId": "%s" }`, string(split), string(payee), group.Id),
			apiServer.POST,
			token,
		)
		bytes, _ = runCommand(t, createExpenseCmd, "Create Expense")
		var exp expense.Expense
		err = json.Unmarshal(bytes, &exp)
		if err != nil {
			t.Error(err)
		}

	}
}
