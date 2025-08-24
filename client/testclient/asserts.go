package testclient

import (
	"strings"
	"testing"
)

func (c *TestClient) AssertNoCommandsCalled(t *testing.T) {
	if len(c.CommandsCalled) > 0 {
		t.Errorf("Expected no commands to be called, but got: %v", c.CommandsCalled)
	}
}

func (c *TestClient) AssertCommandsCalled(t *testing.T, expectedCommands ...string) {
	if len(c.CommandsCalled) != len(expectedCommands) {
		t.Errorf("Expected %d commands to be called, but got %d: %v", len(expectedCommands), len(c.CommandsCalled), c.CommandsCalled)
		return
	}

	for i, cmd := range expectedCommands {
		if !strings.EqualFold(c.CommandsCalled[i], cmd) {
			t.Errorf("Expected command '%s' at index %d, but got '%s'", cmd, i, c.CommandsCalled[i])
		}
	}
}

func (c *TestClient) ResetCommandsCalled() {
	c.CommandsCalled = []string{}
}
