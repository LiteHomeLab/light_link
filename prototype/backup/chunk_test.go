package backup

import (
	"bytes"
	"io"
	"testing"
)

func TestChunkSplitter_SmallData(t *testing.T) {
	data := []byte("hello world")

	splitter := NewChunkSplitter(data)
	chunks, err := splitter.SplitAll()
	if err != nil {
		t.Fatalf("SplitAll failed: %v", err)
	}

	if len(chunks) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(chunks))
	}

	if !bytes.Equal(chunks[0].Data, data) {
		t.Error("Chunk data mismatch")
	}

	// Verify metadata
	metadata := splitter.Metadata()
	if metadata.TotalSize != uint64(len(data)) {
		t.Error("TotalSize mismatch")
	}
	if metadata.TotalChunks != 1 {
		t.Error("TotalChunks mismatch")
	}
}

func TestChunkSplitter_ExactMultiple(t *testing.T) {
	// Create data exactly 2 chunks
	data := make([]byte, ChunkSize*2)
	for i := range data {
		data[i] = byte(i % 256)
	}

	splitter := NewChunkSplitter(data)
	chunks, err := splitter.SplitAll()
	if err != nil {
		t.Fatalf("SplitAll failed: %v", err)
	}

	if len(chunks) != 2 {
		t.Fatalf("Expected 2 chunks, got %d", len(chunks))
	}

	if chunks[0].Size != ChunkSize {
		t.Errorf("First chunk size wrong: got %d, want %d", chunks[0].Size, ChunkSize)
	}

	if chunks[1].Size != ChunkSize {
		t.Errorf("Second chunk size wrong: got %d, want %d", chunks[1].Size, ChunkSize)
	}
}

func TestChunkSplitter_PartialLastChunk(t *testing.T) {
	// Create data 2.5 chunks
	data := make([]byte, ChunkSize*2+ChunkSize/2)
	for i := range data {
		data[i] = byte(i % 256)
	}

	splitter := NewChunkSplitter(data)
	chunks, err := splitter.SplitAll()
	if err != nil {
		t.Fatalf("SplitAll failed: %v", err)
	}

	if len(chunks) != 3 {
		t.Fatalf("Expected 3 chunks, got %d", len(chunks))
	}

	if chunks[2].Size != ChunkSize/2 {
		t.Errorf("Last chunk size wrong: got %d, want %d", chunks[2].Size, ChunkSize/2)
	}
}

func TestChunkAssembler_PerfectOrder(t *testing.T) {
	data := make([]byte, ChunkSize*2+ChunkSize/2)
	for i := range data {
		data[i] = byte(i % 256)
	}

	splitter := NewChunkSplitter(data)
	splitter.SetFileID("test-file")
	chunks, err := splitter.SplitAll()
	if err != nil {
		t.Fatalf("SplitAll failed: %v", err)
	}

	assembler := NewChunkAssembler(splitter.Metadata())

	// Add chunks in order
	for _, chunk := range chunks {
		if err := assembler.AddChunk(chunk); err != nil {
			t.Fatalf("AddChunk failed: %v", err)
		}
	}

	// Verify completeness
	if !assembler.IsComplete() {
		t.Error("Assembler should be complete")
	}

	// Assemble and verify
	result, err := assembler.Assemble()
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	if !bytes.Equal(result, data) {
		t.Error("Assembled data mismatch")
	}
}

func TestChunkAssembler_OutOfOrder(t *testing.T) {
	data := make([]byte, ChunkSize*3)
	for i := range data {
		data[i] = byte(i % 256)
	}

	splitter := NewChunkSplitter(data)
	chunks, err := splitter.SplitAll()
	if err != nil {
		t.Fatalf("SplitAll failed: %v", err)
	}

	assembler := NewChunkAssembler(splitter.Metadata())

	// Add chunks in reverse order
	for i := len(chunks) - 1; i >= 0; i-- {
		if err := assembler.AddChunk(chunks[i]); err != nil {
			t.Fatalf("AddChunk failed: %v", err)
		}
	}

	// Verify completeness
	if !assembler.IsComplete() {
		t.Error("Assembler should be complete")
	}

	// Assemble and verify
	result, err := assembler.Assemble()
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	if !bytes.Equal(result, data) {
		t.Error("Assembled data mismatch with out-of-order chunks")
	}
}

