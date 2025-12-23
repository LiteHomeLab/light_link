package backup

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
)

// Generate test data with realistic patterns
func generateData(size int, changeRate float64) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return data
}

// Modify data randomly
func modifyData(data []byte, changeRate float64) []byte {
	result := make([]byte, len(data))
	copy(result, data)

	numChanges := int(float64(len(data)) * changeRate)
	for i := 0; i < numChanges; i++ {
		pos := rand.Intn(len(data))
		result[pos] = byte(rand.Intn(256))
	}

	return result
}

// Append data
func appendData(data []byte, appendSize int) []byte {
	appendage := make([]byte, appendSize)
	for i := range appendage {
		appendage[i] = byte((len(data) + i) % 256)
	}
	return append(data, appendage...)
}

// Benchmark: BinaryDiff with various change rates
func BenchmarkBinaryDiff_NoChanges(b *testing.B) {
	data := generateData(1024*1024, 0) // 1MB
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := BinaryDiff(data, data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBinaryDiff_1PercentChanges(b *testing.B) {
	base := generateData(1024*1024, 0)
	modified := modifyData(base, 0.01)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := BinaryDiff(base, modified)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBinaryDiff_10PercentChanges(b *testing.B) {
	base := generateData(1024*1024, 0)
	modified := modifyData(base, 0.10)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := BinaryDiff(base, modified)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBinaryDiff_50PercentChanges(b *testing.B) {
	base := generateData(1024*1024, 0)
	modified := modifyData(base, 0.50)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := BinaryDiff(base, modified)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBinaryDiff_AppendOnly(b *testing.B) {
	base := generateData(1024*1024, 0)
	modified := appendData(base, 1024*100) // Add 100KB
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := BinaryDiff(base, modified)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: ApplyDiff
func BenchmarkApplyDiff_Small(b *testing.B) {
	oldData := generateData(1024*100, 0)
	newData := modifyData(oldData, 0.05)
	ops, _ := BinaryDiff(oldData, newData)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ApplyDiff(oldData, ops)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkApplyDiff_Large(b *testing.B) {
	oldData := generateData(1024*1024, 0)
	newData := modifyData(oldData, 0.05)
	ops, _ := BinaryDiff(oldData, newData)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ApplyDiff(oldData, ops)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: ReverseMerge
func BenchmarkReverseMerge_TwoVersions(b *testing.B) {
	v1 := generateData(1024*1024, 0)
	v2 := modifyData(v1, 0.05)
	v3 := modifyData(v2, 0.05)

	ops1, _ := BinaryDiff(v1, v2)
	ops2, _ := BinaryDiff(v2, v3)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ReverseMerge(v1, [][]DiffOp{ops1, ops2})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReverseMerge_TenVersions(b *testing.B) {
	versions := make([][]byte, 10)
	versions[0] = generateData(1024*1024, 0)

	for i := 1; i < 10; i++ {
		versions[i] = modifyData(versions[i-1], 0.05)
	}

	var diffOpsList [][]DiffOp
	for i := 1; i < 10; i++ {
		ops, _ := BinaryDiff(versions[i-1], versions[i])
		diffOpsList = append(diffOpsList, ops)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ReverseMerge(versions[0], diffOpsList)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: Chunk operations
func BenchmarkChunkSplit_1MB(b *testing.B) {
	data := make([]byte, 1024*1024)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		splitter := NewChunkSplitter(data)
		_, err := splitter.SplitAll()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkChunkSplit_10MB(b *testing.B) {
	data := make([]byte, 10*1024*1024)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		splitter := NewChunkSplitter(data)
		_, err := splitter.SplitAll()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkChunkAssembler_1MB(b *testing.B) {
	data := make([]byte, 1024*1024)
	splitter := NewChunkSplitter(data)
	chunks, _ := splitter.SplitAll()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		assembler := NewChunkAssembler(splitter.Metadata())
		for _, chunk := range chunks {
			assembler.AddChunk(chunk)
		}
		_, err := assembler.Assemble()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkChunkAssembler_10MB(b *testing.B) {
	data := make([]byte, 10*1024*1024)
	splitter := NewChunkSplitter(data)
	chunks, _ := splitter.SplitAll()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		assembler := NewChunkAssembler(splitter.Metadata())
		for _, chunk := range chunks {
			assembler.AddChunk(chunk)
		}
		_, err := assembler.Assemble()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: Serialization
func BenchmarkSerializeDiffOps_Small(b *testing.B) {
	oldData := generateData(1024*100, 0)
	newData := modifyData(oldData, 0.05)
	ops, _ := BinaryDiff(oldData, newData)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := SerializeDiffOps(ops)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSerializeDiffOps_Large(b *testing.B) {
	oldData := generateData(1024*1024, 0)
	newData := modifyData(oldData, 0.05)
	ops, _ := BinaryDiff(oldData, newData)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := SerializeDiffOps(ops)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSerializeChunk(b *testing.B) {
	data := make([]byte, ChunkSize)
	chunk := Chunk{
		Index:    0,
		Data:     data,
		Checksum: CalculateChecksum(data),
		Size:     ChunkSize,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := SerializeChunk(chunk)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Memory and diff size analysis test
func TestDiffSizeAnalysis(t *testing.T) {
	testCases := []struct {
		name       string
		size       int
		changeRate float64
	}{
		{"Small_1%", 100 * 1024, 0.01},
		{"Small_5%", 100 * 1024, 0.05},
		{"Small_10%", 100 * 1024, 0.10},
		{"Medium_1%", 1024 * 1024, 0.01},
		{"Medium_5%", 1024 * 1024, 0.05},
		{"Medium_10%", 1024 * 1024, 0.10},
		{"Large_1%", 10 * 1024 * 1024, 0.01},
		{"Large_5%", 10 * 1024 * 1024, 0.05},
	}

	t.Log("Diff Size Analysis:")
	t.Log("==================")

	for _, tc := range testCases {
		base := generateData(tc.size, 0)
		modified := modifyData(base, tc.changeRate)

		ops, err := BinaryDiff(base, modified)
		if err != nil {
			t.Fatalf("BinaryDiff failed: %v", err)
		}

		serialized, err := SerializeDiffOps(ops)
		if err != nil {
			t.Fatalf("SerializeDiffOps failed: %v", err)
		}

		compressionRatio := float64(len(serialized)) / float64(tc.size) * 100

		t.Logf("%s: Original=%d bytes, Diff=%d bytes (%.1f%%)",
			tc.name, tc.size, len(serialized), compressionRatio)
	}
}

// Test multi-version scenario
func TestMultiVersionScenario(t *testing.T) {
	t.Log("Multi-Version Scenario Test:")
	t.Log("============================")

	// Simulate 10 versions with incremental changes
	versions := make([][]byte, 10)
	versions[0] = generateData(1024*500, 0) // Start with 500KB

	for i := 1; i < 10; i++ {
		// Each version adds/modifies 5% of data
		versions[i] = modifyData(versions[i-1], 0.05)
	}

	// Create diffs
	var diffOpsList [][]DiffOp
	totalDiffSize := 0

	for i := 1; i < 10; i++ {
		ops, err := BinaryDiff(versions[i-1], versions[i])
		if err != nil {
			t.Fatalf("BinaryDiff failed for version %d: %v", i, err)
		}

		serialized, err := SerializeDiffOps(ops)
		if err != nil {
			t.Fatalf("SerializeDiffOps failed: %v", err)
		}

		diffOpsList = append(diffOpsList, ops)
		totalDiffSize += len(serialized)

		compressionRatio := float64(len(serialized)) / float64(len(versions[i])) * 100
		t.Logf("Version %d->%d: Diff size=%d bytes (%.1f%% of original)",
			i, i+1, len(serialized), compressionRatio)
	}

	// Compare with storing full versions
	fullStorageSize := len(versions[0]) * 10

	t.Logf("\nStorage Comparison:")
	t.Logf("  Full versions: %d bytes", fullStorageSize)
	t.Logf("  Incremental diffs: %d bytes", totalDiffSize)
	t.Logf("  Space saved: %d bytes (%.1f%%)",
		fullStorageSize-totalDiffSize,
		float64(fullStorageSize-totalDiffSize)/float64(fullStorageSize)*100)

	// Test reverse merge performance
	t.Log("\nTesting reverse merge...")
	result, err := ReverseMerge(versions[0], diffOpsList)
	if err != nil {
		t.Fatalf("ReverseMerge failed: %v", err)
	}

	if !bytes.Equal(result, versions[len(versions)-1]) {
		t.Error("Reverse merge result mismatch")
	}

	t.Log("Reverse merge successful!")
}

// Test realistic database backup scenario
func TestDatabaseBackupScenario(t *testing.T) {
	t.Log("Database Backup Scenario Test:")
	t.Log("==============================")

	// Simulate a JSON database dump
	generateDBDump := func(numRecords int) []byte {
		var buf bytes.Buffer
		buf.WriteString("[\n")
		for i := 0; i < numRecords; i++ {
			if i > 0 {
				buf.WriteString(",\n")
			}
			buf.WriteString(fmt.Sprintf(`  {"id": %d, "name": "user%d", "email": "user%d@example.com", "data": "%s"}`,
				i, i, i, "some fixed data that doesn't change"))
		}
		buf.WriteString("\n]")
		return buf.Bytes()
	}

	// Day 1: Initial backup
	day1 := generateDBDump(1000)

	// Day 2: Add 50 new records
	day2 := generateDBDump(1050)

	// Day 3: Modify 20 records, add 30 new
	day3 := generateDBDump(1080)

	// Create diffs
	ops1to2, _ := BinaryDiff(day1, day2)
	ops2to3, _ := BinaryDiff(day2, day3)

	serialized1to2, _ := SerializeDiffOps(ops1to2)
	serialized2to3, _ := SerializeDiffOps(ops2to3)

	t.Logf("Day 1 size: %d bytes", len(day1))
	t.Logf("Day 2 size: %d bytes", len(day2))
	t.Logf("Day 3 size: %d bytes", len(day3))
	t.Logf("Diff Day1->Day2: %d bytes (%.1f%%)", len(serialized1to2),
		float64(len(serialized1to2))/float64(len(day2))*100)
	t.Logf("Diff Day2->Day3: %d bytes (%.1f%%)", len(serialized2to3),
		float64(len(serialized2to3))/float64(len(day3))*100)

	// Verify we can reconstruct
	reconstructedDay2, _ := ApplyDiff(day1, ops1to2)
	if !bytes.Equal(reconstructedDay2, day2) {
		t.Error("Day 2 reconstruction failed")
	}

	reconstructedDay3, _ := ReverseMerge(day1, [][]DiffOp{ops1to2, ops2to3})
	if !bytes.Equal(reconstructedDay3, day3) {
		t.Error("Day 3 reverse merge failed")
	}

	t.Log("Database backup scenario test passed!")
}
