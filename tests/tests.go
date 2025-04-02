package tests

func TestPasswordHashing(t *testing.T) {
    password := "mysecurepassword"
    hash, err := hashPassword(password)
    if err != nil {
        t.Fatal("Failed to hash password:", err)
    }

    if !checkPasswordHash(password, hash) {
        t.Error("Password should match hash but doesn't")
    }

    if checkPasswordHash("wrongpassword", hash) {
        t.Error("Wrong password matched hash!")
    }
}
func TestLoginHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/login?username=admin&password=secret", nil)
    w := httptest.NewRecorder()

    loginHandler(w, req)

    res := w.Result()
    body, _ := io.ReadAll(res.Body)

    if res.StatusCode != http.StatusOK {
        t.Fatalf("Expected 200 OK, got %d: %s", res.StatusCode, body)
    }

    if !strings.Contains(string(body), "Authenticated") {
        t.Errorf("Unexpected response: %s", body)
    }
}