func TestChunkAssembler_DuplicateChunks(t *testing.T) {
	data := make([]byte, ChunkSize*2)
	for i := range data {
		data[i] = byte(i % 256)
	}

	splitter := NewChunkSplitter(data)
	chunks, err := splitter.SplitAll()
	if err != nil {
		t.Fatalf("SplitAll failed: %v", err)
	}

	assembler := NewChunkAssembler(splitter.Metadata())

	// Add chunk 0
	if err := assembler.AddChunk(chunks[0]); err != nil {
		t.Fatalf("AddChunk failed: %v", err)
	}

	// Add duplicate chunk 0 (should be ignored)
	if err := assembler.AddChunk(chunks[0]); err != nil {
		t.Fatalf("AddChunk of duplicate failed: %v", err)
	}

	// Add chunk 1
	if err := assembler.AddChunk(chunks[1]); err != nil {
		t.Fatalf("AddChunk failed: %v", err)
	}

	if !assembler.IsComplete() {
		t.Error("Assembler should be complete")
	}

	result, err := assembler.Assemble()
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	if !bytes.Equal(result, data) {
		t.Error("Assembled data mismatch with duplicate chunks")
	}
}

func TestChunkAssembler_ChecksumFailure(t *testing.T) {
	data := make([]byte, ChunkSize)

	splitter := NewChunkSplitter(data)
	chunks, err := splitter.SplitAll()
	if err != nil {
		t.Fatalf("SplitAll failed: %v", err)
	}

	assembler := NewChunkAssembler(splitter.Metadata())

	// Corrupt the chunk data
	chunks[0].Data[0] = ^chunks[0].Data[0]

	// Try to add corrupted chunk
	err = assembler.AddChunk(chunks[0])
	if err == nil {
		t.Error("Expected checksum error, got nil")
	}
}

func TestChunkAssembler_MissingChunks(t *testing.T) {
	data := make([]byte, ChunkSize*3)

	splitter := NewChunkSplitter(data)
	chunks, err := splitter.SplitAll()
	if err != nil {
		t.Fatalf("SplitAll failed: %v", err)
	}

	assembler := NewChunkAssembler(splitter.Metadata())

	// Add only first and last chunk
	assembler.AddChunk(chunks[0])
	assembler.AddChunk(chunks[2])

	// Check missing chunks
	missing := assembler.MissingChunks()
	if len(missing) != 1 || missing[0] != 1 {
		t.Errorf("MissingChunks wrong: got %v, want [1]", missing)
	}

	// Try to assemble incomplete data
	_, err = assembler.Assemble()
	if err == nil {
		t.Error("Expected error for incomplete assembly")
	}
}

func TestSerializeDeserializeChunk(t *testing.T) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	chunk := Chunk{
		Index:    5,
		Data:     data,
		Checksum: CalculateChecksum(data),
		Size:     uint32(len(data)),
	}

	// Serialize
	serialized, err := SerializeChunk(chunk)
	if err != nil {
		t.Fatalf("SerializeChunk failed: %v", err)
	}

	// Deserialize
	deserialized, err := DeserializeChunk(serialized)
	if err != nil {
		t.Fatalf("DeserializeChunk failed: %v", err)
	}

	// Verify
	if deserialized.Index != chunk.Index {
		t.Error("Index mismatch")
	}
	if deserialized.Size != chunk.Size {
		t.Error("Size mismatch")
	}
	if !bytes.Equal(deserialized.Data, chunk.Data) {
		t.Error("Data mismatch")
	}
	if !bytes.Equal(deserialized.Checksum, chunk.Checksum) {
		t.Error("Checksum mismatch")
	}
}

