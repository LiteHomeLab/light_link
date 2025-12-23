package backup

import (
	"bytes"
	"fmt"
	"testing"
)

func TestBinaryDiff_IdenticalData(t *testing.T) {
	data := []byte("hello world")

	ops, err := BinaryDiff(data, data)
	if err != nil {
		t.Fatalf("BinaryDiff failed: %v", err)
	}

	// Should have at least one match operation
	if len(ops) == 0 {
		t.Fatal("Expected at least one operation")
	}

	// Reconstruct and verify
	result, err := ApplyDiff(data, ops)
	if err != nil {
		t.Fatalf("ApplyDiff failed: %v", err)
	}

	if !bytes.Equal(result, data) {
		t.Errorf("Result mismatch: got %q, want %q", result, data)
	}
}

func TestBinaryDiff_SmallChange(t *testing.T) {
	oldData := []byte("hello world")
	newData := []byte("hello there")

	ops, err := BinaryDiff(oldData, newData)
	if err != nil {
		t.Fatalf("BinaryDiff failed: %v", err)
	}

	// Should detect changes
	if len(ops) == 0 {
		t.Fatal("Expected at least one operation")
	}

	// Verify reconstruction
	result, err := ApplyDiff(oldData, ops)
	if err != nil {
		t.Fatalf("ApplyDiff failed: %v", err)
	}

	if !bytes.Equal(result, newData) {
		t.Errorf("Result mismatch: got %q, want %q", result, newData)
	}
}

func TestBinaryDiff_AppendData(t *testing.T) {
	oldData := []byte("hello")
	newData := []byte("hello world")

	ops, err := BinaryDiff(oldData, newData)
	if err != nil {
		t.Fatalf("BinaryDiff failed: %v", err)
	}

	result, err := ApplyDiff(oldData, ops)
	if err != nil {
		t.Fatalf("ApplyDiff failed: %v", err)
	}

	if !bytes.Equal(result, newData) {
		t.Errorf("Result mismatch: got %q, want %q", result, newData)
	}
}

func TestBinaryDiff_DeleteData(t *testing.T) {
	oldData := []byte("hello world")
	newData := []byte("hello")

	ops, err := BinaryDiff(oldData, newData)
	if err != nil {
		t.Fatalf("BinaryDiff failed: %v", err)
	}

	result, err := ApplyDiff(oldData, ops)
	if err != nil {
		t.Fatalf("ApplyDiff failed: %v", err)
	}

	if !bytes.Equal(result, newData) {
		t.Errorf("Result mismatch: got %q, want %q", result, newData)
	}
}

func TestBinaryDiff_EmptyToData(t *testing.T) {
	oldData := []byte("")
	newData := []byte("hello world")

	ops, err := BinaryDiff(oldData, newData)
	if err != nil {
		t.Fatalf("BinaryDiff failed: %v", err)
	}

	result, err := ApplyDiff(oldData, ops)
	if err != nil {
		t.Fatalf("ApplyDiff failed: %v", err)
	}

	if !bytes.Equal(result, newData) {
		t.Errorf("Result mismatch: got %q, want %q", result, newData)
	}
}

func TestBinaryDiff_DataToEmpty(t *testing.T) {
	oldData := []byte("hello world")
	newData := []byte("")

	ops, err := BinaryDiff(oldData, newData)
	if err != nil {
		t.Fatalf("BinaryDiff failed: %v", err)
	}

	result, err := ApplyDiff(oldData, ops)
	if err != nil {
		t.Fatalf("ApplyDiff failed: %v", err)
	}

	if !bytes.Equal(result, newData) {
		t.Errorf("Result mismatch: got %q, want %q", result, newData)
	}
}

func TestBinaryDiff_LargeData(t *testing.T) {
	oldData := make([]byte, 100*4096) // 100 blocks
	for i := range oldData {
		oldData[i] = byte(i % 256)
	}

	newData := make([]byte, len(oldData))
	copy(newData, oldData)

	// Modify 10 random positions
	for i := 0; i < 10; i++ {
		pos := (i * 4096) + 100
		newData[pos] = ^newData[pos]
	}

	ops, err := BinaryDiff(oldData, newData)
	if err != nil {
		t.Fatalf("BinaryDiff failed: %v", err)
	}

	result, err := ApplyDiff(oldData, ops)
	if err != nil {
		t.Fatalf("ApplyDiff failed: %v", err)
	}

	if !bytes.Equal(result, newData) {
		t.Errorf("Large data result mismatch")
	}
}

