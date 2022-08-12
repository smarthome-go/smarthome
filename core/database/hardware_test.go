package database

import "testing"

func TestCreateHardwareNodeTable(t *testing.T) {
	if err := createHardwareNodeTable(); err != nil {
		t.Error(err.Error())
		return
	}
}

func TestHardwareNode(t *testing.T) {
	node := HardwareNode{
		Name:  "test",
		Url:   "http://localhost:4242",
		Token: "",
	}
	if err := CreateHardwareNode(node); err != nil {
		t.Error(err.Error())
		return
	}

	nodeCreated, exists, err := GetHardwareNodeByUrl(node.Url)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !exists {
		t.Errorf("Node %s not found after creation", node.Url)
	}
	// Check metadata
	if nodeCreated.Name != node.Name ||
		nodeCreated.Url != node.Url ||
		nodeCreated.Token != node.Token {
		t.Errorf("Created node has different metadata: want: %v got: %v", node, nodeCreated)
		return
	}
	if err := DeleteHardwareNode(node.Url); err != nil {
		t.Error(err.Error())
		return
	}
	_, exists, err = GetHardwareNodeByUrl(node.Url)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if exists {
		t.Errorf("Node %s still exists after deletion", node.Url)
		return
	}
}

func TestSetNodeOnline(t *testing.T) {
	node := HardwareNode{
		Name:    "test1",
		Online:  false,
		Enabled: false,
		Url:     "http://localhost:123",
		Token:   "",
	}
	if err := CreateHardwareNode(node); err != nil {
		t.Error(err.Error())
		return
	}
	if err := SetNodeOnline(node.Url, !node.Online); err != nil {
		t.Error(err.Error())
		return
	}
	nodeAfter, exists, err := GetHardwareNodeByUrl(node.Url)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !exists {
		t.Errorf("Node `%s` not found in dataset", node.Url)
		return
	}
	if nodeAfter.Online == node.Online {
		t.Errorf("Node `%s` did not change online status: want: %t got: %t", node.Url, !node.Online, nodeAfter.Online)
		return
	}
}

func TestModifyNode(t *testing.T) {
	nodeBefore := HardwareNode{
		Name:    "before",
		Enabled: false,
		Online:  true,
		Url:     "http://localhost:before",
		Token:   "before",
	}
	if err := CreateHardwareNode(nodeBefore); err != nil {
		t.Error(err.Error())
		return
	}
	nodeAfter := HardwareNode{
		Name:    "after",
		Enabled: true,
		Online:  true,
		Token:   "after",
	}
	if err := ModifyHardwareNode(
		nodeBefore.Url,
		nodeAfter.Enabled,
		nodeAfter.Name,
		nodeAfter.Token,
	); err != nil {
		t.Error(err.Error())
		return
	}
	nodeFromDb, exists, err := GetHardwareNodeByUrl(nodeBefore.Url)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if !exists {
		t.Errorf("Node %s does not exist after modification", nodeBefore.Url)
		return
	}
	if nodeFromDb.Enabled != nodeAfter.Enabled ||
		nodeFromDb.Name != nodeAfter.Name ||
		nodeFromDb.Token != nodeAfter.Token ||
		nodeFromDb.Online != nodeAfter.Online ||
		nodeFromDb.Url != nodeBefore.Url {
		t.Errorf("Modification did not affect all metadata: want: %v got: %v", nodeAfter, nodeFromDb)
		return
	}
	nodes, err := GetHardwareNodes()
	if err != nil {
		t.Error(err.Error())
		return
	}
	valid := false
	for _, node := range nodes {
		if node.Enabled == nodeAfter.Enabled &&
			node.Name == nodeAfter.Name &&
			node.Token == nodeAfter.Token &&
			node.Online == nodeAfter.Online &&
			node.Url == nodeBefore.Url {
			valid = true
		}
	}
	if !valid {
		t.Errorf("Hardware node not found in nodes want: %v got: {}", nodeAfter)
		return
	}
}
