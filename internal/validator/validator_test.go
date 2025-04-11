package validator

import (
    "testing"
)

func TestValidator_New(t *testing.T) {
    v := New()
    if v.Errors == nil {
        t.Error("Expected Errors map to be initialized")
    }
}

func TestValidator_Valid(t *testing.T) {
    t.Run("NoErrors", func(t *testing.T) {
        v := New()
        if !v.Valid() {
            t.Error("Expected Valid() to return true when no errors")
        }
    })

    t.Run("WithErrors", func(t *testing.T) {
        v := New()
        v.AddError("test", "error")
        if v.Valid() {
            t.Error("Expected Valid() to return false when errors exist")
        }
    })
}

func TestValidator_AddError(t *testing.T) {
    v := New()
    v.AddError("test", "message1")
    if len(v.Errors) != 1 {
        t.Fatalf("Expected 1 error, got %d", len(v.Errors))
    }

    v.AddError("test", "message2")
    if len(v.Errors) != 1 {
        t.Error("AddError should not add duplicate keys")
    }
    if msg := v.Errors["test"]; msg != "message1" {
        t.Errorf("Expected message1, got %s", msg)
    }
}

func TestValidator_Check(t *testing.T) {
    v := New()
    ok := v.Check(false, "key", "error")
    if ok {
        t.Error("Check should return false when condition is false")
    }
    if _, exists := v.Errors["key"]; !exists {
        t.Error("Check should add error when condition is false")
    }

    ok = v.Check(true, "another", "error")
    if !ok {
        t.Error("Check should return true when condition is true")
    }
    if _, exists := v.Errors["another"]; exists {
        t.Error("Check should not add error when condition is true")
    }
}

func TestValidator_GetError(t *testing.T) {
    v := New()
    v.AddError("key", "error message")
    err := v.GetError("key")
    expected := "key error message"
    if err.Error() != expected {
        t.Errorf("Expected %s, got %s", expected, err.Error())
    }
}

func TestIn(t *testing.T) {
    tests := []struct {
        value string
        list  []string
        want  bool
    }{
        {"a", []string{"a", "b"}, true},
        {"c", []string{"a", "b"}, false},
        {"", []string{}, false},
    }

    for _, tt := range tests {
        got := In(tt.value, tt.list...)
        if got != tt.want {
            t.Errorf("In(%q, %v) = %v, want %v", tt.value, tt.list, got, tt.want)
        }
    }
}

func TestMatches(t *testing.T) {
    validEmail := "user@example.com"
    invalidEmail := "user@.com"
    validUsername := "user.name_123"
    invalidUsername := "user^name"

    if !Matches(validEmail, EmailRX) {
        t.Errorf("Expected %s to match email regex", validEmail)
    }
    if Matches(invalidEmail, EmailRX) {
        t.Errorf("Expected %s to not match email regex", invalidEmail)
    }

    if !Matches(validUsername, UsernameRX) {
        t.Errorf("Expected %s to match username regex", validUsername)
    }
    if Matches(invalidUsername, UsernameRX) {
        t.Errorf("Expected %s to not match username regex", invalidUsername)
    }
}

func TestUnique(t *testing.T) {
    tests := []struct {
        values []string
        want   bool
    }{
        {[]string{"a", "b", "c"}, true},
        {[]string{"a", "a", "b"}, false},
        {[]string{}, true},
        {[]string{"unique"}, true},
    }

    for _, tt := range tests {
        got := Unique(tt.values)
        if got != tt.want {
            t.Errorf("Unique(%v) = %v, want %v", tt.values, got, tt.want)
        }
    }
}