func TestBinaryDiff_MultipleVersions(t *testing.T) {
	// Create a series of versions
	versions := [][]byte{
		[]byte("version 1"),
		[]byte("version 1 with more data"),
		[]byte("version 1 with even more"),
		[]byte("version 2 - completely different"),
	}

	// Create diffs between consecutive versions
	var diffOpsList [][]DiffOp
	for i := 1; i < len(versions); i++ {
		ops, err := BinaryDiff(versions[i-1], versions[i])
		if err != nil {
			t.Fatalf("BinaryDiff failed for version %d: %v", i, err)
		}
		diffOpsList = append(diffOpsList, ops)
	}

	// Test reconstruction from first version
	for i, ops := range diffOpsList {
		expected := versions[i+1]
		result, err := ApplyDiff(versions[i], ops)
		if err != nil {
			t.Fatalf("ApplyDiff failed for version %d: %v", i, err)
		}
		if !bytes.Equal(result, expected) {
			t.Errorf("Version %d reconstruction failed", i+1)
		}
	}
}

func TestReverseMerge(t *testing.T) {
	versions := [][]byte{
		[]byte("version 1: initial data"),
		[]byte("version 2: initial data with additions"),
		[]byte("version 3: initial data with additions and changes"),
		[]byte("version 4: final version"),
	}

	// Create diffs between consecutive versions
	var diffOpsList [][]DiffOp
	for i := 1; i < len(versions); i++ {
		ops, err := BinaryDiff(versions[i-1], versions[i])
		if err != nil {
			t.Fatalf("BinaryDiff failed: %v", err)
		}
		diffOpsList = append(diffOpsList, ops)
	}

	// Reverse merge from first version
	result, err := ReverseMerge(versions[0], diffOpsList)
	if err != nil {
		t.Fatalf("ReverseMerge failed: %v", err)
	}

	expected := versions[len(versions)-1]
	if !bytes.Equal(result, expected) {
		t.Errorf("ReverseMerge result mismatch: got %q, want %q", result, expected)
	}
}

func TestSerializeDeserializeDiffOps(t *testing.T) {
	oldData := []byte("hello world")
	newData := []byte("hello there")

	ops, err := BinaryDiff(oldData, newData)
	if err != nil {
		t.Fatalf("BinaryDiff failed: %v", err)
	}

	// Serialize
	serialized, err := SerializeDiffOps(ops)
	if err != nil {
		t.Fatalf("SerializeDiffOps failed: %v", err)
	}

	// Deserialize
	deserialized, err := DeserializeDiffOps(serialized)
	if err != nil {
		t.Fatalf("DeserializeDiffOps failed: %v", err)
	}

	// Verify operations match
	if len(deserialized) != len(ops) {
		t.Fatalf("Length mismatch: got %d, want %d", len(deserialized), len(ops))
	}

	for i := range ops {
		if deserialized[i].Type != ops[i].Type {
			t.Errorf("Operation %d type mismatch: got %v, want %v",
				i, deserialized[i].Type, ops[i].Type)
		}
		if deserialized[i].OldPos != ops[i].OldPos {
			t.Errorf("Operation %d OldPos mismatch", i)
		}
		if deserialized[i].NewPos != ops[i].NewPos {
			t.Errorf("Operation %d NewPos mismatch", i)
		}
		if !bytes.Equal(deserialized[i].Data, ops[i].Data) {
			t.Errorf("Operation %d Data mismatch", i)
		}
		if !bytes.Equal(deserialized[i].Checksum, ops[i].Checksum) {
			t.Errorf("Operation %d Checksum mismatch", i)
		}
	}

	// Verify reconstruction still works
	result, err := ApplyDiff(oldData, deserialized)
	if err != nil {
		t.Fatalf("ApplyDiff with deserialized ops failed: %v", err)
	}

	if !bytes.Equal(result, newData) {
		t.Errorf("Reconstruction from deserialized ops failed")
	}
}