func TestSerializeDeserializeMetadata(t *testing.T) {
	metadata := ChunkMetadata{
		TotalChunks: 5,
		TotalSize:   5 * 1024 * 1024,
		FileID:      "test-file-123",
		Checksum:    []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	// Serialize
	serialized, err := SerializeMetadata(metadata)
	if err != nil {
		t.Fatalf("SerializeMetadata failed: %v", err)
	}

	// Deserialize
	deserialized, err := DeserializeMetadata(serialized)
	if err != nil {
		t.Fatalf("DeserializeMetadata failed: %v", err)
	}

	// Verify
	if deserialized.TotalChunks != metadata.TotalChunks {
		t.Error("TotalChunks mismatch")
	}
	if deserialized.TotalSize != metadata.TotalSize {
		t.Error("TotalSize mismatch")
	}
	if deserialized.FileID != metadata.FileID {
		t.Error("FileID mismatch")
	}
	if !bytes.Equal(deserialized.Checksum, metadata.Checksum) {
		t.Error("Checksum mismatch")
	}
}

func TestCustomChunkSize(t *testing.T) {
	data := make([]byte, 10000)

	splitter := NewChunkSplitter(data)
	splitter.SetChunkSize(1000) // Custom size

	chunks, err := splitter.SplitAll()
	if err != nil {
		t.Fatalf("SplitAll failed: %v", err)
	}

	// Should have 10 chunks
	if len(chunks) != 10 {
		t.Fatalf("Expected 10 chunks, got %d", len(chunks))
	}

	// Each chunk should be 1000 bytes
	for i, chunk := range chunks {
		if chunk.Size != 1000 {
			t.Errorf("Chunk %d size wrong: got %d, want 1000", i, chunk.Size)
		}
	}

	// Metadata should reflect custom size
	metadata := splitter.Metadata()
	if metadata.TotalChunks != 10 {
		t.Error("TotalChunks in metadata wrong")
	}
}

func TestStreamChunkReader(t *testing.T) {
	data := make([]byte, ChunkSize*2+ChunkSize/2)
	for i := range data {
		data[i] = byte(i % 256)
	}

	reader := bytes.NewReader(data)
	streamReader := NewStreamChunkReader(reader, ChunkSize)

	chunkCount := 0
	totalSize := 0

	for {
		chunk, err := streamReader.NextChunk()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("NextChunk failed: %v", err)
		}

		chunkCount++
		totalSize += int(chunk.Size)
	}

	if chunkCount != 3 {
		t.Fatalf("Expected 3 chunks, got %d", chunkCount)
	}

	if totalSize != len(data) {
		t.Errorf("Total size mismatch: got %d, want %d", totalSize, len(data))
	}
}

func TestStreamChunkWriter(t *testing.T) {
	data := make([]byte, ChunkSize*3)
	for i := range data {
		data[i] = byte(i % 256)
	}

	splitter := NewChunkSplitter(data)
	chunks, err := splitter.SplitAll()
	if err != nil {
		t.Fatalf("SplitAll failed: %v", err)
	}

	var buf bytes.Buffer
	streamWriter := NewStreamChunkWriter(&buf, splitter.Metadata())

	// Add chunks out of order: 2, 0, 1
	streamWriter.AddChunk(chunks[2])
	streamWriter.AddChunk(chunks[0])
	streamWriter.AddChunk(chunks[1])

	if !streamWriter.IsComplete() {
		t.Error("Stream writer should be complete")
	}

	if !bytes.Equal(buf.Bytes(), data) {
		t.Error("Stream writer output mismatch")
	}
}

func TestLargeFile(t *testing.T) {
	// Simulate a 10MB file
	size := 10 * 1024 * 1024
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	splitter := NewChunkSplitter(data)
	chunks, err := splitter.SplitAll()
	if err != nil {
		t.Fatalf("SplitAll failed: %v", err)
	}

	expectedChunks := (size + ChunkSize - 1) / ChunkSize
	if len(chunks) != expectedChunks {
		t.Fatalf("Expected %d chunks, got %d", expectedChunks, len(chunks))
	}

	assembler := NewChunkAssembler(splitter.Metadata())

	// Add all chunks
	for _, chunk := range chunks {
		if err := assembler.AddChunk(chunk); err != nil {
			t.Fatalf("AddChunk failed: %v", err)
		}
	}

	result, err := assembler.Assemble()
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	if !bytes.Equal(result, data) {
		t.Error("Large file assembly failed")
	}
}