func TestChecksum(t *testing.T) {
	data := []byte("test data for checksum")

	checksum := CalculateChecksum(data)
	if len(checksum) != 32 { // SHA256 produces 32 bytes
		t.Fatalf("Checksum length wrong: got %d, want 32", len(checksum))
	}

	if !VerifyChecksum(data, checksum) {
		t.Error("VerifyChecksum failed for valid checksum")
	}

	wrongChecksum := make([]byte, 32)
	wrongChecksum[0] = ^checksum[0]
	if VerifyChecksum(data, wrongChecksum) {
		t.Error("VerifyChecksum passed for invalid checksum")
	}
}

// TestBinaryDiff_RealWorldScenario tests a more realistic scenario
func TestBinaryDiff_RealWorldScenario(t *testing.T) {
	// Simulate a database dump with records
	type Record struct {
		ID   uint32
		Name string
		Data []byte
	}

	// Version 1: 100 records
	v1Records := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		record := fmt.Sprintf("record-%d: some data here", i)
		v1Records[i] = []byte(record)
	}
	v1Data := bytes.Join(v1Records, []byte("\n"))

	// Version 2: Add 10 records
	v2Records := make([][]byte, 110)
	copy(v2Records, v1Records)
	for i := 100; i < 110; i++ {
		record := fmt.Sprintf("record-%d: new data", i)
		v2Records[i] = []byte(record)
	}
	v2Data := bytes.Join(v2Records, []byte("\n"))

	// Version 3: Modify 5 records
	v3Records := make([][]byte, 110)
	copy(v3Records, v2Records)
	for i := 0; i < 5; i++ {
		record := fmt.Sprintf("record-%d: modified data", i)
		v3Records[i] = []byte(record)
	}
	v3Data := bytes.Join(v3Records, []byte("\n"))

	// Create diffs
	v1ToV2Ops, err := BinaryDiff(v1Data, v2Data)
	if err != nil {
		t.Fatalf("BinaryDiff v1->v2 failed: %v", err)
	}

	v2ToV3Ops, err := BinaryDiff(v2Data, v3Data)
	if err != nil {
		t.Fatalf("BinaryDiff v2->v3 failed: %v", err)
	}

	// Verify reconstructions
	v2FromV1, err := ApplyDiff(v1Data, v1ToV2Ops)
	if err != nil {
		t.Fatalf("ApplyDiff v1->v2 failed: %v", err)
	}
	if !bytes.Equal(v2FromV1, v2Data) {
		t.Error("v1->v2 reconstruction failed")
	}

	v3FromV2, err := ApplyDiff(v2Data, v2ToV3Ops)
	if err != nil {
		t.Fatalf("ApplyDiff v2->v3 failed: %v", err)
	}
	if !bytes.Equal(v3FromV2, v3Data) {
		t.Error("v2->v3 reconstruction failed")
	}

	// Test reverse merge from v1 to v3
	v3FromV1, err := ReverseMerge(v1Data, [][]DiffOp{v1ToV2Ops, v2ToV3Ops})
	if err != nil {
		t.Fatalf("ReverseMerge v1->v3 failed: %v", err)
	}
	if !bytes.Equal(v3FromV1, v3Data) {
		t.Error("v1->v3 reverse merge failed")
	}

	// Verify diff size is smaller than full data for incremental changes
	v1ToV2Serialized, _ := SerializeDiffOps(v1ToV2Ops)
	v2ToV3Serialized, _ := SerializeDiffOps(v2ToV3Ops)

	t.Logf("v1 data size: %d bytes", len(v1Data))
	t.Logf("v1->v2 diff size: %d bytes (%.1f%% of original)",
		len(v1ToV2Serialized), float64(len(v1ToV2Serialized))/float64(len(v1Data))*100)
	t.Logf("v2->v3 diff size: %d bytes (%.1f%% of original)",
		len(v2ToV3Serialized), float64(len(v2ToV3Serialized))/float64(len(v2Data))*100)

	// For append-only scenario, diff should be much smaller
	if len(v1ToV2Serialized) > len(v1Data)/2 {
		t.Logf("Warning: v1->v2 diff is larger than expected")
	}
}
